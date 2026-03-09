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

## Product Pillars

| Belief | Pillar | CLI command | Status |
|---|---|---|---|
| Composition beats generation | **Compose** | `archway new` | shipped |
| Architecture is code | **Analyze** | `archway analyze` | shipped |
| The gap should be zero | **Enforce** | `archway check` | shipped |

## Features (Roadmap)

| Feature | Status | PRD |
|---------|--------|-----|
| `archway new` — scaffold services | shipped | archway-v1.md |
| `archway analyze` — detect architecture | shipped | archway-v1.md |
| `archway init` — generate archway.yaml | shipped | archway-v1.md |
| `archway check` — validate compliance | shipped | archway-v1.md |
| `archway plan` — compare desired vs actual | planned | |
| `archway apply` — execute migrations | planned | |
| `archway diff` — show drift from declared state | planned | |
| `archway mcp serve` — MCP server | deferred | archway-v1.md |
| Multi-language providers (TypeScript/Node) | planned | PRD-001 |

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

---

*Initialized by keel: 2026-03-08*
*Updated: 2026-03-09 — added vision, mission, core beliefs, product pillars, competitive matrix*
