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
