# Capabilities Matrix

> Everything Archway can compose into your Go service — **36 capabilities**.

Archway scaffolds projects by combining an **architecture** with **capabilities**. Pick what you need — Archway wires it all together.

## Architectures

| Architecture | Structure | Best For |
|-------------|-----------|----------|
| **Hexagonal** (Ports & Adapters) | `domain/` → `port/` → `service/` → `adapter/` | Production APIs, microservices, domain-heavy services |
| **Flat** | Single package, `main.go` + `go.mod` | CLIs, scripts, prototypes, simple tools |

## Capabilities

### Transport

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `http-api` | Chi router, middleware chain, OpenAPI spec | RFC 7807 errors, pagination, structured responses |
| `grpc` | Protocol Buffers, buf tooling, interceptors | Unary/stream handlers, reflection for dev |
| `kafka-consumer` | Consumer group, handler pattern | Graceful shutdown, message routing |
| `websocket` | WebSocket hub with connection management | Upgrade, broadcast, connection lifecycle |

### Data

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `postgres` | pgx connection pool, health checks | Config-driven pool, graceful shutdown |
| `mysql` | Connection pooling, health checks | Repository per aggregate, config-driven DSN |
| `redis` | Connection management | Repository pattern, config-driven connection |
| `migrations` | golang-migrate with SQL stubs | Up/down migrations, version tracking |
| `uuid` | UUIDv7 (time-sortable) ID type | B-tree friendly, no index fragmentation, timestamp extractable |
| `repository` | Generic repository interface (Go generics) | `Repository[T, ID]`, pagination support |

### Resilience

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `circuit-breaker` | Circuit breaker for external calls | Closed/Open/Half-Open states, configurable thresholds |
| `retry` | Exponential backoff with jitter | Configurable attempts, max delay, retryable check |
| `idempotency` | Idempotency key middleware | SHA-256 key hashing, response caching, TTL |
| `health` | `/healthz` and `/readyz` endpoints | Pluggable health checkers, dependency readiness |

### Patterns

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `cqrs` | Command bus + Query bus (generics) | Typed handlers, dispatch by name, separation of reads/writes |
| `event-bus` | In-process domain event pub/sub | Event type routing, synchronous dispatch, handler registration |
| `outbox` | Transactional outbox with relay | Batch processing, retry tracking, outbox SQL migration |

### Security

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `auth-jwt` | JWT middleware for HTTP | Claims extraction, route-level auth |
| `rate-limiting` | Token bucket limiter | Per-endpoint rate limits via middleware |
| `cors` | CORS middleware with defaults | Configurable origins, methods, headers, max-age |
| `validation` | go-playground/validator wrapper | Structured field errors, domain-friendly messages |
| `api-versioning` | URL prefix or header-based versioning | Context propagation, default version fallback |

### Observability

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `observability` | OpenTelemetry tracing + Prometheus metrics | OTLP gRPC export, request counters, duration histograms |
| `request-id` | Request ID middleware | Extract or generate UUID, context propagation, response header |
| `audit-log` | Structured audit trail | Actor/action/resource logging, slog backend, compliance-ready |

### Infrastructure

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `platform` | Config, lifecycle, logging, OTel, PII redaction | `slog` structured logging, OTLP export, graceful shutdown |
| `bootstrap` | Thin `main.go` + `internal/bootstrap/` wiring | Composition root, testable dependency injection |
| `docker` | `docker-compose.yml`, `.env.example` | Local dev with service dependencies |
| `worker` | Background job processor | Worker pool, graceful shutdown, error collection |
| `scheduler` | Cron-style periodic tasks | Interval-based execution, context cancellation |
| `http-client` | Resilient outbound HTTP | Retry, timeout, observability |
| `email-gateway` | Email adapter | Provider abstraction, adapter pattern |

### Quality

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `testing` | Test helpers, example tests | Table-driven tests, test fixtures |
| `linting` | `.golangci.yaml` config | Curated linter set for production Go |
| `pre-commit` | Pre-commit hook config | Automated checks before every commit |
| `ci-github` | Issue/PR templates | Standardized GitHub workflows |

## Dependency & Suggestion Rules

Capabilities can declare relationships:

| Field | Meaning | Example |
|-------|---------|---------|
| `requires` | Must be selected together | `bootstrap` requires `platform` |
| `suggests` | Recommended but optional | `http-api` suggests `rate-limiting` |
| `conflicts` | Warns if both selected (asks if intentional) | *(soft warning, not a blocker)* |

### Smart Suggestions

