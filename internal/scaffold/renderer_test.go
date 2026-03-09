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

func TestValidatePathWithinDir(t *testing.T) {
	base := t.TempDir()

	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{"child path ok", filepath.Join(base, "sub", "file.go"), false},
		{"exact dir ok", base, false},
		{"traversal blocked", filepath.Join(base, "..", "etc", "passwd"), true},
		{"double traversal blocked", filepath.Join(base, "a", "..", "..", "evil"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePathWithinDir(tt.target, base)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePathWithinDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRenderPath_TraversalBlocked(t *testing.T) {
	// Verify that renderPath with a malicious variable produces a path
	// that would be caught by validatePathWithinDir.
	vars := map[string]interface{}{
		"ServiceName": "../../etc",
		"ModulePath":  "github.com/acme/orders",
	}
	rendered, err := renderPath("cmd/__ServiceName__/main.go", vars)
	if err != nil {
		t.Fatalf("renderPath() error = %v", err)
	}
	// The rendered path should contain ".." which validatePathWithinDir would catch.
	outDir := t.TempDir()
	absOut, _ := filepath.Abs(outDir)
	dstPath := filepath.Join(outDir, filepath.FromSlash(rendered))
	if err := validatePathWithinDir(dstPath, absOut); err == nil {
		t.Fatal("expected path traversal to be blocked")
	}
}

var _ fs.FS
