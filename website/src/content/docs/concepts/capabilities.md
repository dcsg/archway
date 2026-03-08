---
title: Capabilities
description: Modular features you can compose into your service
---

Capabilities are self-contained modules that add specific functionality to your project. Each provides template files, configuration, and wiring code.

## Transport

### `http-api`
REST API built on [Chi](https://github.com/go-chi/chi) v5.

**What you get:**
- Router with middleware chain (RequestID, RealIP, Recoverer, OTel)
- RFC 7807 Problem Detail error responses
- Pagination helpers
- OpenAPI spec (`api/openapi.yaml`)
- Handler patterns with request/response types

### `grpc`
gRPC service with [buf](https://buf.build/) tooling.

**What you get:**
- Protocol Buffer definitions
- Buf configuration (`buf.yaml`, `buf.gen.yaml`)
- Server setup with interceptors
- Reflection enabled for development

### `kafka-consumer`
Kafka consumer with handler pattern.

**What you get:**
- Consumer group setup
- Message handler interface
- Graceful shutdown with drain
- Config for brokers, topics, consumer group

## Data

### `mysql`
MySQL database with connection pooling.

**What you get:**
- Connection setup with pool configuration
- Health checks
- Repository pattern scaffold
- Config for DSN, pool sizes, connection lifetime

### `redis`
Redis connection and repository pattern.

**What you get:**
- Connection management
- Repository pattern scaffold
- Config for address, password, DB number

## Security

### `auth-jwt`
JWT authentication middleware for HTTP APIs.

**What you get:**
- Middleware that validates JWT tokens
- Claims extraction into request context
- Route-level authentication

**Requires:** `http-api`

### `rate-limiting`
Token bucket rate limiter for HTTP endpoints.

**What you get:**
- Rate limiting middleware
- Per-endpoint configuration

**Requires:** `http-api`

## Infrastructure

### `platform`
The production infrastructure layer.

**What you get:**
- **Config** — YAML config loading with capability-aware struct fields
- **Lifecycle** — `lifecycle.App` with ordered startup/shutdown of components
- **Logging** — Structured logging via `slog` (JSON in prod, text in dev)
- **OpenTelemetry** — Traces and metrics via OTLP/gRPC
- **PII Redaction** — Log handler that redacts sensitive fields
- **Live Reload** — `.air.toml` for hot reload during development

**Suggests:** `docker`, `bootstrap`

### `bootstrap`
The bootstrap / composition root pattern.

**What you get:**
- Thin `main.go` that just calls `bootstrap.Run(version)`
- `internal/bootstrap/bootstrap.go` with all dependency wiring
- Partial injection points for other capabilities to wire into

The bootstrap pattern keeps your `main.go` clean and makes dependency wiring testable.

**Requires:** `platform`

## Integration

### `email-gateway`
Email sending with provider abstraction.

### `http-client`
Resilient HTTP client with retry, timeout, and observability.

## Quality

### `testing`
Test utilities and example tests.

**What you get:**
- Test helper functions
- Example table-driven tests
- Test fixtures pattern

### `linting`
Go linting configuration.

**What you get:**
- `.golangci.yaml` with curated linter set (errcheck, govet, staticcheck, bodyclose, noctx, and more)

### `pre-commit`
Pre-commit hook configuration for automated checks.

## DevOps

### `docker`
Local development environment.

**What you get:**
- `docker-compose.yml` with service dependencies
- `.env.example` for environment variables

### `ci-github`
GitHub workflow templates.

**What you get:**
- Issue templates (bug report, feature request)
- Pull request template
