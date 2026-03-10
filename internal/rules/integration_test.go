package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRulesDir(t *testing.T, projectDir string) string {
	t.Helper()
	rulesDir := filepath.Join(projectDir, ".archway", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))
	return rulesDir
}

func TestFullPipeline(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", id)
	fmt.Println(query)
}
`,
	})

	rulesDir := setupRulesDir(t, dir)
	writeRule(t, rulesDir, "no-sql-sprintf.yaml", `
id: no-sql-sprintf
engine: grep
description: "Do not use fmt.Sprintf for SQL queries"
severity: error
pattern: 'fmt\.Sprintf.*SELECT'
scope:
  - "**/*.go"
`)

	result, err := RunRules(rulesDir, dir, nil)
	require.NoError(t, err)

	assert.Equal(t, 1, result.ValidRuleCount(), "expected 1 valid rule")
	assert.Equal(t, 0, result.InvalidRuleCount())
	assert.Equal(t, 0, result.StaleRuleCount())
	require.Len(t, result.Violations, 1)
	assert.Equal(t, "no-sql-sprintf", result.Violations[0].RuleID)
	assert.Equal(t, "grep", result.Violations[0].Engine)
	assert.Equal(t, "main.go", result.Violations[0].File)
	assert.Equal(t, 6, result.Violations[0].Line)
	assert.Equal(t, 1, result.ErrorCount())
	assert.Greater(t, result.Duration.Nanoseconds(), int64(0))
}

func TestFullPipelineWithAST(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"state.go": `package main

var GlobalCache = map[string]string{}

func main() {}
`,
	})

	rulesDir := setupRulesDir(t, dir)
	writeRule(t, rulesDir, "no-global-mutable.yaml", `
id: no-global-mutable
engine: ast
description: "No global mutable state"
severity: error
detector: global-mutable-state
scope:
  - "**/*.go"
`)

	result, err := RunRules(rulesDir, dir, nil)
	require.NoError(t, err)

	assert.Equal(t, 1, result.ValidRuleCount())
	require.NotEmpty(t, result.Violations, "expected at least one AST violation for global mutable state")
	assert.Equal(t, "no-global-mutable", result.Violations[0].RuleID)
	assert.Equal(t, "ast", result.Violations[0].Engine)
	assert.Equal(t, "state.go", result.Violations[0].File)
}

func TestMixedEngines(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"handler.go": `package main

import "fmt"

var GlobalCache = map[string]string{}

func handle() {
	query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", id)
	fmt.Println(query)
}
`,
	})

	rulesDir := setupRulesDir(t, dir)
	writeRule(t, rulesDir, "no-sql-sprintf.yaml", `
id: no-sql-sprintf
engine: grep
description: "No SQL string formatting"
severity: error
pattern: 'fmt\.Sprintf.*SELECT'
scope:
  - "**/*.go"
`)

	writeRule(t, rulesDir, "no-global-mutable.yaml", `
id: no-global-mutable
engine: ast
description: "No global mutable state"
severity: warning
detector: global-mutable-state
scope:
  - "**/*.go"
`)

	result, err := RunRules(rulesDir, dir, nil)
	require.NoError(t, err)

	assert.Equal(t, 2, result.ValidRuleCount())
	assert.GreaterOrEqual(t, len(result.Violations), 2, "expected violations from both engines")

	// Verify both engines contributed violations.
	engines := map[string]bool{}
	for _, v := range result.Violations {
		engines[v.Engine] = true
	}
	assert.True(t, engines["grep"], "expected grep violations")
	assert.True(t, engines["ast"], "expected ast violations")

	// Verify severity counts.
	assert.GreaterOrEqual(t, result.ErrorCount(), 1, "expected at least 1 error from grep rule")
	assert.GreaterOrEqual(t, result.WarningCount(), 1, "expected at least 1 warning from ast rule")
}

func TestFullPipelineNoViolations(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

func main() {
	println("hello")
}
`,
	})

	rulesDir := setupRulesDir(t, dir)
	writeRule(t, rulesDir, "no-sql-sprintf.yaml", `
id: no-sql-sprintf
engine: grep
description: "No SQL string formatting"
severity: error
pattern: 'fmt\.Sprintf.*SELECT'
scope:
  - "**/*.go"
`)

	result, err := RunRules(rulesDir, dir, nil)
	require.NoError(t, err)

	assert.Equal(t, 1, result.ValidRuleCount())
	assert.Empty(t, result.Violations)
	assert.Equal(t, 0, result.ErrorCount())
	assert.Equal(t, 0, result.WarningCount())
}

func TestRunRulesNonexistentDir(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".archway", "rules")

	result, err := RunRules(rulesDir, dir, nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Violations)
	assert.Empty(t, result.Statuses)
}
