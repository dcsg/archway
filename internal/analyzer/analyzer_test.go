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
