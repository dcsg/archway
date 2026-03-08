# Soul — Archway

## What This Is

A Go CLI tool for code scaffolding, brownfield analysis, and project governance — "Architecture-aware service composer." It works as a standalone CLI (`archway new`, `archway init`, `archway analyze`, `archway plan`, `archway apply`, `archway check`).

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

Active development. v1 CLI shipped with scaffolding (`archway new`), analysis (`archway analyze`), and init (`archway init`). Templates refactored from monolithic go-hexagonal/go-minimal to composable api/cli. v2 features (plan/apply/check) are designed but not yet implemented.

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
