package golang

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestMigrate(t *testing.T) {
	p := &GoProvider{}
	_, err := p.Migrate(context.Background(), provider.MigrateRequest{})
	if err != provider.ErrNotImplemented {
		t.Fatalf("Migrate() error = %v, want ErrNotImplemented", err)
	}
}

func TestGetTemplateFS(t *testing.T) {
	p := &GoProvider{}
	fs := p.GetTemplateFS()
	if fs == nil {
		t.Fatal("GetTemplateFS() returned nil")
	}
}

func TestScaffoldDefaultTemplate(t *testing.T) {
	p := &GoProvider{}
	out := filepath.Join(t.TempDir(), "default-svc")
	resp, err := p.Scaffold(context.Background(), provider.ScaffoldRequest{
		ProjectName: "default-svc",
		ModulePath:  "github.com/acme/default-svc",
		OutputDir:   out,
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

func TestAnalyze(t *testing.T) {
	// Analyze needs a valid Go project. Use the provider's own directory.
	p := &GoProvider{}
	_, err := p.Analyze(context.Background(), provider.AnalyzeRequest{Path: "."})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
}

func TestAnalyze_EmptyPath(t *testing.T) {
	p := &GoProvider{}
	// Empty path should default to "." and succeed.
	_, err := p.Analyze(context.Background(), provider.AnalyzeRequest{Path: ""})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
}

func TestAnalyze_WhitespacePath(t *testing.T) {
	p := &GoProvider{}
	_, err := p.Analyze(context.Background(), provider.AnalyzeRequest{Path: "  "})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
}

func TestAnalyze_InvalidPath(t *testing.T) {
	p := &GoProvider{}
	_, err := p.Analyze(context.Background(), provider.AnalyzeRequest{Path: "/nonexistent/path"})
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestGetInfo_TemplatesExist(t *testing.T) {
	p := &GoProvider{}
	info, err := p.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if info.Name != "archway-go-provider" {
		t.Errorf("Name = %q, want archway-go-provider", info.Name)
	}
	if info.Version != "v1" {
		t.Errorf("Version = %q, want v1", info.Version)
	}
	if len(info.SupportedArchitectures) == 0 {
		t.Error("expected supported architectures")
	}
	// Verify templates have variables populated.
	for _, tmpl := range info.Templates {
		if tmpl.Name == "" {
			t.Error("template with empty name")
		}
	}
}

func TestScaffold_InvalidArchitecture(t *testing.T) {
	p := &GoProvider{}
	out := filepath.Join(t.TempDir(), "invalid")
	_, err := p.Scaffold(context.Background(), provider.ScaffoldRequest{
		ProjectName:  "invalid",
		ModulePath:   "github.com/acme/invalid",
		TemplateName: "nonexistent-arch",
		OutputDir:    out,
		Options: map[string]string{
			"skip_hooks": "true",
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid architecture")
	}
}

func TestScaffoldWithCapabilities(t *testing.T) {
	p := &GoProvider{}
	out := filepath.Join(t.TempDir(), "orders")
	resp, err := p.Scaffold(context.Background(), provider.ScaffoldRequest{
		ProjectName:  "orders",
		ModulePath:   "github.com/acme/orders",
		TemplateName: "api",
		OutputDir:    out,
		Options: map[string]string{
			"skip_hooks":   "true",
			"capabilities": "platform,bootstrap,http-api,mysql",
		},
	})
	if err != nil {
		t.Fatalf("Scaffold() error = %v", err)
	}
	if len(resp.FilesCreated) == 0 {
		t.Fatal("expected created files")
	}

	// Verify capability files were rendered.
	httpHandler := filepath.Join(out, "adapter", "httphandler", "handler.go")
	if _, err := os.Stat(httpHandler); os.IsNotExist(err) {
		t.Error("expected http handler file to exist")
	}
	mysqlConn := filepath.Join(out, "adapter", "mysqlrepo", "connection.go")
	if _, err := os.Stat(mysqlConn); os.IsNotExist(err) {
		t.Error("expected mysql connection file to exist")
	}

	// Verify archway.yaml includes capabilities.
	archwayPath := filepath.Join(out, "archway.yaml")
	data, err := os.ReadFile(archwayPath)
	if err != nil {
		t.Fatalf("read archway.yaml: %v", err)
	}
	if !strings.Contains(string(data), "http-api") {
		t.Error("archway.yaml should contain http-api capability")
	}
}
