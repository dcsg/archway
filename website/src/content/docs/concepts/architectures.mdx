---
title: Architectures
description: Available architecture patterns and when to use them
---

import { Aside } from '@astrojs/starlight/components';

An architecture defines your project's structure, dependency rules, and base files.

## Hexagonal (Ports & Adapters)

The hexagonal architecture separates your application into layers with strict dependency direction — inner layers never import from outer layers.

```
domain/              → Pure business logic, no external dependencies
port/                → Interfaces (inbound use cases, outbound repositories)
service/             → Use case implementations
adapter/             → External integrations (HTTP, DB, messaging)
config/              → Configuration loading (via platform capability)
platform/            → Cross-cutting concerns (via platform capability)
internal/bootstrap/  → Dependency wiring (via bootstrap capability)
cmd/<name>/          → Entry point
```

### Dependency Rules

| Component | May Depend On |
|-----------|--------------|
| domain | nothing |
| ports | domain |
| service | domain, ports |
| adapters | ports, domain |

These rules are enforced by `archway check`. If your adapter code imports from another adapter, or your domain imports from service — it's caught immediately.

<Aside type="tip">
Every scaffolded project includes an `archway.yaml` with these rules defined. You can customize the rules, add required directories, or set function complexity limits.
</Aside>

### When to Use

- Production APIs and microservices
- Services with complex business logic
- Projects where you want strict separation of concerns
- Teams that need consistent structure across services

## Flat

A single-package structure with just `main.go` and `go.mod`.

```
main.go
go.mod
```

### When to Use

- CLI tools
- Simple scripts
- Quick prototypes
- Projects that don't need layered architecture

<Aside type="note">
You can still add capabilities to a flat architecture (e.g., `http-api` for a simple API server), but the project won't have the layered structure of hexagonal.
</Aside>
