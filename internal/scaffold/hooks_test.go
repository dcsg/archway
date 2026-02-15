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
