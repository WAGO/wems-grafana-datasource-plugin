package plugin

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// labelsOf reduces the option list to its labels, which is what the ordering
// assertions are actually about.
func labelsOf(opts []map[string]string) []string {
	out := make([]string, 0, len(opts))
	for _, o := range opts {
		out = append(out, o["label"])
	}
	return out
}

func idsOf(opts []map[string]string) []string {
	out := make([]string, 0, len(opts))
	for _, o := range opts {
		out = append(out, o["id"])
	}
	return out
}

func TestBuildApplianceOptions(t *testing.T) {
	tests := []struct {
		name        string
		procs       []process
		modelLabels map[int]string
		wantLabels  []string
		wantIDs     []string
	}{
		{
			name: "reference follows its base appliance and carries the model name",
			procs: []process{{
				Name:       "HVAC",
				Appliances: []appliance{{ID: "a1", FriendlyName: "Heat pump", ApplianceReference: 7}},
				ApplianceReferences: []applianceRef{
					{ID: "r1", FriendlyName: "Compressor", Reference: "a1"},
				},
			}},
			modelLabels: map[int]string{7: "WAGO PFC200"},
			wantLabels: []string{
				"[HVAC] Heat pump (WAGO PFC200)",
				"[HVAC] Heat pump (WAGO PFC200) – Compressor",
			},
			wantIDs: []string{"a1", "r1"},
		},
		{
			name: "multiple references to one base are sorted among themselves",
			procs: []process{{
				Name:       "HVAC",
				Appliances: []appliance{{ID: "a1", FriendlyName: "Heat pump", ApplianceReference: 7}},
				ApplianceReferences: []applianceRef{
					{ID: "r2", FriendlyName: "Compressor", Reference: "a1"},
					{ID: "r1", FriendlyName: "Backup heater", Reference: "a1"},
				},
			}},
			modelLabels: map[int]string{7: "WAGO PFC200"},
			wantLabels: []string{
				"[HVAC] Heat pump (WAGO PFC200)",
				"[HVAC] Heat pump (WAGO PFC200) – Backup heater",
				"[HVAC] Heat pump (WAGO PFC200) – Compressor",
			},
			wantIDs: []string{"a1", "r1", "r2"},
		},
		{
			name: "bases sorted by label, each with its own references, across processes",
			procs: []process{
				{
					Name:       "PV",
					Appliances: []appliance{{ID: "b1", FriendlyName: "Grid meter", ApplianceReference: 9}},
					ApplianceReferences: []applianceRef{
						{ID: "s1", FriendlyName: "Phase L1", Reference: "b1"},
					},
				},
				{
					Name:       "HVAC",
					Appliances: []appliance{{ID: "a1", FriendlyName: "Heat pump", ApplianceReference: 7}},
					ApplianceReferences: []applianceRef{
						{ID: "r1", FriendlyName: "Compressor", Reference: "a1"},
					},
				},
			},
			modelLabels: map[int]string{7: "WAGO PFC200", 9: "EM4300"},
			wantLabels: []string{
				"[HVAC] Heat pump (WAGO PFC200)",
				"[HVAC] Heat pump (WAGO PFC200) – Compressor",
				"[PV] Grid meter (EM4300)",
				"[PV] Grid meter (EM4300) – Phase L1",
			},
			wantIDs: []string{"a1", "r1", "b1", "s1"},
		},
		{
			name: "a reference may point at a base appliance in another process",
			procs: []process{
				{
					Name:       "HVAC",
					Appliances: []appliance{{ID: "a1", FriendlyName: "Heat pump"}},
				},
				{
					Name: "PV",
					ApplianceReferences: []applianceRef{
						{ID: "r1", FriendlyName: "Compressor", Reference: "a1"},
					},
				},
			},
			wantLabels: []string{
				"[HVAC] Heat pump",
				"[HVAC] Heat pump – Compressor",
			},
			wantIDs: []string{"a1", "r1"},
		},
		{
			name: "orphan reference stays selectable and is appended last",
			procs: []process{{
				Name:       "HVAC",
				Appliances: []appliance{{ID: "a1", FriendlyName: "Heat pump"}},
				ApplianceReferences: []applianceRef{
					{ID: "r9", FriendlyName: "Dangling", Reference: "does-not-exist"},
					{ID: "r1", FriendlyName: "Compressor", Reference: "a1"},
				},
			}},
			wantLabels: []string{
				"[HVAC] Heat pump",
				"[HVAC] Heat pump – Compressor",
				"[HVAC] Dangling",
			},
			wantIDs: []string{"a1", "r1", "r9"},
		},
		{
			name: "falls back to ids when friendly names are missing",
			procs: []process{{
				Appliances: []appliance{{ID: "a1"}},
				ApplianceReferences: []applianceRef{
					{ID: "r1", Reference: "a1"},
				},
			}},
			wantLabels: []string{"a1", "a1 – r1"},
			wantIDs:    []string{"a1", "r1"},
		},
		{
			name: "unresolved model leaves the label without a model suffix",
			procs: []process{{
				Name:       "HVAC",
				Appliances: []appliance{{ID: "a1", FriendlyName: "Heat pump", ApplianceReference: 7}},
			}},
			modelLabels: map[int]string{},
			wantLabels:  []string{"[HVAC] Heat pump"},
			wantIDs:     []string{"a1"},
		},
		{
			name:       "empty description yields an empty, non-nil list",
			procs:      []process{},
			wantLabels: []string{},
			wantIDs:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildApplianceOptions(tt.procs, tt.modelLabels)
			if gotLabels := labelsOf(got); !reflect.DeepEqual(gotLabels, tt.wantLabels) {
				t.Errorf("labels:\n got %#v\nwant %#v", gotLabels, tt.wantLabels)
			}
			if gotIDs := idsOf(got); !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Errorf("ids:\n got %#v\nwant %#v", gotIDs, tt.wantIDs)
			}
		})
	}
}

