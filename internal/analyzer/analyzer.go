package analyzer

import (
	"context"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"github.com/dcsg/archway/internal/analyzer/detector"
	"github.com/dcsg/archway/internal/analyzer/graph"
	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/provider"
	"golang.org/x/tools/go/packages"
)

type Analyzer struct {
	path         string
	pkgs         []*packages.Package
	goModContent string
}

func New(path string) *Analyzer {
	return &Analyzer{path: path}
}

func (a *Analyzer) LoadPackages(path string) error {
	if path != "" {
		a.path = path
	}
	if a.path == "" {
		a.path = "."
	}
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedSyntax |
			packages.NeedModule |
			packages.NeedCompiledGoFiles,
		Dir: a.path,
	}
	cfg.Env = withSafeGoEnv()

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}
	a.pkgs = pkgs

	goModPath := filepath.Join(a.path, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		a.goModContent = string(data)
	}

	return nil
}

func (a *Analyzer) Analyze(_ context.Context) (*provider.AnalyzeResponse, error) {
	if len(a.pkgs) == 0 {
		if err := a.LoadPackages(a.path); err != nil {
			return nil, err
		}
	}

	lang, _, err := detector.DetectLanguage(a.path)
	if err != nil {
		lang = "unknown"
	}

	depGraph := graph.BuildGraph(a.pkgs)
	arch := detector.DetectArchitecture(depGraph, a.pkgs)
	framework := detector.DetectFramework(a.goModContent, a.pkgs)
	conventions := detector.DetectConventions(a.pkgs)

	result := &provider.AnalyzeResponse{
		Language:        lang,
		Architecture:    arch,
		Framework:       framework,
		Conventions:     conventions,
		DependencyGraph: depGraph,
		PackageCount:    len(depGraph.Nodes),
		FileCount:       countFiles(a.pkgs),
		FunctionCount:   countFunctions(a.pkgs),
		Metadata: map[string]string{
			"analyzer": "go/packages",
		},
	}

	if archwayPath, err := config.FindArchwayYAML(a.path); err == nil {
		if cfg, err := config.LoadArchwayYAML(archwayPath); err == nil {
			result.Violations = graph.LayerViolations(depGraph, cfg.Rules.Dependencies)
		}
	}
	return result, nil
}

func countFiles(pkgs []*packages.Package) int {
	files := map[string]bool{}
	for _, pkg := range pkgs {
		for _, file := range pkg.CompiledGoFiles {
			files[file] = true
		}
	}
	return len(files)
}

func countFunctions(pkgs []*packages.Package) int {
	total := 0
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn != nil {
					total++
				}
			}
		}
	}
	return total
}

func (a *Analyzer) Packages() []*packages.Package {
	return a.pkgs
}

func (a *Analyzer) Path() string {
	if strings.TrimSpace(a.path) == "" {
		return "."
	}
	return a.path
}

func withSafeGoEnv() []string {
	env := os.Environ()
	has := func(key string) bool {
		prefix := key + "="
		for _, entry := range env {
			if strings.HasPrefix(entry, prefix) {
				return true
			}
		}
		return false
	}

	if !has("GOCACHE") {
		cache := filepath.Join(os.TempDir(), "archway-go-build-cache")
		_ = os.MkdirAll(cache, 0o755)
		env = append(env, "GOCACHE="+cache)
	}
	if !has("GOMODCACHE") {
		modcache := filepath.Join(os.TempDir(), "archway-go-mod-cache")
		_ = os.MkdirAll(modcache, 0o755)
		env = append(env, "GOMODCACHE="+modcache)
	}
	return env
}
