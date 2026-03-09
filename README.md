# Archway

**Architecture-aware service composer and enforcer.**

Stop writing boilerplate. Compose production-grade services from an architecture and capabilities — config, logging, graceful shutdown, middleware, database connections, and more — wired correctly from day one.

> **Language support:** Go is fully supported today. TypeScript/Node is next. Archway's provider model makes it straightforward to add any language.

```bash
archway new my-service \
  --arch hexagonal \
  --cap platform,bootstrap,http-api,mysql,docker \
  --module github.com/myorg/my-service \
  --no-wizard
```

**29 files. Production-ready. Under 10 seconds.**

---

## How It Works

Archway composes your project from two building blocks:

```
Architecture  +  Capabilities  =  Your Service
(structure)      (features)       (ready to run)
```

**Architectures** define your project's structure and dependency rules:

| Architecture | Structure | Best For |
|-------------|-----------|----------|
| Hexagonal | `domain/` → `port/` → `service/` → `adapter/` | Production APIs, microservices |
| Flat | Single package | CLIs, scripts, prototypes |

**Capabilities** are modular features you compose in — **38 and growing:**

| Transport | Data | Resilience | Patterns |
|-----------|------|------------|----------|
| `http-api` | `postgres` | `circuit-breaker` | `cqrs` |
| `grpc` | `mysql` | `retry` | `event-bus` |
| `kafka-consumer` | `redis` | `idempotency` | `outbox` |
| `websocket` | `migrations` | `health` | `repository` |
| | `uuid` | | |

| Security | Observability | Infrastructure | Quality |
|----------|---------------|----------------|---------|
| `auth-jwt` | `observability` | `platform` | `testing` |
| `rate-limiting` | `request-id` | `bootstrap` | `linting` |
| `cors` | `audit-log` | `docker` | `pre-commit` |
| `validation` | | `worker` | `ci-github` |
| `api-versioning` | | `scheduler` | |
| | | `http-client` | |
| | | `email-gateway` | |

Each capability provides template files, config, and wiring code. The `bootstrap` capability gives you a thin 15-line `main.go` that delegates to `internal/bootstrap/` for testable dependency injection — and other capabilities automatically wire into it.

See the full [capabilities matrix](docs/reference/capabilities-matrix.md) for details on each.

## What You Get

A scaffolded hexagonal service with `http-api` and `mysql` includes:

```
my-service/
├── cmd/my-service/main.go           # 15 lines — calls bootstrap.Run()
├── internal/bootstrap/bootstrap.go  # All dependency wiring (testable)
├── domain/                          # Pure business logic (zero deps)
│   ├── errors.go                   # Typed errors: NotFound, Validation, Conflict
│   └── clock.go                    # Testable time abstraction
├── port/                            # Interfaces
│   ├── inbound.go                  # Use case contracts
│   └── outbound.go                 # Repository contracts
├── service/                         # Business logic implementations
├── adapter/
│   ├── httphandler/                # Chi router, RFC 7807 errors, middleware
│   └── mysqlrepo/                  # Connection pooling, repository pattern
├── config/                          # YAML config with capability-aware fields
├── platform/
│   ├── lifecycle/                  # Graceful startup/shutdown
│   └── observability/              # slog, OpenTelemetry, PII redaction
├── docs/PROJECT.md                  # Auto-generated project anatomy
├── archway.yaml                     # Architecture rules (enforced by check)
├── Dockerfile, Makefile, go.mod, ...
```

## Smart Suggestions

When you select capabilities, Archway suggests what you might be missing:

```
Based on your selections, you might also want:
  [x] platform        Production services need config, logging, and lifecycle management
  [x] bootstrap       Bootstrap pattern provides testable dependency wiring
  [ ] rate-limiting   HTTP APIs benefit from rate limiting to prevent abuse
  [ ] auth-jwt        HTTP APIs typically need authentication
  [ ] uuid            UUIDv7 provides database-friendly primary keys without index fragmentation
  [ ] health          Health endpoints enable orchestrator readiness probes
```

## Capability Warnings

After suggestions, Archway warns about potentially problematic combinations:

```
⚠  Capability warnings:
  • postgres without uuid: Consider UUIDv7 for database-friendly primary keys (avoids index fragmentation)
  • event-bus without outbox: Without transactional outbox, events can be lost if the process crashes
  • http-api without cors: If this API serves browser clients, CORS headers are required
```

Warnings are advisory — they don't block scaffolding, but flag combinations that often cause issues in production.

## Architecture Enforcement

Every project gets an `archway.yaml` with component boundaries. Run `archway check` to validate:

