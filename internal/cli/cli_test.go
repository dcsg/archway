package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dcsg/archway/internal/provider"
	_ "github.com/dcsg/archway/providers/golang"
)

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})
}

func TestVersionCommand(t *testing.T) {
	_, err := executeCommand(t, "version")
	if err != nil {
		t.Fatalf("version command should not error, got: %v", err)
	}
}

func TestInvalidOutputFlag(t *testing.T) {
	_, err := executeCommand(t, "--output", "invalid", "version")
	if err == nil {
		t.Fatal("expected error for invalid --output flag")
	}
	if !strings.Contains(err.Error(), "invalid --output value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewHexagonalWithHTTPAPI(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new", "test-svc",
		"--arch", "hexagonal",
		"--cap", "http-api",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	svcDir := filepath.Join(tmp, "test-svc")

	// Verify output directory exists.
	if _, err := os.Stat(svcDir); os.IsNotExist(err) {
		t.Fatal("expected service directory to exist")
	}

	// Verify key files and dirs exist.
	for _, p := range []string{
		"go.mod",
		"archway.yaml",
		"domain",
		"adapter",
		"port",
		"service",
	} {
		full := filepath.Join(svcDir, p)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", p)
		}
	}

	// Verify archway.yaml contains architecture: hexagonal.
	data, err := os.ReadFile(filepath.Join(svcDir, "archway.yaml"))
	if err != nil {
		t.Fatalf("failed to read archway.yaml: %v", err)
	}
	if !strings.Contains(string(data), "architecture: hexagonal") {
		t.Errorf("archway.yaml should contain 'architecture: hexagonal', got:\n%s", string(data))
	}
}

func TestNewFlat(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new", "test-flat",
		"--arch", "flat",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	svcDir := filepath.Join(tmp, "test-flat")

	for _, p := range []string{
		"main.go",
		"go.mod",
		"archway.yaml",
	} {
		full := filepath.Join(svcDir, p)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", p)
		}
	}

	data, err := os.ReadFile(filepath.Join(svcDir, "archway.yaml"))
	if err != nil {
		t.Fatalf("failed to read archway.yaml: %v", err)
	}
	if !strings.Contains(string(data), "architecture: flat") {
		t.Errorf("archway.yaml should contain 'architecture: flat', got:\n%s", string(data))
	}
}

func TestNewLayered(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new", "test-layered",
		"--arch", "layered",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	svcDir := filepath.Join(tmp, "test-layered")

	for _, p := range []string{
		"go.mod",
		"archway.yaml",
		filepath.Join("internal", "handler"),
		filepath.Join("internal", "service"),
		filepath.Join("internal", "repository"),
		filepath.Join("internal", "model"),
	} {
		full := filepath.Join(svcDir, p)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", p)
		}
	}

	data, err := os.ReadFile(filepath.Join(svcDir, "archway.yaml"))
	if err != nil {
		t.Fatalf("failed to read archway.yaml: %v", err)
	}
	if !strings.Contains(string(data), "architecture: layered") {
		t.Errorf("archway.yaml should contain 'architecture: layered', got:\n%s", string(data))
	}
}

func TestNewClean(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new", "test-clean",
		"--arch", "clean",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	svcDir := filepath.Join(tmp, "test-clean")

	for _, p := range []string{
		"go.mod",
		"archway.yaml",
		filepath.Join("internal", "entity"),
		filepath.Join("internal", "usecase"),
		filepath.Join("internal", "interface", "handler"),
		filepath.Join("internal", "infrastructure", "config"),
	} {
		full := filepath.Join(svcDir, p)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", p)
		}
	}

	data, err := os.ReadFile(filepath.Join(svcDir, "archway.yaml"))
	if err != nil {
		t.Fatalf("failed to read archway.yaml: %v", err)
	}
	if !strings.Contains(string(data), "architecture: clean") {
		t.Errorf("archway.yaml should contain 'architecture: clean', got:\n%s", string(data))
	}
}

func scaffoldClean(t *testing.T, dir, name string) string {
	t.Helper()
	chdir(t, dir)
	_, err := executeCommand(t,
		"new", name,
		"--arch", "clean",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("scaffold clean failed: %v", err)
	}
	return filepath.Join(dir, name)
}

