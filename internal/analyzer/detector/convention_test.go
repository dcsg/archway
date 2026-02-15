package detector

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestDetectConventions(t *testing.T) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedImports | packages.NeedSyntax | packages.NeedFiles, Dir: filepath.Join("..", "testdata", "hexagonal")}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("packages.Load: %v", err)
	}
	result := DetectConventions(pkgs)
	if result.Logging.Pattern == "" {
		t.Fatal("expected logging pattern")
	}
	if result.Testing.TotalGoFiles == 0 {
		t.Fatal("expected testing file stats")
	}
}
