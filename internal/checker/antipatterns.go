package checker

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// AntiPattern defines a detected anti-pattern with context.
type AntiPattern struct {
	Name     string // Short identifier: "global_state", "init_abuse", etc.
	Category string // "code", "architecture", "security"
	Severity string // "error", "warning", "info"
	File     string
	Line     int
	Message  string
}

// checkAntiPatterns runs all AST-based anti-pattern detectors.
func checkAntiPatterns(pkgs []*packages.Package, projectPath string) []AntiPattern {
	var results []AntiPattern

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			fset := pkg.Fset
			filePath := fset.Position(file.Pos()).Filename

			// Skip test files — anti-patterns in tests are acceptable.
			if strings.HasSuffix(filePath, "_test.go") {
				continue
			}

			// Make path relative for cleaner output.
			relPath := filePath
			if rel, err := filepath.Rel(projectPath, filePath); err == nil {
				relPath = rel
			}

			results = append(results, detectGlobalMutableState(file, fset, relPath)...)
			results = append(results, detectInitAbuse(file, fset, relPath)...)
			results = append(results, detectNakedGoroutines(file, fset, relPath)...)
			results = append(results, detectSwallowedErrors(file, fset, relPath)...)
			results = append(results, detectContextBackground(file, fset, relPath, pkg.PkgPath)...)
			results = append(results, detectSQLConcatenation(file, fset, relPath)...)
			results = append(results, detectUUIDv4AsKey(file, fset, relPath)...)
			results = append(results, detectFatHandlers(file, fset, relPath, pkg.PkgPath)...)
		}
	}

	// Architectural anti-patterns (cross-package).
	results = append(results, detectGodPackages(pkgs, projectPath)...)
	results = append(results, detectDomainImportingAdapters(pkgs, projectPath)...)
	results = append(results, detectMVCInHexagonal(pkgs, projectPath)...)

	return results
}

// detectGlobalMutableState finds package-level vars with mutable types.
func detectGlobalMutableState(file *ast.File, fset *token.FileSet, filePath string) []AntiPattern {
	var results []AntiPattern
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
				// Skip unexported error sentinels (var ErrFoo = errors.New(...)).
				if strings.HasPrefix(name.Name, "Err") || strings.HasPrefix(name.Name, "err") {
					continue
				}
				// Skip blank identifier.
				if name.Name == "_" {
					continue
				}

				if isMutableType(vs.Type) || hasMutableValue(vs.Values) {
					results = append(results, AntiPattern{
						Name:     "global_mutable_state",
						Category: "code",
						Severity: "warning",
						File:     filePath,
						Line:     fset.Position(name.Pos()).Line,
						Message:  fmt.Sprintf("global mutable variable %q — use dependency injection instead", name.Name),
					})
				}
			}
		}
	}
	return results
}

// detectInitAbuse finds init() functions doing real work (not just registration).
func detectInitAbuse(file *ast.File, fset *token.FileSet, filePath string) []AntiPattern {
	var results []AntiPattern
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "init" || fn.Body == nil {
			continue
		}

		// Count statements — simple registrations are 1-3 lines.
		stmts := countStatements(fn.Body)
		if stmts > 5 {
			results = append(results, AntiPattern{
				Name:     "init_abuse",
				Category: "code",
				Severity: "warning",
				File:     filePath,
				Line:     fset.Position(fn.Pos()).Line,
				Message:  fmt.Sprintf("init() has %d statements — move complex logic to explicit setup functions", stmts),
			})
			continue
		}

		// Check for side effects: HTTP calls, DB connections, file I/O.
		if hasHeavySideEffects(fn.Body) {
			results = append(results, AntiPattern{
				Name:     "init_side_effects",
				Category: "code",
				Severity: "warning",
				File:     filePath,
				Line:     fset.Position(fn.Pos()).Line,
				Message:  "init() performs I/O or network calls — use explicit initialization for testability",
			})
		}
	}
	return results
}

