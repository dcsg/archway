package scaffold

import (
	"os"
	"path/filepath"
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
