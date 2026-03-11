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

func TestBuildContent_Layered(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "layered",
		Components: []config.Component{
			{Name: "handler", In: []string{"internal/handler/**"}, MayDependOn: []string{"service", "model"}},
			{Name: "service", In: []string{"internal/service/**"}, MayDependOn: []string{"repository", "model"}},
			{Name: "repository", In: []string{"internal/repository/**"}, MayDependOn: []string{"model"}},
			{Name: "model", In: []string{"internal/model/**"}, MayDependOn: []string{}},
		},
	}

	content := buildContent(opts)

	assert.Contains(t, content, "Architecture: layered")
	assert.Contains(t, content, "handler → service → repository → model")
	assert.Contains(t, content, "## Layer Rules")
	assert.Contains(t, content, "Dependencies: none (innermost layer)")
	assert.Contains(t, content, "## Anti-patterns to Avoid")
	assert.Contains(t, content, "NEVER let handler bypass service and call repository directly")
	assert.Contains(t, content, "NEVER put business logic in `internal/handler/`")
	assert.NotContains(t, content, "NEVER import infrastructure packages from `domain/`")
}

func TestBuildContent_LayeredAddingCode(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "layered",
		Components:   nil,
	}

	content := buildContent(opts)

	assert.Contains(t, content, "## Adding Code")
	assert.Contains(t, content, "internal/model/")
	assert.Contains(t, content, "internal/repository/")
	assert.Contains(t, content, "internal/service/")
	assert.Contains(t, content, "internal/handler/router.go")
}

func TestBuildContent_Clean(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "clean",
		Components: []config.Component{
			{Name: "entity", In: []string{"internal/entity/**"}, MayDependOn: []string{}},
			{Name: "usecase", In: []string{"internal/usecase/**"}, MayDependOn: []string{"entity"}},
			{Name: "interface", In: []string{"internal/interface/**"}, MayDependOn: []string{"usecase", "entity"}},
			{Name: "infrastructure", In: []string{"internal/infrastructure/**"}, MayDependOn: []string{"interface", "usecase", "entity"}},
		},
	}

	content := buildContent(opts)

	assert.Contains(t, content, "Architecture: clean")
	assert.Contains(t, content, "Clean Architecture")
	assert.Contains(t, content, "## Layer Rules")
	assert.Contains(t, content, "entity")
	assert.Contains(t, content, "usecase")
	assert.Contains(t, content, "interface")
	assert.Contains(t, content, "infrastructure")
	assert.Contains(t, content, "Dependencies: none (innermost layer)")
	assert.Contains(t, content, "## Anti-patterns to Avoid")
	assert.Contains(t, content, "NEVER let `internal/entity/` import from usecase")
	assert.Contains(t, content, "NEVER let `internal/usecase/` import from infrastructure")
	assert.NotContains(t, content, "NEVER import infrastructure packages from `domain/`")
}

func TestBuildContent_CleanAddingCode(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "clean",
		Components:   nil,
	}

	content := buildContent(opts)

	assert.Contains(t, content, "## Adding Code")
	assert.Contains(t, content, "internal/entity/")
	assert.Contains(t, content, "internal/usecase/")
	assert.Contains(t, content, "internal/interface/handler/")
	assert.Contains(t, content, "internal/infrastructure/")
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

	// Write (monolithic) still works for CatalogOnly mode.
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

func TestGenerate_EmptyProjectDir(t *testing.T) {
	dir := t.TempDir()
	emptyDir := filepath.Join(dir, "empty")
	require.NoError(t, os.MkdirAll(emptyDir, 0o755))

	opts := GenerateOptions{
		ProjectDir:   emptyDir,
		Target:       "claude",
		Architecture: "flat",
	}

	err := Generate(opts)
	require.NoError(t, err)

	// Should create the index file even in an empty dir.
	path := filepath.Join(emptyDir, ".claude", "rules", "archway-index.md")
	assert.FileExists(t, path)
}

