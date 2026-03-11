package scaffold

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func testCapabilityFS() fs.FS {
	return fstest.MapFS{
		"templates/architectures/hexagonal/manifest.yaml": &fstest.MapFile{
			Data: []byte(`name: hexagonal
description: "Hexagonal architecture"
variables:
  - name: ServiceName
    type: string
    required: true
  - name: ModulePath
    type: string
    required: true
`),
		},
		"templates/architectures/hexagonal/files/go.mod.tmpl": &fstest.MapFile{
			Data: []byte("module {{.ModulePath}}\n"),
		},
		"templates/capabilities/http-api/capability.yaml": &fstest.MapFile{
			Data: []byte(`name: http-api
description: "HTTP API"
requires: []
suggests:
  - rate-limiting
  - auth-jwt
conflicts: []
`),
		},
		"templates/capabilities/http-api/files/handler.go.tmpl": &fstest.MapFile{
			Data: []byte("package httphandler\n"),
		},
		"templates/capabilities/http-api/_partials/main_imports.go.tmpl": &fstest.MapFile{
			Data: []byte(`"{{.ModulePath}}/adapter/httphandler"`),
		},
		"templates/capabilities/mysql/capability.yaml": &fstest.MapFile{
			Data: []byte(`name: mysql
description: "MySQL"
requires: []
suggests:
  - observability
conflicts: []
`),
		},
		"templates/capabilities/mysql/files/repo.go.tmpl": &fstest.MapFile{
			Data: []byte("package mysqlrepo\n"),
		},
		"templates/capabilities/mysql/_partials/main_imports.go.tmpl": &fstest.MapFile{
			Data: []byte(`"{{.ModulePath}}/adapter/mysqlrepo"`),
		},
		"templates/capabilities/mysql/_partials/main_init.go.tmpl": &fstest.MapFile{
			Data: []byte("db := mysqlrepo.New(cfg)"),
		},
		"templates/capabilities/auth-jwt/capability.yaml": &fstest.MapFile{
			Data: []byte(`name: auth-jwt
description: "JWT auth"
requires:
  - http-api
suggests: []
conflicts: []
`),
		},
		"templates/capabilities/auth-jwt/files/auth.go.tmpl": &fstest.MapFile{
			Data: []byte("package auth\n"),
		},
		"templates/capabilities/conflict-a/capability.yaml": &fstest.MapFile{
			Data: []byte(`name: conflict-a
description: "Conflicts with conflict-b"
requires: []
suggests: []
conflicts:
  - conflict-b
`),
		},
		"templates/capabilities/conflict-b/capability.yaml": &fstest.MapFile{
			Data: []byte(`name: conflict-b
description: "Conflicts with conflict-a"
requires: []
suggests: []
conflicts:
  - conflict-a
`),
		},
	}
}

func TestComposeProject(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	}

	plan, err := ComposeProject(tfs, "hexagonal", []string{"http-api", "mysql"}, vars)
	if err != nil {
		t.Fatalf("ComposeProject() error = %v", err)
	}

	if plan.Architecture != "hexagonal" {
		t.Errorf("Architecture = %q, want hexagonal", plan.Architecture)
	}
	if len(plan.Capabilities) != 2 {
		t.Errorf("Capabilities = %d, want 2", len(plan.Capabilities))
	}
	if len(plan.CapManifests) != 2 {
		t.Errorf("CapManifests = %d, want 2", len(plan.CapManifests))
	}

	// Check partials were collected.
	imports, ok := plan.Partials["main_imports"]
	if !ok || len(imports) != 2 {
		t.Errorf("main_imports partials = %d, want 2", len(imports))
	}
	init, ok := plan.Partials["main_init"]
	if !ok || len(init) != 1 {
		t.Errorf("main_init partials = %d, want 1", len(init))
	}
}

func TestComposeProject_AutoResolveDependencies(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	}

	// auth-jwt requires http-api — should be auto-resolved.
	plan, err := ComposeProject(tfs, "hexagonal", []string{"auth-jwt"}, vars)
	if err != nil {
		t.Fatalf("ComposeProject() should auto-resolve deps, got error: %v", err)
	}

	// http-api should have been auto-added.
	capSet := map[string]bool{}
	for _, c := range plan.Capabilities {
		capSet[c] = true
	}
	if !capSet["http-api"] {
		t.Errorf("expected http-api to be auto-resolved, got capabilities: %v", plan.Capabilities)
	}
	if !capSet["auth-jwt"] {
		t.Errorf("expected auth-jwt in capabilities, got: %v", plan.Capabilities)
	}
	// Dependencies should come before dependents.
	httpIdx := -1
	authIdx := -1
	for i, c := range plan.Capabilities {
		if c == "http-api" {
			httpIdx = i
		}
		if c == "auth-jwt" {
			authIdx = i
		}
	}
	if httpIdx > authIdx {
		t.Errorf("http-api (idx %d) should come before auth-jwt (idx %d)", httpIdx, authIdx)
	}
}

