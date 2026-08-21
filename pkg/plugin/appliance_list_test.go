package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// captureSender collects the single CallResourceResponse a handler sends.
type captureSender struct {
	resp *backend.CallResourceResponse
}

func (c *captureSender) Send(resp *backend.CallResourceResponse) error {
	c.resp = resp
	return nil
}

// TestCallResourceApplianceList exercises the whole appliance-list path against a
// stubbed WEMS API: description fetch, parallel model-name resolution, flattening and
// JSON encoding. This is the wiring that the pure buildApplianceOptions tests cannot
// cover.
func TestCallResourceApplianceList(t *testing.T) {
	var modelCalls int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/description"):
			if got := r.URL.Path; got != "/v1/endpoint/ep-1/description" {
				t.Errorf("unexpected description path: %s", got)
			}
			fmt.Fprint(w, `{
			  "building": {"spaces": [{"label": "Room 1"}]},
			  "processes": [
			    {
			      "id": "p1", "name": "HVAC",
			      "appliances": [
			        {"id": "a1", "friendlyName": "Heat pump", "applianceReference": 7},
			        {"id": "a2", "friendlyName": "Wallbox", "applianceReference": 7}
			      ],
			      "applianceReferences": [
			        {"id": "r2", "friendlyName": "Compressor", "reference": "a1"},
			        {"id": "r1", "friendlyName": "Backup heater", "reference": "a1"}
			      ]
			    },
			    {
			      "id": "p2", "name": "PV",
			      "appliances": [{"id": "b1", "friendlyName": "Grid meter", "applianceReference": 9}],
			      "applianceReferences": [{"id": "s1", "friendlyName": "Phase L1", "reference": "b1"}]
			    }
			  ]
			}`)

		case strings.HasPrefix(r.URL.Path, "/v1/component/appliance/"):
			atomic.AddInt64(&modelCalls, 1)
			switch r.URL.Path {
			case "/v1/component/appliance/7":
				fmt.Fprint(w, `{"friendlyName": "WAGO PFC200"}`)
			case "/v1/component/appliance/9":
				fmt.Fprint(w, `{"friendlyName": "EM4300"}`)
			default:
				w.WriteHeader(http.StatusNotFound)
			}

		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// A non-empty token with a far-future expiry keeps getTokenIfNeeded from
	// reaching for the real token endpoint.
	ds := &Datasource{baseURL: srv.URL, token: "test-token"}
	ds.tokenExpiry = farFuture()

	sender := &captureSender{}
	err := ds.CallResource(
		context.Background(),
		&backend.CallResourceRequest{Path: "appliance-list", URL: "appliance-list?endpointId=ep-1"},
		sender,
	)
	if err != nil {
		t.Fatalf("CallResource: %v", err)
	}
	if sender.resp == nil {
		t.Fatal("no response sent")
	}
	if sender.resp.Status != http.StatusOK {
		t.Fatalf("status %d, body: %s", sender.resp.Status, sender.resp.Body)
	}

	var got []map[string]string
	if err := json.Unmarshal(sender.resp.Body, &got); err != nil {
		t.Fatalf("unmarshal response: %v (body %s)", err, sender.resp.Body)
	}

	want := []map[string]string{
		{"id": "a1", "label": "[HVAC] Heat pump (WAGO PFC200)"},
		{"id": "r1", "label": "[HVAC] Heat pump (WAGO PFC200) – Backup heater"},
		{"id": "r2", "label": "[HVAC] Heat pump (WAGO PFC200) – Compressor"},
		{"id": "a2", "label": "[HVAC] Wallbox (WAGO PFC200)"},
		{"id": "b1", "label": "[PV] Grid meter (EM4300)"},
		{"id": "s1", "label": "[PV] Grid meter (EM4300) – Phase L1"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("options:\n got %#v\nwant %#v", got, want)
	}

	// Two distinct models across three appliances: the lookup must be deduplicated.
	if n := atomic.LoadInt64(&modelCalls); n != 2 {
		t.Errorf("expected 2 model lookups (deduplicated), got %d", n)
	}
}

// TestCallResourceApplianceListMissingEndpointID keeps the existing guard covered.
func TestCallResourceApplianceListMissingEndpointID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no upstream request expected, got %s", r.URL.Path)
	}))
	defer srv.Close()

	ds := &Datasource{baseURL: srv.URL, token: "test-token"}
	ds.tokenExpiry = farFuture()

	sender := &captureSender{}
	if err := ds.CallResource(
		context.Background(),
		&backend.CallResourceRequest{Path: "appliance-list", URL: "appliance-list"},
		sender,
	); err != nil {
		t.Fatalf("CallResource: %v", err)
	}
	if sender.resp.Status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", sender.resp.Status)
	}
}