// detectNakedGoroutines finds `go func()` without errgroup/waitgroup/context.
// Skips goroutines inside Run/Start/ListenAndServe methods (server lifecycle pattern).
func detectNakedGoroutines(file *ast.File, fset *token.FileSet, filePath string) []AntiPattern {
	// Collect line ranges for Run/Start/ListenAndServe methods — goroutines
	// inside these are expected (e.g. HTTP server startup).
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

	var results []AntiPattern
	ast.Inspect(file, func(n ast.Node) bool {
		gs, ok := n.(*ast.GoStmt)
		if !ok {
			return true
		}

		line := fset.Position(gs.Pos()).Line

		// Skip goroutines inside server lifecycle methods.
		for _, r := range excluded {
			if line >= r.start && line <= r.end {
				return true
			}
		}

		results = append(results, AntiPattern{
			Name:     "naked_goroutine",
			Category: "code",
			Severity: "warning",
			File:     filePath,
			Line:     line,
			Message:  "bare 'go' statement — use errgroup.Go() or structured concurrency for error propagation and lifecycle",
		})
		return true
	})
	return results
}

// detectSwallowedErrors finds if err != nil blocks that return nil or do nothing.
func detectSwallowedErrors(file *ast.File, fset *token.FileSet, filePath string) []AntiPattern {
	var results []AntiPattern
	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		// Match: if err != nil { return nil } or if err != nil { }
		if !isErrNilCheck(ifStmt.Cond) {
			return true
		}

		body := ifStmt.Body
		if body == nil {
			return true
		}

		// Empty body: if err != nil { }
		if len(body.List) == 0 {
			results = append(results, AntiPattern{
				Name:     "swallowed_error",
				Category: "code",
				Severity: "error",
				File:     filePath,
				Line:     fset.Position(ifStmt.Pos()).Line,
				Message:  "error checked but silently discarded — handle, wrap, or log it",
			})
			return true
		}

		// return nil (swallowing): if err != nil { return nil }
		if len(body.List) == 1 {
			ret, ok := body.List[0].(*ast.ReturnStmt)
			if ok && len(ret.Results) == 1 {
				if ident, ok := ret.Results[0].(*ast.Ident); ok && ident.Name == "nil" {
					results = append(results, AntiPattern{
						Name:     "swallowed_error",
						Category: "code",
						Severity: "error",
						File:     filePath,
						Line:     fset.Position(ifStmt.Pos()).Line,
						Message:  "error checked but return nil discards it — propagate or wrap the error",
					})
				}
			}
		}

		return true
	})
	return results
}

// detectContextBackground finds context.Background() usage in handler/adapter packages.
// Skips shutdown contexts (context.WithTimeout(context.Background(), ...)) which are legitimate.
func detectContextBackground(file *ast.File, fset *token.FileSet, filePath string, pkgPath string) []AntiPattern {
	// Only flag in handler/adapter code where request context should be used.
	if !isHandlerPackage(pkgPath) {
		return nil
	}

	// Collect lines where context.Background() is used legitimately:
	// - Inside context.WithTimeout/WithDeadline (shutdown pattern)
	// - As argument to initialization calls (Fetch, Connect, Open, Dial, Init, New)
	skipLines := map[int]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fnName := callName(call)

		// Shutdown pattern: context.WithTimeout(context.Background(), ...)
		if fnName == "context.WithTimeout" || fnName == "context.WithDeadline" {
			for _, arg := range call.Args {
				if innerCall, ok := arg.(*ast.CallExpr); ok && isContextBackgroundCall(innerCall) {
					skipLines[fset.Position(innerCall.Pos()).Line] = true
				}
			}
		}

		// Initialization calls: jwk.Fetch(context.Background(), ...), sql.Open(...), etc.
		if isInitCall(fnName) {
			for _, arg := range call.Args {
				if innerCall, ok := arg.(*ast.CallExpr); ok && isContextBackgroundCall(innerCall) {
					skipLines[fset.Position(innerCall.Pos()).Line] = true
				}
			}
		}

		return true
	})

	var results []AntiPattern
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isContextBackgroundCall(call) {
			line := fset.Position(call.Pos()).Line
			if skipLines[line] {
				return true // skip shutdown contexts
			}
			results = append(results, AntiPattern{
				Name:     "context_background_in_handler",
				Category: "code",
				Severity: "warning",
				File:     filePath,
				Line:     line,
				Message:  "context.Background() in handler — use request context (r.Context()) for proper cancellation",
			})
		}
		return true
	})
	return results
}

// detectSQLConcatenation finds string concatenation in SQL-like contexts.
func detectSQLConcatenation(file *ast.File, fset *token.FileSet, filePath string) []AntiPattern {
	var results []AntiPattern
	ast.Inspect(file, func(n ast.Node) bool {
		// Look for string concat with SQL keywords.
		binExpr, ok := n.(*ast.BinaryExpr)
		if !ok || binExpr.Op != token.ADD {
			return true
		}

		if containsSQLKeyword(binExpr) {
			results = append(results, AntiPattern{
				Name:     "sql_concatenation",
				Category: "security",
				Severity: "error",
				File:     filePath,
				Line:     fset.Position(binExpr.Pos()).Line,
				Message:  "SQL string concatenation detected — use parameterized queries to prevent injection",
			})
		}
		return true
	})
	return results
}

