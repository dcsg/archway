package detector

import (
	"go/ast"
	"go/parser"
	"go/token"
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

func makeSyntheticPkg(t *testing.T, src string, imports map[string]*packages.Package) *packages.Package {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	pkg := &packages.Package{
		PkgPath: "example.com/test",
		Name:    "test",
		Fset:    fset,
		Syntax:  []*ast.File{file},
		Imports: imports,
	}
	return pkg
}

func TestDetectErrorHandling_Typed(t *testing.T) {
	src := `package foo
type NotFoundError struct{}
`
	pkg := makeSyntheticPkg(t, src, nil)
	result := detectErrorHandling([]*packages.Package{pkg})
	if result.Pattern != "typed" {
		t.Errorf("Pattern = %q, want typed", result.Pattern)
	}
}

func TestDetectErrorHandling_Wrapped(t *testing.T) {
	src := `package foo
import "fmt"
func f() error {
	return fmt.Errorf("wrap: %w", nil)
}
`
	pkg := makeSyntheticPkg(t, src, nil)
	result := detectErrorHandling([]*packages.Package{pkg})
	if result.Pattern != "wrapped" {
		t.Errorf("Pattern = %q, want wrapped", result.Pattern)
	}
}

func TestDetectErrorHandling_Sentinel(t *testing.T) {
	src := `package foo
var ErrNotFound = "not found"
`
	pkg := makeSyntheticPkg(t, src, nil)
	result := detectErrorHandling([]*packages.Package{pkg})
	if result.Pattern != "sentinel" {
		t.Errorf("Pattern = %q, want sentinel", result.Pattern)
	}
}

func TestDetectErrorHandling_Minimal(t *testing.T) {
	src := `package foo
func f() {}
`
	pkg := makeSyntheticPkg(t, src, nil)
	result := detectErrorHandling([]*packages.Package{pkg})
	if result.Pattern != "minimal" {
		t.Errorf("Pattern = %q, want minimal", result.Pattern)
	}
}

func TestDetectLogging_Zap(t *testing.T) {
	src := `package foo`
	imports := map[string]*packages.Package{
		"go.uber.org/zap": {PkgPath: "go.uber.org/zap"},
	}
	pkg := makeSyntheticPkg(t, src, imports)
	result := detectLogging([]*packages.Package{pkg})
	if result.Pattern != "zap/structured" {
		t.Errorf("Pattern = %q, want zap/structured", result.Pattern)
	}
}

func TestDetectLogging_Zerolog(t *testing.T) {
	src := `package foo`
	imports := map[string]*packages.Package{
		"github.com/rs/zerolog": {PkgPath: "github.com/rs/zerolog"},
	}
	pkg := makeSyntheticPkg(t, src, imports)
	result := detectLogging([]*packages.Package{pkg})
	if result.Pattern != "zerolog/structured" {
		t.Errorf("Pattern = %q, want zerolog/structured", result.Pattern)
	}
}

func TestDetectLogging_Slog(t *testing.T) {
	src := `package foo`
	imports := map[string]*packages.Package{
		"log/slog": {PkgPath: "log/slog"},
	}
	pkg := makeSyntheticPkg(t, src, imports)
	result := detectLogging([]*packages.Package{pkg})
	if result.Pattern != "slog/structured" {
		t.Errorf("Pattern = %q, want slog/structured", result.Pattern)
	}
}

func TestDetectLogging_Unstructured(t *testing.T) {
	src := `package foo
import "log"
func f() { log.Printf("msg") }
`
	pkg := makeSyntheticPkg(t, src, nil)
	result := detectLogging([]*packages.Package{pkg})
	if result.Pattern != "stdlib/unstructured" {
		t.Errorf("Pattern = %q, want stdlib/unstructured", result.Pattern)
	}
}

func TestDetectConfig_Viper(t *testing.T) {
	src := `package foo`
	imports := map[string]*packages.Package{
		"github.com/spf13/viper": {PkgPath: "github.com/spf13/viper"},
	}
	pkg := makeSyntheticPkg(t, src, imports)
	result := detectConfig([]*packages.Package{pkg})
	if result.Pattern != "viper" {
		t.Errorf("Pattern = %q, want viper", result.Pattern)
	}
}

func TestDetectConfig_Koanf(t *testing.T) {
	src := `package foo`
	imports := map[string]*packages.Package{
		"github.com/knadh/koanf": {PkgPath: "github.com/knadh/koanf"},
	}
	pkg := makeSyntheticPkg(t, src, imports)
	result := detectConfig([]*packages.Package{pkg})
	if result.Pattern != "koanf" {
		t.Errorf("Pattern = %q, want koanf", result.Pattern)
	}
}

func TestDetectConfig_Envconfig(t *testing.T) {
	src := `package foo`
	imports := map[string]*packages.Package{
		"github.com/kelseyhightower/envconfig": {PkgPath: "github.com/kelseyhightower/envconfig"},
	}
	pkg := makeSyntheticPkg(t, src, imports)
	result := detectConfig([]*packages.Package{pkg})
	if result.Pattern != "envconfig" {
		t.Errorf("Pattern = %q, want envconfig", result.Pattern)
	}
}

func TestDetectConfig_StructTags(t *testing.T) {
	src := "package foo\ntype Config struct {\n\tHost string `yaml:\"host\"`\n}\n"
	pkg := makeSyntheticPkg(t, src, nil)
	result := detectConfig([]*packages.Package{pkg})
	if result.Pattern != "struct-tags" {
		t.Errorf("Pattern = %q, want struct-tags", result.Pattern)
	}
}

func TestDetectConfig_Env(t *testing.T) {
	src := `package foo
func f() {}
`
	pkg := makeSyntheticPkg(t, src, nil)
	result := detectConfig([]*packages.Package{pkg})
	if result.Pattern != "env" {
		t.Errorf("Pattern = %q, want env", result.Pattern)
	}
}

func TestDetectTesting_Minimal(t *testing.T) {
	src := `package foo
func f() {}
`
	pkg := makeSyntheticPkg(t, src, nil)
	result := detectTesting([]*packages.Package{pkg})
	if result.Pattern != "minimal" {
		t.Errorf("Pattern = %q, want minimal", result.Pattern)
	}
}

func TestDetectTesting_BDD(t *testing.T) {
	src := `package foo`
	imports := map[string]*packages.Package{
		"github.com/onsi/ginkgo": {PkgPath: "github.com/onsi/ginkgo"},
	}
	pkg := makeSyntheticPkg(t, src, imports)
	result := detectTesting([]*packages.Package{pkg})
	if result.Pattern != "bdd" {
		t.Errorf("Pattern = %q, want bdd", result.Pattern)
	}
}
