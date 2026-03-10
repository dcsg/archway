package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadRules reads all .yaml files from rulesDir, parses them, and validates each rule.
// Returns the parsed rules and their validation statuses.
func LoadRules(rulesDir, projectRoot string) ([]Rule, []RuleStatus, error) {
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("read rules directory %s: %w", rulesDir, err)
	}

	allRules := make([]Rule, 0, len(entries))
	statuses := make([]RuleStatus, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(rulesDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			statuses = append(statuses, RuleStatus{
				Filename: name,
				Status:   "invalid",
				Error:    fmt.Sprintf("cannot read file: %v", err),
			})
			continue
		}

		var rule Rule
		if err := yaml.Unmarshal(data, &rule); err != nil {
			statuses = append(statuses, RuleStatus{
				Filename: name,
				Status:   "invalid",
				Error:    fmt.Sprintf("invalid YAML: %v", err),
			})
			continue
		}

		status := ValidateRule(rule, name, projectRoot)
		statuses = append(statuses, status)

		if status.Status == "valid" || status.Status == "stale" {
			allRules = append(allRules, rule)
		}
	}

	return allRules, statuses, nil
}

// ValidateRule checks a single rule for correctness and returns its status.
func ValidateRule(r Rule, filename, projectRoot string) RuleStatus {
	status := RuleStatus{
		Rule:     r,
		Filename: filename,
		Status:   "valid",
	}

	// Required fields.
	if strings.TrimSpace(r.ID) == "" {
		status.Status = "invalid"
		status.Error = "missing required field: id"
		return status
	}

	if r.Engine != "grep" && r.Engine != "ast" {
		status.Status = "invalid"
		status.Error = fmt.Sprintf("unknown engine %q (must be grep or ast)", r.Engine)
		return status
	}

	// Default severity.
	if r.Severity == "" {
		r.Severity = "error"
	}
	if r.Severity != "error" && r.Severity != "warning" {
		status.Status = "invalid"
		status.Error = fmt.Sprintf("invalid severity %q (must be error or warning)", r.Severity)
		return status
	}

	if len(r.Scope) == 0 {
		status.Status = "invalid"
		status.Error = "missing required field: scope"
		return status
	}

	// Engine-specific validation.
	switch r.Engine {
	case "grep":
		if r.Pattern == "" && r.FileMustContain == "" {
			status.Status = "invalid"
			status.Error = "grep engine requires at least one of: pattern, file-must-contain"
			return status
		}
		// Validate regex compilation.
		if r.Pattern != "" {
			if _, err := regexp.Compile(r.Pattern); err != nil {
				status.Status = "invalid"
				status.Error = fmt.Sprintf("invalid pattern regex: %v", err)
				return status
			}
		}
		if r.MustContain != "" {
			if _, err := regexp.Compile(r.MustContain); err != nil {
				status.Status = "invalid"
				status.Error = fmt.Sprintf("invalid must-contain regex: %v", err)
				return status
			}
		}
		if r.MustNotContain != "" {
			if _, err := regexp.Compile(r.MustNotContain); err != nil {
				status.Status = "invalid"
				status.Error = fmt.Sprintf("invalid must-not-contain regex: %v", err)
				return status
			}
		}
		if r.FileMustContain != "" {
			if _, err := regexp.Compile(r.FileMustContain); err != nil {
				status.Status = "invalid"
				status.Error = fmt.Sprintf("invalid file-must-contain regex: %v", err)
				return status
			}
		}

	case "ast":
		if strings.TrimSpace(r.Detector) == "" {
			status.Status = "invalid"
			status.Error = "ast engine requires: detector"
			return status
		}
		if !IsKnownDetector(r.Detector) {
			status.Status = "invalid"
			status.Error = fmt.Sprintf("unknown detector %q", r.Detector)
			return status
		}
	}

	// Check for stale scope (matches 0 files).
	if projectRoot != "" {
		files, err := ExpandScope(r.Scope, r.Exclude, projectRoot, nil)
		if err == nil && len(files) == 0 {
			status.Status = "stale"
			status.Error = "scope matches 0 files"
		}
	}

	return status
}
