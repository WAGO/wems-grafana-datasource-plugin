package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/wago/wems-grafana-plugin/pkg/models"
)

const DefaultBaseURL = "https://c1.api.wago.com/wems"

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	clientID     string
	clientSecret string
	baseURL      string
	token        string
	tokenExpiry  time.Time
}

// TokenRequest is the payload for the WEMS token endpoint
// Only the required fields for super_token are included
// See OpenAPI for full structure
type TokenRequest struct {
	ApplicationComponents map[string][]string `json:"application_components"`
	ClientID              string              `json:"client_id"`
	ClientSecret          string              `json:"client_secret"`
	Endpoints             map[string][]string `json:"endpoints"`
	PlatformScopes        []string            `json:"platform_scopes"`
	SuperToken            bool                `json:"super_token"`
}

// DatasourceSettings holds the config from plugin.json
type DatasourceSettings struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	BaseURL      string `json:"base_url"`
}

// NewDatasource creates a new datasource instance.
func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var dsSettings DatasourceSettings
	if err := json.Unmarshal(settings.JSONData, &dsSettings); err != nil {
		return nil, fmt.Errorf("failed to parse datasource settings: %w", err)
	}
	if settings.DecryptedSecureJSONData != nil {
		if v, ok := settings.DecryptedSecureJSONData["client_secret"]; ok {
			dsSettings.ClientSecret = v
		}
	}

	// Use default base URL if not provided
	if dsSettings.BaseURL == "" {
		dsSettings.BaseURL = DefaultBaseURL
	}
	// Remove trailing slash from baseURL if present
	if len(dsSettings.BaseURL) > 0 && dsSettings.BaseURL[len(dsSettings.BaseURL)-1] == '/' {
		dsSettings.BaseURL = dsSettings.BaseURL[:len(dsSettings.BaseURL)-1]
	}
	// Prepare token request
	tokenReq := TokenRequest{
		ApplicationComponents: map[string][]string{},
		ClientID:              dsSettings.ClientID,
		ClientSecret:          dsSettings.ClientSecret,
		Endpoints:             map[string][]string{},
		PlatformScopes:        []string{},
		SuperToken:            true,
	}

	// Request token
	tokenURL := dsSettings.BaseURL + "/v1/token"
	fmt.Println("alo " + tokenURL)
	body, err := json.Marshal(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}
	resp, err := doHTTPPost(tokenURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to get WEMS token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("WEMS token request failed: %s %s", resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}
	token := string(bodyBytes)

	return &Datasource{
		clientID:     dsSettings.ClientID,
		clientSecret: dsSettings.ClientSecret,
		baseURL:      dsSettings.BaseURL,
		token:        token,
		tokenExpiry:  time.Now().Add(20 * time.Minute),
	}, nil
}

