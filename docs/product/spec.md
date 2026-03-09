# Archway — Product Spec

## Identity

"Architecture-aware service composer and enforcer" — a Go CLI for scaffolding, analyzing, and governing code architecture.

## Users

- Go developers working on new and existing projects
- Tech leads enforcing architecture conventions
- Teams adopting hexagonal/clean/DDD patterns
- AI-powered development workflows (via MCP)

## Problems We Solve

- Starting new services requires copying boilerplate and making many manual choices
- Existing codebases drift from intended architecture with no automated detection
- Architecture decisions are made implicitly and lost over time

## Features (Roadmap)

| Feature | Status | PRD |
|---------|--------|-----|
| `archway new` — scaffold services | shipped | archway-v1.md |
| `archway analyze` — detect architecture | shipped | archway-v1.md |
| `archway init` — generate archway.yaml | shipped | archway-v1.md |
| `archway check` — validate compliance | shipped | archway-v1.md |
| `archway mcp serve` — MCP server | deferred | archway-v1.md |
| `archway plan` — compare desired vs actual | planned | |
| `archway apply` — execute migrations | planned | |
| Multi-language providers (TypeScript/Node) | planned | PRD-001 |

## Market Context

Competitors: go-arch-lint, arch-go, copier, cookiecutter. Archway differentiates by combining scaffolding + analysis + governance in one tool with a Terraform-like workflow.

---

*Initialized by keel: 2026-03-08*
