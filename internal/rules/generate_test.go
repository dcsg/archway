package rules

import (
	"fmt"
	"testing"

	"github.com/dcsg/archway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRules_NilConfig(t *testing.T) {
	rules := GenerateRules(nil)
	assert.Empty(t, rules)
}

func TestGenerateRules_EmptyConfig(t *testing.T) {
	cfg := &config.ArchwayConfig{}
	rules := GenerateRules(cfg)
	assert.Empty(t, rules)
}

func TestGenerateRules_FlatArchitecture(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
	}
	rules := GenerateRules(cfg)
	assert.Empty(t, rules)
}

func TestGenerateRules_HexagonalArchitecture(t *testing.T) {
	cfg := config.DefaultArchwayConfig("go", "hexagonal")
	rules := GenerateRules(cfg)

	require.NotEmpty(t, rules)

	rulesByID := indexByID(rules)

	// Domain should have isolation rule (domain may_depend_on is empty, so all others are forbidden).
	domainRule, ok := rulesByID["arch-domain-isolation"]
	require.True(t, ok, "expected arch-domain-isolation rule")
	assert.Equal(t, "error", domainRule.Severity)
	assert.Equal(t, "grep", domainRule.Engine)
	assert.NotEmpty(t, domainRule.Pattern)
	assert.Equal(t, []string{"domain/**/*.go"}, domainRule.Scope)
}

func TestGenerateRules_HexagonalComponentDeps(t *testing.T) {
	cfg := config.DefaultArchwayConfig("go", "hexagonal")
	rules := GenerateRules(cfg)
	rulesByID := indexByID(rules)

	// Service may depend on domain and ports — should NOT have those as forbidden.
	serviceRule, ok := rulesByID["arch-service-isolation"]
	require.True(t, ok, "expected arch-service-isolation rule")
	assert.NotContains(t, serviceRule.Pattern, "domain")
	assert.NotContains(t, serviceRule.Pattern, "ports")
	// Should contain adapters and platform as forbidden.
	assert.Contains(t, serviceRule.Pattern, "adapters")
	assert.Contains(t, serviceRule.Pattern, "platform")
}

func TestGenerateRules_PostgresCapability(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
		Capabilities: []string{"postgres"},
	}
	rules := GenerateRules(cfg)

	require.Len(t, rules, 1)
	assert.Equal(t, "cap-sql-parameterized", rules[0].ID)
	assert.Equal(t, "error", rules[0].Severity)
}

func TestGenerateRules_MysqlCapability(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
		Capabilities: []string{"mysql"},
	}
	rules := GenerateRules(cfg)

	require.Len(t, rules, 1)
	assert.Equal(t, "cap-sql-parameterized", rules[0].ID)
}

func TestGenerateRules_DuplicateSQLCapability(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
		Capabilities: []string{"postgres", "mysql"},
	}
	rules := GenerateRules(cfg)

	// Both produce the same rule ID — should deduplicate.
	require.Len(t, rules, 1)
	assert.Equal(t, "cap-sql-parameterized", rules[0].ID)
}

func TestGenerateRules_HTTPAPICapability(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
		Capabilities: []string{"http-api"},
	}
	rules := GenerateRules(cfg)

	require.Len(t, rules, 1)
	assert.Equal(t, "cap-handler-context", rules[0].ID)
	assert.Equal(t, "warning", rules[0].Severity)
}

func TestGenerateRules_GRPCCapability(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
		Capabilities: []string{"grpc"},
	}
	rules := GenerateRules(cfg)

	require.Len(t, rules, 1)
	assert.Equal(t, "cap-grpc-proto", rules[0].ID)
	assert.NotEmpty(t, rules[0].FileMustContain)
}

func TestGenerateRules_AuthJWTCapability(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
		Capabilities: []string{"auth-jwt"},
	}
	rules := GenerateRules(cfg)

	require.Len(t, rules, 1)
	assert.Equal(t, "cap-auth-check", rules[0].ID)
}

func TestGenerateRules_ObservabilityCapability(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
		Capabilities: []string{"observability"},
	}
	rules := GenerateRules(cfg)

	require.Len(t, rules, 1)
	assert.Equal(t, "cap-tracing-context", rules[0].ID)
	assert.NotEmpty(t, rules[0].MustNotContain)
}

func TestGenerateRules_KafkaCapability(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
		Capabilities: []string{"kafka-consumer"},
	}
	rules := GenerateRules(cfg)

	require.Len(t, rules, 1)
	assert.Equal(t, "cap-kafka-error-handling", rules[0].ID)
	assert.NotEmpty(t, rules[0].MustContain)
}

func TestGenerateRules_UnknownCapabilityIgnored(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Language:     "go",
		Capabilities: []string{"unknown-cap"},
	}
	rules := GenerateRules(cfg)
	assert.Empty(t, rules)
}

