# Soul — Archway

## Tagline

Architecture-aware service composer and enforcer.

## Vision

Code architecture becomes a first-class artifact — declared, composed, and enforced. The gap between intended architecture and actual architecture is zero. Architecture drift becomes as unacceptable as infrastructure drift. AI agents write architecturally correct code from the first line.

## Mission

Give developers the tools to declare their architecture, compose services from proven patterns, guide AI agents with architectural context, and enforce structural integrity — from first commit to production, across every service.

## Core Beliefs

1. **Architecture is code** — declared in config, versioned in git, enforced in CI
2. **Composition beats generation** — modular capabilities snap together; the rules travel with the code
3. **The gap should be zero** — if you can describe your architecture, your tooling should enforce it
4. **AI agents need architectural context** — prevention beats correction; feed the agent before it writes
5. **Tool-agnostic, universally enforced** — Archway works with every AI coding tool. Rules go in `.cursorrules`, `AGENTS.md`, `.github/copilot-instructions.md` — wherever your tool reads. Pre-commit enforcement works regardless of which AI wrote the code. Claude Code gets real-time enforcement via hooks as a bonus.

## What This Is

Architecture-aware service composer and enforcer. Scaffolds production-ready services by composing architecture patterns (hexagonal, flat) with capability modules (http-api, grpc, mysql, redis, kafka, etc.), generating both code AND architectural DNA (archway.yaml). Proactively feeds AI agents with architectural context via `archway guide`, so agents write correct code from the first line. Standalone CLI: `archway guide`, `archway new`, `archway check`, `archway analyze`, `archway init`.

## Product Pillars

| Belief | Pillar | CLI command |
|---|---|---|
| AI agents need context | **Guide** | `archway guide` |
| Composition beats generation | **Compose** | `archway new` |
| Architecture is code | **Analyze** | `archway analyze` |
| The gap should be zero | **Enforce** | `archway check` |

Guide (prevention) + Enforce (detection) = the gap is zero. Future: `archway plan`, `archway diff` complete the lifecycle.

## Stack

- **Language:** Go
- **CLI:** Cobra
- **TUI:** Bubbletea + Huh (wizard forms)
- **Config:** Viper (YAML + env)
- **Templates:** text/template + embed.FS
- **Go Analysis:** go/packages, go/ast/inspector, dst (decorated syntax tree)
- **Architecture Enforcement:** Custom analyzers (inspired by arch-go, go-arch-lint)
- **Distribution:** GoReleaser, Homebrew

## Current State

Active development. v1 CLI ships with composition-based scaffolding (`archway new` with `--arch` + `--cap`), architecture validation (`archway check`), brownfield analysis (`archway analyze`), and init (`archway init`). Templates use composable architecture + capability modules with partial-based main.go assembly. Smart wizard suggests missing capabilities. Complements Keel (AI context layer) — Archway owns code + architecture, Keel owns AI guardrails.

## Users

- Go developers working on new and existing projects
- Tech leads enforcing architecture conventions
- Teams adopting hexagonal/clean/DDD patterns
- AI-powered development workflows (Claude Code, Cursor, Copilot, Windsurf)

## Critical Rules

- **MVP is Go-only:** Full template scaffold, brownfield analysis, standalone CLI
- **Embedded providers first:** Language providers are Go packages compiled into one binary for MVP; gRPC plugins (hashicorp/go-plugin) come post-MVP
- **Brownfield-first philosophy:** Detect existing architecture, don't just validate. Support gradual adoption with thresholds and ignores.
- **Provider interface:** `Scaffold()`, `Analyze()`, `Migrate()`, `GetInfo()` — language-agnostic manifest + language-specific templates

---

*Initialized: 2026-02-14*
*Updated: 2026-03-09 — added vision, mission, core beliefs, product pillars*
