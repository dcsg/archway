# Archway — Product Spec

## Identity

"Architecture-aware service composer and enforcer" — a Go CLI for scaffolding, analyzing, and governing code architecture.

## Vision

Code architecture becomes a first-class artifact — declared, composed, and enforced. The gap between intended architecture and actual architecture is zero. Architecture drift becomes as unacceptable as infrastructure drift.

## Mission

Give developers the tools to declare their architecture, compose services from proven patterns, and enforce structural integrity — from first commit to production, across every service.

## Core Beliefs

1. **Architecture is code** — declared in config, versioned in git, enforced in CI
2. **Composition beats generation** — modular capabilities snap together; the rules travel with the code
3. **The gap should be zero** — if you can describe your architecture, your tooling should enforce it

## Users

- Go developers working on new and existing projects
- Tech leads enforcing architecture conventions
- Teams adopting hexagonal/clean/DDD patterns
- AI-powered development workflows (via MCP)

## Problems We Solve

- Starting new services requires copying boilerplate and making many manual choices
- Existing codebases drift from intended architecture with no automated detection
- Architecture decisions are made implicitly and lost over time
- There's no declarative way to describe and enforce code architecture (unlike infrastructure with Terraform)
- Monorepo teams have no tooling to enforce cross-service boundaries

## Product Pillars

| Belief | Pillar | CLI command | Status |
|---|---|---|---|
| Composition beats generation | **Compose** | `archway new` | shipped |
| Architecture is code | **Analyze** | `archway analyze` | shipped |
| The gap should be zero | **Enforce** | `archway check` | shipped |

## Architectures

| # | Architecture | Structure | Best For | Status |
|---|---|---|---|---|
| 1 | **Hexagonal** | domain → port → service → adapter | Production APIs, microservices | shipped |
| 2 | **Flat** | single package | CLIs, scripts, prototypes | shipped |
| 3 | **Layered** | handler → service → repository | Most common Go pattern, simpler services | planned (v1.1) |
| 4 | **Clean** | entity → usecase → interface → infrastructure | Teams following Uncle Bob, complex business logic | planned (v1.1) |
| 5 | **DDD** | bounded contexts, aggregates, domain events | Complex domains with multiple subdomains | planned (v2.0) |
| 6 | **Modular Monolith** | multiple modules, shared infra, one binary | Teams avoiding premature microservices | planned (v2.0) |
| 7 | **Event-driven** | event sourcing + CQRS as primary pattern | Event-sourced systems, audit-heavy domains | planned (v2.0) |

## Workspace / Monorepo Support (v2.1)

Archway works at two levels in a monorepo:

- **Service-level** — each service has its own `archway.yaml` (works today)
- **Workspace-level** — root `archway.yaml` defines the monorepo structure, enforces cross-service boundaries

```yaml
# archway.yaml (root)
workspace:
  services:
    - path: services/order-api
      architecture: hexagonal
    - path: services/notification-worker
      architecture: event-driven
  shared:
    - path: pkg/shared
      rules: no-business-logic
  rules:
    - services cannot import other services directly
    - shared packages cannot import service packages
```

`archway check` at the root validates service isolation — no service importing another service's internals.

## Roadmap

### v1.0 — Polish & Ship
- Fix doc issues (feature-flags ghost capability, Go version, category count)
- Clean up unused CLI flags (`--verbose`, `--config`) and migrate stub
- CLI integration tests
- First GitHub release via GoReleaser

### v1.1 — More Architectures
- **Layered** architecture (handler → service → repository)
- **Clean Architecture** (entity → usecase → interface → infrastructure)
- Update wizard, docs, capabilities matrix

### v1.2 — Complete the Enforce Pillar
- `archway diff` — show drift from declared state
- CI integration guide (GitHub Actions, GitLab CI)
- Pre-commit hook for `archway check`

### v1.3 — Brownfield Adoption
- `archway add <capability>` — add capability to existing project
- Presets — shareable config bundles ("api-starter", "event-worker")

### v2.0 — Advanced Architectures
- **DDD** — bounded contexts, aggregates, domain events
- **Modular Monolith** — multiple modules, shared infra, enforced boundaries
- **Event-driven** — event sourcing + CQRS as primary structure

### v2.1 — Workspace / Monorepo
- `archway init --workspace` — root-level archway.yaml
- Cross-service boundary enforcement
- `archway new` inside a workspace context
- `archway check` at workspace level

### v2.2 — The Terraform Lifecycle
- `archway plan` — compare desired vs actual, show what would change
- `archway apply` — execute the plan
- `archway diff` enhanced with plan awareness

### v3.0 — Multi-Language
- TypeScript/Node provider (PRD-001)
- Provider plugin system for community languages

### v3.1 — Platform
- `archway mcp serve` — expose to AI agents
- Team dashboards — compliance across repos
- Architecture scoring / badges

## Market Context

Competitors: go-arch-lint, arch-go, copier, cookiecutter, go-blueprint, yeoman. Archway differentiates by combining scaffolding + analysis + governance in one tool — the only scaffolding tool that generates architectural DNA alongside code.

| Feature | Archway | Cookiecutter | go-blueprint | arch-go |
|---|---|---|---|---|
| Scaffolding | yes | yes | yes | no |
| Composition model (arch + caps) | yes | no | no | no |
| Architecture detection | yes | no | no | yes |
| Validation / enforcement | yes (11 detectors) | no | no | yes |
| Generates archway.yaml (desired state) | yes | no | no | no |
| Smart capability suggestions | yes (18 rules) | no | no | no |
| Monorepo / workspace support | planned | no | no | no |
| Multi-language | planned | yes | no | no |

---

*Initialized by keel: 2026-03-08*
*Updated: 2026-03-09 — vision, mission, beliefs, architectures roadmap, monorepo support, full roadmap v2*
