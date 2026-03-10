package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dcsg/archway/internal/checker"
	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/rules"
	"github.com/spf13/cobra"
)

type checkFlags struct {
	projectPath string
	proxyRules  bool
	detectors   bool
	rule        string
	staged      bool
}

func newCheckCommand(opts *globalOptions) *cobra.Command {
	flags := &checkFlags{}

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validate project against archway.yaml rules",
		Long: `Check validates an existing project against its archway.yaml rules.

Reports dependency violations, structure issues, and function complexity.
Runs both built-in detectors and proxy rules by default.
Exits with code 1 if any error-severity violations are found (useful in CI).`,
		Example: `  archway check
  archway check --path ./my-service
  archway check --proxy-rules
  archway check --staged
  archway check --rule cap-sql-parameterized`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(opts, flags)
		},
	}

	cmd.Flags().StringVar(&flags.projectPath, "path", ".", "Project path to check")
	cmd.Flags().BoolVar(&flags.proxyRules, "proxy-rules", false, "Run only proxy rules (skip built-in detectors)")
	cmd.Flags().BoolVar(&flags.detectors, "detectors", false, "Run only built-in detectors (skip proxy rules)")
	cmd.Flags().StringVar(&flags.rule, "rule", "", "Run a single proxy rule by ID")
	cmd.Flags().BoolVar(&flags.staged, "staged", false, "Only check files in git staging area")

	return cmd
}

func runCheck(opts *globalOptions, flags *checkFlags) error {
	projectPath := flags.projectPath

	archwayPath, err := config.FindArchwayYAML(projectPath)
	if err != nil {
		return fmt.Errorf("no archway.yaml found in %s (or parent directories)", projectPath)
	}

	cfg, err := config.LoadArchwayYAML(archwayPath)
	if err != nil {
		return err
	}

	// Get staged files if --staged is set.
	var stagedFiles []string
	if flags.staged {
		stagedFiles, err = getStagedFiles(projectPath)
		if err != nil {
			return fmt.Errorf("get staged files: %w", err)
		}
		if len(stagedFiles) == 0 {
			fmt.Println("No staged files to check.")
			return nil
		}
	}

	var checkerResult *checker.CheckResult
	var ruleResult *rules.RunResult
	hasErrors := false

	// Run built-in detectors unless --proxy-rules or --rule is set.
	if !flags.proxyRules && flags.rule == "" {
		checkerResult, err = checker.Check(cfg, projectPath)
		if err != nil {
			return fmt.Errorf("check failed: %w", err)
		}
		if !checkerResult.Passed() {
			hasErrors = true
		}
	}

	// Run proxy rules unless --detectors is set.
	if !flags.detectors {
		rulesDir := filepath.Join(projectPath, ".archway", "rules")
		ruleResult, err = rules.RunRules(rulesDir, projectPath, stagedFiles)
		if err != nil {
			return fmt.Errorf("proxy rules failed: %w", err)
		}

		// Filter to single rule if --rule is set.
		if flags.rule != "" && ruleResult != nil {
			ruleResult = filterRuleResult(ruleResult, flags.rule)
		}

		if ruleResult != nil && ruleResult.ErrorCount() > 0 {
			hasErrors = true
		}
	}

	if opts.Output == "json" {
		return printCombinedJSON(checkerResult, ruleResult, hasErrors)
	}
	printCombinedTerminal(checkerResult, ruleResult, cfg, flags)

	if hasErrors {
		os.Exit(1)
	}
	return nil
}

// getStagedFiles returns relative file paths in the git staging area.
func getStagedFiles(projectPath string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}

// filterRuleResult keeps only violations and statuses matching a specific rule ID.
func filterRuleResult(r *rules.RunResult, ruleID string) *rules.RunResult {
	filtered := &rules.RunResult{Duration: r.Duration}
	for _, v := range r.Violations {
		if v.RuleID == ruleID {
			filtered.Violations = append(filtered.Violations, v)
		}
	}
	for _, s := range r.Statuses {
		if s.Rule.ID == ruleID {
			filtered.Statuses = append(filtered.Statuses, s)
		}
	}
	return filtered
}