func TestGenerateFromConfig_NilComponents(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Capabilities: nil,
		Components:   nil,
	}

	err := GenerateFromConfig(dir, cfg, "claude")
	require.NoError(t, err)

	// Split output: index file should exist.
	path := filepath.Join(dir, ".claude", "rules", "archway-index.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "Architecture: flat")
}

func TestBuildContent_UnknownArchitecture(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "onion",
		Components:   nil,
	}

	content := buildContent(opts)
	assert.Contains(t, content, "Architecture type: onion")
}

func TestBuildContent_EmptyCapabilities(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "flat",
		Capabilities: []string{},
	}

	content := buildContent(opts)
	assert.Contains(t, content, "No capabilities configured")
}

func TestBuildContent_CatalogOnly(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{"http-api", "mysql"},
		CatalogOnly:  true,
	}

	content := buildContent(opts)

	// Should contain header and catalog-related sections.
	assert.Contains(t, content, "# Archway -- Architecture Guide")

	// Should NOT contain architecture-specific sections.
	assert.NotContains(t, content, "Architecture: hexagonal")
	assert.NotContains(t, content, "## Layer Rules")
	assert.NotContains(t, content, "## Dependency Direction")
	assert.NotContains(t, content, "## Adding Code")
	assert.NotContains(t, content, "## Capabilities")
	assert.NotContains(t, content, "## Anti-patterns to Avoid")
}

func TestBuildContent_CatalogOnlyFalse_FullOutput(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{"http-api"},
		CatalogOnly:  false,
	}

	content := buildContent(opts)

	assert.Contains(t, content, "Architecture: hexagonal")
	assert.Contains(t, content, "## Layer Rules")
	assert.Contains(t, content, "## Anti-patterns to Avoid")
	assert.Contains(t, content, "## Capabilities")
}

func TestWriteRuleSummaries_WithRules(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".archway", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	// Create a valid rule file.
	ruleYAML := `id: no-fmt-println
engine: grep
severity: warning
description: Do not use fmt.Println
pattern: "fmt\\.Println"
scope:
  - "**/*.go"
`
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "no-fmt-println.yaml"), []byte(ruleYAML), 0o644))

	// Create a .go file so scope is not stale.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644))

	var b strings.Builder
	writeRuleSummaries(&b, dir)

	content := b.String()
	assert.Contains(t, content, "## Active Rules")
	assert.Contains(t, content, "no-fmt-println")
	assert.Contains(t, content, "grep")
	assert.Contains(t, content, "warning")
	assert.Contains(t, content, "archway check")
}

func TestWriteRuleSummaries_NoRulesDir(t *testing.T) {
	dir := t.TempDir()

	var b strings.Builder
	writeRuleSummaries(&b, dir)

	assert.Empty(t, b.String())
}

func TestGuideIntegration_FullOutputWithCatalogAndRules(t *testing.T) {
	dir := t.TempDir()

	// Create .archway/rules/ with a proxy rule.
	rulesDir := filepath.Join(dir, ".archway", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))
	ruleYAML := `id: domain-isolation
engine: grep
severity: error
description: Domain must not import infrastructure
pattern: "infrastructure"
scope:
  - "domain/**/*.go"
`
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "domain-isolation.yaml"), []byte(ruleYAML), 0o644))
	// Create a file matching scope so rule isn't stale.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "domain"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "domain", "entity.go"), []byte("package domain\n"), 0o644))

	cfg := &config.ArchwayConfig{
		Architecture: "hexagonal",
		Capabilities: []string{"http-api", "mysql", "auth-jwt"},
		Components: []config.Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{}},
			{Name: "ports", In: []string{"port/**"}, MayDependOn: []string{"domain"}},
			{Name: "service", In: []string{"service/**"}, MayDependOn: []string{"domain", "ports"}},
			{Name: "adapters", In: []string{"adapter/**"}, MayDependOn: []string{"ports", "domain"}},
		},
	}

	err := GenerateFromConfig(dir, cfg, "claude")
	require.NoError(t, err)

	// Claude target now produces split files.
	indexData, err := os.ReadFile(filepath.Join(dir, ".claude", "rules", "archway-index.md"))
	require.NoError(t, err)
	indexContent := string(indexData)

	// Index has architecture summary.
	assert.Contains(t, indexContent, "Architecture: hexagonal")
	assert.Contains(t, indexContent, "## Layer Rules")
	assert.Contains(t, indexContent, "## Active Capabilities")
	assert.Contains(t, indexContent, "http-api")
	assert.Contains(t, indexContent, "mysql")
	assert.Contains(t, indexContent, "auth-jwt")

	// Rule summaries in index.
	assert.Contains(t, indexContent, "## Active Rules")
	assert.Contains(t, indexContent, "domain-isolation")

	// Category files exist.
	assert.FileExists(t, filepath.Join(dir, ".claude", "rules", "archway-http.md"))
	assert.FileExists(t, filepath.Join(dir, ".claude", "rules", "archway-data.md"))
	assert.FileExists(t, filepath.Join(dir, ".claude", "rules", "archway-security.md"))

	// Old monolithic file should not exist.
	assert.NoFileExists(t, filepath.Join(dir, ".claude", "rules", "archway.md"))

	// HTTP category has adding code and warnings.
	httpData, err := os.ReadFile(filepath.Join(dir, ".claude", "rules", "archway-http.md"))
	require.NoError(t, err)
	httpContent := string(httpData)
	assert.Contains(t, httpContent, "## Adding Code")
	assert.Contains(t, httpContent, "adapter/httphandler/")
}

