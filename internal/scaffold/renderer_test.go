package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
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

func TestExecuteTemplate_CustomFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"camelCase", `{{camelCase "my-service"}}`, "myService"},
		{"snakeCase", `{{snakeCase "my-service"}}`, "my_service"},
		{"pascalCase", `{{pascalCase "my-service"}}`, "MyService"},
		{"kebabCase", `{{kebabCase "My Service"}}`, "my-service"},
		{"upper", `{{upper "hello"}}`, "HELLO"},
		{"lower", `{{lower "HELLO"}}`, "hello"},
		{"title", `{{title "hello world"}}`, "Hello World"},
		{"contains true", `{{contains "foobar" "bar"}}`, "true"},
		{"contains false", `{{contains "foobar" "baz"}}`, "false"},
		{"hasPrefix true", `{{hasPrefix "foobar" "foo"}}`, "true"},
		{"hasPrefix false", `{{hasPrefix "foobar" "bar"}}`, "false"},
		{"hasSuffix true", `{{hasSuffix "foobar" "bar"}}`, "true"},
		{"hasSuffix false", `{{hasSuffix "foobar" "foo"}}`, "false"},
		{"camelCase empty", `{{camelCase ""}}`, ""},
		{"snakeCase empty", `{{snakeCase ""}}`, ""},
		{"pascalCase empty", `{{pascalCase ""}}`, ""},
		{"kebabCase empty", `{{kebabCase ""}}`, ""},
		{"camelCase single word", `{{camelCase "hello"}}`, "hello"},
		{"pascalCase single word", `{{pascalCase "hello"}}`, "Hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executeTemplate(tt.template, map[string]interface{}{})
			if err != nil {
				t.Fatalf("executeTemplate() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("executeTemplate(%q) = %q, want %q", tt.template, string(got), tt.want)
			}
		})
	}
}

func TestRenderPath_WithTemplate(t *testing.T) {
	vars := map[string]interface{}{"ServiceName": "orders"}
	got, err := renderPath("cmd/{{.ServiceName}}/main.go", vars)
	if err != nil {
		t.Fatalf("renderPath() error = %v", err)
	}
	if got != "cmd/orders/main.go" {
		t.Errorf("renderPath() = %q, want %q", got, "cmd/orders/main.go")
	}
}

func TestRenderPath_NoTemplate(t *testing.T) {
	vars := map[string]interface{}{"ServiceName": "orders"}
	got, err := renderPath("cmd/server/main.go", vars)
	if err != nil {
		t.Fatalf("renderPath() error = %v", err)
	}
	if got != "cmd/server/main.go" {
		t.Errorf("renderPath() = %q, want %q", got, "cmd/server/main.go")
	}
}

