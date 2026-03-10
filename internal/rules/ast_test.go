package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunAST_UnknownDetector(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": "package main\n",
	})

	rule := Rule{
		ID:       "test-unknown",
		Engine:   "ast",
		Detector: "nonexistent-detector",
		Severity: "error",
		Scope:    []string{"**/*.go"},
	}

	_, err := RunAST(rule, dir, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown detector")
}

func TestRunAST_GlobalMutableState(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"state.go": `package main

var cache = make(map[string]string)
var count int
`,
	})

	rule := Rule{
		ID:       "no-global-state",
		Engine:   "ast",
		Detector: "global-mutable-state",
		Severity: "warning",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunAST(rule, dir, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(violations), 1)
	assert.Equal(t, "no-global-state", violations[0].RuleID)
	assert.Equal(t, "ast", violations[0].Engine)
	assert.Equal(t, "warning", violations[0].Severity)
}

func TestRunAST_InitSideEffects(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"setup.go": `package main

import "os"

func init() {
	data, _ := os.ReadFile("config.yaml")
	_ = data
}
`,
	})

	rule := Rule{
		ID:       "no-init-side-effects",
		Engine:   "ast",
		Detector: "init-side-effects",
		Severity: "warning",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunAST(rule, dir, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(violations), 1)
}

func TestRunAST_NakedGoroutine(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"async.go": `package main

func doWork() {
	go func() {
		println("fire and forget")
	}()
}
`,
	})

	rule := Rule{
		ID:       "no-naked-goroutines",
		Engine:   "ast",
		Detector: "naked-goroutine",
		Severity: "warning",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunAST(rule, dir, nil)
	require.NoError(t, err)
	assert.Len(t, violations, 1)
	assert.Equal(t, "async.go", violations[0].File)
}

func TestRunAST_SwallowedError(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"errors.go": `package main

import "os"

func readFile() {
	_, err := os.Open("test.txt")
	if err != nil {
	}
}
`,
	})

	rule := Rule{
		ID:       "no-swallowed-errors",
		Engine:   "ast",
		Detector: "swallowed-error",
		Severity: "error",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunAST(rule, dir, nil)
	require.NoError(t, err)
	assert.Len(t, violations, 1)
}

func TestRunAST_SQLConcatenation(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"db.go": `package main

func query(id string) string {
	return "SELECT * FROM users WHERE id = " + id
}
`,
	})

	rule := Rule{
		ID:       "no-sql-concat",
		Engine:   "ast",
		Detector: "sql-concatenation",
		Severity: "error",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunAST(rule, dir, nil)
	require.NoError(t, err)
	assert.Len(t, violations, 1)
}

func TestRunAST_AllowedFilesFiltering(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"a.go": `package main

var cache = make(map[string]string)
`,
		"b.go": `package main

var store = make(map[string]int)
`,
	})

	rule := Rule{
		ID:       "no-global-state",
		Engine:   "ast",
		Detector: "global-mutable-state",
		Severity: "warning",
		Scope:    []string{"**/*.go"},
	}

	// Only check a.go.
	violations, err := RunAST(rule, dir, []string{"a.go"})
	require.NoError(t, err)
	assert.Len(t, violations, 1)
	assert.Equal(t, "a.go", violations[0].File)
}

func TestRunAST_NoViolations(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"clean.go": `package main

func main() {
	println("hello")
}
`,
	})

	rule := Rule{
		ID:       "no-naked-goroutines",
		Engine:   "ast",
		Detector: "naked-goroutine",
		Severity: "warning",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunAST(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestRunAST_SkipsTestFiles(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main_test.go": `package main

var cache = make(map[string]string)
`,
	})

	rule := Rule{
		ID:       "no-global-state",
		Engine:   "ast",
		Detector: "global-mutable-state",
		Severity: "warning",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunAST(rule, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestRunAST_InitAbuse(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"init.go": `package main

func init() {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
	_ = a + b + c + d + e + f
}
`,
	})

	rule := Rule{
		ID:       "no-init-abuse",
		Engine:   "ast",
		Detector: "init-abuse",
		Severity: "warning",
		Scope:    []string{"**/*.go"},
	}

	violations, err := RunAST(rule, dir, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(violations), 1)
}
