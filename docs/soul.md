# Soul — Archway

## What This Is

Architecture-aware service composer and enforcer. Scaffolds production-ready services by composing architecture patterns (hexagonal, flat) with capability modules (http-api, grpc, mysql, redis, kafka, etc.), generating both code AND architectural DNA (archway.yaml) that AI agents and `archway check` can enforce. Standalone CLI: `archway new`, `archway check`, `archway analyze`, `archway init`.

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

## Critical Rules

- **MVP is Go-only:** Full template scaffold, brownfield analysis, standalone CLI
- **Embedded providers first:** Language providers are Go packages compiled into one binary for MVP; gRPC plugins (hashicorp/go-plugin) come post-MVP
- **Brownfield-first philosophy:** Detect existing architecture, don't just validate. Support gradual adoption with thresholds and ignores.
- **Provider interface:** `Scaffold()`, `Analyze()`, `Migrate()`, `GetInfo()` — language-agnostic manifest + language-specific templates

---

*Initialized: 2026-02-14*
