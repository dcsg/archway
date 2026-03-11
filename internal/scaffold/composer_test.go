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

func TestComposeProject_UnknownArchitecture(t *testing.T) {
	tfs := testCapabilityFS()
	_, err := ComposeProject(tfs, "nonexistent", nil, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for unknown architecture")
	}
}

func TestComposeProject_InvalidCapabilityManifest(t *testing.T) {
	tfs := fstest.MapFS{
		"templates/architectures/hexagonal/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: hexagonal\n"),
		},
		"templates/capabilities/bad/capability.yaml": &fstest.MapFile{
			Data: []byte(":::invalid"),
		},
	}
	_, err := ComposeProject(tfs, "hexagonal", []string{"bad"}, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for invalid capability manifest")
	}
}

func TestComposeProject_UnknownCapability(t *testing.T) {
	tfs := testCapabilityFS()
	_, err := ComposeProject(tfs, "hexagonal", []string{"nonexistent"}, map[string]interface{}{
		"ServiceName": "test",
		"ModulePath":  "github.com/test/test",
	})
	if err == nil {
		t.Fatal("expected error for unknown capability")
	}
}

func TestComposeProject_NilVars(t *testing.T) {
	tfs := testCapabilityFS()
	_, err := ComposeProject(tfs, "hexagonal", []string{"http-api"}, nil)
	if err == nil {
		t.Fatal("expected error for missing required vars (ServiceName, ModulePath)")
	}
}

func TestComposeProject_HasFlags(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	}
	plan, err := ComposeProject(tfs, "hexagonal", []string{"http-api", "mysql"}, vars)
	if err != nil {
		t.Fatalf("ComposeProject() error = %v", err)
	}
	if plan.Vars["HasHTTP"] != true {
		t.Error("expected HasHTTP = true")
	}
	if plan.Vars["HasMySQL"] != true {
		t.Error("expected HasMySQL = true")
	}
}

func TestComposeProject_CapabilityVariableDefaults(t *testing.T) {
	tfs := fstest.MapFS{
		"templates/architectures/simple/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: simple\ndescription: test\n"),
		},
		"templates/capabilities/with-vars/capability.yaml": &fstest.MapFile{
			Data: []byte(`name: with-vars
description: has variables
variables:
  - name: Port
    type: string
    default: "8080"
  - name: Debug
    type: bool
    default: "true"
`),
		},
	}
	plan, err := ComposeProject(tfs, "simple", []string{"with-vars"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("ComposeProject() error = %v", err)
	}
	if plan.Vars["Port"] != "8080" {
		t.Errorf("Port = %v, want '8080'", plan.Vars["Port"])
	}
	if plan.Vars["Debug"] != true {
		t.Errorf("Debug = %v, want true", plan.Vars["Debug"])
	}
}

func TestComposeProject_BoolCoercion(t *testing.T) {
	tfs := fstest.MapFS{
		"templates/architectures/boolarch/manifest.yaml": &fstest.MapFile{
			Data: []byte(`name: boolarch
description: test
variables:
  - name: EnableFeature
    type: bool
`),
		},
	}
	// Pass a string "true" — should be coerced to bool.
	vars := map[string]interface{}{"EnableFeature": "true"}
	plan, err := ComposeProject(tfs, "boolarch", nil, vars)
	if err != nil {
		t.Fatalf("ComposeProject() error = %v", err)
	}
	if plan.Vars["EnableFeature"] != true {
		t.Errorf("EnableFeature = %v (%T), want bool true", plan.Vars["EnableFeature"], plan.Vars["EnableFeature"])
	}
}

func TestCollectPartials_NoPartialsDir(t *testing.T) {
	tfs := fstest.MapFS{
		"templates/capabilities/nop/capability.yaml": &fstest.MapFile{
			Data: []byte("name: nop\ndescription: no partials\n"),
		},
	}
	partials, err := collectPartials(tfs, []string{"templates/capabilities/nop"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(partials) != 0 {
		t.Errorf("expected 0 partials, got %d", len(partials))
	}
}

func TestCollectPartials_SkipsDirectories(t *testing.T) {
	// The collectPartials function skips directory entries in _partials.
	tfs := testCapabilityFS()
	partials, err := collectPartials(tfs, []string{"templates/capabilities/http-api"}, map[string]interface{}{
		"ModulePath": "github.com/test/test",
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if _, ok := partials["main_imports"]; !ok {
		t.Error("expected main_imports partial")
	}
}

func TestSuggestions_NonexistentCapability(t *testing.T) {
	tfs := testCapabilityFS()
	// Should not error, just skip capabilities with no manifest.
	suggestions := Suggestions(tfs, []string{"nonexistent"})
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions, got %d", len(suggestions))
	}
}

func TestResolveCapabilityDeps_NonexistentDep(t *testing.T) {
	tfs := testCapabilityFS()
	capSet := map[string]bool{"nonexistent": true}
	result := resolveCapabilityDeps(tfs, []string{"nonexistent"}, capSet)
	if len(result) != 1 || result[0] != "nonexistent" {
		t.Errorf("expected [nonexistent], got %v", result)
	}
}

func TestCollectPartials_BadTemplate(t *testing.T) {
	tfs := fstest.MapFS{
		"templates/capabilities/badpartial/_partials/bad.go.tmpl": &fstest.MapFile{
			Data: []byte("{{.X | nonexistent}}"),
		},
	}
	_, err := collectPartials(tfs, []string{"templates/capabilities/badpartial"}, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for bad partial template")
	}
}

func TestCollectPartials_EmptyPartialSkipped(t *testing.T) {
	tfs := fstest.MapFS{
		"templates/capabilities/emptypartial/_partials/empty.go.tmpl": &fstest.MapFile{
			Data: []byte("{{if .Include}}content{{end}}"),
		},
	}
	partials, err := collectPartials(tfs, []string{"templates/capabilities/emptypartial"}, map[string]interface{}{"Include": false})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(partials) != 0 {
		t.Errorf("expected 0 partials for empty render, got %d", len(partials))
	}
}

func TestParseCapabilityManifest_InvalidYAML(t *testing.T) {
	_, err := ParseCapabilityManifest([]byte(":::bad"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseCapabilityManifest_MissingName(t *testing.T) {
	data := []byte(`description: "No name"`)
	_, err := ParseCapabilityManifest(data)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}