// doHTTPPost is a helper to POST JSON and return the response
func doHTTPPost(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type WEMSQueryModel struct {
	EndpointID        string `json:"endpoint_id"`
	ApplianceID       string `json:"appliance_id"`
	ServiceURI        string `json:"service_uri"`
	DataPoint         string `json:"data_point"`
	AggregateFunction string `json:"aggregate_function,omitempty"`
	CreateEmptyValues *bool  `json:"create_empty_values,omitempty"`
}

type TimeSeriesDataPoint struct {
	Time  int64       `json:"time"`
	Value interface{} `json:"value"`
}

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// Unmarshal the JSON into our query model (only for endpoint/appliance/service/datapoint)
	var qm WEMSQueryModel
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	// Validate required fields
	if qm.EndpointID == "" || qm.ApplianceID == "" || qm.ServiceURI == "" || qm.DataPoint == "" {
		return backend.ErrDataResponse(backend.StatusBadRequest, "Missing required query fields: endpoint_id, appliance_id, service_uri, data_point")
	}

	// Build the WEMS API URL
	url := fmt.Sprintf("%s/v1/endpoint/%s/series/%s/%s/%s", d.baseURL, qm.EndpointID, qm.ApplianceID, qm.ServiceURI, qm.DataPoint)

	// Build query params using backend.DataQuery fields
	params := make(map[string]string)
	params["from"] = fmt.Sprintf("%d", query.TimeRange.From.Unix())
	params["to"] = fmt.Sprintf("%d", query.TimeRange.To.Unix())
	if query.MaxDataPoints > 0 {
		params["limit"] = fmt.Sprintf("%d", query.MaxDataPoints)
	}
	if query.Interval > 0 {
		params["aggregateInterval"] = fmt.Sprintf("%ds", int(query.Interval.Seconds()))
	}
	if qm.AggregateFunction != "" {
		params["aggregateFunction"] = qm.AggregateFunction
	}
	if qm.CreateEmptyValues != nil {
		params["createEmptyValues"] = fmt.Sprintf("%v", *qm.CreateEmptyValues)
	}

	// Build the full URL with query params
	qstr := ""
	for k, v := range params {
		if qstr == "" {
			qstr = "?"
		} else {
			qstr += "&"
		}
		qstr += fmt.Sprintf("%s=%s", k, v)
	}
	fullURL := url + qstr

	// Prepare HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, "Failed to create request: "+err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+d.token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, "Request failed: "+err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("WEMS API error: %s %s", resp.Status, string(bodyBytes)))
	}

	var points []TimeSeriesDataPoint
	if err := json.NewDecoder(resp.Body).Decode(&points); err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, "Failed to decode WEMS response: "+err.Error())
	}

	// Convert to Grafana data frame
	times := make([]time.Time, 0, len(points))
	values := make([]float64, 0, len(points))
	for _, p := range points {
		times = append(times, time.Unix(p.Time, 0))
		// Try to convert value to float64
		switch v := p.Value.(type) {
		case float64:
			values = append(values, v)
		case int:
			values = append(values, float64(v))
		case int64:
			values = append(values, float64(v))
		case string:
			// Try to parse string as float
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				values = append(values, f)
			} else {
				values = append(values, 0)
			}
		default:
			values = append(values, 0)
		}
	}

	frame := data.NewFrame("response",
		data.NewField("time", nil, times),
		data.NewField("value", nil, values),
	)
	response.Frames = append(response.Frames, frame)
	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Unable to load settings: " + err.Error(),
		}, nil
	}
	if config.ClientID == "" || config.Secrets.ClientSecret == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Missing client_id or client_secret",
		}, nil
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}

	// Remove trailing slash from baseURL if present
	baseURL := config.BaseURL
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	// Prepare token request
	tokenReq := TokenRequest{
		ApplicationComponents: map[string][]string{},
		ClientID:              config.ClientID,
		ClientSecret:          config.Secrets.ClientSecret,
		Endpoints:             map[string][]string{},
		PlatformScopes:        []string{},
		SuperToken:            true,
	}
	tokenURL := baseURL + "/v1/token"
	body, err := json.Marshal(tokenReq)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Failed to marshal token request: " + err.Error(),
		}, nil
	}
	resp, err := doHTTPPost(tokenURL, body)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Failed to get WEMS token: " + err.Error(),
		}, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("WEMS token request failed: %s %s", resp.Status, string(bodyBytes)),
		}, nil
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

