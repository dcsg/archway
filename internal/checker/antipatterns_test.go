package checker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
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

func TestDetectGodPackages(t *testing.T) {
	// Requires *packages.Package — covered via integration tests.
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
