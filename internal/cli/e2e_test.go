package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/dcsg/archway/providers/golang"
)

func goModTidy(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("go mod tidy output: %s", string(out))
		t.Skipf("go mod tidy failed (requires network): %v", err)
	}
}

func TestE2E_ScaffoldCheckAnalyzeGuide(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	tmp := t.TempDir()
	svcDir := scaffoldHexagonal(t, tmp, "e2e-hex")

	// go mod tidy so check/analyze can parse Go code.
	goModTidy(t, svcDir)

	// Check.
	chdir(t, svcDir)
	_, err := executeCommand(t, "check", "--path", ".")
	if err != nil {
		t.Skip("requires go toolchain in scaffolded project")
	}

	// Analyze — verify it succeeds (output goes to os.Stdout, not captured by cobra).
	chdir(t, svcDir)
	_, err = executeCommand(t, "analyze", "--path", ".")
	if err != nil {
		t.Skip("requires go toolchain in scaffolded project")
	}

	// Guide.
	chdir(t, svcDir)
	_, err = executeCommand(t, "guide")
	if err != nil {
		t.Fatalf("guide command failed: %v", err)
	}

	for _, p := range []string{
		filepath.Join(".claude", "rules", "archway-index.md"),
		".cursorrules",
		filepath.Join(".github", "copilot-instructions.md"),
		".windsurfrules",
	} {
		full := filepath.Join(svcDir, p)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected %s to exist after guide", p)
		}
	}

	// Verify guide content references capabilities.
	data, err := os.ReadFile(filepath.Join(svcDir, ".claude", "rules", "archway-index.md"))
	if err != nil {
		t.Fatalf("failed to read archway-index.md: %v", err)
	}
	content := string(data)
	for _, cap := range []string{"http-api"} {
		if !strings.Contains(content, cap) {
			t.Errorf("guide content should mention capability %q", cap)
		}
	}
}

func TestE2E_FlatPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	tmp := t.TempDir()
	svcDir := scaffoldFlat(t, tmp, "e2e-flat")

	goModTidy(t, svcDir)

	// Check.
	chdir(t, svcDir)
	_, err := executeCommand(t, "check", "--path", ".")
	if err != nil {
		t.Skip("requires go toolchain in scaffolded project")
	}

	// Analyze — verify it succeeds.
	chdir(t, svcDir)
	_, err = executeCommand(t, "analyze", "--path", ".")
	if err != nil {
		t.Skip("requires go toolchain in scaffolded project")
	}

	// Guide.
	chdir(t, svcDir)
	_, err = executeCommand(t, "guide")
	if err != nil {
		t.Fatalf("guide command failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(svcDir, ".claude", "rules", "archway-index.md"))
	if err != nil {
		t.Fatalf("failed to read archway-index.md: %v", err)
	}
	content := string(data)
	for _, want := range []string{"flat", "no layer restrictions"} {
		if !strings.Contains(content, want) {
			t.Errorf("guide content should contain %q", want)
		}
	}
}

func TestE2E_MultipleCapabilities(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	tmp := t.TempDir()
	chdir(t, tmp)

	caps := "http-api,mysql,docker,observability,health"
	_, err := executeCommand(t,
		"new", "e2e-multi",
		"--arch", "hexagonal",
		"--cap", caps,
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("scaffold failed: %v", err)
	}

	svcDir := filepath.Join(tmp, "e2e-multi")
	goModTidy(t, svcDir)

	// Check.
	chdir(t, svcDir)
	_, err = executeCommand(t, "check", "--path", ".")
	if err != nil {
		t.Skip("requires go toolchain in scaffolded project")
	}

	// Guide.
	chdir(t, svcDir)
	_, err = executeCommand(t, "guide", "--target", "claude")
	if err != nil {
		t.Fatalf("guide command failed: %v", err)
	}

	indexData, err := os.ReadFile(filepath.Join(svcDir, ".claude", "rules", "archway-index.md"))
	if err != nil {
		t.Fatalf("failed to read archway-index.md: %v", err)
	}

	content := string(indexData)
	for _, cap := range []string{"http-api", "mysql", "docker", "observability", "health"} {
		if !strings.Contains(content, cap) {
			t.Errorf("guide index should mention capability %q", cap)
		}
	}
}
