package graph

import (
	"path/filepath"
	"testing"

	"github.com/dcsg/archway/internal/config"
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
	rules := []config.DependencyRule{
		{Layer: "domain", Packages: []string{"domain/**"}, MayDependOn: []string{}},
		{Layer: "ports", Packages: []string{"port/**"}, MayDependOn: []string{"domain"}},
		{Layer: "adapters", Packages: []string{"adapter/**"}, MayDependOn: []string{"ports", "domain"}},
	}
	_ = LayerViolations(g, rules)
}
