package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunGrep_PatternMatch(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", id)
	fmt.Println(query)
}
`,
	})

	rule := Rule{
		ID:          "test-sql",
		Engine:      "grep",
		Description: "No SQL string formatting",
		Severity:    "error",
		Pattern:     `fmt\.Sprintf.*SELECT`,
		Scope:       []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Len(t, violations, 1)
	assert.Equal(t, "test-sql", violations[0].RuleID)
	assert.Equal(t, "main.go", violations[0].File)
	assert.Equal(t, 6, violations[0].Line)
}

func TestRunGrep_NoMatch(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

func main() {
	println("hello")
}
`,
	})

	rule := Rule{
		ID:       "test-sql",
		Engine:   "grep",
		Severity: "error",
		Pattern:  `fmt\.Sprintf.*SELECT`,
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestRunGrep_MustContain(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"repo.go": `package repo

func GetUser(id int) {
	db.Query("SELECT * FROM users WHERE id = ?", id)
}

func GetOrder(id int) {
	db.Query("SELECT * FROM orders WHERE id = ?", id)
}
`,
	})

	// Rule: SQL SELECT must contain tenant_id.
	rule := Rule{
		ID:          "tenant-check",
		Engine:      "grep",
		Severity:    "error",
		Pattern:     `SELECT.*FROM`,
		MustContain: `tenant_id`,
		Scope:       []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Len(t, violations, 2) // both queries lack tenant_id
}

func TestRunGrep_MustNotContain(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"handler.go": `package handler

func Handle() {
	log.Info("processing request", "user", user.Email)
	log.Error("failed to process", "password", user.Password)
}
`,
	})

	// Rule: log statements must not contain sensitive data.
	rule := Rule{
		ID:             "no-sensitive-logs",
		Engine:         "grep",
		Severity:       "error",
		Pattern:        `log\.(Info|Error|Warn)`,
		MustNotContain: `password|secret|token`,
		Scope:          []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Len(t, violations, 1) // only the password line
	assert.Equal(t, 5, violations[0].Line)
}

func TestRunGrep_FileMustContain(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"handler.go": `package handler

func HandleOrder() {
	// does something without checking permissions
}
`,
		"secure_handler.go": `package handler

import "auth"

func HandleSecure() {
	auth.Check(ctx)
}
`,
	})

	// Rule: handler files must reference auth.
	rule := Rule{
		ID:              "auth-required",
		Engine:          "grep",
		Severity:        "warning",
		FileMustContain: `auth`,
		Scope:           []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Equal(t, "handler.go", violations[0].File)
	assert.Equal(t, 0, violations[0].Line) // file-level violation
}

