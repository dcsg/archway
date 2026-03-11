# Archway — Product Spec

## Identity

"Architecture-aware service composer and enforcer" — a Go CLI for scaffolding, analyzing, and governing code architecture.

## Vision

Code architecture becomes a first-class artifact — declared, composed, and enforced. The gap between intended architecture and actual architecture is zero. Architecture drift becomes as unacceptable as infrastructure drift. AI agents write architecturally correct code from the first line.

## Mission

Give developers the tools to declare their architecture, compose services from proven patterns, guide AI agents with architectural context, and enforce structural integrity — from first commit to production, across every service.

## Core Beliefs

1. **Architecture is code** — declared in config, versioned in git, enforced in CI
2. **Composition beats generation** — modular capabilities snap together; the rules travel with the code
3. **The gap should be zero** — if you can describe your architecture, your tooling should enforce it
4. **AI agents need architectural context** — prevention beats correction; feed the agent before it writes

## Users

- Go developers working on new and existing projects
- Tech leads enforcing architecture conventions
- Teams adopting hexagonal/clean/DDD patterns
- AI-powered development workflows (Claude Code, Cursor, Copilot, Windsurf)

## Problems We Solve

- Starting new services requires copying boilerplate and making many manual choices
- Existing codebases drift from intended architecture with no automated detection
- Architecture decisions are made implicitly and lost over time
- There's no declarative way to describe and enforce code architecture (unlike infrastructure with Terraform)
- Monorepo teams have no tooling to enforce cross-service boundaries
- **AI agents write code without architectural awareness** — they don't know the project's patterns, layer boundaries, or conventions, leading to violations that humans must fix

## Product Pillars

| Belief | Pillar | CLI command | Status |
|---|---|---|---|
| AI agents need context | **Guide** | `archway guide` | planned (v1.0) |
| Composition beats generation | **Compose** | `archway new` | shipped |
| Architecture is code | **Analyze** | `archway analyze` | shipped |
| The gap should be zero | **Enforce** | `archway check` | shipped |

**Guide → Compose → Analyze → Enforce** — the full lifecycle.
- **Guide** is preventive — feeds AI agents with architecture, patterns, and rules before they write code
- **Compose** is generative — scaffolds services from architecture + capabilities
- **Analyze** is diagnostic — detects architecture in existing codebases
- **Enforce** is reactive — catches violations after code is written

Together, Guide (prevention) + Enforce (detection) = the gap is zero.

## Architectures

| # | Architecture | Structure | Best For | Status |
|---|---|---|---|---|
| 1 | **Hexagonal** | domain → port → service → adapter | Production APIs, microservices | shipped |
| 2 | **Flat** | single package | CLIs, scripts, prototypes | shipped |
| 3 | **Layered** | handler → service → repository | Most common Go pattern, simpler services | shipped |
| 4 | **Clean** | entity → usecase → interface → infrastructure | Teams following Uncle Bob, complex business logic | shipped |
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
- **`archway guide`** — generate AI agent instructions from archway.yaml + templates (PRD-002)
- ~~Fix doc issues (feature-flags ghost capability, Go version, category count)~~ (resolved)
- Clean up unused CLI flags (`--verbose`, `--config`) and migrate stub
- CLI integration tests
- First GitHub release via GoReleaser

### v1.1 — More Architectures
- **Layered** architecture (handler → service → repository)
- **Clean Architecture** (entity → usecase → interface → infrastructure)
- Update wizard, docs, capabilities matrix
- Update `archway guide` output for new architectures

### v1.2 — Complete the Enforce Pillar
- `archway diff` — show drift from declared state
- CI integration guide (GitHub Actions, GitLab CI)
- Pre-commit hook for `archway check`

### v1.3 — Brownfield Adoption
- `archway add <capability>` — add capability to existing project
- Presets — shareable config bundles ("api-starter", "event-worker")

### v2.0 — Rust Engine + Advanced Architectures
- **Rust analysis engine** with tree-sitter (ADR-007)
- Protobuf protocol between Go and Rust
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
- `archway diff` enhanced with plan awareness

### v3.0 — Multi-Language
- TypeScript/Node provider (PRD-001)
- New language = tree-sitter grammar + query files + manifest
- Provider plugin system for community languages

### v3.1 — Platform
- Team dashboards — architecture compliance across repos
- Architecture scoring / badges
- CI integration for `archway guide` regeneration

## Market Context

Competitors: go-arch-lint, arch-go, copier, cookiecutter, go-blueprint, yeoman. Archway differentiates by combining scaffolding + analysis + governance in one tool — the only scaffolding tool that generates architectural DNA alongside code.

**Unique differentiator: `archway guide` — the only architecture tool that proactively feeds AI agents.** No competitor does this. Every other tool outputs to terminals for humans. Archway outputs to AI agent context files, making it the first AI-native architecture platform.

| Feature | Archway | Cookiecutter | go-blueprint | arch-go |
|---|---|---|---|---|
| Scaffolding | yes | yes | yes | no |
| Composition model (arch + caps) | yes | no | no | no |
| AI agent guidance | **yes** | no | no | no |
| Architecture detection | yes | no | no | yes |
| Validation / enforcement | yes (11 detectors) | no | no | yes |
| Generates archway.yaml (desired state) | yes | no | no | no |
| Smart capability suggestions | yes (18 rules) | no | no | no |
| Monorepo / workspace support | planned | no | no | no |
| Multi-language | planned | yes | no | no |

---

*Initialized by keel: 2026-03-08*
*Updated: 2026-03-09 — 4th pillar (Guide), polyglot architecture, updated roadmap, cut archway apply*