func TestGuideIntegration_CatalogOnlyMode(t *testing.T) {
	dir := t.TempDir()

	opts := GenerateOptions{
		ProjectDir:   dir,
		Target:       "claude",
		Capabilities: []string{"http-api"},
		CatalogOnly:  true,
		TemplateFS:   testCapFS(),
	}

	err := Generate(opts)
	require.NoError(t, err)

	// CatalogOnly uses monolithic output even for Claude.
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "rules", "archway.md"))
	require.NoError(t, err)
	content := string(data)

	// Should have catalog.
	assert.Contains(t, content, "## Capability Catalog")

	// Should NOT have architecture sections.
	assert.NotContains(t, content, "## Layer Rules")
	assert.NotContains(t, content, "## Dependency Direction")
	assert.NotContains(t, content, "## Anti-patterns to Avoid")
}

func TestGuideTokenCompliance(t *testing.T) {
	opts := GenerateOptions{
		Architecture: "hexagonal",
		Capabilities: []string{
			"http-api", "mysql", "redis", "docker", "ci-github",
			"observability", "health", "cors", "auth-jwt", "rate-limiting",
		},
		TemplateFS: testCapFS(),
		Components: []config.Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{}},
			{Name: "ports", In: []string{"port/**"}, MayDependOn: []string{"domain"}},
			{Name: "service", In: []string{"service/**"}, MayDependOn: []string{"domain", "ports"}},
			{Name: "adapters", In: []string{"adapter/**"}, MayDependOn: []string{"ports", "domain"}},
		},
	}

	content := buildContent(opts)
	words := len(strings.Fields(content))
	approxTokens := int(float64(words) * 1.3)

	// Guide should stay under 2000 tokens (INV-001 upper bound for rules files is 1500,
	// but guide is a generated composite so we allow up to 2000).
	if approxTokens > 2000 {
		t.Errorf("guide output too large: ~%d tokens (%d words); consider trimming", approxTokens, words)
	}
}

func TestForbiddenDeps_ComponentWithNoForbidden(t *testing.T) {
	// Component that may depend on everything else.
	comp := config.Component{
		Name:        "adapters",
		MayDependOn: []string{"domain", "ports", "service"},
	}
	all := []config.Component{
		{Name: "domain"},
		{Name: "ports"},
		{Name: "service"},
		{Name: "adapters"},
	}

	result := forbiddenDeps(comp, all)
	assert.Empty(t, result)
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
		filepath.Join(dir, ".claude", "rules", "archway-index.md"),
		filepath.Join(dir, ".cursorrules"),
		filepath.Join(dir, ".github", "copilot-instructions.md"),
		filepath.Join(dir, ".windsurfrules"),
	}
	for _, path := range expected {
		assert.FileExists(t, path)
	}
}
