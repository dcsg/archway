package checker_test

import (
	"testing"

	"github.com/dcsg/archway/internal/checker"
	"github.com/dcsg/archway/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCheckDecisions_AllDecided(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "architecture-pattern", Tier: 1, Status: "decided", Choice: "hexagonal"},
		{Topic: "authentication-strategy", Tier: 1, Status: "decided", Choice: "jwt"},
	}

	violations := checker.CheckDecisions(decisions)

	assert.Empty(t, violations)
}

func TestCheckDecisions_Tier1Undecided(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "tenant-isolation", Tier: 1, Status: "pending"},
	}

	violations := checker.CheckDecisions(decisions)

	assert.Len(t, violations, 1)
	assert.Equal(t, "tenant-isolation", violations[0].Topic)
	assert.Equal(t, 1, violations[0].Tier)
	assert.Equal(t, "error", violations[0].Severity)
	assert.Contains(t, violations[0].Message, "UNDECIDED")
}

func TestCheckDecisions_Tier2Undecided(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "failure-strategy", Tier: 2, Status: "pending"},
	}

	violations := checker.CheckDecisions(decisions)

	assert.Len(t, violations, 1)
	assert.Equal(t, "failure-strategy", violations[0].Topic)
	assert.Equal(t, 2, violations[0].Tier)
	assert.Equal(t, "warning", violations[0].Severity)
}

func TestCheckDecisions_Mixed(t *testing.T) {
	decisions := []config.Decision{
		{Topic: "architecture-pattern", Tier: 1, Status: "decided", Choice: "hexagonal"},
		{Topic: "tenant-isolation", Tier: 1, Status: "pending"},
		{Topic: "failure-strategy", Tier: 2, Status: "pending"},
		{Topic: "auth-strategy", Tier: 1, Status: "decided", Choice: "jwt"},
	}

	violations := checker.CheckDecisions(decisions)

	assert.Len(t, violations, 2)

	// Tier 1 undecided should be error.
	assert.Equal(t, "tenant-isolation", violations[0].Topic)
	assert.Equal(t, "error", violations[0].Severity)

	// Tier 2 undecided should be warning.
	assert.Equal(t, "failure-strategy", violations[1].Topic)
	assert.Equal(t, "warning", violations[1].Severity)
}

func TestCheckDecisions_Empty(t *testing.T) {
	violations := checker.CheckDecisions(nil)

	assert.Empty(t, violations)
}
