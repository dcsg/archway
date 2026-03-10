package rules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// KnownDetectors lists all detector names supported by the AST engine.
var KnownDetectors = []string{
	"global-mutable-state",
	"init-abuse",
	"init-side-effects",
	"naked-goroutine",
	"swallowed-error",
	"context-background-in-handler",
	"sql-concatenation",
	"uuid-v4-as-key",
	"fat-handler",
	"god-package",
	"domain-imports-adapter",
	"mvc-in-hexagonal",
}

// detectorFunc is a per-file AST detector that returns violations.
type detectorFunc func(file *ast.File, fset *token.FileSet, relPath string) []RuleViolation

// RunAST executes an AST-engine rule by parsing Go files and running the
// named detector against them.
func RunAST(rule Rule, projectRoot string, allowedFiles []string) ([]RuleViolation, error) {
	fn, ok := detectorRegistry[rule.Detector]
	if !ok {
		return nil, fmt.Errorf("unknown detector %q for rule %s", rule.Detector, rule.ID)
	}

	files, err := ExpandScope(rule.Scope, rule.Exclude, projectRoot, allowedFiles)
	if err != nil {
		return nil, fmt.Errorf("expand scope for rule %s: %w", rule.ID, err)
	}

	var violations []RuleViolation
	fset := token.NewFileSet()

	for _, relPath := range files {
		if strings.HasSuffix(relPath, "_test.go") {
			continue
		}
		if !strings.HasSuffix(relPath, ".go") {
			continue
		}

		absPath := filepath.Join(projectRoot, relPath)
		file, err := parser.ParseFile(fset, absPath, nil, parser.AllErrors)
		if err != nil {
			continue // skip unparseable files
		}

		found := fn(file, fset, relPath)
		for i := range found {
			found[i].RuleID = rule.ID
			found[i].Engine = "ast"
			found[i].Description = rule.Description
			found[i].Severity = severity(rule.Severity)
			found[i].Ref = rule.Ref
		}
		violations = append(violations, found...)
	}

	return violations, nil
}

// detectorRegistry maps detector names to their per-file implementations.
var detectorRegistry = map[string]detectorFunc{
	"global-mutable-state":          detectGlobalMutableStateRule,
	"init-abuse":                    detectInitAbuseRule,
	"init-side-effects":             detectInitSideEffectsRule,
	"naked-goroutine":               detectNakedGoroutineRule,
	"swallowed-error":               detectSwallowedErrorRule,
	"context-background-in-handler": detectContextBackgroundRule,
	"sql-concatenation":             detectSQLConcatenationRule,
	"uuid-v4-as-key":                detectUUIDv4Rule,
	"fat-handler":                   detectFatHandlerRule,
	"god-package":                   unsupportedCrossPackage,
	"domain-imports-adapter":        unsupportedCrossPackage,
	"mvc-in-hexagonal":              unsupportedCrossPackage,
}

func unsupportedCrossPackage(_ *ast.File, _ *token.FileSet, _ string) []RuleViolation {
	return nil // cross-package detectors require full checker pipeline
}

// IsKnownDetector returns true if the detector name is recognized.
func IsKnownDetector(name string) bool {
	_, ok := detectorRegistry[name]
	return ok
}
