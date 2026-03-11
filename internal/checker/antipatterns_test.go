package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"golang.org/x/tools/go/packages"
)

func parseSource(t *testing.T, src string) (*ast.File, *token.FileSet) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return file, fset
}

func TestDetectGlobalMutableState(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		wantN int
	}{
		{
			"mutable map",
			`package foo; var cache = map[string]int{}`,
			1,
		},
		{
			"mutable slice",
			`package foo; var items = []string{}`,
			1,
		},
		{
			"pointer var",
			`package foo; var cfg *Config`,
			1,
		},
		{
			"make call",
			`package foo; var ch = make(chan int)`,
			1,
		},
		{
			"error sentinel ignored",
			`package foo; import "errors"; var ErrNotFound = errors.New("not found")`,
			0,
		},
		{
			"blank identifier ignored",
			`package foo; var _ = func(){}`,
			0,
		},
		{
			"const-like var ignored",
			`package foo; var maxRetries = 3`,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fset := parseSource(t, tt.src)
			results := detectGlobalMutableState(file, fset, "test.go")
			if len(results) != tt.wantN {
				t.Errorf("got %d violations, want %d", len(results), tt.wantN)
			}
		})
	}
}

func TestDetectInitAbuse(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		wantN int
	}{
		{
			"short init ok",
			`package foo; func init() { x := 1; _ = x }`,
			0,
		},
		{
			"long init flagged",
			`package foo
func init() {
	a := 1; b := 2; c := 3; d := 4; e := 5; f := 6
	_ = a; _ = b; _ = c; _ = d; _ = e; _ = f
}`,
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fset := parseSource(t, tt.src)
			results := detectInitAbuse(file, fset, "test.go")
			if len(results) != tt.wantN {
				t.Errorf("got %d violations, want %d", len(results), tt.wantN)
			}
		})
	}
}

func TestDetectNakedGoroutines(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		wantN int
	}{
		{
			"bare goroutine in regular func",
			`package foo
func doWork() {
	go func() { println("hello") }()
}`,
			1,
		},
		{
			"goroutine in Run method skipped",
			`package foo
import "context"
type Server struct{}
func (s *Server) Run(ctx context.Context) error {
	go func() { println("serving") }()
	return nil
}`,
			0,
		},
		{
			"goroutine in Start method skipped",
			`package foo
func Start() {
	go func() { println("starting") }()
}`,
			0,
		},
		{
			"goroutine in ListenAndServe skipped",
			`package foo
func ListenAndServe() {
	go func() { println("listening") }()
}`,
			0,
		},
		{
			"goroutine outside Run still flagged",
			`package foo
import "context"
type Server struct{}
func (s *Server) Run(ctx context.Context) error {
	go func() {}()
	return nil
}
func (s *Server) Handle() {
	go func() { println("bad") }()
}`,
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fset := parseSource(t, tt.src)
			results := detectNakedGoroutines(file, fset, "test.go")
			if len(results) != tt.wantN {
				t.Errorf("got %d violations, want %d", len(results), tt.wantN)
			}
		})
	}
}

func TestDetectSwallowedErrors(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		wantN int
	}{
		{
			"empty body",
			`package foo
func f() {
	var err error
	if err != nil {}
}`,
			1,
		},
		{
			"return nil",
			`package foo
func f() error {
	var err error
	if err != nil { return nil }
	return nil
}`,
			1,
		},
		{
			"proper handling",
			`package foo
func f() error {
	var err error
	if err != nil { return err }
	return nil
}`,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fset := parseSource(t, tt.src)
			results := detectSwallowedErrors(file, fset, "test.go")
			if len(results) != tt.wantN {
				t.Errorf("got %d violations, want %d", len(results), tt.wantN)
			}
		})
	}
}

func TestDetectSQLConcatenation(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		wantN int
	}{
		{
			"concat with SELECT",
			`package foo
func f(id string) string {
	return "SELECT * FROM users WHERE id=" + id
}`,
			1,
		},
		{
			"safe string concat",
			`package foo
func f() string {
	return "hello" + " world"
}`,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fset := parseSource(t, tt.src)
			results := detectSQLConcatenation(file, fset, "test.go")
			if len(results) != tt.wantN {
				t.Errorf("got %d violations, want %d", len(results), tt.wantN)
			}
		})
	}
}

func TestDetectUUIDv4AsKey(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		wantN int
	}{
		{
			"uuid.New flagged",
			`package foo
import "github.com/google/uuid"
func f() { id := uuid.New(); _ = id }`,
			1,
		},
		{
			"uuid.NewString flagged",
			`package foo
import "github.com/google/uuid"
func f() string { return uuid.NewString() }`,
			1,
		},
		{
			"no uuid usage",
			`package foo
func f() string { return "hello" }`,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fset := parseSource(t, tt.src)
			results := detectUUIDv4AsKey(file, fset, "test.go")
			if len(results) != tt.wantN {
				t.Errorf("got %d violations, want %d", len(results), tt.wantN)
			}
		})
	}
}

