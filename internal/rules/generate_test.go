package rules

import (
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

func indexByID(rules []Rule) map[string]Rule {
	m := make(map[string]Rule, len(rules))
	for _, r := range rules {
		m[r.ID] = r
	}
	return m
}
