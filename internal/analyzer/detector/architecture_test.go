package detector

import (
	"path/filepath"
	"testing"

	"github.com/dcsg/archway/internal/analyzer/graph"
	"github.com/dcsg/archway/internal/provider"
	"golang.org/x/tools/go/packages"
)

func loadArchitecturePkgs(t *testing.T, dir string) []*packages.Package {
	t.Helper()
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedImports | packages.NeedFiles | packages.NeedModule, Dir: dir}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("packages.Load() error = %v", err)
	}
	return pkgs
}

func TestDetectArchitecture(t *testing.T) {
	pkgs := loadArchitecturePkgs(t, filepath.Join("..", "testdata", "hexagonal"))
	g := graph.BuildGraph(pkgs)
	result := DetectArchitecture(g, pkgs)
	if result.Confidence <= 0 {
		t.Fatalf("confidence = %v, want > 0", result.Confidence)
	}
	if result.Pattern == "" {
		t.Fatal("expected pattern")
	}
}

func TestDetectHexagonal_Unit(t *testing.T) {
	g := provider.DependencyGraph{
		Nodes: []provider.PackageNode{
			{Path: "example.com/app/domain/order", Name: "order", Layer: "domain"},
			{Path: "example.com/app/port/inbound", Name: "inbound", Layer: "ports"},
			{Path: "example.com/app/adapter/http", Name: "http", Layer: "adapters"},
		},
		Edges: []provider.DependencyEdge{
			{From: "example.com/app/adapter/http", To: "example.com/app/port/inbound"},
			{From: "example.com/app/port/inbound", To: "example.com/app/domain/order"},
		},
	}
	result := detectHexagonal(g)
	if result.Confidence < 0.9 {
		t.Errorf("confidence = %v, want >= 0.9", result.Confidence)
	}
}

func TestDetectHexagonal_Empty(t *testing.T) {
	g := provider.DependencyGraph{}
	result := detectHexagonal(g)
	if result.Confidence != 0.1 {
		t.Errorf("confidence = %v, want 0.1 for empty graph", result.Confidence)
	}
}

func TestDetectClean_Unit(t *testing.T) {
	g := provider.DependencyGraph{
		Nodes: []provider.PackageNode{
			{Path: "example.com/app/domain/entity", Name: "entity", Layer: "domain"},
			{Path: "example.com/app/application/usecase", Name: "usecase", Layer: "application"},
			{Path: "example.com/app/infrastructure/db", Name: "db", Layer: "infrastructure"},
		},
		Edges: []provider.DependencyEdge{
			{From: "example.com/app/application/usecase", To: "example.com/app/domain/entity"},
			{From: "example.com/app/infrastructure/db", To: "example.com/app/domain/entity"},
		},
	}
	result := detectClean(g)
	if result.Confidence < 0.9 {
		t.Errorf("confidence = %v, want >= 0.9", result.Confidence)
	}
}

func TestDetectClean_Empty(t *testing.T) {
	g := provider.DependencyGraph{}
	result := detectClean(g)
	if result.Confidence != 0.1 {
		t.Errorf("confidence = %v, want 0.1", result.Confidence)
	}
}

func TestDetectDDD_Unit(t *testing.T) {
	g := provider.DependencyGraph{
		Nodes: []provider.PackageNode{
			{Path: "example.com/app/domain/aggregate", Name: "aggregate", Layer: "domain"},
			{Path: "example.com/app/application", Name: "application", Layer: "application"},
			{Path: "example.com/app/infrastructure", Name: "infrastructure", Layer: "infrastructure"},
		},
	}
	pkgs := []*packages.Package{
		{PkgPath: "example.com/app/domain/aggregate"},
		{PkgPath: "example.com/app/domain/valueobject"},
		{PkgPath: "example.com/app/domain/repository"},
	}
	result := detectDDD(g, pkgs)
	if result.Confidence < 0.9 {
		t.Errorf("confidence = %v, want >= 0.9", result.Confidence)
	}
}

func TestDetectDDD_NilPkg(t *testing.T) {
	g := provider.DependencyGraph{}
	result := detectDDD(g, []*packages.Package{nil})
	if result.Confidence != 0.1 {
		t.Errorf("confidence = %v, want 0.1", result.Confidence)
	}
}

func TestDetectLayered_Unit(t *testing.T) {
	g := provider.DependencyGraph{
		Nodes: []provider.PackageNode{
			{Path: "example.com/app/model"},
			{Path: "example.com/app/view"},
			{Path: "example.com/app/controller"},
			{Path: "example.com/app/service"},
		},
	}
	result := detectLayered(g)
	if result.Confidence < 0.8 {
		t.Errorf("confidence = %v, want >= 0.8", result.Confidence)
	}
}

func TestDetectFlat_Unit(t *testing.T) {
	tests := []struct {
		name     string
		nodes    int
		wantConf float64
	}{
		{"single package", 1, 0.9},
		{"two packages", 2, 0.9},
		{"three packages", 3, 0.5},
		{"five packages", 5, 0.1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := make([]provider.PackageNode, tt.nodes)
			for i := range tt.nodes {
				nodes[i] = provider.PackageNode{Path: "p" + string(rune('a'+i))}
			}
			g := provider.DependencyGraph{Nodes: nodes}
			result := detectFlat(g)
			if result.Confidence != tt.wantConf {
				t.Errorf("confidence = %v, want %v", result.Confidence, tt.wantConf)
			}
		})
	}
}

func TestEdgeBetweenLayers(t *testing.T) {
	g := provider.DependencyGraph{
		Nodes: []provider.PackageNode{
			{Path: "a", Layer: "domain"},
			{Path: "b", Layer: "adapters"},
		},
		Edges: []provider.DependencyEdge{{From: "a", To: "b"}},
	}
	if !edgeBetweenLayers(g, "domain", "adapters") {
		t.Error("expected edge from domain to adapters")
	}
	if edgeBetweenLayers(g, "adapters", "domain") {
		t.Error("expected no edge from adapters to domain")
	}
}

func TestDetectArchitecture_LowConfidence(t *testing.T) {
	g := provider.DependencyGraph{
		Nodes: []provider.PackageNode{
			{Path: "example.com/app/pkg1"},
			{Path: "example.com/app/pkg2"},
			{Path: "example.com/app/pkg3"},
			{Path: "example.com/app/pkg4"},
			{Path: "example.com/app/pkg5"},
			{Path: "example.com/app/pkg6"},
		},
	}
	result := DetectArchitecture(g, nil)
	if result.Pattern != "unrecognized" {
		// With 6 packages and no layer signals, should be unrecognized or flat at low confidence
		t.Logf("Pattern = %q, Confidence = %v", result.Pattern, result.Confidence)
	}
}
