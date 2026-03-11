package analyzer

import (
	"context"
	"path/filepath"
	"testing"
)

func TestAnalyzerAnalyzeHexagonal(t *testing.T) {
	a := New(filepath.Join("testdata", "hexagonal"))
	if err := a.LoadPackages(""); err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}
	result, err := a.Analyze(context.Background())
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Language != "go" {
		t.Fatalf("Language = %q, want go", result.Language)
	}
	if result.PackageCount == 0 {
		t.Fatal("expected packages")
	}
	if result.Architecture.Confidence <= 0 {
		t.Fatalf("expected architecture confidence > 0, got %v", result.Architecture.Confidence)
	}
}

func TestAnalyzerInvalidPath(t *testing.T) {
	a := New(filepath.Join("testdata", "missing"))
	if err := a.LoadPackages(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestAnalyzerPackages(t *testing.T) {
	a := New(filepath.Join("testdata", "hexagonal"))
	if err := a.LoadPackages(""); err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}
	pkgs := a.Packages()
	if len(pkgs) == 0 {
		t.Fatal("Packages() returned empty after LoadPackages")
	}
}

func TestAnalyzerPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"non-empty path", "/some/path", "/some/path"},
		{"empty path returns dot", "", "."},
		{"whitespace path returns dot", "   ", "."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(tt.path)
			if got := a.Path(); got != tt.want {
				t.Errorf("Path() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAnalyzerAnalyzeWithoutLoadPackages(t *testing.T) {
	// Analyze should auto-load packages when pkgs is empty.
	a := New(filepath.Join("testdata", "hexagonal"))
	result, err := a.Analyze(context.Background())
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.PackageCount == 0 {
		t.Fatal("expected packages from auto-load")
	}
}

func TestAnalyzerLoadPackagesOverridesPath(t *testing.T) {
	a := New("")
	err := a.LoadPackages(filepath.Join("testdata", "hexagonal"))
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}
	if a.Path() != filepath.Join("testdata", "hexagonal") {
		t.Errorf("Path() = %q after override", a.Path())
	}
}

func TestAnalyzerLoadPackagesDefaultPath(t *testing.T) {
	// When both path and arg are empty, defaults to ".".
	a := New("")
	// This will load from "." which should work (we're in the analyzer package).
	err := a.LoadPackages("")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}
	if a.Path() != "." {
		t.Errorf("Path() = %q, want '.'", a.Path())
	}
}
