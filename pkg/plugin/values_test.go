package plugin

import (
	"encoding/json"
	"testing"
)

func TestToNullableFloat(t *testing.T) {
	f := func(v float64) *float64 { return &v }

	tests := []struct {
		name string
		in   interface{}
		want *float64
	}{
		{"float", 12.5, f(12.5)},
		{"zero is a real value", 0.0, f(0)},
		{"negative", -3.25, f(-3.25)},
		{"int", 7, f(7)},
		{"int64", int64(7), f(7)},
		{"bool true", true, f(1)},
		{"bool false", false, f(0)},
		{"numeric string", "42.5", f(42.5)},
		{"negative numeric string", "-1", f(-1)},

		// The cases that matter for alerting: these must be nil, never 0.
		{"json null", nil, nil},
		{"non numeric string", "N/A", nil},
		{"empty string", "", nil},
		{"object", map[string]interface{}{"a": 1}, nil},
		{"array", []interface{}{1, 2}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toNullableFloat(tt.in)
			switch {
			case tt.want == nil && got != nil:
				t.Fatalf("expected nil, got %v", *got)
			case tt.want != nil && got == nil:
				t.Fatalf("expected %v, got nil", *tt.want)
			case tt.want != nil && *got != *tt.want:
				t.Fatalf("expected %v, got %v", *tt.want, *got)
			}
		})
	}
}

// TestToNullableFloatFromJSON goes through the same decode path QueryData uses, since
// encoding/json decodes every JSON number into float64 and null into a nil interface.
func TestToNullableFloatFromJSON(t *testing.T) {
	body := []byte(`[
	  {"time": 1, "value": 12.5},
	  {"time": 2, "value": null},
	  {"time": 3, "value": 0},
	  {"time": 4, "value": "oops"},
	  {"time": 5, "value": true}
	]`)

	var points []TimeSeriesDataPoint
	if err := json.Unmarshal(body, &points); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(points) != 5 {
		t.Fatalf("expected 5 points, got %d", len(points))
	}

	values := make([]*float64, 0, len(points))
	for _, p := range points {
		values = append(values, toNullableFloat(p.Value))
	}

	if values[0] == nil || *values[0] != 12.5 {
		t.Errorf("point 0: expected 12.5, got %v", values[0])
	}
	if values[1] != nil {
		t.Errorf("point 1: a JSON null must stay nil, got %v", *values[1])
	}
	if values[2] == nil || *values[2] != 0 {
		t.Errorf("point 2: a real 0 must survive, got %v", values[2])
	}
	if values[3] != nil {
		t.Errorf("point 3: an unparseable string must be nil, got %v", *values[3])
	}
	if values[4] == nil || *values[4] != 1 {
		t.Errorf("point 4: expected 1 for true, got %v", values[4])
	}

	// The distinction the whole change is about: a missing reading and a genuine
	// zero reading must not collapse onto the same value.
	if (values[1] == nil) == (values[2] == nil) {
		t.Error("null and 0 must be distinguishable")
	}
}
