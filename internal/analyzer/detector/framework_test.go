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

func TestDetectFramework_Gin(t *testing.T) {
	goMod := "require github.com/gin-gonic/gin v1.9.0"
	result := DetectFramework(goMod, nil)
	if result.Name != "gin" {
		t.Fatalf("Name = %q, want gin", result.Name)
	}
	if result.Version != "v1.9.0" {
		t.Errorf("Version = %q, want v1.9.0", result.Version)
	}
}

func TestDetectFramework_Echo(t *testing.T) {
	goMod := "require github.com/labstack/echo/v4 v4.11.0"
	result := DetectFramework(goMod, nil)
	if result.Name != "echo" {
		t.Fatalf("Name = %q, want echo", result.Name)
	}
}

func TestDetectFramework_Fiber(t *testing.T) {
	goMod := "require github.com/gofiber/fiber/v2 v2.50.0"
	result := DetectFramework(goMod, nil)
	if result.Name != "fiber" {
		t.Fatalf("Name = %q, want fiber", result.Name)
	}
}

func TestDetectFramework_GRPC(t *testing.T) {
	goMod := "require google.golang.org/grpc v1.60.0"
	result := DetectFramework(goMod, nil)
	if result.Name != "grpc" {
		t.Fatalf("Name = %q, want grpc", result.Name)
	}
}

func TestDetectFramework_Stdlib(t *testing.T) {
	result := DetectFramework("module myapp\ngo 1.21", nil)
	if result.Name != "stdlib" {
		t.Fatalf("Name = %q, want stdlib", result.Name)
	}
}

func TestDetectFramework_DBLibraries(t *testing.T) {
	goMod := `module myapp
require (
	gorm.io/gorm v1.25.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/jackc/pgx/v5 v5.4.0
)`
	result := DetectFramework(goMod, nil)
	if len(result.Libraries) != 3 {
		t.Fatalf("expected 3 DB libraries, got %d", len(result.Libraries))
	}
}

func TestParseGoModModules(t *testing.T) {
	content := `module myapp

go 1.21

require (
	github.com/go-chi/chi/v5 v5.2.0
	github.com/jmoiron/sqlx v1.3.5
)

// a comment
`
	modules := parseGoModModules(content)
	if modules["github.com/go-chi/chi/v5"] != "v5.2.0" {
		t.Errorf("missing chi module, got %v", modules)
	}
}

func TestCollectImports(t *testing.T) {
	result := collectImports(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestHasImportPrefix(t *testing.T) {
	imports := []string{"github.com/go-chi/chi/v5", "fmt"}
	if !hasImportPrefix(imports, "github.com/go-chi") {
		t.Error("expected true for chi prefix")
	}
	if hasImportPrefix(imports, "github.com/gin") {
		t.Error("expected false for gin prefix")
	}
	if hasImportPrefix(nil, "anything") {
		t.Error("expected false for nil imports")
	}
}

func TestModuleVersionByPrefix(t *testing.T) {
	modules := map[string]string{
		"github.com/go-chi/chi/v5": "v5.2.0",
	}
	v, ok := moduleVersionByPrefix(modules, "github.com/go-chi")
	if !ok || v != "v5.2.0" {
		t.Errorf("expected v5.2.0, got %q (ok=%v)", v, ok)
	}
	_, ok = moduleVersionByPrefix(modules, "github.com/gin")
	if ok {
		t.Error("expected not found for gin")
	}
}
