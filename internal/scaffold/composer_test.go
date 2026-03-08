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

func TestComposeProject_MissingRequirement(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	}

	_, err := ComposeProject(tfs, "hexagonal", []string{"auth-jwt"}, vars)
	if err == nil {
		t.Fatal("expected error for missing requirement")
	}
}

func TestComposeProject_ConflictDetection(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	}

	_, err := ComposeProject(tfs, "hexagonal", []string{"conflict-a", "conflict-b"}, vars)
	if err == nil {
		t.Fatal("expected error for conflicting capabilities")
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
