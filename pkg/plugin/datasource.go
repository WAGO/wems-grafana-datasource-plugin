package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type queryModel struct{}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	// https://grafana.com/developers/plugin-tools/introduction/data-frames
	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
		data.NewField("values", nil, []int64{10, 20}),
	)

	// add the frames to the response.
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
