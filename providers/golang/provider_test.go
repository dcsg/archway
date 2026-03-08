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
