package detector

import (
	"go/ast"
	"strconv"
	"strings"

	"github.com/dcsg/archway/internal/provider"
	"golang.org/x/tools/go/packages"
)

func DetectConventions(pkgs []*packages.Package) provider.ConventionResults {
	return provider.ConventionResults{
		ErrorHandling: detectErrorHandling(pkgs),
		Logging:       detectLogging(pkgs),
		Config:        detectConfig(pkgs),
		Testing:       detectTesting(pkgs),
	}
}

func detectErrorHandling(pkgs []*packages.Package) provider.ConventionFinding {
	sentinel := 0
	typed := 0
	wrapped := 0

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.ValueSpec:
					for _, name := range n.Names {
						if strings.HasPrefix(name.Name, "Err") {
							sentinel++
						}
					}
				case *ast.TypeSpec:
					if strings.HasSuffix(n.Name.Name, "Error") {
						typed++
					}
				case *ast.CallExpr:
					if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
						if x, ok := sel.X.(*ast.Ident); ok && x.Name == "fmt" && sel.Sel.Name == "Errorf" {
							for _, arg := range n.Args {
								if lit, ok := arg.(*ast.BasicLit); ok && strings.Contains(lit.Value, "%w") {
									wrapped++
								}
							}
						}
					}
				}
				return true
			})
		}
	}

	pattern := "minimal"
	confidence := 0.3
	evidence := []string{}
	max := sentinel
	if typed > max {
		max = typed
		pattern = "typed"
		confidence = 0.8
	}
	if wrapped > max {
		_ = max // consumed by comparison above
		pattern = "wrapped"
		confidence = 0.8
	}
	if sentinel > 0 {
		evidence = append(evidence, "sentinel errors detected")
		if pattern == "minimal" {
			pattern = "sentinel"
			confidence = 0.7
		}
	}
	if typed > 0 {
		evidence = append(evidence, "typed errors detected")
	}
	if wrapped > 0 {
		evidence = append(evidence, "error wrapping with %w detected")
	}
	return provider.ConventionFinding{Pattern: pattern, Confidence: confidence, Evidence: evidence}
}

func detectLogging(pkgs []*packages.Package) provider.ConventionFinding {
	slog := 0
	zap := 0
	zerolog := 0
	unstructured := 0
	structured := 0

	for _, pkg := range pkgs {
		for importPath := range pkg.Imports {
			switch {
			case strings.Contains(importPath, "log/slog"):
				slog++
			case strings.Contains(importPath, "go.uber.org/zap"):
				zap++
			case strings.Contains(importPath, "github.com/rs/zerolog"):
				zerolog++
			}
		}
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if sel.Sel.Name == "With" {
					structured++
				}
				if sel.Sel.Name == "Printf" || sel.Sel.Name == "Println" {
					unstructured++
				}
				return true
			})
		}
	}

	pattern := "unstructured"
	if structured >= unstructured {
		pattern = "structured"
	}
	library := "stdlib"
	switch {
	case slog > 0:
		library = "slog"
	case zap > 0:
		library = "zap"
	case zerolog > 0:
		library = "zerolog"
	}

	return provider.ConventionFinding{
		Pattern:    library + "/" + pattern,
		Confidence: 0.8,
		Evidence:   []string{"detected logging library: " + library},
	}
}

func detectConfig(pkgs []*packages.Package) provider.ConventionFinding {
	hits := map[string]int{"viper": 0, "koanf": 0, "godotenv": 0, "envconfig": 0}
	tags := 0
	for _, pkg := range pkgs {
		for importPath := range pkg.Imports {
			switch {
			case strings.Contains(importPath, "spf13/viper"):
				hits["viper"]++
			case strings.Contains(importPath, "knadh/koanf"):
				hits["koanf"]++
			case strings.Contains(importPath, "godotenv"):
				hits["godotenv"]++
			case strings.Contains(importPath, "envconfig"):
				hits["envconfig"]++
			}
		}
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(node ast.Node) bool {
				field, ok := node.(*ast.Field)
				if !ok || field.Tag == nil {
					return true
				}
				tag := field.Tag.Value
				if strings.Contains(tag, "mapstructure") || strings.Contains(tag, "env") || strings.Contains(tag, "yaml") {
					tags++
				}
				return true
			})
		}
	}

	best := "env"
	bestScore := 0
	for k, v := range hits {
		if v > bestScore {
			best = k
			bestScore = v
		}
	}
	if bestScore == 0 && tags > 0 {
		best = "struct-tags"
	}
	confidence := 0.6
	if bestScore > 0 || tags > 0 {
		confidence = 0.8
	}
	return provider.ConventionFinding{Pattern: best, Confidence: confidence, Evidence: []string{"config tags: " + strconv.Itoa(tags)}}
}

func detectTesting(pkgs []*packages.Package) provider.TestingFinding {
	testFiles := 0
	totalGoFiles := 0
	tableDriven := 0
	bdd := 0

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			filename := pkg.Fset.Position(file.Pos()).Filename
			if strings.HasSuffix(filename, ".go") {
				totalGoFiles++
			}
			if strings.HasSuffix(filename, "_test.go") {
				testFiles++
			}
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if ok && sel.Sel.Name == "Run" {
					tableDriven++
				}
				return true
			})
		}
		for importPath := range pkg.Imports {
			if strings.Contains(importPath, "ginkgo") || strings.Contains(importPath, "godog") {
				bdd++
			}
		}
	}

	pattern := "minimal"
	confidence := 0.4
	evidence := []string{}
	if tableDriven > 0 {
		pattern = "table-driven"
		confidence = 0.85
		evidence = append(evidence, "found t.Run usage")
	}
	if bdd > 0 {
		pattern = "bdd"
		confidence = 0.9
		evidence = append(evidence, "found BDD test libraries")
	}

	return provider.TestingFinding{
		Pattern:      pattern,
		Confidence:   confidence,
		Evidence:     evidence,
		TestFiles:    testFiles,
		TotalGoFiles: totalGoFiles,
	}
}
