package guide

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	sentinelStart = "<!-- archway:start -->"
	sentinelEnd   = "<!-- archway:end -->"
	claudeHeader  = "<!-- archway:generated -->\n"
)

// Target defines where and how to write guide output.
type Target interface {
	Name() string
	Write(projectDir, content string) error
}

// resolveTargets returns the targets matching the given selector.
func resolveTargets(selector string) ([]Target, error) {
	all := []Target{
		&claudeTarget{},
		&sentinelTarget{name: "cursor", relPath: ".cursorrules"},
		&sentinelTarget{name: "copilot", relPath: filepath.Join(".github", "copilot-instructions.md")},
		&sentinelTarget{name: "windsurf", relPath: ".windsurfrules"},
	}

	if selector == "" || selector == "all" {
		return all, nil
	}

	for _, t := range all {
		if t.Name() == selector {
			return []Target{t}, nil
		}
	}
	return nil, fmt.Errorf("unknown guide target %q (valid: all, claude, cursor, copilot, windsurf)", selector)
}

// claudeTarget writes directly to .claude/rules/archway.md.
type claudeTarget struct{}

func (t *claudeTarget) Name() string { return "claude" }

func (t *claudeTarget) Write(projectDir, content string) error {
	dir := filepath.Join(projectDir, ".claude", "rules")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create claude rules dir: %w", err)
	}
	path := filepath.Join(dir, "archway.md")
	return os.WriteFile(path, []byte(claudeHeader+content), 0o644)
}

// sentinelTarget uses sentinel-based merging for shared files.
type sentinelTarget struct {
	name    string
	relPath string
}

func (t *sentinelTarget) Name() string { return t.name }

func (t *sentinelTarget) Write(projectDir, content string) error {
	path := filepath.Join(projectDir, t.relPath)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dir for %s: %w", t.name, err)
	}

	wrapped := sentinelStart + "\n" + content + sentinelEnd + "\n"

	existing, err := os.ReadFile(path)
	if err != nil {
		// No existing file: create with sentinels.
		return os.WriteFile(path, []byte(wrapped), 0o644)
	}

	merged := mergeSentinels(string(existing), wrapped)
	return os.WriteFile(path, []byte(merged), 0o644)
}

// mergeSentinels replaces content between sentinels, or appends if none found.
func mergeSentinels(existing, wrapped string) string {
	startIdx := strings.Index(existing, sentinelStart)
	endIdx := strings.Index(existing, sentinelEnd)

	if startIdx < 0 || endIdx < 0 || endIdx < startIdx {
		// No sentinels found: append.
		if !strings.HasSuffix(existing, "\n") && existing != "" {
			existing += "\n"
		}
		return existing + "\n" + wrapped
	}

	// Replace between sentinels (inclusive).
	after := endIdx + len(sentinelEnd)
	// Consume trailing newline if present.
	if after < len(existing) && existing[after] == '\n' {
		after++
	}
	return existing[:startIdx] + wrapped + existing[after:]
}