func TestRemoveEmptyDirs(t *testing.T) {
	root := t.TempDir()
	// Create nested empty dirs.
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a dir with a file (should not be removed).
	withFile := filepath.Join(root, "keep")
	if err := os.MkdirAll(withFile, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(withFile, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	removeEmptyDirs(root)

	// Empty dirs should be gone.
	if _, err := os.Stat(filepath.Join(root, "a")); err == nil {
		t.Error("expected empty dir 'a' to be removed")
	}
	// Dir with file should remain.
	if _, err := os.Stat(withFile); err != nil {
		t.Error("expected 'keep' dir to remain")
	}
}

func TestRenderComposition(t *testing.T) {
	tfs := testCapabilityFS()
	vars := map[string]interface{}{
		"ServiceName": "orders",
		"ModulePath":  "github.com/acme/orders",
	}

	plan, err := ComposeProject(tfs, "hexagonal", []string{"http-api", "mysql"}, vars)
	if err != nil {
		t.Fatalf("ComposeProject() error = %v", err)
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	result, err := renderer.RenderComposition(plan, out)
	if err != nil {
		t.Fatalf("RenderComposition() error = %v", err)
	}
	if len(result.FilesCreated) == 0 {
		t.Fatal("expected files to be created")
	}

	// Check that architecture file was rendered.
	modPath := filepath.Join(out, "go.mod")
	content, err := os.ReadFile(modPath)
	if err != nil {
		t.Fatalf("expected go.mod: %v", err)
	}
	if !strings.Contains(string(content), "github.com/acme/orders") {
		t.Errorf("go.mod missing module path: %s", string(content))
	}

	// Check that capability files were rendered.
	handlerPath := filepath.Join(out, "handler.go")
	if _, err := os.Stat(handlerPath); err != nil {
		t.Fatalf("expected handler.go from http-api capability: %v", err)
	}
	repoPath := filepath.Join(out, "repo.go")
	if _, err := os.Stat(repoPath); err != nil {
		t.Fatalf("expected repo.go from mysql capability: %v", err)
	}
}

func TestRenderComposition_NoFilesDir(t *testing.T) {
	// Capability without a files/ directory should not error.
	tfs := fstest.MapFS{
		"templates/architectures/empty/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: empty\ndescription: test\n"),
		},
		"templates/capabilities/nofiles/capability.yaml": &fstest.MapFile{
			Data: []byte("name: nofiles\ndescription: no files dir\n"),
		},
	}

	plan := &CompositionPlan{
		Architecture: "empty",
		Capabilities: []string{"nofiles"},
		Manifest:     &Manifest{Name: "empty"},
		CapManifests: []CapabilityManifest{{Name: "nofiles"}},
		Vars:         map[string]interface{}{},
		Partials:     map[string][]string{},
		ArchDir:      "templates/architectures/empty",
		CapDirs:      []string{"templates/capabilities/nofiles"},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	result, err := renderer.RenderComposition(plan, out)
	if err != nil {
		t.Fatalf("RenderComposition() error = %v", err)
	}
	if len(result.FilesCreated) != 0 {
		t.Errorf("expected no files created, got %d", len(result.FilesCreated))
	}
}

func TestRenderTemplate_ConditionallyEmptyFile(t *testing.T) {
	tfs := fstest.MapFS{
		"conditional/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: conditional\ndescription: test\n"),
		},
		"conditional/files/maybe.go.tmpl": &fstest.MapFile{
			Data: []byte("{{if .Include}}package main{{end}}"),
		},
	}

	renderer := NewRenderer(tfs)

	// When Include is false, the file should be skipped.
	out1 := t.TempDir()
	result1, err := renderer.RenderTemplate("conditional", out1, map[string]interface{}{"Include": false})
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}
	if len(result1.FilesCreated) != 0 {
		t.Errorf("expected 0 files when template renders empty, got %d", len(result1.FilesCreated))
	}

	// When Include is true, the file should be created.
	out2 := t.TempDir()
	result2, err := renderer.RenderTemplate("conditional", out2, map[string]interface{}{"Include": true})
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}
	if len(result2.FilesCreated) != 1 {
		t.Errorf("expected 1 file, got %d", len(result2.FilesCreated))
	}
}

func TestExecuteTemplate_DateAndNow(t *testing.T) {
	got, err := executeTemplate(`{{now.Year}}`, map[string]interface{}{})
	if err != nil {
		t.Fatalf("executeTemplate() error = %v", err)
	}
	if len(string(got)) != 4 {
		t.Errorf("expected 4-digit year, got %q", string(got))
	}
}

