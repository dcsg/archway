package checker

import (
	"fmt"

	"github.com/dcsg/archway/internal/config"
)

// DecisionViolation represents an undecided architecture decision gate.
type DecisionViolation struct {
	Topic    string `json:"topic"`
	Tier     int    `json:"tier"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error" for Tier 1, "warning" for Tier 2
}

// CheckDecisions validates that all architecture decisions have been decided.
// Tier 1 undecided decisions are errors (blocking), Tier 2 are warnings.
func CheckDecisions(decisions []config.Decision) []DecisionViolation {
	violations := make([]DecisionViolation, 0, len(decisions))
	for _, d := range decisions {
		if d.Status == "decided" {
			continue
		}
		sev := "warning"
		if d.Tier == 1 {
			sev = "error"
		}
		violations = append(violations, DecisionViolation{
			Topic:    d.Topic,
			Tier:     d.Tier,
			Message:  fmt.Sprintf("%s (Tier %d): UNDECIDED", d.Topic, d.Tier),
			Severity: sev,
		})
	}
	return violations
}
