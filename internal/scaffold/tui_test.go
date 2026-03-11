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

func TestBuildFieldInput(t *testing.T) {
	q := WizardQuestion{
		Variable: "Name",
		Prompt:   "Service name?",
		Type:     "input",
	}
	state := map[string]interface{}{"Name": "existing"}
	field, err := buildField(q, VariableDefinition{Name: "Name", Type: "string"}, state)
	if err != nil {
		t.Fatalf("buildField() error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	// State should preserve existing value.
	if state["Name"] != "existing" {
		t.Errorf("state[Name] = %v, want 'existing'", state["Name"])
	}
}

func TestBuildFieldInputWithValidation(t *testing.T) {
	q := WizardQuestion{
		Variable: "Name",
		Prompt:   "Name?",
		Type:     "input",
		Validate: `^[a-z]+$`,
	}
	state := map[string]interface{}{}
	field, err := buildField(q, VariableDefinition{Name: "Name", Type: "string"}, state)
	if err != nil {
		t.Fatalf("buildField() error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

func TestBuildFieldInputInvalidRegex(t *testing.T) {
	q := WizardQuestion{
		Variable: "Name",
		Prompt:   "Name?",
		Type:     "input",
		Validate: `[invalid`,
	}
	_, err := buildField(q, VariableDefinition{}, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestBuildFieldConfirm(t *testing.T) {
	q := WizardQuestion{
		Variable: "UseAuth",
		Prompt:   "Enable auth?",
		Type:     "confirm",
	}
	state := map[string]interface{}{"UseAuth": true}
	field, err := buildField(q, VariableDefinition{Name: "UseAuth", Type: "bool"}, state)
	if err != nil {
		t.Fatalf("buildField() error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

func TestBuildFieldDefaultType(t *testing.T) {
	// Empty type defaults to "input".
	q := WizardQuestion{
		Variable: "Name",
		Prompt:   "Name?",
		Type:     "",
	}
	state := map[string]interface{}{}
	field, err := buildField(q, VariableDefinition{}, state)
	if err != nil {
		t.Fatalf("buildField() error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field (defaulting to input)")
	}
}

func TestBuildFieldInputNilState(t *testing.T) {
	q := WizardQuestion{
		Variable: "Name",
		Prompt:   "Name?",
		Type:     "input",
	}
	state := map[string]interface{}{}
	field, err := buildField(q, VariableDefinition{Name: "Name", Type: "string"}, state)
	if err != nil {
		t.Fatalf("buildField() error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	// state should have empty string for Name.
	if state["Name"] != "" {
		t.Errorf("state[Name] = %v, want ''", state["Name"])
	}
}

func TestBuildFieldConfirmFalseDefault(t *testing.T) {
	q := WizardQuestion{
		Variable: "Flag",
		Prompt:   "Enable?",
		Type:     "confirm",
	}
	state := map[string]interface{}{} // no "Flag" key — should default to false
	field, err := buildField(q, VariableDefinition{Name: "Flag", Type: "bool"}, state)
	if err != nil {
		t.Fatalf("buildField() error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	if state["Flag"] != false {
		t.Errorf("state[Flag] = %v, want false", state["Flag"])
	}
}

func TestBuildFieldSelectWithExistingValue(t *testing.T) {
	q := WizardQuestion{
		Variable: "Lang",
		Prompt:   "Language?",
		Type:     "select",
		Options: []FlexOption{
			{Label: "Go", Value: "go"},
			{Label: "TypeScript", Value: "ts"},
		},
	}
	state := map[string]interface{}{"Lang": "ts"} // pre-set to second option
	field, err := buildField(q, VariableDefinition{}, state)
	if err != nil {
		t.Fatalf("buildField() error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	if state["Lang"] != "ts" {
		t.Errorf("state[Lang] = %v, want 'ts'", state["Lang"])
	}
}

func TestBuildFieldSelectEmptyOptions(t *testing.T) {
	q := WizardQuestion{
		Variable: "Choice",
		Prompt:   "Pick",
		Type:     "select",
		Options:  []FlexOption{},
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

func TestBuildFieldMultiselectWithExisting(t *testing.T) {
	q := WizardQuestion{
		Variable: "Caps",
		Prompt:   "Caps?",
		Type:     "multiselect",
		Options: []FlexOption{
			{Label: "A", Value: "a"},
			{Label: "B", Value: "b"},
		},
	}
	state := map[string]interface{}{"Caps": []string{"a"}}
	field, err := buildField(q, VariableDefinition{}, state)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

func TestBuildFieldUnsupportedType(t *testing.T) {
	q := WizardQuestion{
		Variable: "X",
		Prompt:   "X?",
		Type:     "slider",
	}
	_, err := buildField(q, VariableDefinition{}, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestEvaluateWhen_NonEmptyString(t *testing.T) {
	// A non-empty string value should return true.
	if !evaluateWhen("Foo", map[string]interface{}{"Foo": "bar"}) {
		t.Fatal("expected true for non-empty string")
	}
	// An empty string should return false.
	if evaluateWhen("Foo", map[string]interface{}{"Foo": ""}) {
		t.Fatal("expected false for empty string")
	}
	// Whitespace-only string should return false.
	if evaluateWhen("Foo", map[string]interface{}{"Foo": "  "}) {
		t.Fatal("expected false for whitespace string")
	}
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
