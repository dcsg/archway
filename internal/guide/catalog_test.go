package guide

import (
	"strings"
	"testing"
	"testing/fstest"
)

func testCapFS() fstest.MapFS {
	return fstest.MapFS{
		"templates/capabilities/http-api/capability.yaml": &fstest.MapFile{
			Data: []byte("name: http-api\ndescription: \"HTTP API with Chi router\"\nrequires: []\nsuggests: [rate-limiting, auth-jwt, cors]\nconflicts: []\n"),
		},
		"templates/capabilities/mysql/capability.yaml": &fstest.MapFile{
			Data: []byte("name: mysql\ndescription: \"MySQL database adapter\"\nrequires: []\nsuggests: [uuid, migrations]\nconflicts: []\n"),
		},
		"templates/capabilities/docker/capability.yaml": &fstest.MapFile{
			Data: []byte("name: docker\ndescription: \"Dockerfile and docker-compose\"\nrequires: []\nsuggests: []\nconflicts: []\n"),
		},
		"templates/capabilities/circuit-breaker/capability.yaml": &fstest.MapFile{
			Data: []byte("name: circuit-breaker\ndescription: \"Circuit breaker for resilient calls\"\nrequires: []\nsuggests: [observability]\nconflicts: []\n"),
		},
		"templates/capabilities/saga/capability.yaml": &fstest.MapFile{
			Data: []byte("name: saga\ndescription: \"Saga orchestrator\"\nrequires: [event-bus]\nsuggests: [observability]\nconflicts: []\n"),
		},
		"templates/capabilities/rate-limiting/capability.yaml": &fstest.MapFile{
			Data: []byte("name: rate-limiting\ndescription: \"Rate limiter middleware\"\nrequires: []\nsuggests: []\nconflicts: []\n"),
		},
		"templates/capabilities/cors/capability.yaml": &fstest.MapFile{
			Data: []byte("name: cors\ndescription: \"CORS middleware\"\nrequires: []\nsuggests: []\nconflicts: []\n"),
		},
		"templates/capabilities/health/capability.yaml": &fstest.MapFile{
			Data: []byte("name: health\ndescription: \"Health check endpoints\"\nrequires: []\nsuggests: []\nconflicts: []\n"),
		},
		"templates/capabilities/multi-tenancy/capability.yaml": &fstest.MapFile{
			Data: []byte("name: multi-tenancy\ndescription: \"Multi-tenant isolation middleware\"\nrequires: []\nsuggests: [auth-jwt]\nconflicts: []\n"),
		},
		"templates/capabilities/auth-jwt/capability.yaml": &fstest.MapFile{
			Data: []byte("name: auth-jwt\ndescription: \"JWT authentication middleware\"\nrequires: []\nsuggests: []\nconflicts: []\n"),
		},
	}
}

func TestBuildCatalog_WithFS(t *testing.T) {
	installed := []string{"http-api", "mysql"}
	catalog, err := BuildCatalog(testCapFS(), installed)
	if err != nil {
		t.Fatalf("BuildCatalog() error = %v", err)
	}
	if len(catalog) != 10 {
		t.Fatalf("expected 10 entries, got %d", len(catalog))
	}

	installedCount := 0
	for _, e := range catalog {
		if e.Installed {
			installedCount++
		}
		if e.Category == "" {
			t.Errorf("entry %q has empty category", e.Name)
		}
	}
	if installedCount != 2 {
		t.Errorf("expected 2 installed, got %d", installedCount)
	}
}

