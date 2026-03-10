package rules

import (
	"go/ast"
	"go/token"
	"strings"
)

func isMutableTypeAST(expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	switch expr.(type) {
	case *ast.MapType, *ast.ArrayType, *ast.ChanType, *ast.StarExpr:
		return true
	}
	return false
}

func hasMutableValueAST(values []ast.Expr) bool {
	for _, v := range values {
		if _, ok := v.(*ast.CompositeLit); ok {
			return true
		}
		if call, ok := v.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "make" {
				return true
			}
		}
	}
	return false
}

func countStatementsAST(body *ast.BlockStmt) int {
	count := 0
	ast.Inspect(body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.AssignStmt, *ast.ExprStmt, *ast.ReturnStmt,
			*ast.DeclStmt, *ast.SendStmt, *ast.IncDecStmt,
			*ast.GoStmt, *ast.DeferStmt, *ast.BranchStmt:
			count++
		}
		return true
	})
	return count
}

func hasHeavySideEffectsAST(body *ast.BlockStmt) bool {
	heavy := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fnName := callNameAST(call)
		patterns := []string{
			"http.Get", "http.Post", "http.Do",
			"sql.Open", "pgx.Connect", "mongo.Connect",
			"os.Open", "os.Create", "os.ReadFile",
			"net.Dial", "net.Listen",
		}
		for _, p := range patterns {
			if strings.Contains(fnName, p) {
				heavy = true
				return false
			}
		}
		return true
	})
	return heavy
}

func callNameAST(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		if x, ok := fn.X.(*ast.Ident); ok {
			return x.Name + "." + fn.Sel.Name
		}
		return fn.Sel.Name
	}
	return ""
}

func isErrNilCheckAST(cond ast.Expr) bool {
	bin, ok := cond.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}
	xIdent, xOk := bin.X.(*ast.Ident)
	yIdent, yOk := bin.Y.(*ast.Ident)
	if xOk && xIdent.Name == "err" && yOk && yIdent.Name == "nil" {
		return true
	}
	if yOk && yIdent.Name == "err" && xOk && xIdent.Name == "nil" {
		return true
	}
	return false
}

func isContextBackgroundCallAST(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return x.Name == "context" && sel.Sel.Name == "Background"
}

func containsSQLKeywordAST(binExpr *ast.BinaryExpr) bool {
	sqlKeywords := []string{"SELECT ", "INSERT ", "UPDATE ", "DELETE ", "FROM ", "WHERE ", "JOIN "}
	var check func(ast.Expr) bool
	check = func(expr ast.Expr) bool {
		switch e := expr.(type) {
		case *ast.BasicLit:
			if e.Kind == token.STRING {
				upper := strings.ToUpper(e.Value)
				for _, kw := range sqlKeywords {
					if strings.Contains(upper, kw) {
						return true
					}
				}
			}
		case *ast.BinaryExpr:
			return check(e.X) || check(e.Y)
		}
		return false
	}
	return check(binExpr)
}

func isHTTPHandlerFuncAST(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil {
		return false
	}
	params := fn.Type.Params.List
	if len(params) < 2 {
		return false
	}
	hasWriter := false
	hasRequest := false
	for _, param := range params {
		ts := typeStringAST(param.Type)
		if strings.Contains(ts, "ResponseWriter") {
			hasWriter = true
		}
		if strings.Contains(ts, "Request") {
			hasRequest = true
		}
	}
	return hasWriter && hasRequest
}

func typeStringAST(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok {
			return x.Name + "." + e.Sel.Name
		}
		return e.Sel.Name
	case *ast.StarExpr:
		return "*" + typeStringAST(e.X)
	}
	return ""
}