func TestComposeProject_ConflictDetection(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	}

	plan, err := ComposeProject(tfs, "hexagonal", []string{"conflict-a", "conflict-b"}, vars)
	if err != nil {
		t.Fatalf("conflicts should warn, not error: %v", err)
	}
	if len(plan.Warnings) == 0 {
		t.Fatal("expected warnings for conflicting capabilities")
	}
}

func TestComposeProject_MissingVariable(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{}

	_, err := ComposeProject(tfs, "hexagonal", []string{"http-api"}, vars)
	if err == nil {
		t.Fatal("expected error for missing required variable")
	}
}

func TestComposeProject_NoDuplicatesInAutoResolve(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	}

	// Both auth-jwt and http-api selected — http-api should not be duplicated.
	plan, err := ComposeProject(tfs, "hexagonal", []string{"http-api", "auth-jwt"}, vars)
	if err != nil {
		t.Fatalf("ComposeProject() error = %v", err)
	}

	seen := map[string]int{}
	for _, c := range plan.Capabilities {
		seen[c]++
		if seen[c] > 1 {
			t.Errorf("capability %q appears %d times", c, seen[c])
		}
	}
}

func TestComposeProject_NoDepsMeansNoChange(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	}

	// http-api has no requires — capabilities should stay the same.
	plan, err := ComposeProject(tfs, "hexagonal", []string{"http-api"}, vars)
	if err != nil {
		t.Fatalf("ComposeProject() error = %v", err)
	}

	if len(plan.Capabilities) != 1 || plan.Capabilities[0] != "http-api" {
		t.Errorf("expected [http-api], got %v", plan.Capabilities)
	}
}

func TestSuggestions(t *testing.T) {
	tfs := testCapabilityFS()
	suggestions := Suggestions(tfs, []string{"http-api"})

	if len(suggestions) != 2 {
		t.Fatalf("Suggestions = %v, want [rate-limiting auth-jwt]", suggestions)
	}
}

func TestSuggestions_AlreadySelected(t *testing.T) {
	tfs := testCapabilityFS()
	suggestions := Suggestions(tfs, []string{"http-api", "auth-jwt"})

	// auth-jwt is already selected, should only suggest rate-limiting.
	if len(suggestions) != 1 || suggestions[0] != "rate-limiting" {
		t.Fatalf("Suggestions = %v, want [rate-limiting]", suggestions)
	}
}

func TestComposeProject_BFFAutoResolvesHTTPAPI(t *testing.T) {
	tfs := testCapabilityFS()
	// Add BFF capability to test FS.
	tfs.(fstest.MapFS)["templates/capabilities/bff/capability.yaml"] = &fstest.MapFile{
		Data: []byte("name: bff\ndescription: \"BFF gateway\"\nrequires:\n  - http-api\nsuggests:\n  - circuit-breaker\nconflicts: []\n"),
	}
	tfs.(fstest.MapFS)["templates/capabilities/bff/files/gateway.go.tmpl"] = &fstest.MapFile{
		Data: []byte("package bffgateway\n"),
	}

	vars := map[string]interface{}{
		"ServiceName": "web-bff",
		"ModulePath":  "github.com/acme/web-bff",
	}

	// Selecting only bff should auto-resolve http-api.
	plan, err := ComposeProject(tfs, "hexagonal", []string{"bff"}, vars)
	if err != nil {
		t.Fatalf("ComposeProject() error = %v", err)
	}

	capSet := map[string]bool{}
	for _, c := range plan.Capabilities {
		capSet[c] = true
	}
	if !capSet["http-api"] {
		t.Errorf("expected http-api auto-resolved, got: %v", plan.Capabilities)
	}
	if !capSet["bff"] {
		t.Errorf("expected bff in capabilities, got: %v", plan.Capabilities)
	}

	// http-api (dependency) should come before bff (dependent).
	httpIdx, bffIdx := -1, -1
	for i, c := range plan.Capabilities {
		if c == "http-api" {
			httpIdx = i
		}
		if c == "bff" {
			bffIdx = i
		}
	}
	if httpIdx > bffIdx {
		t.Errorf("http-api (idx %d) should come before bff (idx %d)", httpIdx, bffIdx)
	}

	// HasBFF flag should be set.
	if !plan.Vars["HasBFF"].(bool) {
		t.Error("expected HasBFF flag to be true")
	}
}

func TestParseCapabilityManifest(t *testing.T) {
	data := []byte(`name: http-api
description: "HTTP API"
requires:
  - auth
suggests:
  - rate-limiting
conflicts:
  - grpc
`)
	cm, err := ParseCapabilityManifest(data)
	if err != nil {
		t.Fatalf("ParseCapabilityManifest() error = %v", err)
	}
	if cm.Name != "http-api" {
		t.Errorf("Name = %q", cm.Name)
	}
	if len(cm.Requires) != 1 {
		t.Errorf("Requires = %v", cm.Requires)
	}
	if len(cm.Suggests) != 1 {
		t.Errorf("Suggests = %v", cm.Suggests)
	}
	if len(cm.Conflicts) != 1 {
		t.Errorf("Conflicts = %v", cm.Conflicts)
	}
}

func TestParseCapabilityManifest_MissingName(t *testing.T) {
	data := []byte(`description: "No name"`)
	_, err := ParseCapabilityManifest(data)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}
