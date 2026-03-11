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
	}
}

func TestBuildCatalog_WithFS(t *testing.T) {
	installed := []string{"http-api", "mysql"}
	catalog, err := BuildCatalog(testCapFS(), installed)
	if err != nil {
		t.Fatalf("BuildCatalog() error = %v", err)
	}
	if len(catalog) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(catalog))
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
