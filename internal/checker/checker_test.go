package checker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dcsg/archway/internal/config"
)

func TestCheckStructure_RequiredDirs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "domain"), 0o755)
	// "port/" is missing.

	rules := config.StructureConfig{
		RequiredDirs: []string{"domain/", "port/"},
	}

	violations := checkStructure(rules, dir)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Rule != "required_dir" {
		t.Errorf("expected rule required_dir, got %s", violations[0].Rule)
	}
}

func TestCheckStructure_ForbiddenDirs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "utils"), 0o755)

	rules := config.StructureConfig{
		ForbiddenDirs: []string{"utils/"},
	}

	violations := checkStructure(rules, dir)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Rule != "forbidden_dir" {
		t.Errorf("expected rule forbidden_dir, got %s", violations[0].Rule)
	}
}

func TestCheckStructure_AllPresent(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "domain"), 0o755)
	os.MkdirAll(filepath.Join(dir, "port"), 0o755)

	rules := config.StructureConfig{
		RequiredDirs: []string{"domain/", "port/"},
	}

	violations := checkStructure(rules, dir)
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(violations))
	}
}

func TestCheckResult_Metrics(t *testing.T) {
	result := &CheckResult{
		DependencyViolations: []Violation{{Category: "dependency"}},
		StructureViolations:  []Violation{{Category: "structure"}},
		RulesChecked:         10,
		RulesPassing:         8,
		ComponentsTotal:      5,
		ComponentsCovered:    4,
	}

	if result.TotalViolations() != 2 {
		t.Errorf("TotalViolations = %d, want 2", result.TotalViolations())
	}
	if result.Passed() {
		t.Error("Passed() should be false")
	}
	if result.Compliance() != 0.8 {
		t.Errorf("Compliance = %f, want 0.8", result.Compliance())
	}
}

func TestCheckResult_AllPassing(t *testing.T) {
	result := &CheckResult{
		RulesChecked: 5,
		RulesPassing: 5,
	}

	if !result.Passed() {
		t.Error("Passed() should be true")
	}
	if result.Compliance() != 1.0 {
		t.Errorf("Compliance = %f, want 1.0", result.Compliance())
	}
}

func TestCheckResult_ZeroViolations(t *testing.T) {
	result := &CheckResult{}
	if !result.Passed() {
		t.Error("Passed() should be true when there are zero violations")
	}
	if result.TotalViolations() != 0 {
		t.Errorf("TotalViolations() = %d, want 0", result.TotalViolations())
	}
}

func TestCheckResult_MixedViolations(t *testing.T) {
	result := &CheckResult{
		DependencyViolations: []Violation{{Category: "dependency"}, {Category: "dependency"}},
		StructureViolations:  []Violation{{Category: "structure"}},
		FunctionViolations:   []Violation{{Category: "function"}, {Category: "function"}, {Category: "function"}},
		NamingViolations:     []Violation{{Category: "naming"}},
		AntiPatternViolations: []AntiPattern{
			{Category: "anti-pattern", Name: "test", Message: "test"},
		},
		RulesChecked: 20,
		RulesPassing: 12,
	}

	if result.TotalViolations() != 8 {
		t.Errorf("TotalViolations() = %d, want 8", result.TotalViolations())
	}
	if result.Passed() {
		t.Error("Passed() should be false with violations")
	}
}

func TestCheckResult_ComplianceZeroRulesChecked(t *testing.T) {
	result := &CheckResult{
		RulesChecked: 0,
		RulesPassing: 0,
	}
	if result.Compliance() != 1.0 {
		t.Errorf("Compliance() = %f, want 1.0 when no rules checked", result.Compliance())
	}
}

func TestFunctionRuleCount(t *testing.T) {
	tests := []struct {
		name  string
		rules config.FunctionRules
		want  int
	}{
		{"all set", config.FunctionRules{MaxLines: 80, MaxParams: 4, MaxReturnValues: 2}, 3},
		{"none set", config.FunctionRules{}, 0},
		{"only lines", config.FunctionRules{MaxLines: 50}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := functionRuleCount(tt.rules)
			if got != tt.want {
				t.Errorf("functionRuleCount() = %d, want %d", got, tt.want)
			}
		})
	}
}
