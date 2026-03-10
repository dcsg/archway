package checker

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"github.com/dcsg/archway/internal/analyzer"
	"github.com/dcsg/archway/internal/analyzer/graph"
	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/provider"
	"golang.org/x/tools/go/packages"
)

// Violation represents a single rule violation.
type Violation struct {
	Category string `json:"category"`
	File     string `json:"file"`
	Line     int    `json:"line,omitempty"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
}

// CheckResult holds the outcome of all checks.
type CheckResult struct {
	DependencyViolations  []Violation   `json:"dependency_violations"`
	StructureViolations   []Violation   `json:"structure_violations"`
	FunctionViolations    []Violation   `json:"function_violations"`
	NamingViolations      []Violation   `json:"naming_violations"`
	AntiPatternViolations []AntiPattern `json:"anti_pattern_violations"`
	ComponentsCovered     int           `json:"components_covered"`
	ComponentsTotal       int           `json:"components_total"`
	RulesChecked          int           `json:"rules_checked"`
	RulesPassing          int           `json:"rules_passing"`
}

// TotalViolations returns the count of all violations.
func (r *CheckResult) TotalViolations() int {
	return len(r.DependencyViolations) + len(r.StructureViolations) +
		len(r.FunctionViolations) + len(r.NamingViolations) +
		len(r.AntiPatternViolations)
}

// Passed returns true if no violations were found.
func (r *CheckResult) Passed() bool {
	return r.TotalViolations() == 0
}

// Compliance returns the ratio of passing rules (0.0 to 1.0).
func (r *CheckResult) Compliance() float64 {
	if r.RulesChecked == 0 {
		return 1.0
	}
	return float64(r.RulesPassing) / float64(r.RulesChecked)
}

// Check validates a project at the given path against its archway.yaml config.
func Check(cfg *config.ArchwayConfig, projectPath string) (*CheckResult, error) {
	result := &CheckResult{
		ComponentsTotal: len(cfg.Components),
	}

	// Load and analyze packages.
	a := analyzer.New(projectPath)
	if err := a.LoadPackages(""); err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	depGraph := graph.BuildGraph(a.Packages())

	// Dependency violations.
	depViolations := graph.LayerViolations(depGraph, cfg.Components)
	for _, v := range depViolations {
		result.DependencyViolations = append(result.DependencyViolations, Violation{
			Category: "dependency",
			File:     v.Source,
			Message:  v.Message,
			Rule:     v.Rule,
			Severity: v.Severity,
		})
	}

	// Component coverage.
	result.ComponentsCovered = countCoveredComponents(depGraph, cfg.Components)

	// Structure violations.
	result.StructureViolations = checkStructure(cfg.Rules.Structure, projectPath)

	// Function violations.
	result.FunctionViolations = checkFunctions(cfg.Rules.Functions, a.Packages())

	// Anti-pattern violations.
	result.AntiPatternViolations = checkAntiPatterns(a.Packages(), projectPath)

	// Compute metrics.
	result.RulesChecked = len(cfg.Components) + structureRuleCount(cfg.Rules.Structure) + functionRuleCount(cfg.Rules.Functions)
	result.RulesPassing = result.RulesChecked - result.TotalViolations()
	if result.RulesPassing < 0 {
		result.RulesPassing = 0
	}

	return result, nil
}

func countCoveredComponents(depGraph provider.DependencyGraph, components []config.Component) int {
	covered := 0
	for _, comp := range components {
		for _, node := range depGraph.Nodes {
			if graph.MatchesComponent(node.Path, comp) {
				covered++
				break
			}
		}
	}
	return covered
}

func checkStructure(rules config.StructureConfig, projectPath string) []Violation {
	var violations []Violation
	for _, dir := range rules.RequiredDirs {
		dirPath := filepath.Join(projectPath, filepath.FromSlash(strings.TrimSuffix(dir, "/")))
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			violations = append(violations, Violation{
				Category: "structure",
				Message:  fmt.Sprintf("missing required directory: %s", dir),
				Rule:     "required_dir",
				Severity: "error",
			})
		}
	}
	for _, dir := range rules.ForbiddenDirs {
		dirPath := filepath.Join(projectPath, filepath.FromSlash(strings.TrimSuffix(dir, "/")))
		if _, err := os.Stat(dirPath); err == nil {
			violations = append(violations, Violation{
				Category: "structure",
				Message:  fmt.Sprintf("forbidden directory exists: %s", dir),
				Rule:     "forbidden_dir",
				Severity: "error",
			})
		}
	}
	return violations
}

func checkFunctions(rules config.FunctionRules, pkgs []*packages.Package) []Violation {
	if rules.MaxLines == 0 && rules.MaxParams == 0 && rules.MaxReturnValues == 0 {
		return nil
	}

	var violations []Violation
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			fset := pkg.Fset
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn == nil {
					continue
				}

				filePath := fset.Position(fn.Pos()).Filename
				line := fset.Position(fn.Pos()).Line

				// Check line count.
				if rules.MaxLines > 0 && fn.Body != nil {
					startLine := fset.Position(fn.Body.Lbrace).Line
					endLine := fset.Position(fn.Body.Rbrace).Line
					lines := endLine - startLine
					if lines > rules.MaxLines {
						violations = append(violations, Violation{
							Category: "function",
							File:     filePath,
							Line:     line,
							Message:  fmt.Sprintf("%s — %d lines (max: %d)", fn.Name.Name, lines, rules.MaxLines),
							Rule:     "max_lines",
							Severity: "warning",
						})
					}
				}

				// Check param count.
				if rules.MaxParams > 0 && fn.Type.Params != nil {
					params := fn.Type.Params.NumFields()
					if params > rules.MaxParams {
						violations = append(violations, Violation{
							Category: "function",
							File:     filePath,
							Line:     line,
							Message:  fmt.Sprintf("%s — %d params (max: %d)", fn.Name.Name, params, rules.MaxParams),
							Rule:     "max_params",
							Severity: "warning",
						})
					}
				}

				// Check return value count.
				if rules.MaxReturnValues > 0 && fn.Type.Results != nil {
					results := fn.Type.Results.NumFields()
					if results > rules.MaxReturnValues {
						violations = append(violations, Violation{
							Category: "function",
							File:     filePath,
							Line:     line,
							Message:  fmt.Sprintf("%s — %d return values (max: %d)", fn.Name.Name, results, rules.MaxReturnValues),
							Rule:     "max_return_values",
							Severity: "warning",
						})
					}
				}
			}
		}
	}
	return violations
}

func structureRuleCount(rules config.StructureConfig) int {
	return len(rules.RequiredDirs) + len(rules.ForbiddenDirs)
}

func functionRuleCount(rules config.FunctionRules) int {
	count := 0
	if rules.MaxLines > 0 {
		count++
	}
	if rules.MaxParams > 0 {
		count++
	}
	if rules.MaxReturnValues > 0 {
		count++
	}
	return count
}