func TestExecuteTemplate_JoinAndSplit(t *testing.T) {
	got, err := executeTemplate(`{{join .Items ","}}`, map[string]interface{}{
		"Items": []string{"a", "b", "c"},
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if string(got) != "a,b,c" {
		t.Errorf("join = %q, want 'a,b,c'", string(got))
	}
}

func TestExecuteTemplate_DateFunction(t *testing.T) {
	got, err := executeTemplate(`{{date "2006" now}}`, map[string]interface{}{})
	if err != nil {
		t.Fatalf("executeTemplate() error = %v", err)
	}
	if len(string(got)) != 4 {
		t.Errorf("expected 4-digit year from date func, got %q", string(got))
	}
}

func TestRenderComposition_ArchFileError(t *testing.T) {
	// Test renderFilesDir with a template that has an invalid template syntax.
	tfs := fstest.MapFS{
		"templates/architectures/badtmpl/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: badtmpl\ndescription: test\n"),
		},
		"templates/architectures/badtmpl/files/bad.go.tmpl": &fstest.MapFile{
			Data: []byte("{{.Invalid | nonexistentFunc}}"),
		},
	}

	plan := &CompositionPlan{
		Architecture: "badtmpl",
		Capabilities: []string{},
		Manifest:     &Manifest{Name: "badtmpl"},
		Vars:         map[string]interface{}{},
		Partials:     map[string][]string{},
		ArchDir:      "templates/architectures/badtmpl",
		CapDirs:      []string{},
	}

	renderer := NewRenderer(tfs)
	_, err := renderer.RenderComposition(plan, t.TempDir())
	if err == nil {
		t.Fatal("expected error for bad template in architecture files")
	}
}

func TestRenderFilesDir_PlainAndTemplateFiles(t *testing.T) {
	tfs := fstest.MapFS{
		"dir/files/plain.txt": &fstest.MapFile{
			Data: []byte("plain content"),
		},
		"dir/files/template.go.tmpl": &fstest.MapFile{
			Data: []byte("package {{.Name}}"),
		},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	result, err := renderer.renderFilesDir("dir/files", out, map[string]interface{}{"Name": "main"})
	if err != nil {
		t.Fatalf("renderFilesDir() error = %v", err)
	}
	if len(result.FilesCreated) != 2 {
		t.Errorf("expected 2 files, got %d", len(result.FilesCreated))
	}

	// Check plain file.
	content, err := os.ReadFile(filepath.Join(out, "plain.txt"))
	if err != nil {
		t.Fatalf("read plain.txt: %v", err)
	}
	if string(content) != "plain content" {
		t.Errorf("plain.txt = %q", string(content))
	}

	// Check template file (should strip .tmpl).
	content, err = os.ReadFile(filepath.Join(out, "template.go"))
	if err != nil {
		t.Fatalf("read template.go: %v", err)
	}
	if string(content) != "package main" {
		t.Errorf("template.go = %q", string(content))
	}
}

func TestRenderTemplate_BoolCoercion(t *testing.T) {
	tfs := fstest.MapFS{
		"booltest/manifest.yaml": &fstest.MapFile{
			Data: []byte(`name: booltest
description: test
variables:
  - name: Enable
    type: bool
`),
		},
		"booltest/files/out.txt.tmpl": &fstest.MapFile{
			Data: []byte(`{{if .Enable}}enabled{{else}}disabled{{end}}`),
		},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	result, err := renderer.RenderTemplate("booltest", out, map[string]interface{}{"Enable": "true"})
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}
	if len(result.FilesCreated) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.FilesCreated))
	}
	content, err := os.ReadFile(filepath.Join(out, "out.txt"))
	if err != nil {
		t.Fatalf("read out.txt: %v", err)
	}
	if strings.TrimSpace(string(content)) != "enabled" {
		t.Errorf("got %q, want 'enabled'", string(content))
	}
}

func TestRenderTemplate_NilVars(t *testing.T) {
	tfs := fstest.MapFS{
		"nilvar/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: nilvar\ndescription: test\n"),
		},
		"nilvar/files/hello.txt": &fstest.MapFile{
			Data: []byte("hello"),
		},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	result, err := renderer.RenderTemplate("nilvar", out, nil)
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}
	if len(result.FilesCreated) != 1 {
		t.Errorf("expected 1 file, got %d", len(result.FilesCreated))
	}
}

