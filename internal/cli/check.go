package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dcsg/archway/internal/checker"
	"github.com/dcsg/archway/internal/config"
	"github.com/spf13/cobra"
)

func newCheckCommand(opts *globalOptions) *cobra.Command {
	var projectPath string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validate project against archway.yaml rules",
		Long: `Check validates an existing project against its archway.yaml rules.

Reports dependency violations, structure issues, and function complexity.
Exits with code 1 if any violations are found (useful in CI).`,
		Example: `  archway check
  archway check --path ./my-service
  archway check --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(opts, projectPath)
		},
	}

	cmd.Flags().StringVar(&projectPath, "path", ".", "Project path to check")

	return cmd
}

func runCheck(opts *globalOptions, projectPath string) error {
	archwayPath, err := config.FindArchwayYAML(projectPath)
	if err != nil {
		return fmt.Errorf("no archway.yaml found in %s (or parent directories)", projectPath)
	}

	cfg, err := config.LoadArchwayYAML(archwayPath)
	if err != nil {
		return err
	}

	result, err := checker.Check(cfg, projectPath)
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	if opts.Output == "json" {
		return printCheckJSON(result)
	}
	return printCheckTerminal(result, cfg)
}

func printCheckTerminal(result *checker.CheckResult, cfg *config.ArchwayConfig) error {
	projectName := cfg.Architecture
	if projectName == "" {
		projectName = "project"
	}

	fmt.Printf("\nArchway Check — %s\n", projectName)
	fmt.Println(strings.Repeat("═", 55))

	coverage := float64(0)
	if result.ComponentsTotal > 0 {
		coverage = float64(result.ComponentsCovered) / float64(result.ComponentsTotal) * 100
	}
	fmt.Printf("\nComponents:  %d defined, %d covered (%.0f%% coverage)\n",
		result.ComponentsTotal, result.ComponentsCovered, coverage)
	fmt.Printf("Rules:       %d checked\n", result.RulesChecked)

	printViolationSection("DEPENDENCY VIOLATIONS", result.DependencyViolations)
	printViolationSection("STRUCTURE VIOLATIONS", result.StructureViolations)
	printViolationSection("FUNCTION VIOLATIONS", result.FunctionViolations)
	printViolationSection("NAMING VIOLATIONS", result.NamingViolations)
	printAntiPatternSection("ANTI-PATTERN VIOLATIONS", result.AntiPatternViolations)

	fmt.Println()
	fmt.Println(strings.Repeat("═", 55))

	if result.Passed() {
		fmt.Println("Result: PASS — no violations found")
	} else {
		fmt.Printf("Result: FAIL — %d violations found\n", result.TotalViolations())
	}
	fmt.Printf("  Compliance: %.0f%% (%d/%d rules passing)\n",
		result.Compliance()*100, result.RulesPassing, result.RulesChecked)
	if result.ComponentsTotal > 0 {
		fmt.Printf("  Coverage:   %.0f%% (%d/%d components checked)\n",
			coverage, result.ComponentsCovered, result.ComponentsTotal)
	}

	if !result.Passed() {
		os.Exit(1)
	}
	return nil
}

func printViolationSection(title string, violations []checker.Violation) {
	fmt.Printf("\n%s (%d)\n", title, len(violations))
	if len(violations) == 0 {
		fmt.Println("  ✓ All checks pass")
		return
	}
	for _, v := range violations {
		if v.File != "" && v.Line > 0 {
			fmt.Printf("  ✗ %s:%d %s\n", v.File, v.Line, v.Message)
		} else if v.File != "" {
			fmt.Printf("  ✗ %s — %s\n", v.File, v.Message)
		} else {
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
		if v.File != "" && v.Line > 0 {
			fmt.Printf("  %s [%s] %s:%d %s\n", sev, v.Name, v.File, v.Line, v.Message)
		} else if v.File != "" {
			fmt.Printf("  %s [%s] %s — %s\n", sev, v.Name, v.File, v.Message)
		} else {
			fmt.Printf("  %s [%s] %s\n", sev, v.Name, v.Message)
		}
	}
}

func printCheckJSON(result *checker.CheckResult) error {
	type jsonOutput struct {
		Result       string                `json:"result"`
		Compliance   float64               `json:"compliance"`
		Coverage     float64               `json:"coverage"`
		Violations   []checker.Violation    `json:"violations"`
		AntiPatterns []checker.AntiPattern  `json:"anti_patterns"`
	}

	coverage := float64(0)
	if result.ComponentsTotal > 0 {
		coverage = float64(result.ComponentsCovered) / float64(result.ComponentsTotal)
	}

	status := "pass"
	if !result.Passed() {
		status = "fail"
	}

	var allViolations []checker.Violation
	allViolations = append(allViolations, result.DependencyViolations...)
	allViolations = append(allViolations, result.StructureViolations...)
	allViolations = append(allViolations, result.FunctionViolations...)
	allViolations = append(allViolations, result.NamingViolations...)

	out := jsonOutput{
		Result:       status,
		Compliance:   result.Compliance(),
		Coverage:     coverage,
		Violations:   allViolations,
		AntiPatterns: result.AntiPatternViolations,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))

	if !result.Passed() {
		os.Exit(1)
	}
	return nil
}
