package rules

import "time"

// Rule represents a proxy rule loaded from a YAML file.
type Rule struct {
	ID              string   `yaml:"id"`
	Engine          string   `yaml:"engine"`
	Description     string   `yaml:"description"`
	Severity        string   `yaml:"severity"`
	Ref             string   `yaml:"ref,omitempty"`
	Pattern         string   `yaml:"pattern,omitempty"`
	MustContain     string   `yaml:"must-contain,omitempty"`
	MustNotContain  string   `yaml:"must-not-contain,omitempty"`
	FileMustContain string   `yaml:"file-must-contain,omitempty"`
	Detector        string   `yaml:"detector,omitempty"`
	Scope           []string `yaml:"scope"`
	Exclude         []string `yaml:"exclude,omitempty"`
}

// RuleViolation represents a single violation found by a rule.
type RuleViolation struct {
	RuleID      string `json:"rule_id"`
	Engine      string `json:"engine"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Ref         string `json:"ref,omitempty"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Match       string `json:"match,omitempty"`
}

// RuleStatus reports the validation status of a rule file.
type RuleStatus struct {
	Rule     Rule   `json:"rule"`
	Filename string `json:"filename"`
	Status   string `json:"status"` // "valid", "invalid", "stale"
	Error    string `json:"error,omitempty"`
}

// RunResult holds the outcome of running all proxy rules.
type RunResult struct {
	Violations []RuleViolation `json:"violations"`
	Statuses   []RuleStatus    `json:"statuses"`
	Duration   time.Duration   `json:"duration"`
}

// ErrorCount returns the number of error-severity violations.
func (r *RunResult) ErrorCount() int {
	count := 0
	for _, v := range r.Violations {
		if v.Severity == "error" {
			count++
		}
	}
	return count
}

// WarningCount returns the number of warning-severity violations.
func (r *RunResult) WarningCount() int {
	count := 0
	for _, v := range r.Violations {
		if v.Severity == "warning" {
			count++
		}
	}
	return count
}

// ValidRuleCount returns the number of valid rules.
func (r *RunResult) ValidRuleCount() int {
	count := 0
	for _, s := range r.Statuses {
		if s.Status == "valid" {
			count++
		}
	}
	return count
}

// InvalidRuleCount returns the number of invalid rules.
func (r *RunResult) InvalidRuleCount() int {
	count := 0
	for _, s := range r.Statuses {
		if s.Status == "invalid" {
			count++
		}
	}
	return count
}

// StaleRuleCount returns the number of stale rules.
func (r *RunResult) StaleRuleCount() int {
	count := 0
	for _, s := range r.Statuses {
		if s.Status == "stale" {
			count++
		}
	}
	return count
}
