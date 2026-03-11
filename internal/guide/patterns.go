package guide

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

// capabilityTemplateMap maps capability names to their template file paths
// (relative to templates/capabilities/<capability>/files/).
var capabilityTemplateMap = map[string][]string{
	"http-api":       {"adapter/httphandler/handler.go.tmpl"},
	"mysql":          {"adapter/mysqlrepo/connection.go.tmpl"},
	"postgres":       {"adapter/pgxrepo/connection.go.tmpl"},
	"grpc":           {"adapter/grpchandler/server.go.tmpl"},
	"kafka-consumer": {"adapter/kafkahandler/consumer.go.tmpl"},
	"redis":          {"adapter/redisrepo/connection.go.tmpl"},
	"bff":            {"adapter/bffgateway/gateway.go.tmpl"},
	"ddd":            {"domain/aggregate.go.tmpl", "domain/event.go.tmpl"},
	"graceful":       {"internal/lifecycle/shutdown.go.tmpl"},
	"nats":           {"adapter/natshandler/subscriber.go.tmpl"},
	"elasticsearch":  {"adapter/esrepo/client.go.tmpl"},
	"timeout":        {"adapter/httphandler/timeout.go.tmpl"},
	"bulkhead":       {"platform/resilience/bulkhead.go.tmpl"},
	"oauth2":         {"adapter/httphandler/oauth2.go.tmpl"},
	"encryption":     {"platform/security/encryption.go.tmpl"},
}

// patternLabel returns a human-readable label for a capability pattern.
var patternLabel = map[string]string{
	"http-api":       "HTTP Handler Pattern",
	"mysql":          "MySQL Repository Pattern",
	"postgres":       "PostgreSQL Repository Pattern",
	"grpc":           "gRPC Server Pattern",
	"kafka-consumer": "Kafka Consumer Pattern",
	"redis":          "Redis Repository Pattern",
	"bff":            "BFF Gateway Pattern",
	"ddd":            "DDD Building Blocks",
	"graceful":       "Graceful Shutdown Pattern",
	"nats":           "NATS Messaging Pattern",
	"elasticsearch":  "Elasticsearch Client Pattern",
	"timeout":        "Timeout Middleware Pattern",
	"bulkhead":       "Bulkhead Pattern",
	"oauth2":         "OAuth2 Flow Pattern",
	"encryption":     "Encryption Pattern",
}

// ExtractPatterns reads template files from the embedded FS for the given
// capabilities and returns formatted pattern guidance.
func ExtractPatterns(templateFS fs.FS, capabilities []string) string {
	if templateFS == nil || len(capabilities) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Patterns\n\n")
	b.WriteString("Follow these patterns extracted from the project's templates.\n\n")

	found := false
	for _, cap := range capabilities {
		files, ok := capabilityTemplateMap[cap]
		if !ok {
			continue
		}
		label := patternLabel[cap]
		if label == "" {
			label = cap + " Pattern"
		}

		for _, relPath := range files {
			fullPath := path.Join("templates", "capabilities", cap, "files", relPath)
			content, err := fs.ReadFile(templateFS, fullPath)
			if err != nil {
				continue
			}

			stripped := StripTemplateDirectives(string(content))
			stripped = strings.TrimSpace(stripped)
			if stripped == "" {
				continue
			}

			found = true
			fmt.Fprintf(&b, "### %s\n\n", label)
			fmt.Fprintf(&b, "Source: `%s`\n\n", strings.TrimSuffix(relPath, ".tmpl"))
			b.WriteString("```go\n")
			b.WriteString(stripped)
			b.WriteString("\n```\n\n")
		}
	}

	if !found {
		return ""
	}

	return b.String()
}

// reBlockDirective matches {{ if ... }}...{{ end }} and {{ range ... }}...{{ end }} blocks.
var reBlockDirective = regexp.MustCompile(`(?m)^\s*\{\{-?\s*(?:if|else|range|end|block|define|with).*?\}\}\s*\n?`)

// reInlineDirective matches inline template expressions like {{ .ServiceName }}.
var reInlineDirective = regexp.MustCompile(`\{\{-?\s*\.(\w+)\s*-?\}\}`)

// reTemplateFuncCall matches template function calls like {{ printf ... }}.
var reTemplateFuncCall = regexp.MustCompile(`\{\{-?\s*\w+\s+.*?-?\}\}`)

// StripTemplateDirectives removes Go template directives from content,
// leaving clean Go code with placeholders for template variables.
func StripTemplateDirectives(content string) string {
	// Remove block directives (if/else/range/end/with/block/define).
	result := reBlockDirective.ReplaceAllString(content, "")

	// Replace inline variable references with example placeholders.
	result = reInlineDirective.ReplaceAllStringFunc(result, func(match string) string {
		sub := reInlineDirective.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		return "<" + sub[1] + ">"
	})

	// Remove remaining template function calls.
	result = reTemplateFuncCall.ReplaceAllString(result, "")

	// Collapse multiple blank lines into at most two.
	reMultiBlank := regexp.MustCompile(`\n{3,}`)
	result = reMultiBlank.ReplaceAllString(result, "\n\n")

	return result
}
