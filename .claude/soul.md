# Soul — Archway

## What This Is

A Go CLI tool for code scaffolding, brownfield analysis, and project governance — "Terraform for Code Architecture." It works in two modes: standalone CLI (`archway new`, `archway init`, `archway analyze`, `archway plan`, `archway apply`, `archway check`) and as an MCP Server exposing deterministic tools for AI clients (Claude Code, Cursor, etc.).

## Stack

- **Language:** Go
- **CLI:** Cobra
- **TUI:** Bubbletea + Huh (wizard forms)
- **Config:** Viper (YAML + env)
- **Templates:** text/template + embed.FS
- **Go Analysis:** go/packages, go/ast/inspector, dst (decorated syntax tree)
- **LLM:** sashabaranov/go-openai (OpenAI-compatible) + Ollama for local
- **MCP:** modelcontextprotocol/go-sdk
- **Architecture Enforcement:** Custom analyzers (inspired by arch-go, go-arch-lint)
- **Distribution:** GoReleaser, Homebrew

## Current State

Greenfield. Pre-code phase with extensive research completed (Go CLI ecosystem, LLM integration, Terraform provider model, project learner patterns). No source code yet.

## Users

- Go developers working on new and existing projects
- Tech leads enforcing architecture conventions
- Teams adopting hexagonal/clean/DDD patterns
- AI-powered development workflows (via MCP)

## Critical Rules

- **Dual-mode LLM strategy:** MCP Server mode lets the host LLM reason; standalone mode auto-detects Ollama -> env API keys -> config -> prompt user
- **MVP is Go-only:** Full 66-template scaffold, brownfield analysis, MCP server, standalone CLI
- **Embedded providers first:** Language providers are Go packages compiled into one binary for MVP; gRPC plugins (hashicorp/go-plugin) come post-MVP
- **Brownfield-first philosophy:** Detect existing architecture, don't just validate. Support gradual adoption with thresholds and ignores.
- **Provider interface:** `Scaffold()`, `Analyze()`, `Migrate()`, `GetInfo()` — language-agnostic manifest + language-specific templates

---

*Initialized: 2026-02-14*
