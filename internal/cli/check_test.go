package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/dcsg/archway/internal/checker"
	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout redirects os.Stdout, runs fn, and returns what was printed.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	require.NoError(t, err)
	return buf.String()
}

// --- filterRuleResult ---

func TestCheckFilterRuleResult_MatchingViolationsAndStatuses(t *testing.T) {
	result := &rules.RunResult{
		Duration: 100 * time.Millisecond,
		Violations: []rules.RuleViolation{
			{RuleID: "rule-a", File: "a.go", Severity: "error"},
			{RuleID: "rule-b", File: "b.go", Severity: "warning"},
			{RuleID: "rule-a", File: "c.go", Severity: "warning"},
		},
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "rule-a"}, Filename: "rule-a.yaml", Status: "valid"},
			{Rule: rules.Rule{ID: "rule-b"}, Filename: "rule-b.yaml", Status: "valid"},
		},
	}

	filtered := filterRuleResult(result, "rule-a")

	assert.Len(t, filtered.Violations, 2)
	assert.Len(t, filtered.Statuses, 1)
	assert.Equal(t, result.Duration, filtered.Duration)
	for _, v := range filtered.Violations {
		assert.Equal(t, "rule-a", v.RuleID)
	}
	assert.Equal(t, "rule-a", filtered.Statuses[0].Rule.ID)
}

func TestCheckFilterRuleResult_NoMatch(t *testing.T) {
	result := &rules.RunResult{
		Duration: 50 * time.Millisecond,
		Violations: []rules.RuleViolation{
			{RuleID: "rule-a", File: "a.go"},
		},
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "rule-a"}, Status: "valid"},
		},
	}

	filtered := filterRuleResult(result, "nonexistent")

	assert.Empty(t, filtered.Violations)
	assert.Empty(t, filtered.Statuses)
	assert.Equal(t, result.Duration, filtered.Duration)
}

// --- printProxyRuleSection ---

func TestCheckPrintProxyRuleSection_AllPass(t *testing.T) {
	result := &rules.RunResult{
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "r1"}, Status: "valid"},
			{Rule: rules.Rule{ID: "r2"}, Status: "valid"},
		},
	}

	out := captureStdout(t, func() { printProxyRuleSection(result) })

	assert.Contains(t, out, "PROXY RULES (2 valid, 0 invalid, 0 stale)")
	assert.Contains(t, out, "All proxy rules pass")
}

func TestCheckPrintProxyRuleSection_ViolationsWithLineNumbers(t *testing.T) {
	result := &rules.RunResult{
		Violations: []rules.RuleViolation{
			{RuleID: "cap-sql", Severity: "error", File: "db.go", Line: 42, Description: "raw SQL detected"},
		},
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "cap-sql"}, Status: "valid"},
		},
	}

	out := captureStdout(t, func() { printProxyRuleSection(result) })

	assert.Contains(t, out, "1 errors, 0 warnings")
	assert.Contains(t, out, "db.go:42")
	assert.Contains(t, out, "raw SQL detected")
}

func TestCheckPrintProxyRuleSection_ViolationsWithoutLineNumbers(t *testing.T) {
	result := &rules.RunResult{
		Violations: []rules.RuleViolation{
			{RuleID: "cap-no-utils", Severity: "warning", File: "utils.go", Line: 0, Description: "avoid utils package"},
		},
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "cap-no-utils"}, Status: "valid"},
		},
	}

	out := captureStdout(t, func() { printProxyRuleSection(result) })

	assert.Contains(t, out, "0 errors, 1 warnings")
	// Without line number, uses " — " separator.
	assert.Contains(t, out, "utils.go")
	assert.Contains(t, out, "avoid utils package")
	assert.NotContains(t, out, "utils.go:0")
}

func TestCheckPrintProxyRuleSection_ViolationWithMatch(t *testing.T) {
	result := &rules.RunResult{
		Violations: []rules.RuleViolation{
			{RuleID: "r1", Severity: "error", File: "x.go", Line: 10, Description: "bad", Match: "fmt.Println(secret)"},
		},
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "r1"}, Status: "valid"},
		},
	}

	out := captureStdout(t, func() { printProxyRuleSection(result) })

	assert.Contains(t, out, "> fmt.Println(secret)")
}

func TestCheckPrintProxyRuleSection_InvalidAndStaleRules(t *testing.T) {
	result := &rules.RunResult{
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "r1"}, Filename: "r1.yaml", Status: "valid"},
			{Rule: rules.Rule{ID: "r2"}, Filename: "r2.yaml", Status: "invalid", Error: "missing id field"},
			{Rule: rules.Rule{ID: "r3"}, Filename: "r3.yaml", Status: "stale", Error: "scope glob matches no files"},
		},
	}

	out := captureStdout(t, func() { printProxyRuleSection(result) })

	assert.Contains(t, out, "1 valid, 1 invalid, 1 stale")
	assert.Contains(t, out, "All proxy rules pass")
	// Note: invalid/stale status details are only printed when there are violations,
	// because the function returns early after "All proxy rules pass".
	assert.NotContains(t, out, "r2.yaml")
	assert.NotContains(t, out, "r3.yaml")
}