func TestGuide_CleanArchitecture(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldClean(t, tmp, "guide-clean")

	chdir(t, svcDir)
	_, err := executeCommand(t, "guide", "--target", "claude")
	if err != nil {
		t.Fatalf("guide command failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(svcDir, ".claude", "rules", "archway.md"))
	if err != nil {
		t.Fatalf("failed to read archway.md: %v", err)
	}

	content := string(data)
	for _, want := range []string{"clean", "entity", "usecase", "interface", "infrastructure"} {
		if !strings.Contains(content, want) {
			t.Errorf("guide content should contain %q", want)
		}
	}
	if !strings.Contains(content, "NEVER let `internal/entity/` import from usecase") {
		t.Errorf("guide content should contain clean NEVER rules")
	}
	if !strings.Contains(content, "NEVER let `internal/usecase/` import from infrastructure") {
		t.Errorf("guide content should contain clean usecase NEVER rule")
	}
}

func scaffoldLayered(t *testing.T, dir, name string) string {
	t.Helper()
	chdir(t, dir)
	_, err := executeCommand(t,
		"new", name,
		"--arch", "layered",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("scaffold layered failed: %v", err)
	}
	return filepath.Join(dir, name)
}

func TestGuide_LayeredArchitecture(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldLayered(t, tmp, "guide-layered")

	chdir(t, svcDir)
	_, err := executeCommand(t, "guide", "--target", "claude")
	if err != nil {
		t.Fatalf("guide command failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(svcDir, ".claude", "rules", "archway.md"))
	if err != nil {
		t.Fatalf("failed to read archway.md: %v", err)
	}

	content := string(data)
	for _, want := range []string{"layered", "handler", "service", "repository", "model"} {
		if !strings.Contains(content, want) {
			t.Errorf("guide content should contain %q", want)
		}
	}
	if !strings.Contains(content, "NEVER let handler bypass service") {
		t.Errorf("guide content should contain layered NEVER rules")
	}
}

func TestNewInvalidArchitecture(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new", "test-bad",
		"--arch", "nonexistent",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err == nil {
		t.Fatal("expected error for invalid architecture")
	}
}

func TestNewMultipleCapabilities(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new", "test-multi",
		"--arch", "hexagonal",
		"--cap", "http-api,mysql,docker",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	svcDir := filepath.Join(tmp, "test-multi")
	data, err := os.ReadFile(filepath.Join(svcDir, "archway.yaml"))
	if err != nil {
		t.Fatalf("failed to read archway.yaml: %v", err)
	}

	content := string(data)
	for _, cap := range []string{"http-api", "mysql", "docker"} {
		if !strings.Contains(content, cap) {
			t.Errorf("archway.yaml should list capability %q, got:\n%s", cap, content)
		}
	}
}

func TestNewRequiresNameWithNoWizard(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new",
		"--arch", "hexagonal",
		"--no-wizard",
	)
	if err == nil {
		t.Fatal("expected error when name is missing with --no-wizard")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func scaffoldHexagonal(t *testing.T, dir, name string) string {
	t.Helper()
	chdir(t, dir)
	_, err := executeCommand(t,
		"new", name,
		"--arch", "hexagonal",
		"--cap", "http-api",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("scaffold hexagonal failed: %v", err)
	}
	return filepath.Join(dir, name)
}

func scaffoldFlat(t *testing.T, dir, name string) string {
	t.Helper()
	chdir(t, dir)
	_, err := executeCommand(t,
		"new", name,
		"--arch", "flat",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("scaffold flat failed: %v", err)
	}
	return filepath.Join(dir, name)
}

func TestGuide_GeneratesAllTargets(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldHexagonal(t, tmp, "guide-all")

	chdir(t, svcDir)
	_, err := executeCommand(t, "guide")
	if err != nil {
		t.Fatalf("guide command failed: %v", err)
	}

	for _, p := range []string{
		filepath.Join(".claude", "rules", "archway.md"),
		".cursorrules",
		filepath.Join(".github", "copilot-instructions.md"),
		".windsurfrules",
	} {
		full := filepath.Join(svcDir, p)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", p)
		}
	}
}

func TestGuide_SingleTarget(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldHexagonal(t, tmp, "guide-single")

	// Remove all guide files created by scaffold so we can verify single-target behavior.
	for _, p := range []string{
		filepath.Join(".claude", "rules", "archway.md"),
		".cursorrules",
		".windsurfrules",
		filepath.Join(".github", "copilot-instructions.md"),
	} {
		_ = os.Remove(filepath.Join(svcDir, p))
	}

	chdir(t, svcDir)
	_, err := executeCommand(t, "guide", "--target", "claude")
	if err != nil {
		t.Fatalf("guide --target claude failed: %v", err)
	}

	// Claude target should exist.
	claudePath := filepath.Join(svcDir, ".claude", "rules", "archway.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("expected .claude/rules/archway.md to exist")
	}

	// Other targets should NOT exist.
	for _, p := range []string{".cursorrules", ".windsurfrules", filepath.Join(".github", "copilot-instructions.md")} {
		full := filepath.Join(svcDir, p)
		if _, err := os.Stat(full); err == nil {
			t.Errorf("expected %s to NOT exist for single target", p)
		}
	}
}

func TestGuide_InvalidTarget(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldHexagonal(t, tmp, "guide-invalid")

	chdir(t, svcDir)
	_, err := executeCommand(t, "guide", "--target", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid target")
	}
}

func TestGuide_NoArchwayYAML(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t, "guide")
	if err == nil {
		t.Fatal("expected error when no archway.yaml exists")
	}
	if !strings.Contains(err.Error(), "archway.yaml") {
		t.Fatalf("expected error about archway.yaml, got: %v", err)
	}
}

func TestGuide_ContentContainsArchitecture(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldHexagonal(t, tmp, "guide-content")

	chdir(t, svcDir)
	_, err := executeCommand(t, "guide", "--target", "claude")
	if err != nil {
		t.Fatalf("guide command failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(svcDir, ".claude", "rules", "archway.md"))
	if err != nil {
		t.Fatalf("failed to read archway.md: %v", err)
	}

	content := string(data)
	for _, want := range []string{"hexagonal", "Layer Rules", "Anti-patterns"} {
		if !strings.Contains(content, want) {
			t.Errorf("guide content should contain %q", want)
		}
	}
}

func TestGuide_FlatArchitecture(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldFlat(t, tmp, "guide-flat")

	chdir(t, svcDir)
	_, err := executeCommand(t, "guide", "--target", "claude")
	if err != nil {
		t.Fatalf("guide command failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(svcDir, ".claude", "rules", "archway.md"))
	if err != nil {
		t.Fatalf("failed to read archway.md: %v", err)
	}

	content := string(data)
	for _, want := range []string{"flat", "No layer restrictions"} {
		if !strings.Contains(content, want) {
			t.Errorf("guide content should contain %q", want)
		}
	}
}

// --- Check command tests ---

func TestCheck_CleanHexagonalProject(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldHexagonal(t, tmp, "check-hex")

	_, err := executeCommand(t, "check", "--path", svcDir)
	if err != nil {
		t.Fatalf("check should pass on clean hexagonal project: %v", err)
	}
}

func TestCheck_NoArchwayYAML(t *testing.T) {
	tmp := t.TempDir()

	_, err := executeCommand(t, "check", "--path", tmp)
	if err == nil {
		t.Fatal("expected error when no archway.yaml exists")
	}
	if !strings.Contains(err.Error(), "archway.yaml") {
		t.Fatalf("expected error about archway.yaml, got: %v", err)
	}
}

func TestCheck_FlatProject(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldFlat(t, tmp, "check-flat")

	_, err := executeCommand(t, "check", "--path", svcDir)
	if err != nil {
		t.Fatalf("check should pass on clean flat project: %v", err)
	}
}

// --- Analyze command tests ---

func TestAnalyze_HexagonalProject(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldHexagonal(t, tmp, "analyze-hex")

	_, err := executeCommand(t, "analyze", "--path", svcDir)
	if err != nil {
		t.Fatalf("analyze should succeed on hexagonal project: %v", err)
	}
}

func TestAnalyze_EmptyDir(t *testing.T) {
	tmp := t.TempDir()

	_, err := executeCommand(t, "analyze", "--path", tmp)
	if err == nil {
		t.Fatal("expected error when analyzing empty directory")
	}
}

func TestAnalyze_JsonOutput(t *testing.T) {
	tmp := t.TempDir()
	svcDir := scaffoldHexagonal(t, tmp, "analyze-json")

	// Capture stdout to verify JSON output.
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	_, execErr := executeCommand(t, "analyze", "--path", svcDir, "--output", "json")

	_ = w.Close()
	os.Stdout = oldStdout

	if execErr != nil {
		t.Fatalf("analyze --output json failed: %v", execErr)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// The output may contain a "Detected language:" prefix line before JSON.
	// Find the first '{' to extract the JSON object.
	idx := strings.Index(output, "{")
	if idx < 0 {
		t.Fatalf("expected JSON object in output, got:\n%s", output)
	}
	jsonStr := output[idx:]
	if !json.Valid([]byte(jsonStr)) {
		t.Fatalf("expected valid JSON output, got:\n%s", jsonStr)
	}
}

// --- Init command tests ---

func TestInit_CreatesArchwayYAML(t *testing.T) {
	tmp := t.TempDir()

	// Create a minimal Go project.
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeCommand(t, "init", "--path", tmp, "--no-wizard")
	if err != nil {
		t.Fatalf("init should succeed: %v", err)
	}

	archwayPath := filepath.Join(tmp, "archway.yaml")
	if _, err := os.Stat(archwayPath); os.IsNotExist(err) {
		t.Fatal("expected archway.yaml to be created")
	}

	data, err := os.ReadFile(archwayPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "architecture:") {
		t.Errorf("archway.yaml should contain architecture field, got:\n%s", string(data))
	}
}

func TestInit_WithPreset(t *testing.T) {
	tmp := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeCommand(t, "init", "--path", tmp, "--no-wizard", "--preset", "archway/go-hexagonal-strict")
	if err != nil {
		t.Fatalf("init with preset should succeed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "archway.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "archway/go-hexagonal-strict") {
		t.Errorf("archway.yaml should contain preset, got:\n%s", string(data))
	}
}

func TestInit_ExistingArchwayYAML(t *testing.T) {
	tmp := t.TempDir()

	// Create an existing archway.yaml.
	archwayPath := filepath.Join(tmp, "archway.yaml")
	if err := os.WriteFile(archwayPath, []byte("architecture: flat\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeCommand(t, "init", "--path", tmp, "--no-wizard")
	if err == nil {
		t.Fatal("expected error when archway.yaml already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected 'already exists' error, got: %v", err)
	}

	// With --force, it should succeed.
	_, err = executeCommand(t, "init", "--path", tmp, "--no-wizard", "--force")
	if err != nil {
		t.Fatalf("init --force should succeed: %v", err)
	}
}

// --- Root command structure ---

func TestRootCommand_HasExpectedSubcommands(t *testing.T) {
	cmd := newRootCommand()

	expectedNames := []string{"new", "init", "analyze", "check", "guide", "version"}
	subNames := make([]string, 0, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subNames = append(subNames, sub.Name())
	}

	for _, name := range expectedNames {
		found := false
		for _, sub := range subNames {
			if sub == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found; have: %v", name, subNames)
		}
	}
}

func TestRootCommand_ValidOutputFormats(t *testing.T) {
	for _, format := range []string{"terminal", "json", "markdown"} {
		_, err := executeCommand(t, "--output", format, "version")
		if err != nil {
			t.Errorf("output format %q should be valid, got: %v", format, err)
		}
	}
}

func TestRootCommand_PersistentFlags(t *testing.T) {
	cmd := newRootCommand()

	noColor := cmd.PersistentFlags().Lookup("no-color")
	if noColor == nil {
		t.Error("expected --no-color persistent flag")
	}

	output := cmd.PersistentFlags().Lookup("output")
	if output == nil {
		t.Error("expected --output persistent flag")
	}
	if output.DefValue != "terminal" {
		t.Errorf("expected --output default 'terminal', got %q", output.DefValue)
	}
}

// --- listArchitectures / listCapabilities ---

func TestListArchitectures_ReturnsEntries(t *testing.T) {
	p, err := provider.Get("go")
	if err != nil {
		t.Fatalf("failed to get go provider: %v", err)
	}
	tFS := p.GetTemplateFS()

	archs, err := listArchitectures(tFS)
	if err != nil {
		t.Fatalf("listArchitectures failed: %v", err)
	}

	if len(archs) == 0 {
		t.Fatal("expected at least one architecture")
	}

	// Verify each entry has required fields.
	for _, a := range archs {
		if a.name == "" {
			t.Error("architecture entry has empty name")
		}
		if a.label == "" {
			t.Error("architecture entry has empty label")
		}
	}

	// Verify known architectures exist.
	names := make(map[string]bool)
	for _, a := range archs {
		names[a.name] = true
	}
	for _, expected := range []string{"hexagonal", "flat"} {
		if !names[expected] {
			t.Errorf("expected architecture %q in list", expected)
		}
	}
}

func TestListCapabilities_ReturnsEntries(t *testing.T) {
	p, err := provider.Get("go")
	if err != nil {
		t.Fatalf("failed to get go provider: %v", err)
	}
	tFS := p.GetTemplateFS()

	caps, err := listCapabilities(tFS)
	if err != nil {
		t.Fatalf("listCapabilities failed: %v", err)
	}

	if len(caps) == 0 {
		t.Fatal("expected at least one capability")
	}

	for _, c := range caps {
		if c.name == "" {
			t.Error("capability entry has empty name")
		}
		if c.description == "" {
			t.Errorf("capability %q has empty description", c.name)
		}
	}

	// Verify known capabilities exist.
	names := make(map[string]bool)
	for _, c := range caps {
		names[c.name] = true
	}
	if !names["http-api"] {
		t.Error("expected capability 'http-api' in list")
	}
}

func TestListArchitectures_InvalidFS(t *testing.T) {
	// An empty FS should return an error.
	emptyFS := os.DirFS(t.TempDir())

	_, err := listArchitectures(emptyFS)
	if err == nil {
		t.Fatal("expected error for FS missing templates/architectures")
	}
}

func TestListCapabilities_InvalidFS(t *testing.T) {
	emptyFS := os.DirFS(t.TempDir())

	_, err := listCapabilities(emptyFS)
	if err == nil {
		t.Fatal("expected error for FS missing templates/capabilities")
	}
}

// --- printEquivalentCommand ---

func TestPrintEquivalentCommand_AllFields(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &newCommandOptions{
		Name:         "my-svc",
		Architecture: "hexagonal",
		Capabilities: "http-api,mysql",
		ModulePath:   "github.com/org/my-svc",
	}
	printEquivalentCommand(opts)

	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "--arch hexagonal") {
		t.Errorf("expected --arch flag, got: %s", out)
	}
	if !strings.Contains(out, "--cap http-api,mysql") {
		t.Errorf("expected --cap flag, got: %s", out)
	}
	if !strings.Contains(out, "--module github.com/org/my-svc") {
		t.Errorf("expected --module flag, got: %s", out)
	}
	if !strings.Contains(out, "--no-wizard") {
		t.Errorf("expected --no-wizard flag, got: %s", out)
	}
}

func TestPrintEquivalentCommand_TemplateInsteadOfArch(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &newCommandOptions{
		Name:     "my-svc",
		Template: "custom",
	}
	printEquivalentCommand(opts)

	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "--template custom") {
		t.Errorf("expected --template flag, got: %s", out)
	}
	if strings.Contains(out, "--arch") {
		t.Errorf("should not contain --arch when using template, got: %s", out)
	}
}

func TestPrintEquivalentCommand_ApiTemplateOmitted(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &newCommandOptions{
		Name:     "my-svc",
		Template: "api",
	}
	printEquivalentCommand(opts)

	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	// "api" is the default template, so --template should not appear.
	if strings.Contains(out, "--template") {
		t.Errorf("should not contain --template for default api template, got: %s", out)
	}
}

// --- New command path parsing ---

func TestNewCommand_PathArgSetsOutputDirAndName(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new", filepath.Join(tmp, "my-svc"),
		"--arch", "flat",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("new with path arg failed: %v", err)
	}

	svcDir := filepath.Join(tmp, "my-svc")
	if _, err := os.Stat(svcDir); os.IsNotExist(err) {
		t.Fatal("expected service directory to exist")
	}
}

func TestNewCommand_InvalidSetFlag(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new", "test-bad-set",
		"--arch", "flat",
		"--no-wizard",
		"--set", "invalid-no-equals",
	)
	if err == nil {
		t.Fatal("expected error for invalid --set value")
	}
	if !strings.Contains(err.Error(), "invalid --set") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- New command with --module flag ---

func TestNewCommand_CustomModule(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := executeCommand(t,
		"new", "mod-svc",
		"--arch", "flat",
		"--module", "github.com/custom/mod-svc",
		"--no-wizard",
		"--set", "skip_hooks=true",
	)
	if err != nil {
		t.Fatalf("new with --module failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "mod-svc", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "github.com/custom/mod-svc") {
		t.Errorf("go.mod should contain custom module path, got:\n%s", string(data))
	}
}