func TestDetectUUIDv4AsKey_SkipsRequestIDFiles(t *testing.T) {
	src := `package foo
import "github.com/google/uuid"
func f() string { return uuid.NewString() }`

	tests := []struct {
		name     string
		filePath string
		wantN    int
	}{
		{"middleware_requestid.go skipped", "middleware_requestid.go", 0},
		{"request_id.go skipped", "request_id.go", 0},
		{"handler.go flagged", "handler.go", 1},
		{"repo.go flagged", "repo.go", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fset := parseSource(t, src)
			results := detectUUIDv4AsKey(file, fset, tt.filePath)
			if len(results) != tt.wantN {
				t.Errorf("got %d violations, want %d", len(results), tt.wantN)
			}
		})
	}
}

func TestDetectContextBackground(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		pkgPath string
		wantN   int
	}{
		{
			"context.Background in handler flagged",
			`package httphandler
import "context"
func Handle() {
	ctx := context.Background()
	_ = ctx
}`,
			"github.com/acme/orders/adapter/httphandler",
			1,
		},
		{
			"context.Background in non-handler skipped",
			`package service
import "context"
func Do() {
	ctx := context.Background()
	_ = ctx
}`,
			"github.com/acme/orders/service",
			0,
		},
		{
			"shutdown context.WithTimeout(context.Background()) skipped",
			`package httphandler
import "context"
import "time"
func Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = ctx
}`,
			"github.com/acme/orders/adapter/httphandler",
			0,
		},
		{
			"context.WithDeadline(context.Background()) skipped",
			`package httphandler
import "context"
import "time"
func Shutdown() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()
	_ = ctx
}`,
			"github.com/acme/orders/adapter/httphandler",
			0,
		},
		{
			"init call like jwk.Fetch(context.Background()) skipped",
			`package httphandler
import "context"
func Setup() {
	keys := jwk.Fetch(context.Background(), "https://example.com/.well-known/jwks.json")
	_ = keys
}`,
			"github.com/acme/orders/adapter/httphandler",
			0,
		},
		{
			"bare context.Background alongside shutdown still flagged",
			`package httphandler
import "context"
import "time"
func Handle() {
	bad := context.Background()
	_ = bad
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = ctx
}`,
			"github.com/acme/orders/adapter/httphandler",
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fset := parseSource(t, tt.src)
			results := detectContextBackground(file, fset, "test.go", tt.pkgPath)
			if len(results) != tt.wantN {
				t.Errorf("got %d violations, want %d", len(results), tt.wantN)
			}
		})
	}
}