func TestCheckPrintProxyRuleSection_InvalidStaleWithViolations(t *testing.T) {
	result := &rules.RunResult{
		Violations: []rules.RuleViolation{
			{RuleID: "r1", Severity: "error", File: "x.go", Line: 1, Description: "bad"},
		},
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "r1"}, Filename: "r1.yaml", Status: "valid"},
			{Rule: rules.Rule{ID: "r2"}, Filename: "r2.yaml", Status: "invalid", Error: "missing id"},
			{Rule: rules.Rule{ID: "r3"}, Filename: "r3.yaml", Status: "stale", Error: "no matches"},
		},
	}

	out := captureStdout(t, func() { printProxyRuleSection(result) })

	assert.Contains(t, out, "r2.yaml")
	assert.Contains(t, out, "invalid")
	assert.Contains(t, out, "missing id")
	assert.Contains(t, out, "r3.yaml")
	assert.Contains(t, out, "stale")
	assert.Contains(t, out, "no matches")
}

// --- printCombinedJSON ---

func TestCheckPrintCombinedJSON_PassNoChecker(t *testing.T) {
	ruleResult := &rules.RunResult{
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "r1"}, Status: "valid"},
		},
	}

	out := captureStdout(t, func() {
		err := printCombinedJSON(nil, ruleResult, nil, false)
		require.NoError(t, err)
	})

	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))

	// result should be "pass"
	var result string
	require.NoError(t, json.Unmarshal(parsed["result"], &result))
	assert.Equal(t, "pass", result)

	// No violations key when checker is nil.
	_, hasViolations := parsed["violations"]
	assert.False(t, hasViolations)

	// proxy_rules should be present.
	_, hasProxy := parsed["proxy_rules"]
	assert.True(t, hasProxy)
}

func TestCheckPrintCombinedJSON_FailWithCheckerViolations(t *testing.T) {
	checkerResult := &checker.CheckResult{
		DependencyViolations: []checker.Violation{
			{Category: "dependency", File: "a.go", Message: "illegal import"},
		},
		AntiPatternViolations: []checker.AntiPattern{
			{Name: "global_state", File: "b.go", Message: "global var"},
		},
	}

	out := captureStdout(t, func() {
		err := printCombinedJSON(checkerResult, nil, nil, true)
		require.NoError(t, err)
	})

	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))

	var result string
	require.NoError(t, json.Unmarshal(parsed["result"], &result))
	assert.Equal(t, "fail", result)

	var violations []json.RawMessage
	require.NoError(t, json.Unmarshal(parsed["violations"], &violations))
	assert.Len(t, violations, 1)

	var antiPatterns []json.RawMessage
	require.NoError(t, json.Unmarshal(parsed["anti_patterns"], &antiPatterns))
	assert.Len(t, antiPatterns, 1)
}

func TestCheckPrintCombinedJSON_MultipleViolationTypes(t *testing.T) {
	checkerResult := &checker.CheckResult{
		DependencyViolations: []checker.Violation{{Message: "dep"}},
		StructureViolations:  []checker.Violation{{Message: "struct"}},
		FunctionViolations:   []checker.Violation{{Message: "func"}},
		NamingViolations:     []checker.Violation{{Message: "name"}},
	}

	out := captureStdout(t, func() {
		err := printCombinedJSON(checkerResult, nil, nil, true)
		require.NoError(t, err)
	})

	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))

	var violations []json.RawMessage
	require.NoError(t, json.Unmarshal(parsed["violations"], &violations))
	assert.Len(t, violations, 4)
}

// --- printCombinedTerminal ---

func TestCheckPrintCombinedTerminal_NilCheckerResult(t *testing.T) {
	cfg := &config.ArchwayConfig{Architecture: "hexagonal"}
	flags := &checkFlags{}

	out := captureStdout(t, func() {
		printCombinedTerminal(nil, nil, nil, cfg, flags)
	})

	assert.Contains(t, out, "Archway Check — hexagonal")
	// Should not contain component info when checkerResult is nil.
	assert.NotContains(t, out, "Components:")
}

func TestCheckPrintCombinedTerminal_NilRuleResult(t *testing.T) {
	cfg := &config.ArchwayConfig{Architecture: "flat"}
	flags := &checkFlags{}

	out := captureStdout(t, func() {
		printCombinedTerminal(nil, nil, nil, cfg, flags)
	})

	assert.Contains(t, out, "Archway Check — flat")
	assert.NotContains(t, out, "PROXY RULES")
}

func TestCheckPrintCombinedTerminal_StagedModeTip(t *testing.T) {
	cfg := &config.ArchwayConfig{Architecture: "test"}
	flags := &checkFlags{staged: true}

	out := captureStdout(t, func() {
		printCombinedTerminal(nil, nil, nil, cfg, flags)
	})

	assert.Contains(t, out, "Tip: Add to .git/hooks/pre-commit:")
	assert.Contains(t, out, "archway check --staged")
}

