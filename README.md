# Archway

**Architecture-aware service composer.**

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

**Capabilities** are modular features you compose in:

| | | | |
|---|---|---|---|
| `http-api` | `grpc` | `kafka-consumer` | `mysql` |
| `redis` | `auth-jwt` | `rate-limiting` | `docker` |
| `platform` | `bootstrap` | `linting` | `testing` |
| `ci-github` | `pre-commit` | `email-gateway` | `http-client` |

Each capability provides template files, config, and wiring code. The `bootstrap` capability gives you a thin 15-line `main.go` that delegates to `internal/bootstrap/` for testable dependency injection — and other capabilities automatically wire into it.

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
```

## Architecture Enforcement

Every project gets an `archway.yaml` with component boundaries. Run `archway check` to validate:

```bash
$ archway check

Architecture: hexagonal
Components: 4 defined

Dependency Violations: 0
Structure Issues: 0
Function Issues: 0
Component Coverage: 100% (4/4)
Compliance: 100%

All checks passed
```

Domain importing adapter code? Service bypassing ports? Functions over 80 lines? Caught.

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
  --cap platform,bootstrap,http-api,mysql,auth-jwt,rate-limiting,docker,linting \
  --no-wizard

# gRPC microservice
archway new my-grpc --arch hexagonal \
  --cap platform,bootstrap,grpc,redis,docker \
  --no-wizard

# Simple CLI
archway new my-cli --arch flat --no-wizard

# Event-driven worker
archway new my-worker --arch hexagonal \
  --cap platform,bootstrap,kafka-consumer,mysql,docker \
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

| Pattern | Where |
|---------|-------|
| Hexagonal / Ports & Adapters | Architecture layer structure |
| Composition Root / Bootstrap | `internal/bootstrap/bootstrap.go` |
| Repository Pattern | `adapter/*/` implements `port/outbound.go` |
| RFC 7807 Problem Detail | `adapter/httphandler/response.go` |
| Sentinel + Typed Errors | `domain/errors.go` |
| Chain of Responsibility | HTTP/gRPC middleware stacks |
| Structured Logging | `slog` with JSON/text handlers |
| PII Redaction | Log handler that strips sensitive fields |
| Graceful Shutdown | Ordered hooks with timeout |
| Distributed Tracing | OpenTelemetry auto-instrumentation |

## Language Providers

Archway is built on a **provider model** — each language is a self-contained provider with its own architectures, capabilities, and templates.

| Language | Status | Architectures | Capabilities |
|----------|--------|---------------|-------------|
| **Go** | Stable | hexagonal, flat | 16 capabilities |
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