func TestRunGrep_ExcludePattern(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main
var cache = map[string]int{}
`,
		"main_test.go": `package main
var cache = map[string]int{}
`,
	})

	rule := Rule{
		ID:       "no-global-map",
		Engine:   "grep",
		Severity: "error",
		Pattern:  `var.*=.*map\[`,
		Scope:    []string{"**/*.go"},
		Exclude:  []string{"*_test.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Len(t, violations, 1)
	assert.Equal(t, "main.go", violations[0].File)
}

func TestRunGrep_BinaryFileSkipped(t *testing.T) {
	dir := t.TempDir()
	// Create a binary file.
	err := os.WriteFile(filepath.Join(dir, "binary.go"), []byte("package main\x00\x00\x00"), 0o644)
	require.NoError(t, err)

	rule := Rule{
		ID:       "test",
		Engine:   "grep",
		Severity: "error",
		Pattern:  "package",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestRunGrep_AllowedFilesFilter(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"a.go": "package main\nvar bad = map[string]int{}\n",
		"b.go": "package main\nvar bad = map[string]int{}\n",
	})

	rule := Rule{
		ID:       "test",
		Engine:   "grep",
		Severity: "error",
		Pattern:  `var bad`,
		Scope:    []string{"**/*.go"},
	}

	// Only allow a.go (simulates --staged).
	violations, err := RunGrep(rule, dir, []string{"a.go"})
	require.NoError(t, err)
	assert.Len(t, violations, 1)
	assert.Equal(t, "a.go", violations[0].File)
}

func TestRunGrep_PatternWithMustContainAndMustNotContain(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"repo.go": `package repo

func Good() {
	db.Query("SELECT * FROM users WHERE tenant_id = ?", tid)
}

func Bad() {
	db.Query("SELECT * FROM users WHERE id = ?", id)
}

func Worse() {
	db.Query("SELECT password FROM users WHERE tenant_id = ?", tid)
}
`,
	})

	rule := Rule{
		ID:             "tenant-no-secrets",
		Engine:         "grep",
		Severity:       "error",
		Pattern:        `SELECT.*FROM`,
		MustContain:    `tenant_id`,
		MustNotContain: `password`,
		Scope:          []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	// Line 4: has tenant_id, no password → OK
	// Line 8: no tenant_id → violation (must-contain fails)
	// Line 12: has tenant_id but also password → violation (must-not-contain fails)
	assert.Len(t, violations, 2)
}

func TestRunGrep_EmptyFile(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"empty.go": "",
	})

	rule := Rule{
		ID:       "test",
		Engine:   "grep",
		Severity: "error",
		Pattern:  "package",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestRunGrep_FileWithOnlyNewlines(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"newlines.go": "\n\n\n\n\n",
	})

	rule := Rule{
		ID:       "test",
		Engine:   "grep",
		Severity: "error",
		Pattern:  `\S`,
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestRunGrep_BinaryDetectionAtBoundary(t *testing.T) {
	dir := t.TempDir()
	// Create a file with a null byte at position 511 (within the 512-byte check window).
	data := make([]byte, 512)
	for i := range data {
		data[i] = 'a'
	}
	data[511] = 0
	err := os.WriteFile(filepath.Join(dir, "boundary.go"), data, 0o644)
	require.NoError(t, err)

	rule := Rule{
		ID:       "test",
		Engine:   "grep",
		Severity: "error",
		Pattern:  "a",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations, "file with null byte at position 511 should be detected as binary")
}

func TestRunGrep_EmptyScope(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": "package main\n",
	})

	rule := Rule{
		ID:       "test",
		Engine:   "grep",
		Severity: "error",
		Pattern:  "package",
		Scope:    []string{},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestRunGrep_FileMustContainMissing(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"handler.go": "package handler\n\nfunc Handle() {}\n",
	})

	rule := Rule{
		ID:              "require-license",
		Engine:          "grep",
		Severity:        "error",
		FileMustContain: `Copyright`,
		Scope:           []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Equal(t, "handler.go", violations[0].File)
	assert.Equal(t, 0, violations[0].Line)
	assert.Contains(t, violations[0].Match, "Copyright")
}

func TestRunGrep_FileMustContainPresent(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"handler.go": "// Copyright 2026\npackage handler\n\nfunc Handle() {}\n",
	})

	rule := Rule{
		ID:              "require-license",
		Engine:          "grep",
		Severity:        "error",
		FileMustContain: `Copyright`,
		Scope:           []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestRunGrep_MustNotContainSuppresses(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"handler.go": `package handler

func Handle() {
	log.Info("safe message", "key", "value")
}
`,
	})

	// Pattern matches log.Info, must-not-contain is "password".
	// Since the line does NOT contain "password", no violation should be produced.
	rule := Rule{
		ID:             "no-sensitive-logs",
		Engine:         "grep",
		Severity:       "error",
		Pattern:        `log\.Info`,
		MustNotContain: `password`,
		Scope:          []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations, "line matching pattern but NOT matching must-not-contain should not violate")
}

func TestRunGrep_PatternMustContainCombo_MustContainAbsent(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"repo.go": `package repo

func Get() {
	db.Query("SELECT * FROM users WHERE id = ?", id)
}
`,
	})

	rule := Rule{
		ID:          "tenant-required",
		Engine:      "grep",
		Severity:    "error",
		Pattern:     `SELECT.*FROM`,
		MustContain: `tenant_id`,
		Scope:       []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Equal(t, 4, violations[0].Line)
}

func TestRunGrep_LongLineTruncation(t *testing.T) {
	longLine := "var x = " + string(make([]byte, 200)) // will have null bytes — use a real string
	// Build a line > 120 chars of valid text.
	longContent := "package main\n" + "var x = \"" + repeatChar('a', 200) + "\"\n"
	dir := setupTestProject(t, map[string]string{
		"long.go": longContent,
	})

	_ = longLine // unused

	rule := Rule{
		ID:       "test",
		Engine:   "grep",
		Severity: "error",
		Pattern:  `var x`,
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunGrep(rule, dir, nil)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.True(t, len(violations[0].Match) <= 123, "match should be truncated to 120 chars + '...'")
	assert.True(t, len(violations[0].Match) > 120, "match should end with '...'")
}

func repeatChar(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}

// --- helpers ---

func setupTestProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		err := os.MkdirAll(filepath.Dir(path), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(path, []byte(content), 0o644)
		require.NoError(t, err)
	}
	return dir
}
