package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	renderer := NewRenderer(os.DirFS("testdata"))
	out := t.TempDir()

	result, err := renderer.RenderTemplate("minimal", out, map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	})
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}
	if len(result.FilesCreated) == 0 {
		t.Fatal("expected files to be created")
	}

	mainPath := filepath.Join(out, "cmd", "orders", "main.go")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("expected rendered main.go, got err: %v", err)
	}

	modBytes, err := os.ReadFile(filepath.Join(out, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.Contains(string(modBytes), "github.com/acme/orders") {
		t.Fatalf("go.mod missing variable substitution: %s", string(modBytes))
	}
}

func TestRenderTemplateFunctions(t *testing.T) {
	got := camelCase("my-service_name")
	if got != "myServiceName" {
		t.Fatalf("camelCase() = %q, want myServiceName", got)
	}
	if kebabCase("My Service") != "my-service" {
		t.Fatalf("kebabCase conversion failed")
	}
}

func TestRenderTemplateMissingRequiredVariable(t *testing.T) {
	renderer := NewRenderer(os.DirFS("testdata"))
	_, err := renderer.RenderTemplate("minimal", t.TempDir(), map[string]interface{}{"ServiceName": "orders"})
	if err == nil {
		t.Fatal("expected error for missing required variable")
	}
}

func TestRendererCopiesPlainFiles(t *testing.T) {
	renderer := NewRenderer(os.DirFS("testdata"))
	out := t.TempDir()
	_, err := renderer.RenderTemplate("minimal", out, map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	})
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}
	ignorePath := filepath.Join(out, ".gitignore")
	content, err := os.ReadFile(ignorePath)
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	if strings.TrimSpace(string(content)) != "bin/" {
		t.Fatalf("copied file content mismatch: %q", string(content))
	}
}

var _ fs.FS
