package guide

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/dcsg/archway/internal/scaffold"
)

// CatalogEntry describes a single capability in the catalog.
type CatalogEntry struct {
	Name        string
	Category    string
	Description string
	WhenToUse   string
	Installed   bool
	Suggests    []string
}

// capabilityCategories maps each capability to its category.
var capabilityCategories = map[string]string{
	// Transport
	"http-api":       "transport",
	"grpc":           "transport",
	"graphql":        "transport",
	"websocket":      "transport",
	"kafka-consumer": "transport",
	"sse":            "transport",
	"bff":            "transport",
	// Data
	"mysql":      "data",
	"postgres":   "data",
	"redis":      "data",
	"mongodb":    "data",
	"sqlite":     "data",
	"dynamodb":   "data",
	"s3":         "data",
	"repository": "data",
	"migrations": "data",
	// Resilience
	"circuit-breaker": "resilience",
	"retry":           "resilience",
	"rate-limiting":   "resilience",
	"idempotency":     "resilience",
	// Patterns
	"event-bus": "patterns",
	"cqrs":      "patterns",
	"outbox":    "patterns",
	"saga":      "patterns",
	"worker":    "patterns",
	"scheduler": "patterns",
	// Security
	"auth-jwt":      "security",
	"cors":          "security",
	"audit-log":     "security",
	"multi-tenancy": "security",
	"feature-flags": "security",
	// Observability
	"observability": "observability",
	"health":        "observability",
	"request-id":    "observability",
	// Infrastructure
	"docker":       "infrastructure",
	"ci-github":    "infrastructure",
	"ci-gitlab":    "infrastructure",
	"makefile":     "infrastructure",
	"devcontainer": "infrastructure",
	"pre-commit":   "infrastructure",
	// Quality
	"linting":    "quality",
	"testing":    "quality",
	"validation": "quality",
	"uuid":       "quality",
	// Platform
	"platform":  "platform",
	"bootstrap": "platform",
	// Frontend
	"templ":         "frontend",
	"htmx":          "frontend",
	"static-assets": "frontend",
	// Other
	"http-client":    "transport",
	"email-gateway":  "patterns",
	"mailpit":        "infrastructure",
	"i18n":           "patterns",
	"api-versioning": "transport",
}

// whenToUse provides guidance on when each capability is appropriate.
var whenToUse = map[string]string{
	"http-api":        "REST APIs, webhooks, web applications",
	"grpc":            "Internal microservice communication, streaming, high-performance RPC",
	"graphql":         "Client-facing APIs with complex querying needs, mobile apps",
	"websocket":       "Real-time bidirectional communication (chat, live updates)",
	"kafka-consumer":  "Event-driven architectures, async message processing",
	"sse":             "Server-to-client streaming (live feeds, notifications)",
	"bff":             "Backend-for-Frontend gateway (service aggregation, response shaping)",
	"mysql":           "Relational data with SQL queries, transactional workloads",
	"postgres":        "Advanced relational data, JSONB, full-text search",
	"redis":           "Caching, session storage, pub/sub, rate limiting backend",
	"mongodb":         "Document-oriented data, flexible schemas, rapid prototyping",
	"sqlite":          "Embedded databases, single-binary deployments, testing",
	"dynamodb":        "Serverless workloads, key-value at scale, AWS-native apps",
	"s3":              "Object storage, file uploads, static asset hosting",
	"repository":      "Data access abstraction, testable persistence layer",
	"migrations":      "Versioned database schema changes, reproducible deployments",
	"circuit-breaker": "External service calls that may fail or slow down",
	"retry":           "Transient failures from external services or networks",
	"rate-limiting":   "Public APIs needing abuse prevention",
	"idempotency":     "Operations that must be safe to retry (payments, webhooks)",
	"event-bus":       "Decoupled components communicating via domain events",
	"cqrs":            "Separate read/write models for complex query patterns",
	"outbox":          "Guaranteed event delivery with transactional consistency",
	"saga":            "Multi-service transactions that need rollback capability",
	"worker":          "Background job processing, async task queues",
	"scheduler":       "Periodic tasks, cron-like scheduling",
	"auth-jwt":        "JWT-based authentication for stateless APIs",
	"cors":            "Browser-facing APIs needing cross-origin access",
	"audit-log":       "Compliance, change tracking, security forensics",
	"multi-tenancy":   "SaaS applications serving multiple organizations",
	"feature-flags":   "Gradual rollouts, A/B testing, kill switches",
	"observability":   "Distributed tracing, metrics, structured logging",
	"health":          "Orchestrator readiness/liveness probes",
	"request-id":      "End-to-end request tracing across services",
	"docker":          "Containerized deployments, local dev with dependencies",
	"ci-github":       "GitHub Actions CI/CD pipeline",
	"ci-gitlab":       "GitLab CI/CD pipeline",
	"makefile":        "Standard build targets (build, test, lint, run)",
	"devcontainer":    "VS Code Dev Containers for reproducible environments",
	"pre-commit":      "Git hooks for code quality checks before commit",
	"linting":         "Static analysis and code style enforcement",
	"testing":         "Test infrastructure and helper utilities",
	"validation":      "Input validation for API request payloads",
	"uuid":            "UUIDv7 primary keys (time-sortable, index-friendly)",
	"platform":        "Config, logging, and lifecycle management",
	"bootstrap":       "Testable dependency wiring with thin main.go",
	"templ":           "Type-safe HTML templates with Go (templ)",
	"htmx":            "Hypermedia-driven UI with HTMX partials",
	"static-assets":   "Embedded static files served via embed.FS",
	"http-client":     "Outbound HTTP calls to external services",
	"email-gateway":   "Transactional email sending",
	"mailpit":         "Local email testing (dev-only SMTP)",
	"i18n":            "Internationalization and localization",
	"api-versioning":  "API version management (URL or header-based)",
}