func printCombinedTerminal(checkerResult *checker.CheckResult, ruleResult *rules.RunResult, cfg *config.ArchwayConfig, flags *checkFlags) {
	projectName := cfg.Architecture
	if projectName == "" {
		projectName = "project"
	}

	fmt.Printf("\nArchway Check — %s\n", projectName)
	fmt.Println(strings.Repeat("═", 55))

	// Built-in detector results.
	if checkerResult != nil {
		coverage := float64(0)
		if checkerResult.ComponentsTotal > 0 {
			coverage = float64(checkerResult.ComponentsCovered) / float64(checkerResult.ComponentsTotal) * 100
		}
		fmt.Printf("\nComponents:  %d defined, %d covered (%.0f%% coverage)\n",
			checkerResult.ComponentsTotal, checkerResult.ComponentsCovered, coverage)

		printViolationSection("DEPENDENCY VIOLATIONS", checkerResult.DependencyViolations)
		printViolationSection("STRUCTURE VIOLATIONS", checkerResult.StructureViolations)
		printViolationSection("FUNCTION VIOLATIONS", checkerResult.FunctionViolations)
		printViolationSection("NAMING VIOLATIONS", checkerResult.NamingViolations)
		printAntiPatternSection("ANTI-PATTERN VIOLATIONS", checkerResult.AntiPatternViolations)
	}

	// Proxy rule results.
	if ruleResult != nil {
		printProxyRuleSection(ruleResult)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("═", 55))

	if flags.staged {
		fmt.Println("\nTip: Add to .git/hooks/pre-commit:")
		fmt.Println("  #!/bin/sh")
		fmt.Println("  archway check --staged")
	}
}

func printProxyRuleSection(result *rules.RunResult) {
	fmt.Printf("\nPROXY RULES (%d valid, %d invalid, %d stale)\n",
		result.ValidRuleCount(), result.InvalidRuleCount(), result.StaleRuleCount())

	if len(result.Violations) == 0 {
		fmt.Println("  ✓ All proxy rules pass")
		return
	}

	errors := result.ErrorCount()
	warnings := result.WarningCount()
	fmt.Printf("  %d errors, %d warnings\n", errors, warnings)

	for _, v := range result.Violations {
		sev := "⚠"
		if v.Severity == "error" {
			sev = "✗"
		}
		if v.Line > 0 {
			fmt.Printf("  %s [%s] %s:%d %s\n", sev, v.RuleID, v.File, v.Line, v.Description)
		} else {
			fmt.Printf("  %s [%s] %s — %s\n", sev, v.RuleID, v.File, v.Description)
		}
		if v.Match != "" {
			fmt.Printf("    > %s\n", v.Match)
		}
	}

	// Report invalid/stale rules.
	for _, s := range result.Statuses {
		switch s.Status {
		case "invalid", "stale":
			fmt.Printf("  ⚠ [%s] %s: %s\n", s.Filename, s.Status, s.Error)
		}
	}
}

func printViolationSection(title string, violations []checker.Violation) {
	fmt.Printf("\n%s (%d)\n", title, len(violations))
	if len(violations) == 0 {
		fmt.Println("  ✓ All checks pass")
		return
	}
	for _, v := range violations {
		switch {
		case v.File != "" && v.Line > 0:
			fmt.Printf("  ✗ %s:%d %s\n", v.File, v.Line, v.Message)
		case v.File != "":
			fmt.Printf("  ✗ %s — %s\n", v.File, v.Message)
		default:
			fmt.Printf("  ✗ %s\n", v.Message)
		}
	}
}

func printAntiPatternSection(title string, violations []checker.AntiPattern) {
	fmt.Printf("\n%s (%d)\n", title, len(violations))
	if len(violations) == 0 {
		fmt.Println("  ✓ All checks pass")
		return
	}
	for _, v := range violations {
		sev := "⚠"
		if v.Severity == "error" {
			sev = "✗"
		}
		switch {
		case v.File != "" && v.Line > 0:
			fmt.Printf("  %s [%s] %s:%d %s\n", sev, v.Name, v.File, v.Line, v.Message)
		case v.File != "":
			fmt.Printf("  %s [%s] %s — %s\n", sev, v.Name, v.File, v.Message)
		default:
			fmt.Printf("  %s [%s] %s\n", sev, v.Name, v.Message)
		}
	}
}

func printCombinedJSON(checkerResult *checker.CheckResult, ruleResult *rules.RunResult, hasErrors bool) error {
	type jsonOutput struct {
		Result       string                `json:"result"`
		Violations   []checker.Violation   `json:"violations,omitempty"`
		AntiPatterns []checker.AntiPattern `json:"anti_patterns,omitempty"`
		ProxyRules   *rules.RunResult      `json:"proxy_rules,omitempty"`
	}

	status := "pass"
	if hasErrors {
		status = "fail"
	}

	out := jsonOutput{
		Result:     status,
		ProxyRules: ruleResult,
	}

	if checkerResult != nil {
		var allViolations []checker.Violation
		allViolations = append(allViolations, checkerResult.DependencyViolations...)
		allViolations = append(allViolations, checkerResult.StructureViolations...)
		allViolations = append(allViolations, checkerResult.FunctionViolations...)
		allViolations = append(allViolations, checkerResult.NamingViolations...)
		out.Violations = allViolations
		out.AntiPatterns = checkerResult.AntiPatternViolations
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
