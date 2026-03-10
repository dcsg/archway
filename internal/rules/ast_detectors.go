package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func detectGlobalMutableStateRule(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation {
	var results []RuleViolation
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name == "_" || strings.HasPrefix(name.Name, "Err") || strings.HasPrefix(name.Name, "err") {
					continue
				}
				if isMutableTypeAST(vs.Type) || hasMutableValueAST(vs.Values) {
					results = append(results, RuleViolation{
						File:  relPath,
						Line:  fset.Position(name.Pos()).Line,
						Match: fmt.Sprintf("global mutable variable %q", name.Name),
					})
				}
			}
		}
	}
	return results
}

func detectInitAbuseRule(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation {
	var results []RuleViolation
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "init" || fn.Body == nil {
			continue
		}
		stmts := countStatementsAST(fn.Body)
		if stmts > 5 {
			results = append(results, RuleViolation{
				File:  relPath,
				Line:  fset.Position(fn.Pos()).Line,
				Match: fmt.Sprintf("init() has %d statements", stmts),
			})
		}
	}
	return results
}

func detectInitSideEffectsRule(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation {
	var results []RuleViolation
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "init" || fn.Body == nil {
			continue
		}
		if hasHeavySideEffectsAST(fn.Body) {
			results = append(results, RuleViolation{
				File:  relPath,
				Line:  fset.Position(fn.Pos()).Line,
				Match: "init() performs I/O or network calls",
			})
		}
	}
	return results
}

func detectNakedGoroutineRule(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation {
	serverMethods := map[string]bool{"Run": true, "Start": true, "ListenAndServe": true, "Serve": true}
	type lineRange struct{ start, end int }
	var excluded []lineRange

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if serverMethods[fn.Name.Name] {
			excluded = append(excluded, lineRange{
				start: fset.Position(fn.Body.Pos()).Line,
				end:   fset.Position(fn.Body.End()).Line,
			})
		}
	}

	var results []RuleViolation
	ast.Inspect(file, func(n ast.Node) bool {
		gs, ok := n.(*ast.GoStmt)
		if !ok {
			return true
		}
		line := fset.Position(gs.Pos()).Line
		for _, r := range excluded {
			if line >= r.start && line <= r.end {
				return true
			}
		}
		results = append(results, RuleViolation{
			File:  relPath,
			Line:  line,
			Match: "bare 'go' statement without structured concurrency",
		})
		return true
	})
	return results
}

func detectSwallowedErrorRule(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation {
	var results []RuleViolation
	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if !isErrNilCheckAST(ifStmt.Cond) || ifStmt.Body == nil {
			return true
		}
		if len(ifStmt.Body.List) == 0 {
			results = append(results, RuleViolation{
				File:  relPath,
				Line:  fset.Position(ifStmt.Pos()).Line,
				Match: "error checked but silently discarded",
			})
			return true
		}
		if len(ifStmt.Body.List) == 1 {
			ret, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
			if ok && len(ret.Results) == 1 {
				if ident, ok := ret.Results[0].(*ast.Ident); ok && ident.Name == "nil" {
					results = append(results, RuleViolation{
						File:  relPath,
						Line:  fset.Position(ifStmt.Pos()).Line,
						Match: "error checked but return nil discards it",
					})
				}
			}
		}
		return true
	})
	return results
}

func detectSQLConcatenationRule(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation {
	var results []RuleViolation
	ast.Inspect(file, func(n ast.Node) bool {
		binExpr, ok := n.(*ast.BinaryExpr)
		if !ok || binExpr.Op != token.ADD {
			return true
		}
		if containsSQLKeywordAST(binExpr) {
			results = append(results, RuleViolation{
				File:  relPath,
				Line:  fset.Position(binExpr.Pos()).Line,
				Match: "SQL string concatenation detected",
			})
		}
		return true
	})
	return results
}

func detectContextBackgroundRule(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation {
	// Without package path info, detect context.Background() in all files.
	var results []RuleViolation
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isContextBackgroundCallAST(call) {
			results = append(results, RuleViolation{
				File:  relPath,
				Line:  fset.Position(call.Pos()).Line,
				Match: "context.Background() usage",
			})
		}
		return true
	})
	return results
}

func detectUUIDv4Rule(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation {
	base := strings.ToLower(relPath)
	if strings.Contains(base, "requestid") || strings.Contains(base, "request_id") {
		return nil
	}
	var results []RuleViolation
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fnName := callNameAST(call)
		if fnName == "uuid.New" || fnName == "uuid.NewString" {
			results = append(results, RuleViolation{
				File:  relPath,
				Line:  fset.Position(call.Pos()).Line,
				Match: "uuid.New() generates UUIDv4 — consider UUIDv7 for DB keys",
			})
		}
		return true
	})
	return results
}

func detectFatHandlerRule(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation {
	var results []RuleViolation
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if !isHTTPHandlerFuncAST(fn) {
			continue
		}
		stmts := countStatementsAST(fn.Body)
		if stmts > 40 {
			results = append(results, RuleViolation{
				File:  relPath,
				Line:  fset.Position(fn.Pos()).Line,
				Match: fmt.Sprintf("handler %s has %d statements", fn.Name.Name, stmts),
			})
		}
	}
	return results
}
