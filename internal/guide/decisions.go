package guide

import (
	"fmt"
	"strings"

	"github.com/dcsg/archway/internal/config"
)

// DecisionTemplate defines a standard architecture decision topic.
type DecisionTemplate struct {
	Topic    string
	Tier     int
	Question string
	Options  []string
}

var standardDecisions = []DecisionTemplate{
	// Tier 1 — Foundational
	{Topic: "architecture-pattern", Tier: 1, Question: "How is code organized?", Options: []string{"hexagonal", "layered", "clean", "flat"}},
	{Topic: "authentication-strategy", Tier: 1, Question: "How are users/services authenticated?", Options: []string{"jwt", "oauth2", "api-keys", "gateway-auth", "none"}},
	{Topic: "data-model-ownership", Tier: 1, Question: "Who owns the data?", Options: []string{"service-owned-db", "shared-db", "event-sourced"}},
	{Topic: "tenant-isolation", Tier: 1, Question: "How are tenants separated?", Options: []string{"row-level", "schema-per-tenant", "db-per-tenant", "not-applicable"}},

	// Tier 2 — Infrastructure
	{Topic: "failure-strategy", Tier: 2, Question: "What happens when external calls fail?", Options: []string{"fail-fast", "retry-with-backoff", "queue-and-retry", "graceful-degradation"}},
	{Topic: "migration-strategy", Tier: 2, Question: "How are DB migrations managed?", Options: []string{"golang-migrate", "goose", "atlas", "manual"}},
	{Topic: "deployment-model", Tier: 2, Question: "Where does this run?", Options: []string{"kubernetes", "ecs", "vm", "serverless"}},
	{Topic: "observability-stack", Tier: 2, Question: "How is the system observed?", Options: []string{"opentelemetry", "datadog", "cloudwatch", "prometheus-grafana"}},
}

// StandardDecisions returns the full standard decision catalog.
func StandardDecisions() []DecisionTemplate {
	result := make([]DecisionTemplate, len(standardDecisions))
	copy(result, standardDecisions)
	return result
}

// FindDecisionTemplate returns the template for a given topic.
func FindDecisionTemplate(topic string) (DecisionTemplate, bool) {
	for _, dt := range standardDecisions {
		if dt.Topic == topic {
			return dt, true
		}
	}
	return DecisionTemplate{}, false
}

// AutoPopulateDecisions generates decisions based on architecture and capabilities.
func AutoPopulateDecisions(arch string, capabilities []string) []config.Decision {
	capSet := make(map[string]bool, len(capabilities))
	for _, c := range capabilities {
		capSet[c] = true
	}

	hasDB := capSet["mysql"] || capSet["postgres"] || capSet["sqlite"] || capSet["mongodb"] || capSet["dynamodb"]

	decisions := make([]config.Decision, 0, len(standardDecisions))

	// All Tier 1 decisions are always included.
	for _, dt := range standardDecisions {
		if dt.Tier != 1 {
			continue
		}
		d := config.Decision{
			Topic:  dt.Topic,
			Tier:   dt.Tier,
			Status: "undecided",
		}

		// Auto-resolve from config.
		switch dt.Topic {
		case "architecture-pattern":
			if arch != "" {
				d.Status = "decided"
				d.Choice = arch
			}
		case "authentication-strategy":
			if capSet["auth-jwt"] {
				d.Status = "decided"
				d.Choice = "jwt"
			} else if capSet["auth-oauth2"] {
				d.Status = "decided"
				d.Choice = "oauth2"
			}
		case "tenant-isolation":
			if !capSet["multi-tenancy"] {
				d.Status = "decided"
				d.Choice = "not-applicable"
			}
		}

		decisions = append(decisions, d)
	}

	// Tier 2 decisions included when relevant capabilities exist.
	for _, dt := range standardDecisions {
		if dt.Tier != 2 {
			continue
		}

		include := false
		switch dt.Topic {
		case "migration-strategy":
			include = hasDB
		case "observability-stack":
			include = capSet["observability"]
		case "failure-strategy":
			// Relevant when there are external integrations.
			include = capSet["http-api"] || capSet["grpc"] || capSet["kafka"] || capSet["kafka-consumer"] || hasDB
		case "deployment-model":
			include = capSet["docker"] || capSet["ci-github"] || capSet["ci-gitlab"]
		default:
			include = true
		}

		if !include {
			continue
		}

		decisions = append(decisions, config.Decision{
			Topic:  dt.Topic,
			Tier:   dt.Tier,
			Status: "undecided",
		})
	}

	return decisions
}

// ResolveDecision updates a decision with the given choice.
func ResolveDecision(decisions []config.Decision, topic, choice, rationale, decidedBy string) ([]config.Decision, error) {
	found := false
	result := make([]config.Decision, len(decisions))
	copy(result, decisions)

	for i := range result {
		if result[i].Topic == topic {
			found = true
			result[i].Status = "decided"
			result[i].Choice = choice
			result[i].Rationale = rationale
			result[i].DecidedBy = decidedBy
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("unknown decision topic: %q", topic)
	}

	return result, nil
}

// UndecidedDecisions returns decisions that haven't been resolved.
func UndecidedDecisions(decisions []config.Decision) []config.Decision {
	var result []config.Decision
	for _, d := range decisions {
		if d.Status != "decided" {
			result = append(result, d)
		}
	}
	return result
}

// DecisionsByTier groups decisions by tier.
func DecisionsByTier(decisions []config.Decision) map[int][]config.Decision {
	result := make(map[int][]config.Decision)
	for _, d := range decisions {
		result[d.Tier] = append(result[d.Tier], d)
	}
	return result
}

// tierName returns a human-readable name for a decision tier.
func tierName(tier int) string {
	switch tier {
	case 1:
		return "Foundational"
	case 2:
		return "Infrastructure"
	default:
		return fmt.Sprintf("Tier %d", tier)
	}
}

// writeDecisionStatus writes the decision status section to the guide.
func writeDecisionStatus(b *strings.Builder, decisions []config.Decision) {
	if len(decisions) == 0 {
		return
	}

	b.WriteString("## Decision Status\n\n")

	byTier := DecisionsByTier(decisions)

	// Output tiers in order.
	for tier := 1; tier <= 2; tier++ {
		tierDecisions, ok := byTier[tier]
		if !ok {
			continue
		}

		fmt.Fprintf(b, "**Tier %d (%s):**\n", tier, tierName(tier))
		for _, d := range tierDecisions {
			if d.Status == "decided" {
				fmt.Fprintf(b, "- V %s: %s\n", d.Topic, d.Choice)
			} else {
				fmt.Fprintf(b, "- X %s: UNDECIDED\n", d.Topic)
			}
		}
		b.WriteString("\n")
	}

	undecided := UndecidedDecisions(decisions)
	if len(undecided) > 0 {
		b.WriteString("**Action:** Run `archway decide` to resolve open decisions.\n")
		b.WriteString("**Rule:** Do NOT implement features that depend on undecided Tier 1 topics.\n\n")
	}
}
