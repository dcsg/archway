package rules

import (
	"fmt"
	"os"
	"time"
)

// RunRules loads rules from rulesDir, validates them, runs valid grep rules,
// and returns the combined results. AST rules are skipped with a warning
// (handled by checker integration in Phase 2).
func RunRules(rulesDir, projectRoot string, allowedFiles []string) (*RunResult, error) {
	start := time.Now()

	result := &RunResult{}

	// Check if rules directory exists.
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		result.Duration = time.Since(start)
		return result, nil
	}

	allRules, statuses, err := LoadRules(rulesDir, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}
	result.Statuses = statuses

	// Run valid rules.
	for _, rule := range allRules {
		// Find the status for this rule — skip stale rules.
		isStale := false
		for _, s := range statuses {
			if s.Rule.ID == rule.ID && s.Status == "stale" {
				isStale = true
				break
			}
		}
		if isStale {
			continue
		}

		switch rule.Engine {
		case "grep":
			violations, err := RunGrep(rule, projectRoot, allowedFiles)
			if err != nil {
				// Mark as failed but continue with other rules.
				for i, s := range result.Statuses {
					if s.Rule.ID == rule.ID {
						result.Statuses[i].Status = "invalid"
						result.Statuses[i].Error = fmt.Sprintf("execution error: %v", err)
						break
					}
				}
				continue
			}
			result.Violations = append(result.Violations, violations...)

		case "ast":
			// AST rules are handled by checker integration (Phase 2).
			// For now, skip with no error.
			continue
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}
