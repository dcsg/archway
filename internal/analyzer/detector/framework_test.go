package detector

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestDetectFramework(t *testing.T) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedImports, Dir: filepath.Join("..", "testdata", "hexagonal")}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("packages.Load: %v", err)
	}
	result := DetectFramework("require github.com/go-chi/chi/v5 v5.2.0", pkgs)
	if result.Name != "chi" {
		t.Fatalf("Name = %q, want chi", result.Name)
	}
}