// detectFatHandlers finds HTTP handlers with too much logic (> 40 statements).
func detectFatHandlers(file *ast.File, fset *token.FileSet, filePath string, pkgPath string) []AntiPattern {
	if !isHandlerPackage(pkgPath) {
		return nil
	}

	var results []AntiPattern
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		// Check if function signature matches handler pattern:
		// func(w http.ResponseWriter, r *http.Request)
		if !isHTTPHandlerFunc(fn) {
			continue
		}

		stmts := countStatements(fn.Body)
		if stmts > 40 {
			results = append(results, AntiPattern{
				Name:     "fat_handler",
				Category: "architecture",
				Severity: "warning",
				File:     filePath,
				Line:     fset.Position(fn.Pos()).Line,
				Message:  fmt.Sprintf("handler %s has %d statements — extract business logic to a service layer", fn.Name.Name, stmts),
			})
		}
	}
	return results
}

// detectGodPackages finds packages with too many exported symbols.
func detectGodPackages(pkgs []*packages.Package, projectPath string) []AntiPattern {
	var results []AntiPattern

	for _, pkg := range pkgs {
		if strings.Contains(pkg.PkgPath, "/vendor/") {
			continue
		}

		exported := 0
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Name.IsExported() {
						exported++
					}
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if s.Name.IsExported() {
								exported++
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if name.IsExported() {
									exported++
								}
							}
						}
					}
				}
			}
		}

		if exported > 40 {
			relPath := pkg.PkgPath
			if rel, err := filepath.Rel(projectPath, pkg.PkgPath); err == nil {
				relPath = rel
			}
			results = append(results, AntiPattern{
				Name:     "god_package",
				Category: "architecture",
				Severity: "warning",
				File:     relPath,
				Message:  fmt.Sprintf("package has %d exported symbols — consider splitting by responsibility", exported),
			})
		}
	}
	return results
}

// detectUUIDv4AsKey flags uuid.New() / uuid.NewString() usage and suggests UUIDv7 for DB keys.
// Skips files related to request ID generation where UUIDv4 is appropriate.
func detectUUIDv4AsKey(file *ast.File, fset *token.FileSet, filePath string) []AntiPattern {
	// UUIDv4 is appropriate for request IDs — skip those files.
	baseName := strings.ToLower(filepath.Base(filePath))
	if strings.Contains(baseName, "requestid") || strings.Contains(baseName, "request_id") {
		return nil
	}

	var results []AntiPattern
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fnName := callName(call)
		if fnName == "uuid.New" || fnName == "uuid.NewString" {
			results = append(results, AntiPattern{
				Name:     "uuid_v4_as_key",
				Category: "code",
				Severity: "info",
				File:     filePath,
				Line:     fset.Position(call.Pos()).Line,
				Message:  "uuid.New() generates UUIDv4 (random) — use UUIDv7 for database primary keys to avoid index fragmentation",
			})
		}
		return true
	})
	return results
}

// detectDomainImportingAdapters finds domain packages that import adapter/infrastructure packages.
func detectDomainImportingAdapters(pkgs []*packages.Package, projectPath string) []AntiPattern {
	var results []AntiPattern
	for _, pkg := range pkgs {
		if !isDomainPackage(pkg.PkgPath) {
			continue
		}
		for _, imp := range pkg.Imports {
			if isAdapterPackage(imp.PkgPath) {
				relPath := pkg.PkgPath
				if rel, err := filepath.Rel(projectPath, pkg.PkgPath); err == nil {
					relPath = rel
				}
				results = append(results, AntiPattern{
					Name:     "domain_imports_adapter",
					Category: "architecture",
					Severity: "error",
					File:     relPath,
					Message:  fmt.Sprintf("domain package imports adapter %q — dependencies must point inward", imp.PkgPath),
				})
			}
		}
	}
	return results
}

