package detector

import (
	"bufio"
	"strings"

	"github.com/dcsg/archway/internal/provider"
	"golang.org/x/tools/go/packages"
)

func DetectFramework(goModContent string, pkgs []*packages.Package) provider.FrameworkResult {
	modules := parseGoModModules(goModContent)
	imports := collectImports(pkgs)

	frameworks := []struct {
		Module     string
		Name       string
		Confidence float64
	}{
		{Module: "github.com/go-chi/chi", Name: "chi", Confidence: 0.95},
		{Module: "github.com/gin-gonic/gin", Name: "gin", Confidence: 0.95},
		{Module: "github.com/labstack/echo", Name: "echo", Confidence: 0.95},
		{Module: "github.com/gofiber/fiber", Name: "fiber", Confidence: 0.95},
		{Module: "google.golang.org/grpc", Name: "grpc", Confidence: 0.9},
	}

	result := provider.FrameworkResult{Name: "stdlib", Confidence: 0.6}
	for _, fw := range frameworks {
		if version, ok := moduleVersionByPrefix(modules, fw.Module); ok || hasImportPrefix(imports, fw.Module) {
			result.Name = fw.Name
			result.Confidence = fw.Confidence
			result.Version = version
			break
		}
	}

	dbModules := []struct {
		Module string
		Name   string
	}{
		{Module: "gorm.io/gorm", Name: "gorm"},
		{Module: "github.com/jmoiron/sqlx", Name: "sqlx"},
		{Module: "github.com/jackc/pgx", Name: "pgx"},
		{Module: "database/sql", Name: "database/sql"},
	}

	for _, db := range dbModules {
		if version, ok := moduleVersionByPrefix(modules, db.Module); ok {
			result.Libraries = append(result.Libraries, provider.LibraryVersion{Name: db.Name, Version: version})
			continue
		}
		if hasImportPrefix(imports, db.Module) {
			result.Libraries = append(result.Libraries, provider.LibraryVersion{Name: db.Name, Version: ""})
		}
	}

	return result
}

func parseGoModModules(content string) map[string]string {
	out := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "require (") || line == ")" {
			continue
		}
		if strings.HasPrefix(line, "require ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "require "))
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			out[fields[0]] = fields[1]
		}
	}
	return out
}

func collectImports(pkgs []*packages.Package) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, pkg := range pkgs {
		for importPath := range pkg.Imports {
			if seen[importPath] {
				continue
			}
			seen[importPath] = true
			out = append(out, importPath)
		}
	}
	return out
}

func hasImportPrefix(imports []string, prefix string) bool {
	for _, importPath := range imports {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	return false
}

func moduleVersionByPrefix(modules map[string]string, prefix string) (string, bool) {
	for module, version := range modules {
		if strings.HasPrefix(module, prefix) {
			return version, true
		}
	}
	return "", false
}