func TestIsDomainPackage(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"github.com/acme/orders/domain", true},
		{"github.com/acme/orders/core", true},
		{"github.com/acme/orders/port", true},
		{"github.com/acme/orders/adapter/httphandler", false},
		{"github.com/acme/orders/service", false},
	}
	for _, tt := range tests {
		if got := isDomainPackage(tt.path); got != tt.want {
			t.Errorf("isDomainPackage(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestIsAdapterPackage(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"github.com/acme/orders/adapter/httphandler", true},
		{"github.com/acme/orders/infrastructure/postgres", true},
		{"github.com/acme/orders/handler", true},
		{"github.com/acme/orders/controller", true},
		{"github.com/acme/orders/domain", false},
		{"github.com/acme/orders/service", false},
	}
	for _, tt := range tests {
		if got := isAdapterPackage(tt.path); got != tt.want {
			t.Errorf("isAdapterPackage(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestIsErrNilCheck(t *testing.T) {
	src := `package foo
func f() {
	var err error
	if err != nil {}
	if nil != err {}
}`
	file, fset := parseSource(t, src)
	_ = fset
	count := 0
	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if isErrNilCheck(ifStmt.Cond) {
			count++
		}
		return true
	})
	if count != 2 {
		t.Errorf("isErrNilCheck matched %d, want 2", count)
	}
}

func TestHasHeavySideEffects(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			"http.Get is heavy",
			`package foo
import "net/http"
func f() { http.Get("http://example.com") }`,
			true,
		},
		{
			"sql.Open is heavy",
			`package foo
import "database/sql"
func f() { sql.Open("postgres", "dsn") }`,
			true,
		},
		{
			"os.ReadFile is heavy",
			`package foo
import "os"
func f() { os.ReadFile("file.txt") }`,
			true,
		},
		{
			"net.Dial is heavy",
			`package foo
import "net"
func f() { net.Dial("tcp", "localhost:80") }`,
			true,
		},
		{
			"no heavy calls",
			`package foo
func f() { x := 1; _ = x }`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := parseSource(t, tt.src)
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				got := hasHeavySideEffects(fn.Body)
				if got != tt.want {
					t.Errorf("hasHeavySideEffects() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestIsHTTPHandlerFunc(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			"standard handler",
			`package foo
import "net/http"
func Handle(w http.ResponseWriter, r *http.Request) {}`,
			true,
		},
		{
			"no params",
			`package foo
func Handle() {}`,
			false,
		},
		{
			"one param",
			`package foo
func Handle(x int) {}`,
			false,
		},
		{
			"non-handler two params",
			`package foo
func Handle(a int, b string) {}`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := parseSource(t, tt.src)
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				got := isHTTPHandlerFunc(fn)
				if got != tt.want {
					t.Errorf("isHTTPHandlerFunc() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestTypeString(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"ident",
			`package foo; func f(x int) {}`,
			"int",
		},
		{
			"selector",
			`package foo; import "net/http"; func f(w http.ResponseWriter) {}`,
			"http.ResponseWriter",
		},
		{
			"star expr",
			`package foo; import "net/http"; func f(r *http.Request) {}`,
			"*http.Request",
		},
		{
			"unknown expr returns empty",
			`package foo; func f(x []int) {}`,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := parseSource(t, tt.src)
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Type.Params == nil {
					continue
				}
				for _, param := range fn.Type.Params.List {
					got := typeString(param.Type)
					if got != tt.want {
						t.Errorf("typeString() = %q, want %q", got, tt.want)
					}
				}
			}
		})
	}
}

func TestDetectFatHandlers(t *testing.T) {
	// Build a handler with > 40 statements.
	var stmts string
	for i := range 42 {
		stmts += fmt.Sprintf("\tx%d := %d; _ = x%d\n", i, i, i)
	}
	src := fmt.Sprintf(`package foo
import "net/http"
func Handle(w http.ResponseWriter, r *http.Request) {
%s}`, stmts)

	file, fset := parseSource(t, src)
	results := detectFatHandlers(file, fset, "handler.go", "github.com/acme/orders/adapter/httphandler")
	if len(results) != 1 {
		t.Errorf("expected 1 fat_handler violation, got %d", len(results))
	}

	// Non-handler package should not flag.
	results = detectFatHandlers(file, fset, "handler.go", "github.com/acme/orders/service")
	if len(results) != 0 {
		t.Errorf("expected 0 violations for non-handler package, got %d", len(results))
	}
}

func TestIsHandlerPackage(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"github.com/acme/orders/adapter/httphandler", true},
		{"github.com/acme/orders/controller", true},
		{"github.com/acme/orders/transport", true},
		{"github.com/acme/orders/api", true},
		{"github.com/acme/orders/domain", false},
	}
	for _, tt := range tests {
		if got := isHandlerPackage(tt.path); got != tt.want {
			t.Errorf("isHandlerPackage(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestCallName(t *testing.T) {
	src := `package foo
func f() {
	println("hello")
	http.Get("url")
}
`
	file, _ := parseSource(t, src)
	names := []string{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if ok {
			names = append(names, callName(call))
		}
		return true
	})
	if len(names) < 2 {
		t.Fatalf("expected at least 2 calls, got %d", len(names))
	}
}

func TestDetectGodPackages(t *testing.T) {
	// Build a package with > 40 exported symbols.
	var decls string
	for i := range 42 {
		decls += fmt.Sprintf("func Exported%d() {}\n", i)
	}
	src := "package foo\n" + decls
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "big.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	pkg := &packages.Package{
		PkgPath: "example.com/app/bigpkg",
		Name:    "foo",
		Fset:    fset,
		Syntax:  []*ast.File{file},
	}
	results := detectGodPackages([]*packages.Package{pkg}, "example.com/app")
	if len(results) != 1 {
		t.Errorf("expected 1 god_package violation, got %d", len(results))
	}
}

func TestDetectGodPackages_Small(t *testing.T) {
	src := `package foo
func Exported1() {}
func Exported2() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "small.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	pkg := &packages.Package{
		PkgPath: "example.com/app/smallpkg",
		Name:    "foo",
		Fset:    fset,
		Syntax:  []*ast.File{file},
	}
	results := detectGodPackages([]*packages.Package{pkg}, "example.com/app")
	if len(results) != 0 {
		t.Errorf("expected 0 violations, got %d", len(results))
	}
}

func TestDetectGodPackages_VendorSkipped(t *testing.T) {
	src := `package foo
func Exported1() {}
`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "vendor.go", src, 0)
	pkg := &packages.Package{
		PkgPath: "example.com/app/vendor/lib",
		Name:    "foo",
		Fset:    fset,
		Syntax:  []*ast.File{file},
	}
	results := detectGodPackages([]*packages.Package{pkg}, "example.com/app")
	if len(results) != 0 {
		t.Errorf("expected vendor to be skipped, got %d", len(results))
	}
}

func TestDetectDomainImportingAdapters(t *testing.T) {
	domainPkg := &packages.Package{
		PkgPath: "example.com/app/domain/order",
		Name:    "order",
		Imports: map[string]*packages.Package{
			"example.com/app/adapter/http": {PkgPath: "example.com/app/adapter/http"},
		},
	}
	results := detectDomainImportingAdapters([]*packages.Package{domainPkg}, "example.com/app")
	if len(results) != 1 {
		t.Errorf("expected 1 domain_imports_adapter violation, got %d", len(results))
	}
}

func TestDetectDomainImportingAdapters_Clean(t *testing.T) {
	domainPkg := &packages.Package{
		PkgPath: "example.com/app/domain/order",
		Name:    "order",
		Imports: map[string]*packages.Package{
			"fmt": {PkgPath: "fmt"},
		},
	}
	results := detectDomainImportingAdapters([]*packages.Package{domainPkg}, "example.com/app")
	if len(results) != 0 {
		t.Errorf("expected 0 violations, got %d", len(results))
	}
}

func TestDetectDomainImportingAdapters_NonDomainSkipped(t *testing.T) {
	servicePkg := &packages.Package{
		PkgPath: "example.com/app/service",
		Name:    "service",
		Imports: map[string]*packages.Package{
			"example.com/app/adapter/http": {PkgPath: "example.com/app/adapter/http"},
		},
	}
	results := detectDomainImportingAdapters([]*packages.Package{servicePkg}, "example.com/app")
	if len(results) != 0 {
		t.Errorf("expected non-domain pkg to be skipped, got %d", len(results))
	}
}

func TestDetectMVCInHexagonal(t *testing.T) {
	pkgs := []*packages.Package{
		{PkgPath: "example.com/app/domain/order"},
		{PkgPath: "example.com/app/port/inbound"},
		{PkgPath: "example.com/app/models"},
		{PkgPath: "example.com/app/controllers"},
	}
	results := detectMVCInHexagonal(pkgs, "example.com/app")
	if len(results) != 2 {
		t.Errorf("expected 2 mvc_in_hexagonal violations, got %d", len(results))
	}
}

func TestDetectMVCInHexagonal_NoHexagonal(t *testing.T) {
	pkgs := []*packages.Package{
		{PkgPath: "example.com/app/models"},
		{PkgPath: "example.com/app/controllers"},
	}
	results := detectMVCInHexagonal(pkgs, "example.com/app")
	if len(results) != 0 {
		t.Errorf("expected 0 violations without hexagonal markers, got %d", len(results))
	}
}

func TestDetectMVCInHexagonal_NoMVC(t *testing.T) {
	pkgs := []*packages.Package{
		{PkgPath: "example.com/app/domain/order"},
		{PkgPath: "example.com/app/port/inbound"},
		{PkgPath: "example.com/app/adapter/http"},
	}
	results := detectMVCInHexagonal(pkgs, "example.com/app")
	if len(results) != 0 {
		t.Errorf("expected 0 violations without MVC packages, got %d", len(results))
	}
}

func TestIsContextBackgroundCall(t *testing.T) {
	src := `package foo
import "context"
func f() {
	context.Background()
	context.TODO()
}
`
	file, _ := parseSource(t, src)
	bgCount := 0
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if ok && isContextBackgroundCall(call) {
			bgCount++
		}
		return true
	})
	if bgCount != 1 {
		t.Errorf("expected 1 context.Background call, got %d", bgCount)
	}
}

func TestIsContextBackgroundCall_NotSelector(t *testing.T) {
	src := `package foo
func f() {
	println("hello")
}
`
	file, _ := parseSource(t, src)
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if ok && isContextBackgroundCall(call) {
			t.Error("println should not be detected as context.Background")
		}
		return true
	})
}

func TestIsMutableType(t *testing.T) {
	tests := []struct {
		src  string
		want bool
	}{
		{`package foo; var x map[string]int`, true},
		{`package foo; var x []int`, true},
		{`package foo; var x chan int`, true},
		{`package foo; var x *int`, true},
		{`package foo; var x int`, false},
	}
	for _, tt := range tests {
		file, _ := parseSource(t, tt.src)
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				got := isMutableType(vs.Type)
				if got != tt.want {
					t.Errorf("isMutableType(%q) = %v, want %v", tt.src, got, tt.want)
				}
			}
		}
	}
}
