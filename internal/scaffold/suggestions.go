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
	{IfAny: []string{"postgres", "mysql"}, Missing: "uuid", Reason: "UUIDv7 provides database-friendly primary keys without index fragmentation"},
	{IfAny: []string{"postgres", "mysql"}, Missing: "migrations", Reason: "Database schema changes should be versioned and reproducible"},
	{IfAny: []string{"event-bus"}, Missing: "outbox", Reason: "Transactional outbox prevents event loss on process crash"},
	{IfAny: []string{"http-client"}, Missing: "circuit-breaker", Reason: "Circuit breakers prevent cascade failures from flaky external services"},
	{IfAny: []string{"http-client"}, Missing: "retry", Reason: "Retry with backoff handles transient failures gracefully"},
	{IfAny: []string{"http-api"}, Missing: "cors", Reason: "Browser-facing APIs require CORS headers"},
	{IfAny: []string{"http-api"}, Missing: "health", Reason: "Health endpoints enable orchestrator readiness probes"},
	{IfAny: []string{"http-api", "grpc", "kafka-consumer"}, Missing: "observability", Reason: "Distributed tracing and metrics for production debugging"},
	{IfAny: []string{"http-api", "grpc"}, Missing: "request-id", Reason: "Request ID propagation enables end-to-end request tracing"},

	// Transport — GraphQL
	{IfAny: []string{"graphql"}, Missing: "auth-jwt", Reason: "GraphQL APIs need authentication"},
	{IfAny: []string{"graphql"}, Missing: "observability", Reason: "GraphQL resolvers benefit from tracing"},

	// Data — NoSQL and embedded databases
	{IfAny: []string{"mongodb", "dynamodb"}, Missing: "health", Reason: "Database connections need health checks"},
	{IfAny: []string{"mongodb", "sqlite", "dynamodb"}, Missing: "config", Reason: "Database connections need configuration"},

	// Patterns
	{IfAny: []string{"saga"}, Missing: "observability", Reason: "Distributed transactions need tracing"},
	{IfAny: []string{"multi-tenancy"}, Missing: "auth-jwt", Reason: "Tenant isolation requires authentication"},
	{IfAny: []string{"feature-flags"}, Missing: "observability", Reason: "Track feature flag usage"},

	// Frontend
	{IfAny: []string{"templ"}, Missing: "htmx", Reason: "templ + HTMX is the standard Go full-stack pattern"},
	{IfAny: []string{"templ"}, Missing: "static-assets", Reason: "Server-rendered apps need CSS/JS serving"},
	{IfAny: []string{"htmx"}, Missing: "static-assets", Reason: "HTMX needs the htmx.js library served"},
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

// CapabilityWarning represents a potential issue with the selected capability combination.
type CapabilityWarning struct {
	Message string
}

// warningRule defines when to warn about a capability combination.
type warningRule struct {
	ifHas   string   // capability that triggers the check
	missing []string // warn if NONE of these are selected
	message string
}

var defaultWarningRules = []warningRule{
	{ifHas: "postgres", missing: []string{"uuid"}, message: "postgres without uuid: Consider UUIDv7 for database-friendly primary keys (avoids index fragmentation)"},
	{ifHas: "mysql", missing: []string{"uuid"}, message: "mysql without uuid: Consider UUIDv7 for database-friendly primary keys (avoids index fragmentation)"},
	{ifHas: "http-api", missing: []string{"health"}, message: "http-api without health: Production APIs should have /healthz and /readyz endpoints"},
	{ifHas: "event-bus", missing: []string{"outbox"}, message: "event-bus without outbox: Without transactional outbox, events can be lost if the process crashes"},
	{ifHas: "http-api", missing: []string{"cors"}, message: "http-api without cors: If this API serves browser clients, CORS headers are required"},
	{ifHas: "http-api", missing: []string{"observability", "request-id"}, message: "http-api without observability: Consider adding tracing and request ID propagation for debugging"},
	{ifHas: "http-client", missing: []string{"circuit-breaker", "retry"}, message: "http-client without resilience: External calls should have circuit breakers or retry logic"},
	{ifHas: "kafka-consumer", missing: []string{"health"}, message: "kafka-consumer without health: Consumers need health checks for orchestrator readiness probes"},
	{ifHas: "grpc", missing: []string{"health"}, message: "grpc without health: gRPC services should implement the health checking protocol"},
	{ifHas: "multi-tenancy", missing: []string{"auth-jwt"}, message: "multi-tenancy without auth-jwt: Multi-tenancy without authentication risks tenant data leaks"},
	{ifHas: "saga", missing: []string{"observability"}, message: "saga without observability: Sagas without tracing make distributed debugging very difficult"},
}

// CapabilityWarnings returns warnings about potentially problematic capability combinations.
func CapabilityWarnings(selected []string) []CapabilityWarning {
	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[s] = true
	}

	var warnings []CapabilityWarning
	for _, rule := range defaultWarningRules {
		if !selectedSet[rule.ifHas] {
			continue
		}
		// Warn only if NONE of the missing capabilities are selected.
		hasSome := false
		for _, m := range rule.missing {
			if selectedSet[m] {
				hasSome = true
				break
			}
		}
		if !hasSome {
			warnings = append(warnings, CapabilityWarning{Message: rule.message})
		}
	}
	return warnings
}