// TestBuildApplianceOptionsWithoutReferences guards the regression that matters most:
// a description carrying no applianceReferences must produce exactly what the previous
// implementation produced — the same {id,label} pairs, sorted by label.
func TestBuildApplianceOptionsWithoutReferences(t *testing.T) {
	procs := []process{
		{
			Name: "PV",
			Appliances: []appliance{
				{ID: "b1", FriendlyName: "Grid meter", ApplianceReference: 9},
			},
		},
		{
			Name: "HVAC",
			Appliances: []appliance{
				{ID: "a2", FriendlyName: "Wallbox", ApplianceReference: 7},
				{ID: "a1", FriendlyName: "Heat pump", ApplianceReference: 7},
			},
		},
	}
	modelLabels := map[int]string{7: "WAGO PFC200", 9: "EM4300"}

	got := buildApplianceOptions(procs, modelLabels)
	want := []string{
		"[HVAC] Heat pump (WAGO PFC200)",
		"[HVAC] Wallbox (WAGO PFC200)",
		"[PV] Grid meter (EM4300)",
	}
	if gotLabels := labelsOf(got); !reflect.DeepEqual(gotLabels, want) {
		t.Errorf("labels:\n got %#v\nwant %#v", gotLabels, want)
	}
}

// TestDescRespUnmarshal pins the JSON contract against the new endpoint description
// format, including the fields this plugin deliberately ignores.
func TestDescRespUnmarshal(t *testing.T) {
	body := []byte(`{
	  "building": {"metadata": {}, "spaces": [{"label": "Room 1", "metadata": {}}]},
	  "processes": [{
	    "id": "p1",
	    "name": "HVAC",
	    "category": "heating",
	    "applianceReferences": [{
	      "friendlyName": "Compressor",
	      "functionalProfileConfiguration": [{"channel": 2, "localOnly": true, "updateInterval": 60, "uri": "u"}],
	      "id": "r1",
	      "locationSpaceRef": "Room 1",
	      "metadata": {},
	      "reference": "a1"
	    }],
	    "appliances": [{
	      "applianceReference": 7,
	      "component": "c",
	      "configuration": {"k": "v"},
	      "friendlyName": "Heat pump",
	      "id": "a1",
	      "portMappings": [{"protocol": "TCP", "source": 1, "target": 2}],
	      "version": "1.0"
	    }]
	  }]
	}`)

	var desc descResp
	if err := json.Unmarshal(body, &desc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(desc.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(desc.Processes))
	}
	proc := desc.Processes[0]
	if len(proc.Appliances) != 1 || proc.Appliances[0].ID != "a1" || proc.Appliances[0].ApplianceReference != 7 {
		t.Errorf("appliances not parsed: %#v", proc.Appliances)
	}
	if len(proc.ApplianceReferences) != 1 {
		t.Fatalf("expected 1 appliance reference, got %d", len(proc.ApplianceReferences))
	}
	ref := proc.ApplianceReferences[0]
	if ref.ID != "r1" || ref.FriendlyName != "Compressor" || ref.Reference != "a1" {
		t.Errorf("appliance reference not parsed: %#v", ref)
	}

	// End to end through the flattener, with the model name resolved.
	got := labelsOf(buildApplianceOptions(desc.Processes, map[int]string{7: "WAGO PFC200"}))
	want := []string{
		"[HVAC] Heat pump (WAGO PFC200)",
		"[HVAC] Heat pump (WAGO PFC200) – Compressor",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("labels:\n got %#v\nwant %#v", got, want)
	}
}

// farFuture returns a token expiry far enough ahead that getTokenIfNeeded treats the
// preset test token as valid and never calls the token endpoint.
func farFuture() time.Time {
	return time.Now().Add(24 * time.Hour)
}
