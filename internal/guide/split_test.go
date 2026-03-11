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

func TestBuildSplitContent_CorrectFileCount(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{"http-api", "mysql", "auth-jwt"},
		Components: []config.Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{}},
		},
	}

	sc := buildSplitContent(opts)

	assert.NotEmpty(t, sc.Index)
	// http-api -> http, mysql -> data, auth-jwt -> security
	assert.Len(t, sc.Categories, 3)
	assert.Contains(t, sc.Categories, "http")
	assert.Contains(t, sc.Categories, "data")
	assert.Contains(t, sc.Categories, "security")
}

func TestBuildIndexContent_CompactSummary(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{"http-api", "mysql"},
		Components: []config.Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{}},
			{Name: "service", In: []string{"service/**"}, MayDependOn: []string{"domain"}},
		},
	}

	content := buildIndexContent(opts)

	assert.Contains(t, content, "globs: [\"**/*\"]")
	assert.Contains(t, content, "Architecture: hexagonal")
	assert.Contains(t, content, "Hexagonal (ports & adapters)")
	assert.Contains(t, content, "## Layer Rules")
	assert.Contains(t, content, "**domain**: depends on nothing (innermost)")
	assert.Contains(t, content, "## Active Capabilities")
	assert.Contains(t, content, "**http-api**")
	assert.Contains(t, content, "**mysql**")
	assert.Contains(t, content, "## Critical Warnings")

	// Should NOT contain full adding code or anti-patterns (those go in category files).
	assert.NotContains(t, content, "## Adding Code")
	assert.NotContains(t, content, "## Anti-patterns")
}

func TestBuildCategoryContent_CorrectGlobs(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{"http-api"},
	}

	content := buildCategoryContent("http", []string{"http-api"}, opts)

	assert.Contains(t, content, `"internal/handler/**"`)
	assert.Contains(t, content, `"adapter/httphandler/**"`)
	assert.Contains(t, content, "# Archway — HTTP & Transport")
	assert.Contains(t, content, "http-api")
	assert.Contains(t, content, "## Adding Code")
}

func TestInstalledInCategory(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		installed []string
		want      []string
	}{
		{
			name:      "http caps filtered",
			category:  "http",
			installed: []string{"http-api", "mysql", "grpc"},
			want:      []string{"http-api", "grpc"},
		},
		{
			name:      "no match",
			category:  "frontend",
			installed: []string{"http-api", "mysql"},
			want:      nil,
		},
		{
			name:      "unknown category",
			category:  "nonexistent",
			installed: []string{"http-api"},
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := installedInCategory(tt.category, tt.installed)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSplitVsMonolithic_NonClaude(t *testing.T) {
	dir := t.TempDir()
	opts := GenerateOptions{
		ProjectDir:   dir,
		Target:       "cursor",
		Architecture: "hexagonal",
		Capabilities: []string{"http-api", "mysql"},
	}

	err := Generate(opts)
	require.NoError(t, err)

	// Cursor gets monolithic output with sentinels.
	data, err := os.ReadFile(filepath.Join(dir, ".cursorrules"))
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "Architecture: hexagonal")
	assert.Contains(t, content, "## Layer Rules")
	assert.Contains(t, content, "## Adding Code")
}

func TestSplitCleanup_RemovesOldFile(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".claude", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	// Create old monolithic file.
	oldPath := filepath.Join(rulesDir, "archway.md")
	require.NoError(t, os.WriteFile(oldPath, []byte("old content"), 0o644))

	opts := GenerateOptions{
		ProjectDir:   dir,
		Target:       "claude",
		Architecture: "hexagonal",
		Capabilities: []string{"http-api"},
	}

	err := Generate(opts)
	require.NoError(t, err)

	// Old file should be deleted.
	assert.NoFileExists(t, oldPath)
	// New split files should exist.
	assert.FileExists(t, filepath.Join(rulesDir, "archway-index.md"))
	assert.FileExists(t, filepath.Join(rulesDir, "archway-http.md"))
}

func TestCatalogOnly_StaysMonolithic(t *testing.T) {
	dir := t.TempDir()
	opts := GenerateOptions{
		ProjectDir:   dir,
		Target:       "claude",
		Architecture: "hexagonal",
		Capabilities: []string{"http-api"},
		CatalogOnly:  true,
		TemplateFS:   testCapFS(),
	}

	err := Generate(opts)
	require.NoError(t, err)

	// CatalogOnly uses monolithic output.
	assert.FileExists(t, filepath.Join(dir, ".claude", "rules", "archway.md"))
	// No split files.
	assert.NoFileExists(t, filepath.Join(dir, ".claude", "rules", "archway-index.md"))
}

func TestGlobsForCategory(t *testing.T) {
	globs := globsForCategory("http")
	assert.Contains(t, globs, "internal/handler/**")
	assert.Contains(t, globs, "adapter/httphandler/**")

	// Unknown category returns wildcard.
	globs = globsForCategory("nonexistent")
	assert.Equal(t, []string{"**/*"}, globs)
}

func TestBuildSplitContent_OnlyInstalledCategories(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "flat",
		Capabilities: []string{"testing"},
	}

	sc := buildSplitContent(opts)

	// Only quality category should be present.
	assert.Len(t, sc.Categories, 1)
	assert.Contains(t, sc.Categories, "quality")
}

func TestSplitFileTokenCompliance(t *testing.T) {
	dir := t.TempDir()
	opts := GenerateOptions{
		ProjectDir:   dir,
		Target:       "claude",
		Architecture: "hexagonal",
		Capabilities: []string{
			"http-api", "mysql", "redis", "auth-jwt", "cors",
			"rate-limiting", "health", "observability", "docker", "testing",
		},
		Components: []config.Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{}},
			{Name: "service", In: []string{"service/**"}, MayDependOn: []string{"domain"}},
			{Name: "adapters", In: []string{"adapter/**"}, MayDependOn: []string{"domain"}},
		},
	}

	require.NoError(t, Generate(opts))

	entries, err := os.ReadDir(filepath.Join(dir, ".claude", "rules"))
	require.NoError(t, err)

	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "archway-") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, ".claude", "rules", e.Name()))
		require.NoError(t, err)
		words := len(strings.Fields(string(data)))
		approxTokens := int(float64(words) * 1.3)
		if approxTokens > 1500 {
			t.Errorf("file %s too large: ~%d tokens (%d words)", e.Name(), approxTokens, words)
		}
	}
}

func TestCategoryContent_HasGlobsFrontmatter(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{"mysql"},
	}

	content := buildCategoryContent("data", []string{"mysql"}, opts)

	// Should start with YAML frontmatter.
	assert.True(t, strings.HasPrefix(content, "---\nglobs: ["))
	assert.Contains(t, content, `"internal/repository/**"`)
	assert.Contains(t, content, `"adapter/*repo/**"`)
}
