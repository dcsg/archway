package scaffold

import "testing"

func TestEvaluateWhen(t *testing.T) {
	vars := map[string]interface{}{"Transport": "http", "UseAuth": true}
	if !evaluateWhen(`Transport=="http"`, vars) {
		t.Fatal("expected true")
	}
	if evaluateWhen(`Transport=="grpc"`, vars) {
		t.Fatal("expected false")
	}
	if !evaluateWhen("UseAuth", vars) {
		t.Fatal("expected true for bool")
	}
}

func TestEvaluateWhenNotEqual(t *testing.T) {
	vars := map[string]interface{}{"Stage": "production"}
	if !evaluateWhen(`Stage != "prototype"`, vars) {
		t.Fatal("expected true")
	}
	if evaluateWhen(`Stage != "production"`, vars) {
		t.Fatal("expected false")
	}
}

func TestEvaluateWhenEmpty(t *testing.T) {
	if !evaluateWhen("", nil) {
		t.Fatal("empty expr should return true")
	}
}

func TestEvaluateWhenMissing(t *testing.T) {
	if evaluateWhen("Missing", map[string]interface{}{}) {
		t.Fatal("missing var should return false")
	}
}

func TestBuildWizardGroups(t *testing.T) {
	cfg := &WizardConfig{Steps: []WizardStep{{
		ID:        "basics",
		Questions: []WizardQuestion{{Variable: "ServiceName", Prompt: "Service name", Type: "input"}},
	}}}
	manifest := &Manifest{Name: "x", Language: "go", Variables: []VariableDefinition{{Name: "ServiceName", Type: "string", Required: true}}}
	groups, err := buildWizardGroups(cfg, manifest, map[string]interface{}{})
	if err != nil {
		t.Fatalf("buildWizardGroups() error = %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("groups len = %d, want 1", len(groups))
	}
}

func TestBuildWizardGroupsSkipsWhenFalse(t *testing.T) {
	cfg := &WizardConfig{Steps: []WizardStep{{
		ID: "conditional",
		Questions: []WizardQuestion{{
			Variable: "UseAuth",
			Prompt:   "Enable auth?",
			Type:     "confirm",
			When:     "HasHTTP",
		}},
	}}}
	manifest := &Manifest{Name: "x", Language: "go"}
	state := map[string]interface{}{"HasHTTP": false}
	groups, err := buildWizardGroups(cfg, manifest, state)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("groups len = %d, want 0 (question skipped)", len(groups))
	}
}

func TestComputeDerived(t *testing.T) {
	tests := []struct {
		name          string
		caps          []string
		wantHasAPI    bool
		wantHasWorker bool
		wantHasCLI    bool
		wantCLIOnly   bool
		wantHasAPIOr  bool
	}{
		{
			name:       "api only",
			caps:       []string{"api"},
			wantHasAPI: true, wantHasAPIOr: true,
		},
		{
			name:       "cli only",
			caps:       []string{"cli"},
			wantHasCLI: true, wantCLIOnly: true,
		},
		{
			name:       "api and cli",
			caps:       []string{"api", "cli"},
			wantHasAPI: true, wantHasCLI: true, wantHasAPIOr: true,
		},
		{
			name:       "all three",
			caps:       []string{"api", "worker", "cli"},
			wantHasAPI: true, wantHasWorker: true, wantHasCLI: true, wantHasAPIOr: true,
		},
		{
			name: "empty",
			caps: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := map[string]interface{}{}
			if tt.caps != nil {
				state["ProjectCapabilities"] = tt.caps
			}
			computeDerived(state)
			if state["HasAPI"] != tt.wantHasAPI {
				t.Fatalf("HasAPI = %v, want %v", state["HasAPI"], tt.wantHasAPI)
			}
			if state["HasWorker"] != tt.wantHasWorker {
				t.Fatalf("HasWorker = %v, want %v", state["HasWorker"], tt.wantHasWorker)
			}
			if state["HasCLI"] != tt.wantHasCLI {
				t.Fatalf("HasCLI = %v, want %v", state["HasCLI"], tt.wantHasCLI)
			}
			if state["CLIOnly"] != tt.wantCLIOnly {
				t.Fatalf("CLIOnly = %v, want %v", state["CLIOnly"], tt.wantCLIOnly)
			}
			if state["HasAPIOrWorker"] != tt.wantHasAPIOr {
				t.Fatalf("HasAPIOrWorker = %v, want %v", state["HasAPIOrWorker"], tt.wantHasAPIOr)
			}
		})
	}
}

func TestSliceContains(t *testing.T) {
	if !sliceContains([]string{"a", "b", "c"}, "b") {
		t.Fatal("expected true for 'b'")
	}
	if sliceContains([]string{"a", "b"}, "c") {
		t.Fatal("expected false for 'c'")
	}
	if sliceContains(nil, "a") {
		t.Fatal("expected false for nil slice")
	}
}

func TestFlexOptionsForQuestion(t *testing.T) {
	t.Run("from question options", func(t *testing.T) {
		q := WizardQuestion{
			Options: []FlexOption{
				{Label: "Foo Bar", Value: "foo"},
				{Label: "Baz", Value: "baz"},
			},
		}
		opts := flexOptionsForQuestion(q, VariableDefinition{})
		if len(opts) != 2 || opts[0].Value != "foo" {
			t.Fatalf("opts = %+v", opts)
		}
	})

	t.Run("from manifest choices", func(t *testing.T) {
		q := WizardQuestion{}
		def := VariableDefinition{Choices: []string{"alpha", "beta"}}
		opts := flexOptionsForQuestion(q, def)
		if len(opts) != 2 || opts[0].Label != "alpha" || opts[0].Value != "alpha" {
			t.Fatalf("opts = %+v", opts)
		}
	})
}

func TestBuildFieldSelect(t *testing.T) {
	q := WizardQuestion{
		Variable: "Lang",
		Prompt:   "Language?",
		Type:     "select",
		Options: []FlexOption{
			{Label: "Go", Value: "go"},
			{Label: "TypeScript", Value: "ts"},
		},
	}
	state := map[string]interface{}{}
	field, err := buildField(q, VariableDefinition{}, state)
	if err != nil {
		t.Fatalf("buildField() error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	// State should be pre-seeded with first option.
	if state["Lang"] != "go" {
		t.Fatalf("state[Lang] = %v, want go", state["Lang"])
	}
}

func TestBuildFieldMultiselect(t *testing.T) {
	q := WizardQuestion{
		Variable: "Caps",
		Prompt:   "Capabilities?",
		Type:     "multiselect",
		Options: []FlexOption{
			{Label: "API", Value: "api"},
			{Label: "Worker", Value: "worker"},
		},
	}
	state := map[string]interface{}{}
	field, err := buildField(q, VariableDefinition{}, state)
	if err != nil {
		t.Fatalf("buildField() error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}