func TestRenderFilesDir_WithSubdirectories(t *testing.T) {
	tfs := fstest.MapFS{
		"dir/files/sub/nested/file.txt": &fstest.MapFile{
			Data: []byte("nested content"),
		},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	result, err := renderer.renderFilesDir("dir/files", out, map[string]interface{}{})
	if err != nil {
		t.Fatalf("renderFilesDir() error = %v", err)
	}
	if len(result.FilesCreated) != 1 {
		t.Errorf("expected 1 file, got %d", len(result.FilesCreated))
	}
	content, err := os.ReadFile(filepath.Join(out, "sub", "nested", "file.txt"))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(content) != "nested content" {
		t.Errorf("content = %q", string(content))
	}
}

func TestRenderFilesDir_EmptyTemplate(t *testing.T) {
	// Templates that render to empty should be skipped.
	tfs := fstest.MapFS{
		"dir/files/empty.go.tmpl": &fstest.MapFile{
			Data: []byte("{{if .Include}}content{{end}}"),
		},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	result, err := renderer.renderFilesDir("dir/files", out, map[string]interface{}{"Include": false})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(result.FilesCreated) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.FilesCreated))
	}
}

func TestRenderFilesDir_NonexistentDir(t *testing.T) {
	tfs := fstest.MapFS{}
	renderer := NewRenderer(tfs)
	result, err := renderer.renderFilesDir("nonexistent/files", t.TempDir(), map[string]interface{}{})
	if err != nil {
		t.Fatalf("expected no error for missing dir, got: %v", err)
	}
	if len(result.FilesCreated) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.FilesCreated))
	}
}

func TestRenderTemplate_MissingManifest(t *testing.T) {
	tfs := fstest.MapFS{}
	renderer := NewRenderer(tfs)
	_, err := renderer.RenderTemplate("nomanifest", t.TempDir(), nil)
	if err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestRenderTemplate_InvalidManifest(t *testing.T) {
	tfs := fstest.MapFS{
		"bad/manifest.yaml": &fstest.MapFile{
			Data: []byte(":::invalid"),
		},
	}
	renderer := NewRenderer(tfs)
	_, err := renderer.RenderTemplate("bad", t.TempDir(), nil)
	if err == nil {
		t.Fatal("expected error for invalid manifest")
	}
}

func TestRenderTemplate_WithDefaults(t *testing.T) {
	tfs := fstest.MapFS{
		"withdefault/manifest.yaml": &fstest.MapFile{
			Data: []byte(`name: withdefault
description: test
variables:
  - name: Port
    type: string
    default: "3000"
`),
		},
		"withdefault/files/config.txt.tmpl": &fstest.MapFile{
			Data: []byte("port={{.Port}}"),
		},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	result, err := renderer.RenderTemplate("withdefault", out, map[string]interface{}{})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(result.FilesCreated) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.FilesCreated))
	}
	content, err := os.ReadFile(filepath.Join(out, "config.txt"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(content) != "port=3000" {
		t.Errorf("content = %q, want 'port=3000'", string(content))
	}
}

func TestRenderTemplate_PathRendering(t *testing.T) {
	tfs := fstest.MapFS{
		"pathtest/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: pathtest\ndescription: test\n"),
		},
		"pathtest/files/__Name__/main.go.tmpl": &fstest.MapFile{
			Data: []byte("package main"),
		},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	_, err := renderer.RenderTemplate("pathtest", out, map[string]interface{}{"Name": "myapp"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "myapp", "main.go")); err != nil {
		t.Fatalf("expected myapp/main.go: %v", err)
	}
}

