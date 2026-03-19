package typescript

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
	"hexagonal": {Category: "Architecture", Pattern: "Hexagonal (Ports & Adapters)", Description: "Domain-centric with strict dependency direction: domain → application → infrastructure → transport"},
	"flat":      {Category: "Architecture", Pattern: "Flat", Description: "Single-directory structure for simple services and scripts"},

	// Platform
	"platform":  {Category: "Platform", Pattern: "Platform Layer", Description: "Config, structured logging (pino), and lifecycle management"},
	"bootstrap": {Category: "Platform", Pattern: "Bootstrap / Composition Root", Description: "Thin index.ts delegates to wiring layer for testable dependency injection"},
	"graceful":  {Category: "Platform", Pattern: "Graceful Shutdown", Description: "SIGTERM/SIGINT handler with registered shutdown hooks and 30s timeout"},

	// Transport
	"http-api":    {Category: "Transport", Pattern: "HTTP API", Description: "Express, Fastify, or Hono — typed routes, JSON responses, error middleware"},
	"http-client": {Category: "Transport", Pattern: "HTTP Client", Description: "Typed fetch wrapper with timeout and base URL — Node 20 built-in fetch"},
	"grpc":        {Category: "Transport", Pattern: "gRPC", Description: "@grpc/grpc-js server and client with TypeScript proto types"},

	// Observability
	"health":        {Category: "Observability", Pattern: "Health Endpoints", Description: "GET /healthz and GET /readyz endpoints for orchestrator probes"},
	"request-id":    {Category: "Observability", Pattern: "Request ID", Description: "UUID per request via AsyncLocalStorage, propagated across the call chain"},
	"observability": {Category: "Observability", Pattern: "OpenTelemetry", Description: "Distributed tracing and metrics via @opentelemetry/sdk-node"},

	// Data
	"postgres":   {Category: "Data", Pattern: "PostgreSQL", Description: "Prisma or TypeORM — connection pool, health check, repository pattern"},
	"mysql":      {Category: "Data", Pattern: "MySQL", Description: "Prisma or TypeORM — connection pool, health check, repository pattern"},
	"mongodb":    {Category: "Data", Pattern: "MongoDB", Description: "Mongoose connection with typed schemas"},
	"redis":      {Category: "Data", Pattern: "Redis", Description: "ioredis client with connection management"},
	"sqlite":     {Category: "Data", Pattern: "SQLite", Description: "better-sqlite3 — synchronous, embedded, ideal for testing"},
	"migrations": {Category: "Data", Pattern: "Database Migrations", Description: "Prisma migrate or TypeORM migrations — versioned schema changes"},
	"repository": {Category: "Data", Pattern: "Repository Pattern", Description: "TypeScript interface in domain, implementation in infrastructure"},
	"uuid":       {Category: "Data", Pattern: "UUID v7", Description: "Time-sortable UUIDv7 primary keys via the uuid package"},

	// Resilience
	"circuit-breaker": {Category: "Resilience", Pattern: "Circuit Breaker", Description: "opossum — prevents cascade failures from slow external services"},
	"retry":           {Category: "Resilience", Pattern: "Retry with Backoff", Description: "p-retry — exponential backoff with jitter for transient failures"},
	"timeout":         {Category: "Resilience", Pattern: "Timeout", Description: "AbortController-based timeout wrapper for any async operation"},
	"idempotency":     {Category: "Resilience", Pattern: "Idempotency Keys", Description: "Redis-backed or in-memory store — prevents duplicate side effects on retry"},

	// Security
	"rate-limiting": {Category: "Security", Pattern: "Rate Limiting", Description: "Per-route or global rate limiter — express-rate-limit / @fastify/rate-limit / Hono"},
	"auth-jwt":      {Category: "Security", Pattern: "JWT Authentication", Description: "jose — IETF-standard JWT sign/verify + HTTP middleware"},
	"cors":          {Category: "Security", Pattern: "CORS", Description: "Cross-origin headers — framework-specific middleware"},
	"audit-log":     {Category: "Security", Pattern: "Audit Log", Description: "Append-only structured audit trail for compliance and forensics"},
	"encryption":    {Category: "Security", Pattern: "Encryption", Description: "AES-256-GCM via Node crypto — no external deps"},

	// Patterns
	"event-bus":     {Category: "Patterns", Pattern: "Event Bus", Description: "eventemitter3 — typed domain events with publish/subscribe"},
	"worker":        {Category: "Patterns", Pattern: "Background Worker", Description: "bullmq (Redis) or in-process queue for async job processing"},
	"scheduler":     {Category: "Patterns", Pattern: "Scheduler", Description: "node-cron — periodic tasks with graceful shutdown integration"},
	"outbox":        {Category: "Patterns", Pattern: "Transactional Outbox", Description: "Guaranteed event delivery — DB-first, then publish via polling processor"},
	"cqrs":          {Category: "Patterns", Pattern: "CQRS", Description: "Typed CommandBus and QueryBus with handler registration"},
	"email-gateway": {Category: "Patterns", Pattern: "Email Gateway", Description: "Abstract EmailGateway interface with nodemailer SMTP implementation"},

	// Messaging
	"kafka-consumer": {Category: "Messaging", Pattern: "Kafka Consumer", Description: "kafkajs consumer group with typed message handlers and graceful shutdown"},

	// Infrastructure / DevOps
	"docker":       {Category: "DevOps", Pattern: "Docker", Description: "Multi-stage Dockerfile (build + production) with docker-compose for dev"},
	"ci-github":    {Category: "DevOps", Pattern: "GitHub CI", Description: "GitHub Actions: install, typecheck, test, lint, build"},
	"pre-commit":   {Category: "DevOps", Pattern: "Pre-commit Hooks", Description: "husky + lint-staged: ESLint + Prettier on staged TypeScript files"},
	"devcontainer": {Category: "DevOps", Pattern: "Dev Container", Description: "VS Code Dev Container with Node 20 + ESLint + Prettier extensions"},
	"makefile":     {Category: "DevOps", Pattern: "Makefile", Description: "Standard targets: build, test, lint, run, dev, clean"},

	// Quality
	"testing": {Category: "Quality", Pattern: "Vitest", Description: "TypeScript-native test runner with coverage — no config needed"},
	"linting": {Category: "Quality", Pattern: "ESLint + Prettier", Description: "@typescript-eslint/recommended + Prettier with strict settings"},
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
	categoryOrder := []string{"Architecture", "Platform", "Language", "Transport", "Messaging", "Observability", "Data", "Security", "Resilience", "Patterns", "Quality", "DevOps"}
	categories := map[string][]matrixEntry{}

	// Add architecture.
	if entry, ok := capabilityMeta[architecture]; ok {
		categories[entry.Category] = append(categories[entry.Category], entry)
	}

	// Add language.
	categories["Language"] = append(categories["Language"], matrixEntry{
		Category:    "Language",
		Pattern:     "TypeScript / Node.js",
		Description: "Strict TypeScript (strict: true), Node 20 LTS, ES2022 target",
	})

	// Add capabilities.
	for _, cap := range capabilities {
		if entry, ok := capabilityMeta[cap]; ok {
			categories[entry.Category] = append(categories[entry.Category], entry)
		}
	}

	var sb strings.Builder
	sb.WriteString("# Project Architecture Matrix\n\n")
	fmt.Fprintf(&sb, "**Generated:** %s\n\n", time.Now().Format("2006-01-02"))
	fmt.Fprintf(&sb, "**Language:** TypeScript / Node.js  \n**Architecture:** %s\n\n", architecture)

	if archwayCfg != nil && len(archwayCfg.Capabilities) > 0 {
		fmt.Fprintf(&sb, "**Capabilities:** %s\n\n", strings.Join(archwayCfg.Capabilities, ", "))
	}

	sb.WriteString("## Composition\n\n")
	sb.WriteString("| Category | Pattern | Description |\n")
	sb.WriteString("|----------|---------|-------------|\n")

	for _, cat := range categoryOrder {
		for _, entry := range categories[cat] {
			fmt.Fprintf(&sb, "| %s | %s | %s |\n", entry.Category, entry.Pattern, entry.Description)
		}
	}

	matrixPath := filepath.Join(outputDir, "docs", "PROJECT_ANATOMY.md")
	if err := os.WriteFile(matrixPath, []byte(sb.String()), 0o644); err != nil {
		return "", fmt.Errorf("write project matrix: %w", err)
	}
	return matrixPath, nil
}