// BuildCatalog reads all capability manifests and returns an organized catalog.
func BuildCatalog(templateFS fs.FS, installedCaps []string) ([]CatalogEntry, error) {
	if templateFS == nil {
		return nil, nil
	}

	installedSet := make(map[string]bool, len(installedCaps))
	for _, c := range installedCaps {
		installedSet[c] = true
	}

	capDir := path.Join("templates", "capabilities")
	entries, err := fs.ReadDir(templateFS, capDir)
	if err != nil {
		return nil, fmt.Errorf("read capabilities dir: %w", err)
	}

	catalog := make([]CatalogEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()

		data, err := fs.ReadFile(templateFS, path.Join(capDir, name, "capability.yaml"))
		if err != nil {
			continue
		}
		cm, err := scaffold.ParseCapabilityManifest(data)
		if err != nil {
			continue
		}

		catalog = append(catalog, CatalogEntry{
			Name:        name,
			Category:    categoryFor(name),
			Description: cm.Description,
			WhenToUse:   whenToUseFor(name),
			Installed:   installedSet[name],
			Suggests:    cm.Suggests,
		})
	}

	sort.Slice(catalog, func(i, j int) bool {
		if catalog[i].Category != catalog[j].Category {
			return catalog[i].Category < catalog[j].Category
		}
		return catalog[i].Name < catalog[j].Name
	})

	return catalog, nil
}

func categoryFor(name string) string {
	if c, ok := capabilityCategories[name]; ok {
		return c
	}
	return "other"
}

func whenToUseFor(name string) string {
	if w, ok := whenToUse[name]; ok {
		return w
	}
	return ""
}

// writeCatalog writes the capability catalog section to the guide.
func writeCatalog(b *strings.Builder, catalog []CatalogEntry, installedCaps []string) {
	if len(catalog) == 0 {
		return
	}

	b.WriteString("## Capability Catalog\n\n")

	// Installed capabilities.
	var installed, available []CatalogEntry
	for _, e := range catalog {
		if e.Installed {
			installed = append(installed, e)
		} else {
			available = append(available, e)
		}
	}

	if len(installed) > 0 {
		b.WriteString("### Installed\n")
		b.WriteString("| Capability | Category | Purpose |\n")
		b.WriteString("|-----------|----------|--------|\n")
		for _, e := range installed {
			purpose := e.WhenToUse
			if purpose == "" {
				purpose = e.Description
			}
			fmt.Fprintf(b, "| %s | %s | %s |\n", e.Name, e.Category, purpose)
		}
		b.WriteString("\n")
	}

	if len(available) > 0 {
		b.WriteString("### Available (not installed)\n")
		b.WriteString("| Capability | Category | When to Consider |\n")
		b.WriteString("|-----------|----------|------------------|\n")
		for _, e := range available {
			when := e.WhenToUse
			if when == "" {
				when = e.Description
			}
			fmt.Fprintf(b, "| %s | %s | %s |\n", e.Name, e.Category, when)
		}
		b.WriteString("\n")
	}

}

// writeWarnings writes capability interaction warnings to the guide.
func writeWarnings(b *strings.Builder, installedCaps []string) {
	warnings := scaffold.CapabilityWarnings(installedCaps)
	if len(warnings) == 0 {
		return
	}

	// Classify warnings into critical and regular.
	var critical, regular []string
	for _, w := range warnings {
		msg := w.Message
		if isCriticalWarning(msg) {
			critical = append(critical, "CRITICAL: "+msg)
		} else {
			regular = append(regular, "WARNING: "+msg)
		}
	}

	b.WriteString("## Interaction Warnings\n\n")
	for _, w := range critical {
		fmt.Fprintf(b, "- %s\n", w)
	}
	for _, w := range regular {
		fmt.Fprintf(b, "- %s\n", w)
	}
	b.WriteString("\n")
}

// isCriticalWarning returns true if the warning message indicates a critical issue.
func isCriticalWarning(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "critical") ||
		strings.Contains(lower, "auth") ||
		strings.Contains(lower, "security") ||
		strings.Contains(lower, "tenant")
}

// writeSuggestions writes architecture suggestions to the guide.
func writeSuggestions(b *strings.Builder, installedCaps []string) {
	suggestions := scaffold.ComputeSuggestions(installedCaps)
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}
	if len(suggestions) == 0 {
		return
	}

	b.WriteString("## Architecture Suggestions\n\n")
	for i, s := range suggestions {
		fmt.Fprintf(b, "%d. **Add %s** — %s\n", i+1, s.Capability, s.Reason)
	}
	b.WriteString("\n")
}
