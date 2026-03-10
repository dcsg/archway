package guide

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dcsg/archway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildContent_Hexagonal(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{"http-api", "mysql"},
		Components: []config.Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{}},
			{Name: "ports", In: []string{"port/**"}, MayDependOn: []string{"domain"}},
			{Name: "service", In: []string{"service/**"}, MayDependOn: []string{"domain", "ports"}},
			{Name: "adapters", In: []string{"adapter/**"}, MayDependOn: []string{"ports", "domain"}},
		},
	}

	content := buildContent(opts)

	assert.Contains(t, content, "Architecture: hexagonal")
	assert.Contains(t, content, "hexagonal (ports & adapters)")
	assert.Contains(t, content, "## Layer Rules")
	assert.Contains(t, content, "domain")
	assert.Contains(t, content, "Dependencies: none (innermost layer)")
	assert.Contains(t, content, "## Anti-patterns to Avoid")
	assert.Contains(t, content, "NEVER import infrastructure packages from `domain/`")
	assert.Contains(t, content, "http-api")
	assert.Contains(t, content, "mysql")
}

func TestBuildContent_Flat(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "flat",
		Capabilities: nil,
		Components:   nil,
	}

	content := buildContent(opts)

	assert.Contains(t, content, "Architecture: flat")
	assert.Contains(t, content, "No layer restrictions")
	assert.Contains(t, content, "No dependency restrictions")
	assert.NotContains(t, content, "NEVER import infrastructure packages from `domain/`")
}

func TestMergeSentinels_NoExistingFile(t *testing.T) {
	dir := t.TempDir()
	target := &sentinelTarget{name: "test", relPath: "test.md"}

	err := target.Write(dir, "guide content\n")
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "test.md"))
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, sentinelStart)
	assert.Contains(t, content, sentinelEnd)
	assert.Contains(t, content, "guide content")
}

func TestMergeSentinels_ExistingWithoutSentinels(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	require.NoError(t, os.WriteFile(path, []byte("user content\n"), 0o644))

	target := &sentinelTarget{name: "test", relPath: "test.md"}
	err := target.Write(dir, "guide content\n")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	assert.True(t, strings.HasPrefix(content, "user content\n"))
	assert.Contains(t, content, sentinelStart)
	assert.Contains(t, content, "guide content")
	assert.Contains(t, content, sentinelEnd)
}

func TestMergeSentinels_ExistingWithSentinels(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	existing := "before\n" + sentinelStart + "\nold content\n" + sentinelEnd + "\nafter\n"
	require.NoError(t, os.WriteFile(path, []byte(existing), 0o644))

	target := &sentinelTarget{name: "test", relPath: "test.md"}
	err := target.Write(dir, "new content\n")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "before\n")
	assert.Contains(t, content, "new content")
	assert.NotContains(t, content, "old content")
	assert.Contains(t, content, "after\n")
}

func TestMergeSentinels_Idempotent(t *testing.T) {
	dir := t.TempDir()
	target := &sentinelTarget{name: "test", relPath: "test.md"}

	require.NoError(t, target.Write(dir, "guide content\n"))
	first, err := os.ReadFile(filepath.Join(dir, "test.md"))
	require.NoError(t, err)

	require.NoError(t, target.Write(dir, "guide content\n"))
	second, err := os.ReadFile(filepath.Join(dir, "test.md"))
	require.NoError(t, err)

	assert.Equal(t, string(first), string(second))
}

func TestClaudeTarget_WritesCorrectPath(t *testing.T) {
	dir := t.TempDir()
	target := &claudeTarget{}

	err := target.Write(dir, "test content\n")
	require.NoError(t, err)

	path := filepath.Join(dir, ".claude", "rules", "archway.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	assert.True(t, strings.HasPrefix(content, claudeHeader))
	assert.Contains(t, content, "test content")
}

func TestResolveTargets(t *testing.T) {
	tests := []struct {
		selector string
		count    int
		wantErr  bool
	}{
		{"all", 4, false},
		{"", 4, false},
		{"claude", 1, false},
		{"cursor", 1, false},
		{"copilot", 1, false},
		{"windsurf", 1, false},
		{"invalid", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.selector, func(t *testing.T) {
			targets, err := resolveTargets(tt.selector)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, targets, tt.count)
		})
	}
}

func TestOutputTargetPaths(t *testing.T) {
	dir := t.TempDir()
	opts := GenerateOptions{
		ProjectDir:   dir,
		Target:       "all",
		Architecture: "flat",
	}

	require.NoError(t, Generate(opts))

	expected := []string{
		filepath.Join(dir, ".claude", "rules", "archway.md"),
		filepath.Join(dir, ".cursorrules"),
		filepath.Join(dir, ".github", "copilot-instructions.md"),
		filepath.Join(dir, ".windsurfrules"),
	}
	for _, path := range expected {
		assert.FileExists(t, path)
	}
}
