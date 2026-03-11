package rules

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMutableTypeAST(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want bool
	}{
		{"nil", nil, false},
		{"MapType", &ast.MapType{}, true},
		{"ArrayType", &ast.ArrayType{Elt: &ast.Ident{Name: "int"}}, true},
		{"ChanType", &ast.ChanType{Value: &ast.Ident{Name: "int"}}, true},
		{"StarExpr", &ast.StarExpr{X: &ast.Ident{Name: "int"}}, true},
		{"Ident", &ast.Ident{Name: "int"}, false},
		{"SelectorExpr", &ast.SelectorExpr{X: &ast.Ident{Name: "pkg"}, Sel: &ast.Ident{Name: "Type"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isMutableTypeAST(tt.expr))
		})
	}
}

func TestHasMutableValueAST(t *testing.T) {
	tests := []struct {
		name   string
		values []ast.Expr
		want   bool
	}{
		{"nil slice", nil, false},
		{"empty slice", []ast.Expr{}, false},
		{"plain ident", []ast.Expr{&ast.Ident{Name: "x"}}, false},
		{"CompositeLit", []ast.Expr{&ast.CompositeLit{}}, true},
		{"make call", []ast.Expr{
			&ast.CallExpr{Fun: &ast.Ident{Name: "make"}, Args: []ast.Expr{&ast.Ident{Name: "map"}}},
		}, true},
		{"non-make call", []ast.Expr{
			&ast.CallExpr{Fun: &ast.Ident{Name: "len"}},
		}, false},
		{"selector call", []ast.Expr{
			&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "pkg"}, Sel: &ast.Ident{Name: "New"}}},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasMutableValueAST(tt.values))
		})
	}
}

func TestCountStatementsAST(t *testing.T) {
	tests := []struct {
		name string
		body *ast.BlockStmt
		want int
	}{
		{"nil block", nil, 0},
		{"empty block", &ast.BlockStmt{}, 0},
		{"three assigns", &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{Tok: token.DEFINE},
				&ast.AssignStmt{Tok: token.DEFINE},
				&ast.ReturnStmt{},
			},
		}, 3},
		{"expr and defer", &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{X: &ast.Ident{Name: "x"}},
				&ast.DeferStmt{Call: &ast.CallExpr{Fun: &ast.Ident{Name: "f"}}},
			},
		}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, countStatementsAST(tt.body))
		})
	}
}

func TestHasHeavySideEffectsAST(t *testing.T) {
	httpGetCall := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Get"}},
			}},
		},
	}
	osOpenCall := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "os"}, Sel: &ast.Ident{Name: "Open"}},
			}},
		},
	}
	cleanBlock := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{X: &ast.CallExpr{
				Fun: &ast.Ident{Name: "println"},
			}},
		},
	}

	tests := []struct {
		name string
		body *ast.BlockStmt
		want bool
	}{
		{"http.Get", httpGetCall, true},
		{"os.Open", osOpenCall, true},
		{"clean block", cleanBlock, false},
		{"empty block", &ast.BlockStmt{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasHeavySideEffectsAST(tt.body))
		})
	}
}

func TestCallNameAST(t *testing.T) {
	tests := []struct {
		name string
		call *ast.CallExpr
		want string
	}{
		{"Ident", &ast.CallExpr{Fun: &ast.Ident{Name: "println"}}, "println"},
		{"SelectorExpr", &ast.CallExpr{
			Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Println"}},
		}, "fmt.Println"},
		{"nested selector", &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.SelectorExpr{X: &ast.Ident{Name: "a"}, Sel: &ast.Ident{Name: "b"}},
				Sel: &ast.Ident{Name: "c"},
			},
		}, "c"},
		{"other expr", &ast.CallExpr{Fun: &ast.IndexExpr{}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, callNameAST(tt.call))
		})
	}
}