func TestRenderFilesDir_PathTraversal(t *testing.T) {
	tfs := fstest.MapFS{
		"dir/files/__Name__/file.txt": &fstest.MapFile{
			Data: []byte("content"),
		},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	_, err := renderer.renderFilesDir("dir/files", out, map[string]interface{}{"Name": "../../etc"})
	if err == nil {
		t.Fatal("expected path traversal error")
	}
}

func TestRenderFilesDir_BadTemplate(t *testing.T) {
	tfs := fstest.MapFS{
		"dir/files/bad.go.tmpl": &fstest.MapFile{
			Data: []byte("{{.X | badFunc}}"),
		},
	}

	renderer := NewRenderer(tfs)
	_, err := renderer.renderFilesDir("dir/files", t.TempDir(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for bad template")
	}
}

func TestRenderTemplate_BadTemplateInWalk(t *testing.T) {
	tfs := fstest.MapFS{
		"badwalk/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: badwalk\ndescription: test\n"),
		},
		"badwalk/files/bad.go.tmpl": &fstest.MapFile{
			Data: []byte("{{.X | badFunc}}"),
		},
	}

	renderer := NewRenderer(tfs)
	_, err := renderer.RenderTemplate("badwalk", t.TempDir(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for bad template during walk")
	}
}

func TestRenderTemplate_DirCreation(t *testing.T) {
	tfs := fstest.MapFS{
		"dirtest/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: dirtest\ndescription: test\n"),
		},
		"dirtest/files/a/b/c/file.txt": &fstest.MapFile{
			Data: []byte("deep"),
		},
	}

	renderer := NewRenderer(tfs)
	out := t.TempDir()
	result, err := renderer.RenderTemplate("dirtest", out, map[string]interface{}{})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(result.FilesCreated) != 1 {
		t.Errorf("expected 1 file, got %d", len(result.FilesCreated))
	}
}

func TestRenderFilesDir_BadPathTemplate(t *testing.T) {
	tfs := fstest.MapFS{
		"dir/files/{{.Bad/file.txt": &fstest.MapFile{
			Data: []byte("content"),
		},
	}

	renderer := NewRenderer(tfs)
	_, err := renderer.renderFilesDir("dir/files", t.TempDir(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for bad path template")
	}
}

func TestRenderTemplate_PathTraversal(t *testing.T) {
	tfs := fstest.MapFS{
		"traversal/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: traversal\ndescription: test\n"),
		},
		"traversal/files/__Evil__/file.txt": &fstest.MapFile{
			Data: []byte("content"),
		},
	}

	renderer := NewRenderer(tfs)
	_, err := renderer.RenderTemplate("traversal", t.TempDir(), map[string]interface{}{"Evil": "../../etc"})
	if err == nil {
		t.Fatal("expected path traversal error")
	}
}

func TestRenderComposition_CapabilityError(t *testing.T) {
	tfs := fstest.MapFS{
		"templates/architectures/ok/files/file.txt": &fstest.MapFile{
			Data: []byte("ok"),
		},
		"templates/capabilities/badcap/files/bad.go.tmpl": &fstest.MapFile{
			Data: []byte("{{.X | badFunc}}"),
		},
	}

	plan := &CompositionPlan{
		Architecture: "ok",
		Capabilities: []string{"badcap"},
		Manifest:     &Manifest{Name: "ok"},
		Vars:         map[string]interface{}{},
		Partials:     map[string][]string{},
		ArchDir:      "templates/architectures/ok",
		CapDirs:      []string{"templates/capabilities/badcap"},
	}

	renderer := NewRenderer(tfs)
	_, err := renderer.RenderComposition(plan, t.TempDir())
	if err == nil {
		t.Fatal("expected error from bad capability template")
	}
}

func TestExecuteTemplate_MissingKey(t *testing.T) {
	// missingkey=zero should not error on missing keys.
	got, err := executeTemplate("{{.Missing}}", map[string]interface{}{})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if string(got) != "<no value>" {
		t.Errorf("got %q", string(got))
	}
}

func TestExecuteTemplate_InvalidTemplate(t *testing.T) {
	_, err := executeTemplate(`{{invalid`, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestWords(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"hello-world", []string{"hello", "world"}},
		{"CamelCase", []string{"camelcase"}},
		{"my_service_name", []string{"my", "service", "name"}},
		{"", nil},
		{"with spaces", []string{"with", "spaces"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := words(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("words(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("words(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

var _ fs.FS