When you select capabilities, Archway suggests what you might be missing:

| If you select... | Archway suggests... | Why |
|-----------------|--------------------|----|
| Any transport | `platform` | Production services need config, logging, lifecycle |
| `platform` | `bootstrap` | Testable wiring with thin main.go |
| `http-api` | `rate-limiting`, `auth-jwt`, `cors`, `health` | API security, reliability, and ops |
| `postgres`, `mysql` | `docker`, `migrations`, `uuid` | Local dev, schema management, performant IDs |
| `event-bus` | `outbox` | Reliable event publishing |
| `http-client` | `circuit-breaker`, `retry` | Resilient external calls |
| Any transport | `observability`, `request-id` | Tracing and correlation |

## Anti-Pattern Detection

`archway check` detects anti-patterns across three categories:

### Code Anti-Patterns

| Detector | Severity | What It Catches |
|----------|----------|----------------|
| `global_mutable_state` | warning | Package-level vars with mutable types (maps, slices, pointers) |
| `init_abuse` | warning | `init()` with 5+ statements or heavy I/O |
| `naked_goroutine` | warning | Bare `go` statements without errgroup/structured concurrency |
| `swallowed_error` | error | `if err != nil {}` or `if err != nil { return nil }` |
| `uuid_v4_as_key` | info | `uuid.New()` for DB keys — suggests UUIDv7 |

### Architecture Anti-Patterns

| Detector | Severity | What It Catches |
|----------|----------|----------------|
| `fat_handler` | warning | HTTP handlers with 40+ statements |
| `god_package` | warning | Packages with 30+ exported symbols |
| `domain_imports_adapter` | error | Domain/core/port packages importing adapters |
| `mvc_in_hexagonal` | warning | `models/`, `controllers/`, `views/` in hexagonal projects |

### Security Anti-Patterns

| Detector | Severity | What It Catches |
|----------|----------|----------------|
| `sql_concatenation` | error | String concatenation with SQL keywords |
| `context_background_in_handler` | warning | `context.Background()` in handler/adapter code |

## Design Patterns by Category

| Category | Pattern | Capability |
|----------|---------|------------|
| **Architecture** | Hexagonal / Ports & Adapters | Architecture layer |
| | Composition Root / Bootstrap | `bootstrap` |
| | CQRS (Command/Query Separation) | `cqrs` |
| | Transactional Outbox | `outbox` |
| **Data** | Generic Repository (generics) | `repository` |
| | UUIDv7 (time-sortable IDs) | `uuid` |
| | Database Migrations | `migrations` |
| **Resilience** | Circuit Breaker | `circuit-breaker` |
| | Retry with Exponential Backoff | `retry` |
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
| **Realtime** | WebSocket Hub | `websocket` |

## Language & Framework Specifics

| Aspect | Choice | Notes |
|--------|--------|-------|
| Language | Go 1.23+ | Modules, generics available |
| HTTP Router | Chi v5 | Lightweight, stdlib-compatible |
| Logging | `log/slog` | Stdlib structured logging |
| Tracing | OpenTelemetry | OTLP/gRPC export |
| Metrics | Prometheus | `promhttp` handler |
| Config | `gopkg.in/yaml.v3` | YAML file loading |
| Lifecycle | `golang.org/x/sync/errgroup` | Concurrent component management |
| PostgreSQL | pgx v5 | Native driver, connection pooling |
| MySQL | `go-sql-driver/mysql` | Standard `database/sql` |
| Migrations | golang-migrate | File-based SQL migrations |
| Validation | go-playground/validator | Struct tag validation |
| Protobuf | `buf` | Modern protobuf tooling |
| IDs | UUIDv7 (RFC 9562) | Time-sortable, zero-dependency |

## Example Compositions

```bash
# Full production API
archway new my-api --arch hexagonal \
  --cap platform,bootstrap,http-api,postgres,uuid,migrations,auth-jwt,rate-limiting,cors,health,observability,request-id,docker,linting

# gRPC microservice
archway new my-grpc --arch hexagonal \
  --cap platform,bootstrap,grpc,redis,docker,linting,health,observability

# Event-driven worker
archway new my-worker --arch hexagonal \
  --cap platform,bootstrap,kafka-consumer,postgres,event-bus,outbox,worker,docker

# CQRS service
archway new my-cqrs --arch hexagonal \
  --cap platform,bootstrap,http-api,postgres,uuid,cqrs,event-bus,outbox,repository,docker

# Simple CLI tool
archway new my-cli --arch flat
```
