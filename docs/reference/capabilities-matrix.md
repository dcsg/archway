# Capabilities Matrix

> Everything Archway can compose into your Go service.

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

### Data

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `mysql` | Connection pooling, health checks | Repository per aggregate, config-driven DSN |
| `redis` | Connection management | Repository pattern, config-driven connection |

### Security

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `auth-jwt` | JWT middleware for HTTP | Claims extraction, route-level auth |
| `rate-limiting` | Token bucket limiter | Per-endpoint rate limits via middleware |

### Integration

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `email-gateway` | Email adapter | Provider abstraction, adapter pattern |
| `http-client` | Resilient outbound HTTP | Retry, timeout, observability |

### Infrastructure

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `platform` | Config, lifecycle, logging, OTel, PII redaction | `slog` structured logging, OTLP export, graceful shutdown |
| `bootstrap` | Thin `main.go` + `internal/bootstrap/` wiring | Composition root, testable dependency injection |

### Quality

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `testing` | Test helpers, example tests | Table-driven tests, test fixtures |
| `linting` | `.golangci.yaml` config | Curated linter set for production Go |
| `pre-commit` | Pre-commit hook config | Automated checks before every commit |

### DevOps

| Capability | What You Get | Key Patterns |
|-----------|-------------|--------------|
| `docker` | `docker-compose.yml`, `.env.example` | Local dev with service dependencies |
| `ci-github` | Issue/PR templates | Standardized GitHub workflows |

## Dependency & Suggestion Rules

Capabilities can declare relationships:

| Field | Meaning | Example |
|-------|---------|---------|
| `requires` | Must be selected together | `bootstrap` requires `platform` |
| `suggests` | Recommended but optional | `http-api` suggests `rate-limiting` |
| `conflicts` | Cannot coexist | *(none currently)* |

### Smart Suggestions

When you select capabilities, Archway suggests what you might be missing:

| If you select... | Archway suggests... | Why |
|-----------------|--------------------|----|
| Any transport (`http-api`, `grpc`, `kafka-consumer`) | `platform` | Production services need config, logging, lifecycle |
| `platform` | `bootstrap` | Testable wiring with thin main.go |
| `http-api` | `rate-limiting`, `auth-jwt`, `testing` | API security and reliability |
| `mysql`, `redis` | `docker` | Local dev with dependencies |
| Any transport | `ci-github`, `linting` | Code quality and CI/CD |
| `http-api`, `grpc` | `docker` | Containerized deployment |

## Design Patterns by Category

| Category | Pattern | Where It Lives |
|----------|---------|---------------|
| **Architecture** | Hexagonal / Ports & Adapters | Architecture layer structure |
| **Architecture** | Dependency Inversion | Domain defines interfaces, adapters implement |
| **DI** | Composition Root / Bootstrap | `internal/bootstrap/bootstrap.go` |
| **Error Handling** | RFC 7807 Problem Detail | `adapter/httphandler/response.go` |
| **Error Handling** | Sentinel Errors | `domain/errors.go` |
| **Error Handling** | Typed Validation Errors | `domain/errors.go` |
| **Data** | Repository Pattern | `adapter/*/` implements `port/outbound.go` |
| **Middleware** | Chain of Responsibility | HTTP/gRPC middleware stacks |
| **Observability** | Structured Logging | `slog` with JSON/text handlers |
| **Observability** | Distributed Tracing | OpenTelemetry auto-instrumentation |
| **Observability** | PII Redaction | Log handler that strips sensitive fields |
| **Lifecycle** | Graceful Shutdown | Ordered shutdown hooks with timeout |
| **Config** | File-based Config | YAML config with env-specific overrides |
| **Security** | Token-based Auth | JWT middleware with claims extraction |
| **Security** | Rate Limiting | Token bucket per-endpoint limiting |

## Language & Framework Specifics

| Aspect | Choice | Notes |
|--------|--------|-------|
| Language | Go 1.23+ | Modules, generics available |
| HTTP Router | Chi v5 | Lightweight, stdlib-compatible |
| Logging | `log/slog` | Stdlib structured logging |
| Tracing | OpenTelemetry | OTLP/gRPC export |
| Config | `gopkg.in/yaml.v3` | YAML file loading |
| Lifecycle | `golang.org/x/sync/errgroup` | Concurrent component management |
| MySQL Driver | `go-sql-driver/mysql` | Standard `database/sql` |
| Protobuf | `buf` | Modern protobuf tooling |

## Example Compositions

```bash
# Full production API
archway new my-api --arch hexagonal \
  --cap platform,bootstrap,http-api,mysql,auth-jwt,rate-limiting,docker,linting,ci-github

# gRPC microservice
archway new my-grpc --arch hexagonal \
  --cap platform,bootstrap,grpc,redis,docker,linting

# Simple CLI tool
archway new my-cli --arch flat

# Event-driven worker
archway new my-worker --arch hexagonal \
  --cap platform,bootstrap,kafka-consumer,mysql,docker
```