func TestCheckPrintCombinedTerminal_NoStagedTipByDefault(t *testing.T) {
	cfg := &config.ArchwayConfig{Architecture: "test"}
	flags := &checkFlags{staged: false}

	out := captureStdout(t, func() {
		printCombinedTerminal(nil, nil, nil, cfg, flags)
	})

	assert.NotContains(t, out, "pre-commit")
}

func TestCheckPrintCombinedTerminal_EmptyArchitectureFallback(t *testing.T) {
	cfg := &config.ArchwayConfig{Architecture: ""}
	flags := &checkFlags{}

	out := captureStdout(t, func() {
		printCombinedTerminal(nil, nil, nil, cfg, flags)
	})

	assert.Contains(t, out, "Archway Check — project")
}

// --- getStagedFiles parsing ---

func TestCheckGetStagedFiles_ParsesOutput(t *testing.T) {
	// We can't easily mock exec.Command, but we test the parsing logic
	// by testing the function in a git repo with no staged files.
	tmp := t.TempDir()

	// Init a git repo so the command doesn't fail.
	initGitRepo(t, tmp)

	files, err := getStagedFiles(tmp)
	require.NoError(t, err)
	assert.Nil(t, files, "no staged files should return nil")
}

func TestCheckGetStagedFiles_WithStagedFile(t *testing.T) {
	tmp := t.TempDir()
	initGitRepo(t, tmp)

	// Create and stage a file.
	require.NoError(t, os.WriteFile(tmp+"/test.go", []byte("package main\n"), 0o644))
	gitCmd(t, tmp, "add", "test.go")

	files, err := getStagedFiles(tmp)
	require.NoError(t, err)
	assert.Equal(t, []string{"test.go"}, files)
}

func TestCheckGetStagedFiles_MultipleStagedFiles(t *testing.T) {
	tmp := t.TempDir()
	initGitRepo(t, tmp)

	for _, name := range []string{"a.go", "b.go", "c.go"} {
		require.NoError(t, os.WriteFile(tmp+"/"+name, []byte("package main\n"), 0o644))
	}
	gitCmd(t, tmp, "add", "a.go", "b.go", "c.go")

	files, err := getStagedFiles(tmp)
	require.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Contains(t, files, "a.go")
	assert.Contains(t, files, "b.go")
	assert.Contains(t, files, "c.go")
}

func TestCheckGetStagedFiles_NotAGitRepo(t *testing.T) {
	tmp := t.TempDir()

	_, err := getStagedFiles(tmp)
	assert.Error(t, err)
}

// --- Edge cases ---

func TestCheckFilterRuleResult_NilRunResult(t *testing.T) {
	// Passing a nil RunResult should not panic — the function dereferences r,
	// so this documents the current behavior (expected panic).
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when filterRuleResult receives nil, but got none")
		}
	}()
	filterRuleResult(nil, "any-rule")
}

func TestCheckFilterRuleResult_EmptyRuleID(t *testing.T) {
	result := &rules.RunResult{
		Duration: 10 * time.Millisecond,
		Violations: []rules.RuleViolation{
			{RuleID: "rule-a", File: "a.go", Severity: "error"},
			{RuleID: "rule-b", File: "b.go", Severity: "warning"},
		},
		Statuses: []rules.RuleStatus{
			{Rule: rules.Rule{ID: "rule-a"}, Status: "valid"},
		},
	}

	filtered := filterRuleResult(result, "")

	// Empty rule ID matches nothing since no violation has RuleID == "".
	assert.Empty(t, filtered.Violations)
	assert.Empty(t, filtered.Statuses)
	assert.Equal(t, result.Duration, filtered.Duration)
}

func TestCheckPrintCombinedJSON_NilBothResults(t *testing.T) {
	out := captureStdout(t, func() {
		err := printCombinedJSON(nil, nil, nil, false)
		require.NoError(t, err)
	})

	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))

	var result string
	require.NoError(t, json.Unmarshal(parsed["result"], &result))
	assert.Equal(t, "pass", result)

	// No violations or proxy_rules keys when both are nil.
	_, hasViolations := parsed["violations"]
	assert.False(t, hasViolations)
	_, hasProxy := parsed["proxy_rules"]
	assert.False(t, hasProxy)
}

func TestCheckPrintProxyRuleSection_ZeroCounts(t *testing.T) {
	result := &rules.RunResult{
		Statuses: []rules.RuleStatus{},
	}

	out := captureStdout(t, func() { printProxyRuleSection(result) })

	assert.Contains(t, out, "PROXY RULES (0 valid, 0 invalid, 0 stale)")
	assert.Contains(t, out, "All proxy rules pass")
}

// helpers

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitCmd(t, dir, "init")
	gitCmd(t, dir, "config", "user.email", "test@test.com")
	gitCmd(t, dir, "config", "user.name", "Test")
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.CommandContext(context.Background(), "git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
}