func TestBuildCatalog_NilFS(t *testing.T) {
	catalog, err := BuildCatalog(nil, []string{"http-api"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if catalog != nil {
		t.Error("expected nil catalog for nil FS")
	}
}

func TestBuildCatalog_EmptyInstalled(t *testing.T) {
	catalog, err := BuildCatalog(testCapFS(), nil)
	if err != nil {
		t.Fatalf("BuildCatalog() error = %v", err)
	}
	for _, e := range catalog {
		if e.Installed {
			t.Errorf("entry %q should not be installed", e.Name)
		}
	}
}

func TestBuildCatalog_SortedByCategoryThenName(t *testing.T) {
	catalog, err := BuildCatalog(testCapFS(), nil)
	if err != nil {
		t.Fatalf("BuildCatalog() error = %v", err)
	}
	for i := 1; i < len(catalog); i++ {
		prev, curr := catalog[i-1], catalog[i]
		if prev.Category > curr.Category {
			t.Errorf("not sorted by category: %s/%s before %s/%s",
				prev.Category, prev.Name, curr.Category, curr.Name)
		}
		if prev.Category == curr.Category && prev.Name > curr.Name {
			t.Errorf("not sorted by name: %s before %s in %s",
				prev.Name, curr.Name, prev.Category)
		}
	}
}

func TestWriteCatalog_Output(t *testing.T) {
	catalog, _ := BuildCatalog(testCapFS(), []string{"http-api", "mysql"})
	var b strings.Builder
	writeCatalog(&b, catalog, []string{"http-api", "mysql"})
	output := b.String()

	checks := []string{
		"## Capability Catalog",
		"### Installed",
		"### Available (not installed)",
		"http-api",
		"mysql",
		"circuit-breaker",
	}
	for _, want := range checks {
		if !strings.Contains(output, want) {
			t.Errorf("missing %q in output", want)
		}
	}
}

func TestWriteCatalog_EmptyCatalog(t *testing.T) {
	var b strings.Builder
	writeCatalog(&b, nil, nil)
	if b.Len() != 0 {
		t.Error("expected empty output for nil catalog")
	}
}

func TestWriteCatalog_SuggestionsMaxFive(t *testing.T) {
	catalog, _ := BuildCatalog(testCapFS(), []string{"http-api"})
	var b strings.Builder
	writeCatalog(&b, catalog, []string{"http-api"})
	output := b.String()
	count := strings.Count(output, "- Consider **")
	if count > 5 {
		t.Errorf("suggestions capped at 5, got %d", count)
	}
}

func TestCategoryFor_Known(t *testing.T) {
	if got := categoryFor("http-api"); got != "transport" {
		t.Errorf("categoryFor(http-api) = %q, want transport", got)
	}
}

func TestCategoryFor_BFF(t *testing.T) {
	if got := categoryFor("bff"); got != "transport" {
		t.Errorf("categoryFor(bff) = %q, want transport", got)
	}
}

func TestCategoryFor_Unknown(t *testing.T) {
	if got := categoryFor("nonexistent"); got != "other" {
		t.Errorf("categoryFor(nonexistent) = %q, want other", got)
	}
}

func TestWhenToUseFor_Known(t *testing.T) {
	if got := whenToUseFor("saga"); got == "" {
		t.Error("expected non-empty whenToUse for saga")
	}
}

func TestWhenToUseFor_Unknown(t *testing.T) {
	if got := whenToUseFor("nonexistent"); got != "" {
		t.Errorf("whenToUseFor(nonexistent) = %q, want empty", got)
	}
}

func TestCatalogInBuildContent(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{"http-api", "mysql"},
		TemplateFS:   testCapFS(),
	}
	content := buildContent(opts)
	if !strings.Contains(content, "## Capability Catalog") {
		t.Error("buildContent should include capability catalog")
	}
}

func TestWriteWarnings_DangerousCombination(t *testing.T) {
	var b strings.Builder
	writeWarnings(&b, []string{"http-api"})
	output := b.String()
	if !strings.Contains(output, "## Interaction Warnings") {
		t.Error("expected Interaction Warnings section")
	}
	if !strings.Contains(output, "rate-limiting") {
		t.Error("expected warning about rate-limiting")
	}
}

func TestWriteWarnings_SafeCombination(t *testing.T) {
	var b strings.Builder
	writeWarnings(&b, []string{"http-api", "rate-limiting", "cors", "health", "observability", "request-id"})
	output := b.String()
	if strings.Contains(output, "## Interaction Warnings") {
		t.Errorf("expected no warnings for safe combination, got: %s", output)
	}
}

func TestWriteWarnings_CriticalFirst(t *testing.T) {
	var b strings.Builder
	// multi-tenancy without auth triggers critical; http-api without health triggers regular.
	writeWarnings(&b, []string{"multi-tenancy", "http-api"})
	output := b.String()

	critIdx := strings.Index(output, "CRITICAL:")
	warnIdx := strings.Index(output, "WARNING:")
	if critIdx < 0 {
		t.Fatal("expected CRITICAL warning")
	}
	if warnIdx < 0 {
		t.Fatal("expected WARNING")
	}
	if critIdx > warnIdx {
		t.Error("CRITICAL warnings should appear before WARNING")
	}
}

func TestWriteSuggestions_ReflectsActualMissing(t *testing.T) {
	var b strings.Builder
	writeSuggestions(&b, []string{"http-api"})
	output := b.String()
	if !strings.Contains(output, "## Architecture Suggestions") {
		t.Error("expected Architecture Suggestions section")
	}
	if !strings.Contains(output, "**Add ") {
		t.Error("expected numbered suggestion format")
	}
	// rate-limiting is a suggestion for http-api
	if !strings.Contains(output, "rate-limiting") {
		t.Error("expected rate-limiting suggestion for http-api")
	}
}

func TestWriteSuggestions_EmptyWhenAllInstalled(t *testing.T) {
	var b strings.Builder
	// Provide capabilities that satisfy all suggestion rules triggered by http-api.
	writeSuggestions(&b, []string{
		"http-api", "platform", "bootstrap", "rate-limiting", "auth-jwt",
		"testing", "docker", "ci-github", "linting", "cors", "health",
		"observability", "request-id",
	})
	output := b.String()
	if strings.Contains(output, "## Architecture Suggestions") {
		t.Errorf("expected no suggestions, got: %s", output)
	}
}

func TestBuildContent_IncludesWarningsAndSuggestions(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{"http-api", "mysql"},
		TemplateFS:   testCapFS(),
	}
	content := buildContent(opts)
	if !strings.Contains(content, "## Interaction Warnings") {
		t.Error("buildContent should include Interaction Warnings")
	}
	if !strings.Contains(content, "## Architecture Suggestions") {
		t.Error("buildContent should include Architecture Suggestions")
	}
}