// CallResource handles resource calls from the frontend (e.g., /resources/endpoint-list, /resources/appliance-list)
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	if req.Path == "endpoint-list" {
		// Build WEMS endpoint list URL
		url := d.baseURL + "/v1/endpoint/"
		request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to create request: " + err.Error()),
			})
		}
		request.Header.Set("Authorization", "Bearer "+d.token)
		request.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(request)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Request failed: " + err.Error()),
			})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to read response: " + err.Error()),
			})
		}

		if resp.StatusCode != 200 {
			return sender.Send(&backend.CallResourceResponse{
				Status: resp.StatusCode,
				Body:   body,
			})
		}

		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   body,
		})
	}

	if req.Path == "appliance-list" {
		endpointId := ""
		if req.URL != "" {
			if parsedUrl, err := url.Parse(req.URL); err == nil {
				endpointId = parsedUrl.Query().Get("endpointId")
			}
		}
		if endpointId == "" {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusBadRequest,
				Body:   []byte("Missing endpointId parameter"),
			})
		}
		url := fmt.Sprintf("%s/v1/endpoint/%s/description?includeApplianceConfiguration=false&draft=false", d.baseURL, endpointId)
		req2, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to create request: " + err.Error()),
			})
		}
		req2.Header.Set("Authorization", "Bearer "+d.token)
		req2.Header.Set("Accept", "application/json")
		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req2)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Request failed: " + err.Error()),
			})
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to read response: " + err.Error()),
			})
		}
		if resp.StatusCode != 200 {
			return sender.Send(&backend.CallResourceResponse{
				Status: resp.StatusCode,
				Body:   body,
			})
		}
		// Parse and flatten appliances from processes
		type appliance struct {
			ID           string `json:"id"`
			FriendlyName string `json:"friendlyName"`
		}
		type process struct {
			ID         string      `json:"id"`
			Name       string      `json:"name"`
			Appliances []appliance `json:"appliances"`
		}
		type descResp struct {
			Processes []process `json:"processes"`
		}
		var desc descResp
		if err := json.Unmarshal(body, &desc); err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to parse appliances: " + err.Error()),
			})
		}
		// Flatten appliances with process info
		var result []map[string]string
		for _, proc := range desc.Processes {
			for _, app := range proc.Appliances {
				label := app.FriendlyName
				if label == "" {
					label = app.ID
				}
				if proc.Name != "" {
					label = fmt.Sprintf("%s (%s)", label, proc.Name)
				}
				result = append(result, map[string]string{
					"id":    app.ID,
					"label": label,
				})
			}
		}
		respBytes, _ := json.Marshal(result)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   respBytes,
		})
	}

	if req.Path == "service-list" {
		endpointId := ""
		applianceId := ""
		if req.URL != "" {
			if parsedUrl, err := url.Parse(req.URL); err == nil {
				endpointId = parsedUrl.Query().Get("endpointId")
				applianceId = parsedUrl.Query().Get("applianceId")
			}
		}
		if endpointId == "" || applianceId == "" {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusBadRequest,
				Body:   []byte("Missing endpointId or applianceId parameter"),
			})
		}
		url := fmt.Sprintf("%s/v1/endpoint/%s/values/%s", d.baseURL, endpointId, applianceId)
		req2, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to create request: " + err.Error()),
			})
		}
		req2.Header.Set("Authorization", "Bearer "+d.token)
		req2.Header.Set("Accept", "application/json")
		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req2)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Request failed: " + err.Error()),
			})
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to read response: " + err.Error()),
			})
		}
		if resp.StatusCode != 200 {
			return sender.Send(&backend.CallResourceResponse{
				Status: resp.StatusCode,
				Body:   body,
			})
		}
		// Parse JSON keys as service URIs
		var raw map[string]interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to parse service list: " + err.Error()),
			})
		}
		var result []map[string]string
		for k := range raw {
			result = append(result, map[string]string{
				"uri":   k,
				"label": k,
			})
		}
		respBytes, _ := json.Marshal(result)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   respBytes,
		})
	}

	if req.Path == "datapoint-list" {
		endpointId := ""
		applianceId := ""
		serviceUri := ""
		if req.URL != "" {
			if parsedUrl, err := url.Parse(req.URL); err == nil {
				endpointId = parsedUrl.Query().Get("endpointId")
				applianceId = parsedUrl.Query().Get("applianceId")
				serviceUri = parsedUrl.Query().Get("serviceUri")
			}
		}
		if endpointId == "" || applianceId == "" || serviceUri == "" {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusBadRequest,
				Body:   []byte("Missing endpointId, applianceId, or serviceUri parameter"),
			})
		}
		url := fmt.Sprintf("%s/v1/endpoint/%s/values/%s/%s", d.baseURL, endpointId, applianceId, serviceUri)
		req2, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to create request: " + err.Error()),
			})
		}
		req2.Header.Set("Authorization", "Bearer "+d.token)
		req2.Header.Set("Accept", "application/json")
		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req2)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Request failed: " + err.Error()),
			})
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("Failed to read response: " + err.Error()),
			})
		}
		if resp.StatusCode != 200 {
			return sender.Send(&backend.CallResourceResponse{
				Status: resp.StatusCode,
				Body:   body,
			})
		}
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   body,
		})
	}

	// Unknown resource
	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusNotFound,
		Body:   []byte("Not found"),
	})
}
