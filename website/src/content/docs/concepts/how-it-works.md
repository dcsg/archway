---
title: How It Works
description: Understanding Archway's composition model
---

import { Aside, Steps } from '@astrojs/starlight/components';

## The Composition Model

```
Architecture + Capabilities = Your Service
```

Every Archway service is built from two types of building blocks that get composed together.

### Architecture

An architecture defines your project's **structure and dependency rules**:

- Directory layout (where domain, ports, adapters, etc. live)
- Base files (go.mod, Makefile, Dockerfile, .gitignore)
- Component definitions with dependency constraints
- A fallback `main.go`

| Architecture | Structure | Best For |
|-------------|-----------|----------|
| **Hexagonal** | `domain/` → `port/` → `service/` → `adapter/` | Production APIs, microservices |
| **Flat** | Single package | CLIs, scripts, prototypes |

### Capabilities

Capabilities are **modular features** that plug into your architecture. Each capability provides:

- Template files that get rendered into your project
- **Partials** — code snippets that inject into the bootstrap wiring (imports, init, shutdown hooks)
- A `capability.yaml` manifest with metadata, requirements, and suggestions

<Aside type="note">
Archway ships with **38 capabilities** across 8 categories: transport, data, security, resilience, patterns, observability, infrastructure, and cross-cutting.
</Aside>

## How Composition Works

When you run `archway new`, here's what happens:

<Steps>

1. **Load architecture** — Read the manifest, set up base variables

2. **Load capabilities** — Read each capability manifest, validate requirements and conflicts

3. **Collect partials** — Gather code snippets from each capability's `_partials/` directory

4. **Render architecture files** — Lay down the base project structure

5. **Render capability files** — Add capability-specific files (may overwrite architecture defaults)

6. **Inject partials** — Capability partials get injected into `bootstrap.go` at marked points

7. **Run hooks** — `go mod tidy`, `git init`, etc.

8. **Generate metadata** — `archway.yaml` and `docs/PROJECT.md`

</Steps>

### Partial Injection Points

The bootstrap pattern provides four injection points where capabilities wire themselves in:

| Partial | Purpose | Example |
|---------|---------|---------|
| `main_imports` | Import statements | `"github.com/org/svc/adapter/httphandler"` |
| `main_init` | Initialize connections | `db, err := mysqlrepo.NewConnection(...)` |
| `main_register` | Register with lifecycle | `app.Register("http", httpServer)` |
| `main_shutdown` | Cleanup hooks | `app.OnShutdown("mysql", ...)` |

## Dependency Rules

Capabilities can declare relationships:

- **`requires`** — Must be selected together. `bootstrap` requires `platform`.
- **`suggests`** — Recommended but optional. Archway prompts you in the wizard.
- **`conflicts`** — Cannot coexist.

## Smart Suggestions

After you select capabilities, Archway checks **18 suggestion rules** with cross-referencing and recommends what you might be missing:

```
Based on your selections, you might also want:
  [x] platform        Production services need config, logging, and lifecycle management
  [x] bootstrap       Bootstrap pattern provides testable dependency wiring
  [ ] rate-limiting   HTTP APIs benefit from rate limiting to prevent abuse
  [ ] cors            Cross-origin resource sharing for browser clients
```

<Aside type="caution">
Archway also warns about **problematic combinations** — like using MySQL without migrations, or running an HTTP API without health checks. These are warnings, not blockers.
</Aside>

## Architecture Enforcement

Every scaffolded project includes an `archway.yaml` that defines component boundaries:

```yaml
components:
  - name: domain
    in: ["domain/**"]
    may_depend_on: []
  - name: ports
    in: ["port/**"]
    may_depend_on: [domain]
  - name: service
    in: ["service/**"]
    may_depend_on: [domain, ports]
  - name: adapters
    in: ["adapter/**"]
    may_depend_on: [ports, domain]
```

Run `archway check` at any time to validate these rules. It uses **11 AST-based detectors** to catch dependency violations, missing directories, function complexity issues, and anti-patterns.
