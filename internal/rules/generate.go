package rules

import (
	"fmt"

	"github.com/dcsg/archway/internal/config"
)

// GenerateRules creates proxy rules based on the project's architecture and capabilities.
func GenerateRules(cfg *config.ArchwayConfig) []Rule {
	if cfg == nil {
		return nil
	}

	archRules := generateArchRules(cfg)
	capRules := generateCapRules(cfg)
	rules := make([]Rule, 0, len(archRules)+len(capRules))
	rules = append(rules, archRules...)
	rules = append(rules, capRules...)
	return rules
}

// generateArchRules creates layer isolation rules based on the declared architecture.
func generateArchRules(cfg *config.ArchwayConfig) []Rule {
	switch cfg.Architecture {
	case "hexagonal":
		return hexagonalRules(cfg)
	case "flat":
		return nil
	default:
		return nil
	}
}

// hexagonalRules generates layer isolation rules from component dependency declarations.
func hexagonalRules(cfg *config.ArchwayConfig) []Rule {
	if len(cfg.Components) == 0 {
		return nil
	}

	allowed := buildAllowedMap(cfg.Components)
	allNames := componentNames(cfg.Components)
	rules := make([]Rule, 0, len(cfg.Components))

	for _, comp := range cfg.Components {
		forbidden := forbiddenDeps(comp.Name, allNames, allowed[comp.Name])
		if len(forbidden) == 0 {
			continue
		}

		rule := buildIsolationRule(comp, forbidden)
		rules = append(rules, rule)
	}

	return rules
}

// buildAllowedMap creates a map of component name to its allowed dependencies.
func buildAllowedMap(components []config.Component) map[string]map[string]bool {
	allowed := make(map[string]map[string]bool, len(components))
	for _, c := range components {
		m := make(map[string]bool, len(c.MayDependOn))
		for _, dep := range c.MayDependOn {
			m[dep] = true
		}
		allowed[c.Name] = m
	}
	return allowed
}

// componentNames extracts names from a slice of components.
func componentNames(components []config.Component) []string {
	names := make([]string, 0, len(components))
	for _, c := range components {
		names = append(names, c.Name)
	}
	return names
}

// forbiddenDeps returns component names that are not in the allowed set (excluding self).
func forbiddenDeps(self string, all []string, allowed map[string]bool) []string {
	var forbidden []string
	for _, name := range all {
		if name == self {
			continue
		}
		if !allowed[name] {
			forbidden = append(forbidden, name)
		}
	}
	return forbidden
}

// buildIsolationRule creates a grep rule that catches imports from forbidden layers.
func buildIsolationRule(comp config.Component, forbidden []string) Rule {
	pattern := buildForbiddenPattern(forbidden)
	return Rule{
		ID:          fmt.Sprintf("arch-%s-isolation", comp.Name),
		Engine:      "grep",
		Description: fmt.Sprintf("%s layer must not import from: %s", comp.Name, joinNames(forbidden)),
		Severity:    "error",
		Ref:         "archway.yaml",
		Pattern:     pattern,
		Scope:       scopeFromPaths(comp.In),
		Exclude:     []string{"*_test.go"},
	}
}

// buildForbiddenPattern creates a regex matching imports from any forbidden component.
func buildForbiddenPattern(forbidden []string) string {
	if len(forbidden) == 1 {
		return fmt.Sprintf(`"[^"]*/%s/[^"]*"`, forbidden[0])
	}
	return fmt.Sprintf(`"[^"]*/(%s)/[^"]*"`, joinNames(forbidden))
}

// scopeFromPaths converts component In paths to scope globs with Go file suffix.
func scopeFromPaths(paths []string) []string {
	scopes := make([]string, 0, len(paths))
	for _, p := range paths {
		scopes = append(scopes, p+"/*.go")
	}
	return scopes
}

// joinNames joins names with | for regex alternation.
func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	result := names[0]
	for _, n := range names[1:] {
		result += "|" + n
	}
	return result
}

// capabilityRuleMap defines rule templates for known capabilities.
var capabilityRuleMap = map[string]func() Rule{
	"postgres":       sqlRule,
	"mysql":          sqlRule,
	"http-api":       httpHandlerContextRule,
	"grpc":           grpcRule,
	"auth-jwt":       authRule,
	"observability":  observabilityRule,
	"kafka-consumer": kafkaRule,
}

// generateCapRules creates best-practice rules for each enabled capability.
func generateCapRules(cfg *config.ArchwayConfig) []Rule {
	seen := make(map[string]bool)
	rules := make([]Rule, 0, len(cfg.Capabilities))

	for _, cap := range cfg.Capabilities {
		fn, ok := capabilityRuleMap[cap]
		if !ok {
			continue
		}
		rule := fn()
		if seen[rule.ID] {
			continue
		}
		seen[rule.ID] = true
		rules = append(rules, rule)
	}

	return rules
}

func sqlRule() Rule {
	return Rule{
		ID:          "cap-sql-parameterized",
		Engine:      "grep",
		Description: "SQL queries must use parameterized queries, not string concatenation",
		Severity:    "error",
		Ref:         "archway.yaml",
		Pattern:     `(fmt\.Sprintf|"\+.*SELECT|"\+.*INSERT|"\+.*UPDATE|"\+.*DELETE)`,
		Scope:       []string{"**/*.go"},
		Exclude:     []string{"*_test.go", "vendor/**"},
	}
}

func httpHandlerContextRule() Rule {
	return Rule{
		ID:          "cap-handler-context",
		Engine:      "grep",
		Description: "HTTP handlers should use request context, not context.Background()",
		Severity:    "warning",
		Ref:         "archway.yaml",
		Pattern:     `context\.Background\(\)`,
		Scope:       []string{"adapter/httphandler/**/*.go", "internal/handler/**/*.go"},
		Exclude:     []string{"*_test.go"},
	}
}

func grpcRule() Rule {
	return Rule{
		ID:              "cap-grpc-proto",
		Engine:          "grep",
		Description:     "gRPC services should have proto file definitions",
		Severity:        "warning",
		Ref:             "archway.yaml",
		FileMustContain: `syntax\s*=\s*"proto3"`,
		Scope:           []string{"**/*.proto"},
	}
}

func authRule() Rule {
	return Rule{
		ID:              "cap-auth-check",
		Engine:          "grep",
		Description:     "Handler files should reference auth middleware or JWT validation",
		Severity:        "warning",
		Ref:             "archway.yaml",
		FileMustContain: `(middleware|auth|jwt|token)`,
		Scope:           []string{"adapter/httphandler/**/*.go", "internal/handler/**/*.go"},
		Exclude:         []string{"*_test.go", "**/router.go"},
	}
}

func observabilityRule() Rule {
	return Rule{
		ID:             "cap-tracing-context",
		Engine:         "grep",
		Description:    "Functions making external calls should pass context for tracing",
		Severity:       "warning",
		Ref:            "archway.yaml",
		Pattern:        `(http\.Get|http\.Post|sql\.Query|sql\.Exec)`,
		MustNotContain: "ctx",
		Scope:          []string{"**/*.go"},
		Exclude:        []string{"*_test.go", "vendor/**"},
	}
}

func kafkaRule() Rule {
	return Rule{
		ID:          "cap-kafka-error-handling",
		Engine:      "grep",
		Description: "Kafka consumers must handle errors, not silently discard them",
		Severity:    "error",
		Ref:         "archway.yaml",
		Pattern:     `\.ReadMessage\(`,
		MustContain: `(err|error|Error)`,
		Scope:       []string{"**/*.go"},
		Exclude:     []string{"*_test.go", "vendor/**"},
	}
}