func TestGenerateRules_CombinedArchAndCaps(t *testing.T) {
	cfg := config.DefaultArchwayConfig("go", "hexagonal")
	cfg.Capabilities = []string{"postgres", "http-api"}
	rules := GenerateRules(cfg)

	rulesByID := indexByID(rules)

	// Should have arch rules.
	assert.Contains(t, rulesByID, "arch-domain-isolation")

	// Should have cap rules.
	assert.Contains(t, rulesByID, "cap-sql-parameterized")
	assert.Contains(t, rulesByID, "cap-handler-context")
}

func TestGenerateRules_AllRulesAreValid(t *testing.T) {
	cfg := config.DefaultArchwayConfig("go", "hexagonal")
	cfg.Capabilities = []string{"postgres", "http-api", "grpc", "auth-jwt", "observability", "kafka-consumer"}
	rules := GenerateRules(cfg)

	for _, r := range rules {
		status := ValidateRule(r, r.ID+".yaml", "")
		assert.Equalf(t, "valid", status.Status, "rule %s invalid: %s", r.ID, status.Error)
	}
}

func TestGenerateRules_ComponentWithEmptyName(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "hexagonal",
		Components: []config.Component{
			{Name: "", In: []string{"pkg/**"}, MayDependOn: []string{}},
			{Name: "service", In: []string{"service/**"}, MayDependOn: []string{}},
		},
	}
	rules := GenerateRules(cfg)

	// Should still generate rules (empty name is treated as any other component).
	require.NotEmpty(t, rules)
}

func TestGenerateRules_MayDependOnSelf(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "hexagonal",
		Components: []config.Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{"domain"}},
			{Name: "service", In: []string{"service/**"}, MayDependOn: []string{}},
		},
	}
	rules := GenerateRules(cfg)
	rulesByID := indexByID(rules)

	// domain's forbidden list should contain service but not itself.
	domainRule, ok := rulesByID["arch-domain-isolation"]
	require.True(t, ok)
	assert.Contains(t, domainRule.Pattern, "service")
	// The forbidden names (in joinNames) should not include "domain" itself.
	// Pattern uses forbidden names, so "domain" should not appear in the regex pattern.
	assert.NotContains(t, domainRule.Pattern, "/domain/")
}

func TestGenerateRules_MayDependOnNonExistentComponent(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "hexagonal",
		Components: []config.Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{"nonexistent"}},
			{Name: "service", In: []string{"service/**"}, MayDependOn: []string{}},
		},
	}
	rules := GenerateRules(cfg)

	// Should still work; domain allows "nonexistent" (which doesn't exist) but
	// service is still forbidden.
	rulesByID := indexByID(rules)
	domainRule, ok := rulesByID["arch-domain-isolation"]
	require.True(t, ok)
	assert.Contains(t, domainRule.Pattern, "service")
}

func TestGenerateRules_LargeForbiddenList(t *testing.T) {
	components := make([]config.Component, 12)
	for i := range components {
		components[i] = config.Component{
			Name: fmt.Sprintf("layer%d", i),
			In:   []string{fmt.Sprintf("layer%d/**", i)},
		}
	}
	// First component depends on nothing, so all 11 others are forbidden.
	cfg := &config.ArchwayConfig{
		Architecture: "hexagonal",
		Components:   components,
	}
	rules := GenerateRules(cfg)
	rulesByID := indexByID(rules)

	rule, ok := rulesByID["arch-layer0-isolation"]
	require.True(t, ok)
	// Pattern should contain alternation of all 11 forbidden layers.
	for i := 1; i < 12; i++ {
		assert.Contains(t, rule.Pattern, fmt.Sprintf("layer%d", i))
	}
}

func TestGenerateRules_DuplicateCapabilities(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Architecture: "flat",
		Capabilities: []string{"postgres", "postgres"},
	}
	rules := GenerateRules(cfg)

	// Deduplication via seen map: should produce exactly 1 rule.
	require.Len(t, rules, 1)
	assert.Equal(t, "cap-sql-parameterized", rules[0].ID)
}

func TestBuildForbiddenPattern_EmptyList(t *testing.T) {
	// Defensive: buildForbiddenPattern shouldn't be called with empty list,
	// but test that it doesn't panic.
	pattern := buildForbiddenPattern([]string{})
	assert.NotEmpty(t, pattern) // Returns pattern with empty alternation.
}

func TestJoinNames_SingleName(t *testing.T) {
	result := joinNames([]string{"domain"})
	assert.Equal(t, "domain", result)
}

func TestJoinNames_EmptySlice(t *testing.T) {
	result := joinNames([]string{})
	assert.Equal(t, "", result)
}

func TestScopeFromPaths_EmptyPaths(t *testing.T) {
	result := scopeFromPaths([]string{})
	assert.Empty(t, result)
}

func TestGenerateCapRules_UnknownCapabilitiesOnly(t *testing.T) {
	cfg := &config.ArchwayConfig{
		Capabilities: []string{"unknown1", "unknown2", "unknown3"},
	}
	rules := generateCapRules(cfg)
	assert.Empty(t, rules)
}

func indexByID(rules []Rule) map[string]Rule {
	m := make(map[string]Rule, len(rules))
	for _, r := range rules {
		m[r.ID] = r
	}
	return m
}
