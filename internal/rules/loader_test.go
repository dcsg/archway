package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRules_ValidRule(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "valid.yaml", `
id: test-rule
engine: grep
description: "Test rule"
severity: error
pattern: "badPattern"
scope:
  - "**/*.go"
`)
	// Create a .go file so scope is not stale.
	writeFile(t, dir, "main.go", "package main\n")

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Len(t, statuses, 1)
	assert.Equal(t, "valid", statuses[0].Status)
	assert.Equal(t, "test-rule", rules[0].ID)
}

func TestLoadRules_InvalidNoEngine(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "bad.yaml", `
id: bad-rule
description: "Missing engine"
severity: error
pattern: "something"
scope:
  - "**/*.go"
`)

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Empty(t, rules)
	assert.Len(t, statuses, 1)
	assert.Equal(t, "invalid", statuses[0].Status)
	assert.Contains(t, statuses[0].Error, "unknown engine")
}

func TestLoadRules_InvalidNoScope(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "bad.yaml", `
id: no-scope
engine: grep
description: "Missing scope"
severity: error
pattern: "something"
`)

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Empty(t, rules)
	assert.Len(t, statuses, 1)
	assert.Equal(t, "invalid", statuses[0].Status)
	assert.Contains(t, statuses[0].Error, "scope")
}

func TestLoadRules_InvalidRegex(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "bad-regex.yaml", `
id: bad-regex
engine: grep
description: "Invalid regex"
severity: error
pattern: "[invalid"
scope:
  - "**/*.go"
`)

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Empty(t, rules)
	assert.Equal(t, "invalid", statuses[0].Status)
	assert.Contains(t, statuses[0].Error, "pattern regex")
}

func TestLoadRules_StaleScope(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "stale.yaml", `
id: stale-rule
engine: grep
description: "Stale rule"
severity: error
pattern: "something"
scope:
  - "nonexistent/**/*.go"
`)

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Len(t, rules, 1) // stale rules are still returned
	assert.Equal(t, "stale", statuses[0].Status)
	assert.Contains(t, statuses[0].Error, "0 files")
}

func TestLoadRules_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "malformed.yaml", `
id: [this is not valid yaml
  broken: {{}
`)

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Empty(t, rules)
	assert.Equal(t, "invalid", statuses[0].Status)
	assert.Contains(t, statuses[0].Error, "invalid YAML")
}

func TestLoadRules_ASTRule(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "ast.yaml", `
id: ast-rule
engine: ast
description: "AST rule"
severity: error
detector: global-mutable-state
scope:
  - "**/*.go"
`)
	writeFile(t, dir, "main.go", "package main\n")

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Equal(t, "valid", statuses[0].Status)
	assert.Equal(t, "ast", rules[0].Engine)
}

func TestLoadRules_ASTMissingDetector(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "bad-ast.yaml", `
id: bad-ast
engine: ast
description: "Missing detector"
severity: error
scope:
  - "**/*.go"
`)

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Empty(t, rules)
	assert.Equal(t, "invalid", statuses[0].Status)
	assert.Contains(t, statuses[0].Error, "detector")
}

func TestLoadRules_SkipsNonYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Not a rule\n")
	writeFile(t, dir, "notes.txt", "just notes\n")

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Empty(t, rules)
	assert.Empty(t, statuses)
}

func TestLoadRules_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Empty(t, rules)
	assert.Empty(t, statuses)
}

func TestValidateRule_DefaultSeverity(t *testing.T) {
	r := Rule{
		ID:      "test",
		Engine:  "grep",
		Pattern: "foo",
		Scope:   []string{"**/*.go"},
	}
	status := ValidateRule(r, "test.yaml", "")
	assert.Equal(t, "valid", status.Status)
}

func TestLoadRules_NumericID(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "numeric.yaml", `
id: 123
engine: grep
description: "Numeric ID rule"
severity: error
pattern: "badPattern"
scope:
  - "**/*.go"
`)
	writeFile(t, dir, "main.go", "package main\n")

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Equal(t, "valid", statuses[0].Status)
	assert.Equal(t, "123", rules[0].ID)
}

func TestValidateRule_EmptyScope(t *testing.T) {
	r := Rule{
		ID:      "test",
		Engine:  "grep",
		Pattern: "foo",
		Scope:   []string{},
	}
	status := ValidateRule(r, "test.yaml", "")
	assert.Equal(t, "invalid", status.Status)
	assert.Contains(t, status.Error, "scope")
}

func TestValidateRule_PatternAndFileMustContain(t *testing.T) {
	r := Rule{
		ID:              "test",
		Engine:          "grep",
		Pattern:         `SELECT`,
		FileMustContain: `package`,
		Scope:           []string{"**/*.go"},
	}
	status := ValidateRule(r, "test.yaml", "")
	assert.Equal(t, "valid", status.Status)
}

func TestValidateRule_ASTValidDetectorInvalidSeverity(t *testing.T) {
	r := Rule{
		ID:       "test",
		Engine:   "ast",
		Detector: "global-mutable-state",
		Severity: "critical", // invalid
		Scope:    []string{"**/*.go"},
	}
	status := ValidateRule(r, "test.yaml", "")
	assert.Equal(t, "invalid", status.Status)
	assert.Contains(t, status.Error, "severity")
}

func TestLoadRules_DirectoryWithNoYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "readme.txt", "not a yaml file\n")
	writeFile(t, dir, "config.json", "{}\n")

	rules, statuses, err := LoadRules(dir, dir)
	require.NoError(t, err)
	assert.Empty(t, rules)
	assert.Empty(t, statuses)
}

func TestValidateRule_EmptyProjectRootSkipsStaleCheck(t *testing.T) {
	r := Rule{
		ID:      "test",
		Engine:  "grep",
		Pattern: "foo",
		Scope:   []string{"nonexistent/**/*.go"},
	}
	// With empty projectRoot, stale check is skipped → should be valid.
	status := ValidateRule(r, "test.yaml", "")
	assert.Equal(t, "valid", status.Status)
}

// --- helpers ---

func writeRule(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	require.NoError(t, err)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
}
