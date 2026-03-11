package graph

import (
	"path/filepath"
	"testing"

	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/provider"
	"golang.org/x/tools/go/packages"
)

func loadPkgs(t *testing.T, dir string) []*packages.Package {
	t.Helper()
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedImports | packages.NeedFiles | packages.NeedModule, Dir: dir}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("packages.Load() error = %v", err)
	}
	return pkgs
}

func TestBuildGraph(t *testing.T) {
	pkgs := loadPkgs(t, filepath.Join("..", "testdata", "hexagonal"))
	g := BuildGraph(pkgs)
	if len(g.Nodes) == 0 {
		t.Fatal("expected nodes")
	}
	if len(g.Edges) == 0 {
		t.Fatal("expected edges")
	}
}

func TestLayerViolations(t *testing.T) {
	pkgs := loadPkgs(t, filepath.Join("..", "testdata", "hexagonal"))
	g := BuildGraph(pkgs)
	components := []config.Component{
		{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{}},
		{Name: "ports", In: []string{"port/**"}, MayDependOn: []string{"domain"}},
		{Name: "adapters", In: []string{"adapter/**"}, MayDependOn: []string{"ports", "domain"}},
	}
	_ = LayerViolations(g, components)
}

func TestGuessLayer(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"example.com/app/domain/order", "domain"},
		{"example.com/app/port/inbound", "ports"},
		{"example.com/app/ports/http", "ports"},
		{"example.com/app/adapter/postgres", "adapters"},
		{"example.com/app/adapters/grpc", "adapters"},
		{"example.com/app/application/service", "application"},
		{"example.com/app/usecase/create", "application"},
		{"example.com/app/infrastructure/db", "infrastructure"},
		{"example.com/app/unknown/thing", ""},
		{"example.com/internal/handler", ""},
		{"example.com/internal/service", ""},
		{"example.com/internal/repository", ""},
		{"example.com/internal/model", ""},
	}
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			got := guessLayer(tc.path)
			if got != tc.want {
				t.Errorf("guessLayer(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestFindCycles(t *testing.T) {
	g := provider.DependencyGraph{
		Nodes: []provider.PackageNode{
			{Path: "a", Name: "a"},
			{Path: "b", Name: "b"},
			{Path: "c", Name: "c"},
		},
		Edges: []provider.DependencyEdge{
			{From: "a", To: "b"},
			{From: "b", To: "c"},
			{From: "c", To: "a"},
		},
	}
	cycles := FindCycles(g)
	if len(cycles) == 0 {
		t.Fatal("expected at least one cycle")
	}
	found := false
	for _, cycle := range cycles {
		if len(cycle) == 3 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a 3-node cycle, got %v", cycles)
	}
}

func TestFindCycles_NoCycles(t *testing.T) {
	g := provider.DependencyGraph{
		Nodes: []provider.PackageNode{
			{Path: "a", Name: "a"},
			{Path: "b", Name: "b"},
			{Path: "c", Name: "c"},
		},
		Edges: []provider.DependencyEdge{
			{From: "a", To: "b"},
			{From: "b", To: "c"},
		},
	}
	cycles := FindCycles(g)
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
}

func TestMatchesComponent(t *testing.T) {
	comp := config.Component{Name: "domain", In: []string{"domain/**"}}
	if !MatchesComponent("example.com/app/domain/order", comp) {
		t.Error("expected MatchesComponent to return true")
	}
	if MatchesComponent("example.com/app/adapter/http", comp) {
		t.Error("expected MatchesComponent to return false")
	}
}

func TestMatchesAnyRule(t *testing.T) {
	tests := []struct {
		name     string
		pkgPath  string
		patterns []string
		want     bool
	}{
		{"glob suffix match", "example.com/app/domain/order", []string{"domain/**"}, true},
		{"exact match", "domain", []string{"domain"}, true},
		{"no match", "example.com/app/other/thing", []string{"domain/**"}, false},
		{"empty pattern skipped", "anything", []string{""}, false},
		{"multiple patterns", "example.com/app/adapter/http", []string{"domain/**", "adapter/**"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesAnyRule(tc.pkgPath, tc.patterns)
			if got != tc.want {
				t.Errorf("matchesAnyRule(%q, %v) = %v, want %v", tc.pkgPath, tc.patterns, got, tc.want)
			}
		})
	}
}

func TestDedupeCycles(t *testing.T) {
	cycles := [][]string{
		{"a", "b", "c"},
		{"a", "b", "c"},
		{"x", "y"},
	}
	result := dedupeCycles(cycles)
	if len(result) != 2 {
		t.Fatalf("expected 2 unique cycles, got %d: %v", len(result), result)
	}
}
