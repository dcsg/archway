---
title: Quick Start
description: Create your first service in 2 minutes
---

## Interactive Mode

The fastest way to get started — the wizard guides you through every choice:

```bash
archway new my-service
```

You'll be asked:
1. **Service name** and Go module path
2. **Architecture** — hexagonal or flat
3. **Capabilities** — multi-select what your service needs
4. **Smart suggestions** — accept or skip recommended additions
5. **Confirm** and scaffold

## Non-Interactive Mode

Know what you want? Skip the wizard:

```bash
archway new my-service \
  --arch hexagonal \
  --cap platform,bootstrap,http-api,mysql \
  --module github.com/myorg/my-service \
  --no-wizard
```

## What You Get

After scaffolding, your project contains:

```
my-service/
├── cmd/my-service/main.go        # Thin entry point
├── internal/bootstrap/bootstrap.go # Dependency wiring
├── domain/                        # Business logic (no dependencies)
│   ├── errors.go                 # Typed domain errors
│   └── clock.go                  # Testable time abstraction
├── port/                          # Interfaces
│   ├── inbound.go                # Use case interfaces
│   └── outbound.go               # Repository interfaces
├── service/                       # Use case implementations
├── adapter/                       # External integrations
│   ├── httphandler/              # REST API (Chi router)
│   └── mysqlrepo/                # MySQL repositories
├── config/                        # YAML config loading
├── platform/                      # Cross-cutting concerns
│   ├── lifecycle/                # Graceful shutdown
│   └── observability/            # Logging, OTel, PII redaction
├── docs/PROJECT.md               # What's in your project
├── archway.yaml                  # Architecture rules
└── go.mod
```

## Run It

```bash
cd my-service
go run ./cmd/my-service
```

## Check Architecture

Validate your project follows its own rules:

```bash
archway check
```

This catches:
- Dependency violations (domain importing adapter code)
- Missing required directories
- Function complexity issues
- Component coverage gaps

## Next Steps

- [How It Works](/archway/concepts/how-it-works/) — understand the composition model
- [Capabilities Matrix](/archway/reference/capabilities-matrix/) — see everything available
- [Building a REST API](/archway/guides/rest-api/) — step-by-step guide