func TestIsErrNilCheckAST(t *testing.T) {
	errIdent := &ast.Ident{Name: "err"}
	nilIdent := &ast.Ident{Name: "nil"}
	otherIdent := &ast.Ident{Name: "x"}

	tests := []struct {
		name string
		cond ast.Expr
		want bool
	}{
		{"err != nil", &ast.BinaryExpr{X: errIdent, Op: token.NEQ, Y: nilIdent}, true},
		{"nil != err", &ast.BinaryExpr{X: nilIdent, Op: token.NEQ, Y: errIdent}, true},
		{"x != nil", &ast.BinaryExpr{X: otherIdent, Op: token.NEQ, Y: nilIdent}, false},
		{"err == nil", &ast.BinaryExpr{X: errIdent, Op: token.EQL, Y: nilIdent}, false},
		{"not binary", errIdent, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isErrNilCheckAST(tt.cond))
		})
	}
}

func TestIsContextBackgroundCallAST(t *testing.T) {
	tests := []struct {
		name string
		call *ast.CallExpr
		want bool
	}{
		{"context.Background", &ast.CallExpr{
			Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "context"}, Sel: &ast.Ident{Name: "Background"}},
		}, true},
		{"context.TODO", &ast.CallExpr{
			Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "context"}, Sel: &ast.Ident{Name: "TODO"}},
		}, false},
		{"plain ident", &ast.CallExpr{Fun: &ast.Ident{Name: "Background"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isContextBackgroundCallAST(tt.call))
		})
	}
}

func TestContainsSQLKeywordAST(t *testing.T) {
	tests := []struct {
		name string
		expr *ast.BinaryExpr
		want bool
	}{
		{"SELECT keyword", &ast.BinaryExpr{
			Op: token.ADD,
			X:  &ast.BasicLit{Kind: token.STRING, Value: `"SELECT * FROM users WHERE id = "`},
			Y:  &ast.Ident{Name: "id"},
		}, true},
		{"no SQL", &ast.BinaryExpr{
			Op: token.ADD,
			X:  &ast.BasicLit{Kind: token.STRING, Value: `"hello "`},
			Y:  &ast.BasicLit{Kind: token.STRING, Value: `"world"`},
		}, false},
		{"nested binary with INSERT", &ast.BinaryExpr{
			Op: token.ADD,
			X: &ast.BinaryExpr{
				Op: token.ADD,
				X:  &ast.BasicLit{Kind: token.STRING, Value: `"INSERT INTO users "`},
				Y:  &ast.Ident{Name: "cols"},
			},
			Y: &ast.Ident{Name: "vals"},
		}, true},
		{"lowercase select", &ast.BinaryExpr{
			Op: token.ADD,
			X:  &ast.BasicLit{Kind: token.STRING, Value: `"select * from users where "`},
			Y:  &ast.Ident{Name: "id"},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, containsSQLKeywordAST(tt.expr))
		})
	}
}

func TestIsHTTPHandlerFuncAST(t *testing.T) {
	handler := &ast.FuncDecl{
		Name: &ast.Ident{Name: "handleUsers"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Type: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "ResponseWriter"}}},
					{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Request"}}}},
				},
			},
		},
	}
	nonHandler := &ast.FuncDecl{
		Name: &ast.Ident{Name: "doStuff"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Type: &ast.Ident{Name: "string"}},
				},
			},
		},
	}
	noParams := &ast.FuncDecl{
		Name: &ast.Ident{Name: "noParams"},
		Type: &ast.FuncType{Params: nil},
	}

	tests := []struct {
		name string
		fn   *ast.FuncDecl
		want bool
	}{
		{"handler signature", handler, true},
		{"non-handler", nonHandler, false},
		{"no params", noParams, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isHTTPHandlerFuncAST(tt.fn))
		})
	}
}

func TestTypeStringAST(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{"nil", nil, ""},
		{"Ident", &ast.Ident{Name: "string"}, "string"},
		{"SelectorExpr", &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Request"}}, "http.Request"},
		{"StarExpr", &ast.StarExpr{X: &ast.Ident{Name: "User"}}, "*User"},
		{"StarExpr selector", &ast.StarExpr{
			X: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Request"}},
		}, "*http.Request"},
		{"nested selector (no X ident)", &ast.SelectorExpr{
			X:   &ast.SelectorExpr{X: &ast.Ident{Name: "a"}, Sel: &ast.Ident{Name: "b"}},
			Sel: &ast.Ident{Name: "c"},
		}, "c"},
		{"other type", &ast.MapType{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, typeStringAST(tt.expr))
		})
	}
}
