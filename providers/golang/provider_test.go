package golang

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dcsg/archway/internal/provider"
)

func TestImplementsLanguageProvider(t *testing.T) {
	var _ provider.LanguageProvider = (*GoProvider)(nil)
}

func TestGetInfo(t *testing.T) {
	p := &GoProvider{}
	info, err := p.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if info.Language != "go" {
		t.Fatalf("Language = %q, want go", info.Language)
	}
	if len(info.Templates) == 0 {
		t.Fatal("expected templates")
	}
}

func TestScaffold(t *testing.T) {
	p := &GoProvider{}
	out := filepath.Join(t.TempDir(), "orders")
	resp, err := p.Scaffold(context.Background(), provider.ScaffoldRequest{
		ProjectName:  "orders",
		ModulePath:   "github.com/acme/orders",
		TemplateName: "cli",
		OutputDir:    out,
		Options: map[string]string{
			"skip_hooks": "true",
		},
	})
	if err != nil {
		t.Fatalf("Scaffold() error = %v", err)
	}
	if len(resp.FilesCreated) == 0 {
		t.Fatal("expected created files")
	}
}
