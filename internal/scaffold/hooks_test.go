package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPostScaffoldHooksSuccess(t *testing.T) {
	dir := t.TempDir()
	hooks := []string{"echo hello > hook.txt"}
	if err := RunPostScaffoldHooks(dir, hooks, nil); err != nil {
		t.Fatalf("RunPostScaffoldHooks() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "hook.txt")); err != nil {
		t.Fatalf("expected hook output file: %v", err)
	}
}

func TestRunPostScaffoldHooksFailure(t *testing.T) {
	err := RunPostScaffoldHooks(t.TempDir(), []string{"exit 42"}, nil)
	if err == nil {
		t.Fatal("expected hook error")
	}
}

func TestRunPostScaffoldHooks_RejectsShellInjection(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]interface{}
	}{
		{"semicolon injection", map[string]interface{}{"ServiceName": "foo; rm -rf /"}},
		{"pipe injection", map[string]interface{}{"ServiceName": "foo | cat /etc/passwd"}},
		{"backtick injection", map[string]interface{}{"ServiceName": "foo`whoami`"}},
		{"dollar expansion", map[string]interface{}{"ServiceName": "foo$(id)"}},
		{"ampersand", map[string]interface{}{"ServiceName": "foo && echo pwned"}},
		{"newline", map[string]interface{}{"ServiceName": "foo\necho pwned"}},
		{"single quote", map[string]interface{}{"ServiceName": "foo'bar"}},
		{"double quote", map[string]interface{}{"ServiceName": `foo"bar`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunPostScaffoldHooks(t.TempDir(), []string{"echo {{.ServiceName}}"}, tt.vars)
			if err == nil {
				t.Fatal("expected error for shell metacharacters")
			}
		})
	}
}

func TestRunPostScaffoldHooks_AllowsSafeValues(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]interface{}
	}{
		{"simple name", map[string]interface{}{"ServiceName": "my-service"}},
		{"with dots", map[string]interface{}{"ServiceName": "my.service"}},
		{"with underscores", map[string]interface{}{"ServiceName": "my_service"}},
		{"module path", map[string]interface{}{"ModulePath": "github.com/acme/orders"}},
		{"with at sign", map[string]interface{}{"ModulePath": "github.com/acme/orders@v1"}},
		{"nil vars", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunPostScaffoldHooks(t.TempDir(), []string{"echo hello"}, tt.vars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunPostScaffoldHooks_SkipsEmptyHooks(t *testing.T) {
	// Empty/whitespace hooks should be skipped silently.
	err := RunPostScaffoldHooks(t.TempDir(), []string{"", "  ", "echo ok"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPostScaffoldHooks_GitInitSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	// Create a .git directory to simulate existing repo.
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	// git init should be skipped when .git already exists.
	err := RunPostScaffoldHooks(dir, []string{"git init"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPostScaffoldHooks_TemplatedHook(t *testing.T) {
	dir := t.TempDir()
	vars := map[string]interface{}{"ServiceName": "myapp"}
	err := RunPostScaffoldHooks(dir, []string{"echo {{.ServiceName}} > name.txt"}, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "name.txt"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if got := strings.TrimSpace(string(content)); got != "myapp" {
		t.Errorf("got %q, want 'myapp'", got)
	}
}

func TestRenderHook_NilVars(t *testing.T) {
	got, err := renderHook("echo hello", nil)
	if err != nil {
		t.Fatalf("renderHook() error = %v", err)
	}
	if got != "echo hello" {
		t.Errorf("got %q, want 'echo hello'", got)
	}
}

func TestRenderHook_WithVars(t *testing.T) {
	got, err := renderHook("echo {{.Name}}", map[string]interface{}{"Name": "world"})
	if err != nil {
		t.Fatalf("renderHook() error = %v", err)
	}
	if got != "echo world" {
		t.Errorf("got %q, want 'echo world'", got)
	}
}

func TestRenderHook_InvalidTemplate(t *testing.T) {
	_, err := renderHook("echo {{.Invalid", nil)
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestValidateHookVars_NilValue(t *testing.T) {
	vars := map[string]interface{}{
		"Foo": nil,
	}
	if err := validateHookVars(vars); err != nil {
		t.Fatalf("should skip nil values: %v", err)
	}
}

func TestValidateHookVars_EmptyStringValue(t *testing.T) {
	vars := map[string]interface{}{
		"Foo": "",
	}
	if err := validateHookVars(vars); err != nil {
		t.Fatalf("should skip empty strings: %v", err)
	}
}

func TestValidateHookVars_SkipsBoolsAndMaps(t *testing.T) {
	vars := map[string]interface{}{
		"HasHTTP":  true,
		"HasRedis": false,
		"Partials": map[string]interface{}{"main_imports": []string{"foo"}},
	}
	if err := validateHookVars(vars); err != nil {
		t.Fatalf("should skip non-string types: %v", err)
	}
}
