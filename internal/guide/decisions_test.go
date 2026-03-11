package guide

import (
	"strings"
	"testing"

	"github.com/dcsg/archway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestAutoPopulateDecisions_HexagonalWithAuth(t *testing.T) {
	decisions := AutoPopulateDecisions("hexagonal", []string{"http-api", "auth-jwt", "postgres"})

	byTopic := decisionMap(decisions)

	// Architecture should be auto-decided.
	arch := byTopic["architecture-pattern"]
	assert.Equal(t, "decided", arch.Status)
	assert.Equal(t, "hexagonal", arch.Choice)

	// Auth should be auto-decided as jwt.
	auth := byTopic["authentication-strategy"]
	assert.Equal(t, "decided", auth.Status)
	assert.Equal(t, "jwt", auth.Choice)

	// Tenant isolation should be decided as not-applicable (no multi-tenancy cap).
	tenant := byTopic["tenant-isolation"]
	assert.Equal(t, "decided", tenant.Status)
	assert.Equal(t, "not-applicable", tenant.Choice)

	// DB capability present, so migration-strategy should be included.
	_, hasMigration := byTopic["migration-strategy"]
	assert.True(t, hasMigration, "migration-strategy should be included when DB capability exists")

	// failure-strategy should be included (http-api present).
	_, hasFailure := byTopic["failure-strategy"]
	assert.True(t, hasFailure)
}

func TestAutoPopulateDecisions_MinimalProject(t *testing.T) {
	decisions := AutoPopulateDecisions("flat", nil)

	// All Tier 1 should be present.
	byTopic := decisionMap(decisions)
	assert.Contains(t, byTopic, "architecture-pattern")
	assert.Contains(t, byTopic, "authentication-strategy")
	assert.Contains(t, byTopic, "data-model-ownership")
	assert.Contains(t, byTopic, "tenant-isolation")

	// Architecture auto-decided.
	assert.Equal(t, "decided", byTopic["architecture-pattern"].Status)
	assert.Equal(t, "flat", byTopic["architecture-pattern"].Choice)

	// Auth should be undecided.
	assert.Equal(t, "undecided", byTopic["authentication-strategy"].Status)

	// No DB capabilities, so no migration-strategy.
	_, hasMigration := byTopic["migration-strategy"]
	assert.False(t, hasMigration)
}

func TestResolveDecision_UpdatesExisting(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "authentication-strategy", Tier: 1, Status: "undecided"},
		{Topic: "data-model-ownership", Tier: 1, Status: "undecided"},
	}

	updated, err := ResolveDecision(decisions, "authentication-strategy", "jwt", "Industry standard", "alice")
	require.NoError(t, err)

	assert.Equal(t, "decided", updated[0].Status)
	assert.Equal(t, "jwt", updated[0].Choice)
	assert.Equal(t, "Industry standard", updated[0].Rationale)
	assert.Equal(t, "alice", updated[0].DecidedBy)

	// Original should be unchanged.
	assert.Equal(t, "undecided", decisions[0].Status)
}

func TestResolveDecision_UnknownTopic(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "authentication-strategy", Tier: 1, Status: "undecided"},
	}

	_, err := ResolveDecision(decisions, "nonexistent", "value", "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown decision topic")
}

func TestUndecidedDecisions_FiltersCorrectly(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "a", Status: "decided"},
		{Topic: "b", Status: "undecided"},
		{Topic: "c", Status: "decided"},
		{Topic: "d", Status: "undecided"},
	}

	undecided := UndecidedDecisions(decisions)
	assert.Len(t, undecided, 2)
	assert.Equal(t, "b", undecided[0].Topic)
	assert.Equal(t, "d", undecided[1].Topic)
}

func TestDecisionsByTier_GroupsCorrectly(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "a", Tier: 1},
		{Topic: "b", Tier: 2},
		{Topic: "c", Tier: 1},
		{Topic: "d", Tier: 2},
	}

	byTier := DecisionsByTier(decisions)
	assert.Len(t, byTier[1], 2)
	assert.Len(t, byTier[2], 2)
	assert.Equal(t, "a", byTier[1][0].Topic)
	assert.Equal(t, "c", byTier[1][1].Topic)
}

func TestWriteDecisionStatus_FormatsCorrectly(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "architecture-pattern", Tier: 1, Status: "decided", Choice: "hexagonal"},
		{Topic: "authentication-strategy", Tier: 1, Status: "undecided"},
		{Topic: "deployment-model", Tier: 2, Status: "decided", Choice: "kubernetes"},
	}

	var b strings.Builder
	writeDecisionStatus(&b, decisions)
	output := b.String()

	assert.Contains(t, output, "## Decision Status")
	assert.Contains(t, output, "Tier 1 (Foundational)")
	assert.Contains(t, output, "V architecture-pattern: hexagonal")
	assert.Contains(t, output, "X authentication-strategy: UNDECIDED")
	assert.Contains(t, output, "Tier 2 (Infrastructure)")
	assert.Contains(t, output, "V deployment-model: kubernetes")
	assert.Contains(t, output, "archway decide")
}

func TestWriteDecisionStatus_EmptyDecisions(t *testing.T) {
	var b strings.Builder
	writeDecisionStatus(&b, nil)
	assert.Empty(t, b.String())
}

func TestWriteDecisionStatus_AllDecided(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "architecture-pattern", Tier: 1, Status: "decided", Choice: "hexagonal"},
	}

	var b strings.Builder
	writeDecisionStatus(&b, decisions)
	output := b.String()

	assert.NotContains(t, output, "archway decide")
}

func TestDecisionConfigParsing(t *testing.T) {
	yamlData := `
language: go
architecture: hexagonal
components:
  - name: domain
    in: ["domain/**"]
    may_depend_on: []
decisions:
  - topic: architecture-pattern
    tier: 1
    status: decided
    choice: hexagonal
  - topic: authentication-strategy
    tier: 1
    status: undecided
`
	var cfg config.ArchwayConfig
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	require.NoError(t, err)

	assert.Len(t, cfg.Decisions, 2)
	assert.Equal(t, "architecture-pattern", cfg.Decisions[0].Topic)
	assert.Equal(t, "decided", cfg.Decisions[0].Status)
	assert.Equal(t, "hexagonal", cfg.Decisions[0].Choice)
	assert.Equal(t, "authentication-strategy", cfg.Decisions[1].Topic)
	assert.Equal(t, "undecided", cfg.Decisions[1].Status)
}

func decisionMap(decisions []config.Decision) map[string]config.Decision {
	m := make(map[string]config.Decision, len(decisions))
	for _, d := range decisions {
		m[d.Topic] = d
	}
	return m
}