```bash
$ archway check

Archway Check — hexagonal
═══════════════════════════════════════════════════════

Components:  4 defined, 4 covered (100% coverage)
Rules:       12 checked

DEPENDENCY VIOLATIONS (0)
  ✓ All checks pass

STRUCTURE VIOLATIONS (0)
  ✓ All checks pass

FUNCTION VIOLATIONS (0)
  ✓ All checks pass

ANTI-PATTERN VIOLATIONS (2)
  ⚠ [uuid_v4_as_key] service/order.go:15 uuid.New() generates UUIDv4 (random) — use UUIDv7
  ✗ [sql_concatenation] adapter/repo.go:42 SQL string concatenation — use parameterized queries

═══════════════════════════════════════════════════════
Result: FAIL — 2 violations found
  Compliance: 83% (10/12 rules passing)
  Coverage:   100% (4/4 components checked)
```

**11 anti-pattern detectors** catch issues across code, architecture, and security:

| Category | Detectors |
|----------|-----------|
| **Code** | Global mutable state, init() abuse, naked goroutines, swallowed errors, UUIDv4 as DB key |
| **Architecture** | Fat handlers, god packages, domain importing adapters, MVC in hexagonal |
| **Security** | SQL string concatenation, context.Background() in handlers |

## Install

```bash
go install github.com/dcsg/archway/cmd/archway@latest
```

Or from source:

```bash
git clone https://github.com/dcsg/archway.git
cd archway
go build -o archway ./cmd/archway
```

## Quick Start

### Interactive (recommended for first time)

```bash
archway new my-service
```

The wizard walks you through architecture, capabilities, and suggestions.

### Non-Interactive

```bash
# Full production API
archway new my-api --arch hexagonal \
  --cap platform,bootstrap,http-api,postgres,uuid,migrations,health,cors,auth-jwt,rate-limiting,observability,request-id,docker,linting \
  --no-wizard

# gRPC microservice
archway new my-grpc --arch hexagonal \
  --cap platform,bootstrap,grpc,redis,health,observability,docker \
  --no-wizard

# Simple CLI
archway new my-cli --arch flat --no-wizard

# Event-driven worker
archway new my-worker --arch hexagonal \
  --cap platform,bootstrap,kafka-consumer,postgres,event-bus,outbox,worker,docker \
  --no-wizard

# CQRS service
archway new my-cqrs --arch hexagonal \
  --cap platform,bootstrap,http-api,postgres,uuid,cqrs,event-bus,outbox,repository,docker \
  --no-wizard
```

### After Scaffolding

```bash
cd my-service
go run ./cmd/my-service    # Run it
archway check              # Validate architecture
```

## Commands

| Command | Description |
|---------|-------------|
| `archway new [name]` | Scaffold a new project (interactive or with flags) |
| `archway check` | Validate project against architecture rules |
| `archway analyze` | Analyze an existing project's architecture |

## Design Patterns Included (Go)

| Category | Pattern | Capability |
|----------|---------|------------|
| **Architecture** | Hexagonal / Ports & Adapters | Architecture layer |
| | Composition Root / Bootstrap | `bootstrap` |
| | CQRS (Command/Query Separation) | `cqrs` |
| | Transactional Outbox | `outbox` |
| **Data** | Repository Pattern (generics) | `repository` |
| | UUIDv7 (time-sortable IDs) | `uuid` |
| | Database Migrations | `migrations` |
| **Resilience** | Circuit Breaker | `circuit-breaker` |
| | Retry with Backoff + Jitter | `retry` |
| | Idempotency Keys | `idempotency` |
| **Events** | Domain Event Bus (pub/sub) | `event-bus` |
| | Reliable Event Publishing | `outbox` |
| **HTTP** | RFC 7807 Problem Detail | `http-api` |
| | Chain of Responsibility (middleware) | `http-api` |
| | CORS | `cors` |
| | Request Validation | `validation` |
| | API Versioning | `api-versioning` |
| **Observability** | Structured Logging (`slog`) | `platform` |
| | Distributed Tracing (OTel) | `observability` |
| | Prometheus Metrics | `observability` |
| | Request ID Propagation | `request-id` |
| | Audit Trail | `audit-log` |
| | PII Redaction | `platform` |
| **Lifecycle** | Graceful Shutdown | `platform` |
| | Background Workers | `worker` |
| | Scheduled Tasks | `scheduler` |
| **Security** | JWT Authentication | `auth-jwt` |
| | Rate Limiting | `rate-limiting` |
| | Health/Readiness Probes | `health` |

## Language Providers

Archway is built on a **provider model** — each language is a self-contained provider with its own architectures, capabilities, and templates.

| Language | Status | Architectures | Capabilities |
|----------|--------|---------------|-------------|
| **Go** | Stable | hexagonal, flat | 38 capabilities |
| **TypeScript/Node** | Planned | — | — |

Want to add a language? See the [provider guide](https://dcsg.github.io/archway/) in the docs.

## Documentation

Full docs at **[dcsg.github.io/archway](https://dcsg.github.io/archway/)**

## Development

```bash
go test ./...        # Run tests
go build ./cmd/archway  # Build
```

## Contributing

1. Fork the repo
2. Create a branch
3. Add tests for behavior changes
4. Open a PR

## License

MIT
