package detector

import (
	"path/filepath"
	"testing"

	"github.com/dcsg/archway/internal/analyzer/graph"
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
