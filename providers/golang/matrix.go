package golang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dcsg/archway/internal/config"
)

// capabilityMeta maps capability names to their categorized metadata for the matrix doc.
var capabilityMeta = map[string]matrixEntry{
	// Architecture
	"hexagonal": {Category: "Architecture", Pattern: "Hexagonal (Ports & Adapters)", Description: "Domain-centric with strict dependency direction: domain → ports → adapters"},
	"flat":      {Category: "Architecture", Pattern: "Flat", Description: "Single-package structure for simple CLIs and scripts"},

	// Design Patterns
	"bootstrap": {Category: "Design Patterns", Pattern: "Bootstrap / Composition Root", Description: "Thin main.go delegates to internal/bootstrap for testable dependency wiring"},
	"platform":  {Category: "Design Patterns", Pattern: "Platform Layer", Description: "Config, lifecycle management, structured logging, OpenTelemetry, PII redaction"},

	// Transport
	"http-api": {Category: "Transport", Pattern: "REST API (Chi)", Description: "Chi router, middleware chain, RFC 7807 error responses, OpenAPI spec"},
	"grpc":     {Category: "Transport", Pattern: "gRPC", Description: "Protocol Buffers, buf tooling, unary/stream interceptors"},

	// Messaging
	"kafka-consumer": {Category: "Messaging", Pattern: "Kafka Consumer", Description: "Consumer group with handler pattern, graceful shutdown"},

	// Data
	"mysql": {Category: "Data", Pattern: "MySQL Repository", Description: "Connection pooling, health checks, repository pattern"},
	"redis": {Category: "Data", Pattern: "Redis Repository", Description: "Connection management, repository pattern"},

	// Security
	"auth-jwt":      {Category: "Security", Pattern: "JWT Authentication", Description: "Middleware-based JWT validation with claims extraction"},
	"rate-limiting": {Category: "Security", Pattern: "Rate Limiting", Description: "Token bucket rate limiter middleware for HTTP endpoints"},

	// Integration
	"email-gateway": {Category: "Integration", Pattern: "Email Gateway", Description: "Adapter pattern for email sending with provider abstraction"},
	"http-client":   {Category: "Integration", Pattern: "HTTP Client", Description: "Resilient HTTP client with retry, timeout, and observability"},

	// Testing
	"testing": {Category: "Testing", Pattern: "Test Helpers", Description: "Test utilities, fixtures, and example table-driven tests"},

	// DevOps / CI
	"ci-github":  {Category: "DevOps", Pattern: "GitHub CI", Description: "Issue templates, PR template for standardized workflows"},
	"docker":     {Category: "DevOps", Pattern: "Docker Compose", Description: "Local development environment with service dependencies"},
	"linting":    {Category: "DevOps", Pattern: "golangci-lint", Description: "Comprehensive Go linting with curated rule set"},
	"pre-commit": {Category: "DevOps", Pattern: "Pre-commit Hooks", Description: "Automated code quality checks before every commit"},

	// Resilience
	"postgres": {Category: "Data", Pattern: "PostgreSQL Repository", Description: "Connection pooling, health checks, repository pattern"},

	// Transport — additional
	"graphql": {Category: "Transport", Pattern: "GraphQL (gqlgen)", Description: "Code-generated GraphQL server with resolver pattern"},
	"sse":     {Category: "Transport", Pattern: "Server-Sent Events", Description: "HTTP streaming for real-time server-to-client updates"},

	// Data — additional
	"mongodb":  {Category: "Data", Pattern: "MongoDB Repository", Description: "Document database with repository pattern"},
	"sqlite":   {Category: "Data", Pattern: "SQLite Repository", Description: "Embedded SQL database for lightweight persistence"},
	"s3":       {Category: "Data", Pattern: "S3 Object Storage", Description: "AWS S3-compatible object storage adapter"},
	"dynamodb": {Category: "Data", Pattern: "DynamoDB Repository", Description: "AWS DynamoDB NoSQL database with repository pattern"},

	// Patterns
	"saga":          {Category: "Design Patterns", Pattern: "Saga Orchestrator", Description: "Distributed transaction coordination with compensating actions"},
	"feature-flags": {Category: "Design Patterns", Pattern: "Feature Flags", Description: "Runtime feature toggling for progressive rollouts"},
	"multi-tenancy": {Category: "Design Patterns", Pattern: "Multi-Tenancy", Description: "Tenant isolation with per-tenant configuration and data separation"},

	// DevOps — additional
	"ci-gitlab":    {Category: "DevOps", Pattern: "GitLab CI", Description: "GitLab CI/CD pipeline with test, lint, and build stages"},
	"makefile":     {Category: "DevOps", Pattern: "Makefile", Description: "Task automation with standard build, test, and lint targets"},
	"devcontainer": {Category: "DevOps", Pattern: "Dev Container", Description: "VS Code dev container with Go toolchain and dependencies"},

	// Frontend
	"templ":         {Category: "Frontend", Pattern: "Templ Components", Description: "Type-safe HTML templating with Go templ"},
	"htmx":          {Category: "Frontend", Pattern: "HTMX", Description: "Hypermedia-driven interactivity without JavaScript frameworks"},
	"static-assets": {Category: "Frontend", Pattern: "Static Assets", Description: "Embedded static file serving for CSS, JS, and images"},

	// Language
	"go": {Category: "Language", Pattern: "Go", Description: "Idiomatic Go with modules, standard project layout"},
}

