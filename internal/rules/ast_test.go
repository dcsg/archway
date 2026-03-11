package rules

import (
	"fmt"
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

func TestRunAST_ContextBackground(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantCount int
	}{
		{
			name: "detects context.Background",
			code: `package main

import "context"

func do() {
	ctx := context.Background()
	_ = ctx
}
`,
			wantCount: 1,
		},
		{
			name: "ignores context.TODO",
			code: `package main

import "context"

func do() {
	ctx := context.TODO()
	_ = ctx
}
`,
			wantCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupTestProject(t, map[string]string{"main.go": tt.code})
			rule := Rule{
				ID:       "ctx-bg",
				Engine:   "ast",
				Detector: "context-background-in-handler",
				Severity: "warning",
				Scope:    []string{"**/*.go"},
			}
			violations, err := RunAST(rule, dir, nil)
			require.NoError(t, err)
			assert.Len(t, violations, tt.wantCount)
		})
	}
}

func TestRunAST_UUIDv4(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		code      string
		wantCount int
	}{
		{
			name: "detects uuid.New",
			file: "id.go",
			code: `package main

import "github.com/google/uuid"

func genID() string {
	return uuid.NewString()
}
`,
			wantCount: 1,
		},
		{
			name: "skips requestid file",
			file: "requestid.go",
			code: `package main

import "github.com/google/uuid"

func reqID() string {
	return uuid.NewString()
}
`,
			wantCount: 0,
		},
		{
			name: "no uuid calls",
			file: "clean.go",
			code: `package main

func clean() string {
	return "hello"
}
`,
			wantCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupTestProject(t, map[string]string{tt.file: tt.code})
			rule := Rule{
				ID:       "uuid-v4",
				Engine:   "ast",
				Detector: "uuid-v4-as-key",
				Severity: "warning",
				Scope:    []string{"**/*.go"},
			}
			violations, err := RunAST(rule, dir, nil)
			require.NoError(t, err)
			assert.Len(t, violations, tt.wantCount)
		})
	}
}

func TestRunAST_FatHandler(t *testing.T) {
	// Build a fat handler with >40 statements.
	stmts := ""
	for i := range 45 {
		stmts += fmt.Sprintf("\t_ = %d\n", i)
	}

	tests := []struct {
		name      string
		code      string
		wantCount int
	}{
		{
			name: "fat handler detected",
			code: `package main

import "net/http"

func handleUsers(w http.ResponseWriter, r *http.Request) {
` + stmts + `}
`,
			wantCount: 1,
		},
		{
			name: "thin handler ok",
			code: `package main

import "net/http"

func handleUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
`,
			wantCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupTestProject(t, map[string]string{"handler.go": tt.code})
			rule := Rule{
				ID:       "fat-handler",
				Engine:   "ast",
				Detector: "fat-handler",
				Severity: "warning",
				Scope:    []string{"**/*.go"},
			}
			violations, err := RunAST(rule, dir, nil)
			require.NoError(t, err)
			assert.Len(t, violations, tt.wantCount)
		})
	}
}

func TestIsKnownDetector(t *testing.T) {
	tests := []struct {
		name string
		det  string
		want bool
	}{
		{"known - global-mutable-state", "global-mutable-state", true},
		{"known - fat-handler", "fat-handler", true},
		{"known - god-package", "god-package", true},
		{"unknown", "nonexistent", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsKnownDetector(tt.det))
		})
	}
}
