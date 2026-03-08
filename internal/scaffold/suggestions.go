package scaffold

// Suggestion represents a capability the user might want to add.
type Suggestion struct {
	Capability string
	Reason     string
}

// SuggestionRule defines when to suggest a capability.
type SuggestionRule struct {
	// IfAny triggers the rule when any of these capabilities are selected.
	IfAny []string
	// Missing is the capability to suggest (skipped if already selected).
	Missing string
	// Reason explains why the suggestion is relevant.
	Reason string
}

var defaultSuggestionRules = []SuggestionRule{
	{IfAny: []string{"http-api", "grpc", "kafka-consumer", "mysql", "redis"}, Missing: "platform", Reason: "Production services need config, logging, and lifecycle management"},
	{IfAny: []string{"platform"}, Missing: "bootstrap", Reason: "Bootstrap pattern provides testable dependency wiring with thin main.go"},
	{IfAny: []string{"http-api"}, Missing: "rate-limiting", Reason: "HTTP APIs benefit from rate limiting to prevent abuse"},
	{IfAny: []string{"http-api"}, Missing: "auth-jwt", Reason: "HTTP APIs typically need authentication"},
	{IfAny: []string{"http-api"}, Missing: "testing", Reason: "APIs need handler tests for reliability"},
	{IfAny: []string{"mysql", "redis"}, Missing: "docker", Reason: "docker-compose simplifies local development with external dependencies"},
	{IfAny: []string{"http-api", "grpc", "kafka-consumer"}, Missing: "ci-github", Reason: "CI/CD catches issues before they reach production"},
	{IfAny: []string{"http-api", "grpc", "kafka-consumer"}, Missing: "linting", Reason: "Linting catches code quality issues early"},
	{IfAny: []string{"http-api", "grpc"}, Missing: "docker", Reason: "Docker simplifies deployment and local development"},
}

// ComputeSuggestions returns capabilities the user might want based on their selections.
func ComputeSuggestions(selected []string) []Suggestion {
	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[s] = true
	}

	seen := map[string]bool{}
	var suggestions []Suggestion

	for _, rule := range defaultSuggestionRules {
		if selectedSet[rule.Missing] || seen[rule.Missing] {
			continue
		}
		for _, trigger := range rule.IfAny {
			if selectedSet[trigger] {
				seen[rule.Missing] = true
				suggestions = append(suggestions, Suggestion{
					Capability: rule.Missing,
					Reason:     rule.Reason,
				})
				break
			}
		}
	}
	return suggestions
}