type matrixEntry struct {
	Category    string
	Pattern     string
	Description string
}

func generateProjectMatrix(outputDir, architecture string, capabilities []string, archwayCfg *config.ArchwayConfig) (string, error) {
	if err := os.MkdirAll(filepath.Join(outputDir, "docs"), 0o755); err != nil {
		return "", fmt.Errorf("create docs dir: %w", err)
	}

	// Build ordered entries by category.
	categoryOrder := []string{"Architecture", "Design Patterns", "Language", "Transport", "Messaging", "Data", "Security", "Integration", "Frontend", "Testing", "DevOps"}
	categories := map[string][]matrixEntry{}

	// Add architecture.
	if entry, ok := capabilityMeta[architecture]; ok {
		categories[entry.Category] = append(categories[entry.Category], entry)
	}

	// Add language.
	categories["Language"] = append(categories["Language"], capabilityMeta["go"])

	// Add capabilities.
	for _, cap := range capabilities {
		if entry, ok := capabilityMeta[cap]; ok {
			categories[entry.Category] = append(categories[entry.Category], entry)
		}
	}

	var b strings.Builder
	b.WriteString("# Project Anatomy\n\n")
	b.WriteString(fmt.Sprintf("> Generated by [Archway](https://github.com/dcsg/archway) on %s\n\n", time.Now().Format("2006-01-02")))

	// Summary matrix table.
	b.WriteString("## What's In This Project\n\n")
	b.WriteString("| Category | Pattern | Description |\n")
	b.WriteString("|----------|---------|-------------|\n")

	for _, cat := range categoryOrder {
		entries, ok := categories[cat]
		if !ok {
			continue
		}
		for _, entry := range entries {
			b.WriteString(fmt.Sprintf("| **%s** | %s | %s |\n", entry.Category, entry.Pattern, entry.Description))
		}
	}

	// Capability set for quick lookup.
	capSet := map[string]bool{}
	for _, c := range capabilities {
		capSet[c] = true
	}

	// Architecture components.
	if archwayCfg != nil && len(archwayCfg.Components) > 0 {
		b.WriteString("\n## Architecture Components\n\n")
		b.WriteString("| Component | Location | May Depend On |\n")
		b.WriteString("|-----------|----------|---------------|\n")
		for _, comp := range archwayCfg.Components {
			deps := "—"
			if len(comp.MayDependOn) > 0 {
				deps = strings.Join(comp.MayDependOn, ", ")
			}
			locs := strings.Join(comp.In, ", ")
			b.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", comp.Name, locs, deps))
		}
		b.WriteString("\nDependency violations are checked by `archway check`.\n")
	}

	// Middleware stack.
	if capSet["http-api"] {
		b.WriteString("\n## HTTP Middleware Stack\n\n")
		b.WriteString("Applied in order (outermost first):\n\n")
		b.WriteString("| # | Middleware | Source |\n")
		b.WriteString("|---|-----------|--------|\n")
		middlewares := []struct{ name, source string }{
			{"RequestID", "chi/middleware"},
			{"RealIP", "chi/middleware"},
			{"Recoverer", "chi/middleware"},
			{"otelhttp", "OpenTelemetry"},
		}
		if capSet["rate-limiting"] {
			middlewares = append(middlewares, struct{ name, source string }{"RateLimiter", "rate-limiting capability"})
		}
		if capSet["auth-jwt"] {
			middlewares = append(middlewares, struct{ name, source string }{"JWTAuth", "auth-jwt capability"})
		}
		for i, mw := range middlewares {
			b.WriteString(fmt.Sprintf("| %d | %s | %s |\n", i+1, mw.name, mw.source))
		}
	}

	// Observability.
	if capSet["platform"] {
		b.WriteString("\n## Observability\n\n")
		b.WriteString("| Signal | Implementation | Notes |\n")
		b.WriteString("|--------|---------------|-------|\n")
		b.WriteString("| Logging | `slog` (structured) | JSON in production, text in dev |\n")
		b.WriteString("| Traces | OpenTelemetry → OTLP/gRPC | Auto-instrumented HTTP/gRPC |\n")
		b.WriteString("| Metrics | OpenTelemetry → OTLP/gRPC | Runtime + custom metrics |\n")
		b.WriteString("| PII Redaction | RedactHandler | Strips secrets from log output |\n")
		if capSet["bootstrap"] {
			b.WriteString("| Lifecycle | lifecycle.App | Graceful shutdown with ordered hooks |\n")
		}
	}

	// Error handling.
	if capSet["http-api"] {
		b.WriteString("\n## Error Handling\n\n")
		b.WriteString("| Layer | Pattern | Format |\n")
		b.WriteString("|-------|---------|--------|\n")
		b.WriteString("| Domain | Sentinel errors + typed validation errors | `domain.ErrNotFound`, `domain.ValidationError` |\n")
		b.WriteString("| HTTP | RFC 7807 Problem Detail | `application/problem+json` |\n")
		b.WriteString("| Mapping | `MapError()` auto-maps domain → HTTP status | 404, 409, 401, 403, 429, 500 |\n")
	}

	// Data layer.
	if capSet["mysql"] || capSet["redis"] {
		b.WriteString("\n## Data Layer\n\n")
		b.WriteString("| Store | Connection | Pattern |\n")
		b.WriteString("|-------|-----------|--------|\n")
		if capSet["mysql"] {
			b.WriteString("| MySQL | Pooled (`sql.DB`) | Repository per aggregate |\n")
		}
		if capSet["redis"] {
			b.WriteString("| Redis | Single connection | Repository pattern |\n")
		}
	}

	// DevOps checklist.
	b.WriteString("\n## DevOps Checklist\n\n")
	checks := []struct {
		item, status string
	}{
		{"Dockerfile", boolToCheck(fileExists(outputDir, "Dockerfile"))},
		{"Docker Compose", boolToCheck(capSet["docker"])},
		{"CI Pipeline", boolToCheck(capSet["ci-github"])},
		{"Linter Config", boolToCheck(capSet["linting"])},
		{"Pre-commit Hooks", boolToCheck(capSet["pre-commit"])},
		{"OpenAPI Spec", boolToCheck(capSet["http-api"])},
		{"Live Reload (.air)", boolToCheck(capSet["platform"])},
		{"Makefile", boolToCheck(fileExists(outputDir, "Makefile"))},
		{"Dependency Graph (`archway check`)", boolToCheck(true)},
	}
	for _, c := range checks {
		b.WriteString(fmt.Sprintf("- %s %s\n", c.status, c.item))
	}

	// Key commands.
	b.WriteString("\n## Key Commands\n\n")
	b.WriteString("```bash\n")
	if architecture == "flat" {
		b.WriteString("# Run\ngo run .\n\n")
	} else {
		b.WriteString("# Run\ngo run ./cmd/<service-name>\n\n")
	}
	if capSet["platform"] {
		b.WriteString("# Live reload\nair\n\n")
	}
	b.WriteString("# Test\ngo test ./...\n\n")
	if capSet["linting"] {
		b.WriteString("# Lint\ngolangci-lint run\n\n")
	}
	b.WriteString("# Check architecture\narchway check\n")
	if capSet["docker"] {
		b.WriteString("\n# Start dependencies\ndocker compose up -d\n")
	}
	b.WriteString("```\n")

	docPath := filepath.Join(outputDir, "docs", "PROJECT.md")
	if err := os.WriteFile(docPath, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("write project matrix: %w", err)
	}

	return docPath, nil
}

func boolToCheck(v bool) string {
	if v {
		return "[x]"
	}
	return "[ ]"
}

func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}