// detectMVCInHexagonal finds MVC-style packages (models/, controllers/, views/) in hexagonal projects.
func detectMVCInHexagonal(pkgs []*packages.Package, projectPath string) []AntiPattern {
	// Only flag if hexagonal markers exist (domain/, port/, adapter/).
	hasHexagonal := false
	mvcPackages := []string{}

	for _, pkg := range pkgs {
		path := pkg.PkgPath
		if strings.Contains(path, "/domain") || strings.Contains(path, "/port") {
			hasHexagonal = true
		}
		lastSegment := path[strings.LastIndex(path, "/")+1:]
		if lastSegment == "models" || lastSegment == "controllers" || lastSegment == "views" {
			mvcPackages = append(mvcPackages, path)
		}
	}

	if !hasHexagonal || len(mvcPackages) == 0 {
		return nil
	}

	results := make([]AntiPattern, 0, len(mvcPackages))
	for _, pkgPath := range mvcPackages {
		relPath := pkgPath
		if rel, err := filepath.Rel(projectPath, pkgPath); err == nil {
			relPath = rel
		}
		results = append(results, AntiPattern{
			Name:     "mvc_in_hexagonal",
			Category: "architecture",
			Severity: "warning",
			File:     relPath,
			Message:  fmt.Sprintf("MVC package %q in hexagonal project — use domain/port/adapter layers instead", relPath),
		})
	}
	return results
}

func isDomainPackage(pkgPath string) bool {
	return strings.Contains(pkgPath, "/domain") ||
		strings.Contains(pkgPath, "/core") ||
		strings.Contains(pkgPath, "/port")
}

func isAdapterPackage(pkgPath string) bool {
	return strings.Contains(pkgPath, "/adapter") ||
		strings.Contains(pkgPath, "/infrastructure") ||
		strings.Contains(pkgPath, "/infra") ||
		strings.Contains(pkgPath, "/handler") ||
		strings.Contains(pkgPath, "/repository") ||
		strings.Contains(pkgPath, "/controller")
}

// --- Helper functions ---

func isMutableType(expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	switch expr.(type) {
	case *ast.MapType, *ast.ArrayType, *ast.ChanType:
		return true
	case *ast.StarExpr:
		return true
	}
	return false
}

func hasMutableValue(values []ast.Expr) bool {
	for _, v := range values {
		if _, ok := v.(*ast.CompositeLit); ok {
			return true
		}
		// Check for make() calls.
		if call, ok := v.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "make" {
				return true
			}
		}
	}
	return false
}

func countStatements(body *ast.BlockStmt) int {
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

func hasHeavySideEffects(body *ast.BlockStmt) bool {
	heavy := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		fnName := callName(call)
		heavyPatterns := []string{
			"http.Get", "http.Post", "http.Do",
			"sql.Open", "pgx.Connect", "mongo.Connect",
			"os.Open", "os.Create", "os.ReadFile",
			"net.Dial", "net.Listen",
		}
		for _, pattern := range heavyPatterns {
			if strings.Contains(fnName, pattern) {
				heavy = true
				return false
			}
		}
		return true
	})
	return heavy
}

func callName(call *ast.CallExpr) string {
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

func isErrNilCheck(cond ast.Expr) bool {
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

func isHandlerPackage(pkgPath string) bool {
	return strings.Contains(pkgPath, "handler") ||
		strings.Contains(pkgPath, "controller") ||
		strings.Contains(pkgPath, "adapter") ||
		strings.Contains(pkgPath, "transport") ||
		strings.Contains(pkgPath, "api")
}

func isInitCall(fnName string) bool {
	initSuffixes := []string{"Fetch", "Connect", "Open", "Dial", "Init", "Listen", "Setup", "Configure"}
	for _, suffix := range initSuffixes {
		if strings.HasSuffix(fnName, suffix) || strings.HasSuffix(fnName, "."+suffix) {
			return true
		}
	}
	return false
}

func isContextBackgroundCall(call *ast.CallExpr) bool {
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

func containsSQLKeyword(binExpr *ast.BinaryExpr) bool {
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

func isHTTPHandlerFunc(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil {
		return false
	}
	params := fn.Type.Params.List
	if len(params) < 2 {
		return false
	}

	// Check for http.ResponseWriter and *http.Request patterns.
	hasWriter := false
	hasRequest := false
	for _, param := range params {
		typeStr := typeString(param.Type)
		if strings.Contains(typeStr, "ResponseWriter") {
			hasWriter = true
		}
		if strings.Contains(typeStr, "Request") {
			hasRequest = true
		}
	}
	return hasWriter && hasRequest
}

func typeString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok {
			return x.Name + "." + e.Sel.Name
		}
		return e.Sel.Name
	case *ast.StarExpr:
		return "*" + typeString(e.X)
	}
	return ""
}
