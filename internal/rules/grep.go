package rules

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"unicode/utf8"
)

const (
	maxFileSize = 1 << 20 // 1MB — skip larger files
)

// RunGrep executes a grep-engine rule against files matching its scope.
func RunGrep(rule Rule, projectRoot string, allowedFiles []string) ([]RuleViolation, error) {
	files, err := ExpandScope(rule.Scope, rule.Exclude, projectRoot, allowedFiles)
	if err != nil {
		return nil, fmt.Errorf("expand scope for rule %s: %w", rule.ID, err)
	}

	var patternRe, mustContainRe, mustNotContainRe, fileMustContainRe *regexp.Regexp

	if rule.Pattern != "" {
		patternRe, err = regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, fmt.Errorf("compile pattern for rule %s: %w", rule.ID, err)
		}
	}
	if rule.MustContain != "" {
		mustContainRe, err = regexp.Compile(rule.MustContain)
		if err != nil {
			return nil, fmt.Errorf("compile must-contain for rule %s: %w", rule.ID, err)
		}
	}
	if rule.MustNotContain != "" {
		mustNotContainRe, err = regexp.Compile(rule.MustNotContain)
		if err != nil {
			return nil, fmt.Errorf("compile must-not-contain for rule %s: %w", rule.ID, err)
		}
	}
	if rule.FileMustContain != "" {
		fileMustContainRe, err = regexp.Compile(rule.FileMustContain)
		if err != nil {
			return nil, fmt.Errorf("compile file-must-contain for rule %s: %w", rule.ID, err)
		}
	}

	var violations []RuleViolation

	for _, relPath := range files {
		absPath := filepath.Join(projectRoot, relPath)

		// Skip oversized files.
		info, err := os.Stat(absPath)
		if err != nil || info.Size() > maxFileSize {
			continue
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		// Skip binary files (check first 512 bytes for null bytes).
		if isBinary(data) {
			continue
		}

		// File-level check: file-must-contain.
		if fileMustContainRe != nil {
			if !fileMustContainRe.Match(data) {
				violations = append(violations, RuleViolation{
					RuleID:      rule.ID,
					Engine:      "grep",
					Description: rule.Description,
					Severity:    severity(rule.Severity),
					Ref:         rule.Ref,
					File:        relPath,
					Line:        0,
					Match:       fmt.Sprintf("file does not contain pattern: %s", rule.FileMustContain),
				})
			}
		}

		// Line-level check: pattern + must-contain + must-not-contain.
		if patternRe != nil {
			lineViolations := scanLines(rule, relPath, data, patternRe, mustContainRe, mustNotContainRe)
			violations = append(violations, lineViolations...)
		}
	}

	return violations, nil
}

func scanLines(rule Rule, relPath string, data []byte, patternRe, mustContainRe, mustNotContainRe *regexp.Regexp) []RuleViolation {
	var violations []RuleViolation

	scanner := bufio.NewScanner(stringReader(data))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if !patternRe.MatchString(line) {
			continue
		}

		// Pattern matched. Now check must-contain / must-not-contain.
		if mustContainRe != nil && !mustContainRe.MatchString(line) {
			// Line matches pattern but does NOT contain required pattern → violation.
			violations = append(violations, RuleViolation{
				RuleID:      rule.ID,
				Engine:      "grep",
				Description: rule.Description,
				Severity:    severity(rule.Severity),
				Ref:         rule.Ref,
				File:        relPath,
				Line:        lineNum,
				Match:       truncate(line, 120),
			})
			continue
		}

		if mustNotContainRe != nil && mustNotContainRe.MatchString(line) {
			// Line matches pattern AND contains forbidden pattern → violation.
			violations = append(violations, RuleViolation{
				RuleID:      rule.ID,
				Engine:      "grep",
				Description: rule.Description,
				Severity:    severity(rule.Severity),
				Ref:         rule.Ref,
				File:        relPath,
				Line:        lineNum,
				Match:       truncate(line, 120),
			})
			continue
		}

		// If neither must-contain nor must-not-contain, the pattern match itself is the violation.
		if mustContainRe == nil && mustNotContainRe == nil {
			violations = append(violations, RuleViolation{
				RuleID:      rule.ID,
				Engine:      "grep",
				Description: rule.Description,
				Severity:    severity(rule.Severity),
				Ref:         rule.Ref,
				File:        relPath,
				Line:        lineNum,
				Match:       truncate(line, 120),
			})
		}
	}

	return violations
}

// isBinary checks if data appears to be a binary file by looking for null bytes
// in the first 512 bytes.
func isBinary(data []byte) bool {
	check := data
	if len(check) > 512 {
		check = check[:512]
	}
	if !utf8.Valid(check) {
		return true
	}
	return slices.Contains(check, 0)
}

func severity(s string) string {
	if s == "" {
		return "error"
	}
	return s
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

type byteReader struct {
	data []byte
	pos  int
}

func stringReader(data []byte) *byteReader {
	return &byteReader{data: data}
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
