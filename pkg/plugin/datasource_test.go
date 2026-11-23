package plugin

import (
	"context"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestQueryData(t *testing.T) {
	ds := Datasource{}

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A"},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
}

func TestIsValidServiceOrDataPoint(t *testing.T) {
	valid := []string{"Meter", "my-service.data_point", "a_b-c.d-e_f", "_x.-y"}
	for _, s := range valid {
		if !isValidServiceOrDataPoint(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}

	invalid := []string{"", "a..b", ".a", "a.", "a/b", "../a", "a b", "a%2Fb"}
	for _, s := range invalid {
		if isValidServiceOrDataPoint(s) {
			t.Errorf("expected %q to be invalid", s)
		}
	}
}
