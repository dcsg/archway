# Archway — Pre-Planning Context

This document captures all decisions made during the research/interview phase.
Feed this to `/gsd:new-project` when asked "What do you want to build?"

## What Is Archway

A Go CLI tool for code scaffolding, brownfield analysis, and project governance. It works in two modes:
1. **Standalone CLI** — `archway new`, `archway analyze`, `archway plan`, `archway apply`, `archway check`
2. **MCP Server** — exposes tools for AI clients (Claude Code, Cursor, etc.)

## Core Mental Model: Terraform for Code Architecture

Archway follows the same conceptual model as Terraform, applied to code architecture instead of infrastructure:

| Terraform | Archway | What it does |
|-----------|---------|--------------|
| Provider (aws, gcp) | Language provider (go, php, node) | Knows how to work with a specific domain |
| Resource (aws_s3_bucket) | Template + Rules | Defines what a project component looks like |
| State (current infra) | Analysis result | What the codebase currently is |
| Plan (desired vs actual) | Migration plan | What needs to change to match desired state |
| Apply (make it so) | Apply migrations | Execute changes safely, step by step |
| Registry | Template/preset registry | Discover and share templates + rule presets |
| HCL config | archway.yaml | Declare desired architecture + rules |

### The Workflow

```bash
# Declare desired state
archway init                    # creates archway.yaml with desired architecture

# Understand current state
archway analyze                 # scans codebase, produces analysis (like terraform show)

# Compare desired vs actual
archway plan                    # shows what would change (like terraform plan)

# Execute changes
archway apply                   # applies migrations step by step (like terraform apply)

# Validate ongoing compliance
archway check                   # CI-friendly validation (like terraform validate)
```

## Extensibility Model

### Three layers, three extensibility approaches:

**1. Templates (data, not code) — anyone can create**
- A template is a directory of files + `manifest.yaml` + `wizard.yaml`
- No Go knowledge needed — just text files with `{{.Variable}}` placeholders
- Distribution: embedded in binary (official), git repos, local directories
- Like Cookiecutter templates but with a standard manifest

**2. Rules & Best Practices (YAML config) — shareable presets**
- Users declare desired state in `archway.yaml`
- Dependency rules, naming conventions, structure requirements, function limits
- Shareable presets like ESLint configs: `archway/go-hexagonal-strict`, `archway/php-symfony`
- Users can create and share presets as git repos

**3. Language Analyzers (code, embedded) — one per language**
- Each language has an analyzer that understands its AST, types, imports
- Go: native `go/ast` + `go/packages` (fast, no deps)
- PHP/Node/Python: exec bridge to language-native tools
- All compiled into one binary — no separate plugin binaries
- If community demand grows beyond ~10 languages, extract to plugin model later

### archway.yaml Example

```yaml
language: go
architecture: hexagonal

# What "good" looks like for this project
rules:
  dependencies:
    - layer: domain
      packages: ["domain/**"]
      may_depend_on: []  # domain depends on nothing

    - layer: ports
      packages: ["port/**"]
      may_depend_on: [domain]

    - layer: adapters
      packages: ["adapter/**"]
      may_depend_on: [ports, domain]

  naming:
    - pattern: "adapter/*repo/*"
      must_end_with: "Repo"

  structure:
    required_dirs: [cmd/, domain/, adapter/, port/]
    forbidden_dirs: [utils/, helpers/]

  functions:
    max_lines: 80
    max_params: 4

# Shareable preset (like eslint-config-airbnb)
extends:
  - archway/go-hexagonal-strict

# Template source for scaffolding
templates:
  source: archway/go-hexagonal  # built-in, or git repo URL
```

## Core Capabilities

### Greenfield: Scaffold New Projects (`archway new`)
- Interactive TUI wizard (Bubbletea/Huh) that interviews the developer
- Template-based code generation (text/template + embed.FS)
- Language providers with embedded templates (Go first, then PHP, Node, Python, Rust, Java/Kotlin, Ruby)
- Post-scaffold hooks (go mod init, git init, pre-commit install)
- Generates `archway.yaml` as part of scaffold (desired state is set from day one)

### Brownfield: Analyze Existing Projects (`archway analyze`)
- **Deterministic layer**: AST parsing, import analysis, dependency graphs, structure detection
- **LLM layer** (optional): Semantic analysis, ADR generation, invariant extraction
- Auto-detect language from manifest files (go.mod, composer.json, package.json, etc.)
- Architecture detection (hexagonal, clean, DDD, flat)
- Framework detection (Chi, Gin, Echo, Fiber, stdlib — for Go)
- Convention extraction (naming patterns, error handling, test patterns)

### Migration: Plan & Apply Changes (`archway plan`, `archway apply`)
- Compare current codebase against desired state in `archway.yaml`
- Generate migration plan (what files to create/move/modify)
- Each step is a git commit (fully reversible)
- Uses `dst` (Decorated Syntax Tree) for comment-preserving Go transformations
- Never destructive — plan first, apply after user approval

### Governance: Validate Architecture (`archway check`)
- Validate code against `archway.yaml` rules (dependency rules, naming, structure)
- CI-friendly: `archway check` with JSON/SARIF output
- Gradual adoption with compliance thresholds

## Architecture Decisions

### Dual-Mode LLM Strategy
- **MCP Server mode**: Archway exposes deterministic tools. The host LLM (Claude, Cursor, GPT) does reasoning.
- **Standalone mode**: Auto-detect chain — Ollama local → env API keys → config → prompt user.
- Uses `sashabaranov/go-openai` for OpenAI-compatible API (works with OpenAI, Ollama, Groq, etc.)
- MCP server via `github.com/modelcontextprotocol/go-sdk`

### Language Provider Architecture
- **MVP**: Embedded providers (Go packages compiled into one binary)
- **Post-MVP**: Extract to gRPC plugins (hashicorp/go-plugin) when community demand exists
- Provider interface: `Scaffold()`, `Analyze()`, `Migrate()`, `GetInfo()`
- Language-agnostic manifest + language-specific template sets

### Go Provider (v1)
- Ports all 66 templates from existing `/dof:scaffold-go` DOF skill
- Templates live at `~/dotfiles/claude/templates/go-service/` (already created)
- Hexagonal architecture, CQRS, Chi, slog, OTel, koanf, franz-go
- Full wizard: service type, transports, data stores, auth, email, etc.

### Tech Stack
- **CLI**: Cobra
- **TUI**: Bubbletea + Huh (wizard forms)
- **Config**: Viper (YAML + env)
- **Templates**: text/template + embed.FS
- **Go Analysis**: go/packages, go/ast/inspector, dst (decorated syntax tree)
- **LLM**: sashabaranov/go-openai (OpenAI-compatible)
- **MCP**: modelcontextprotocol/go-sdk
- **Architecture Enforcement**: Custom analyzers (inspired by arch-go, go-arch-lint)
- **Distribution**: GoReleaser, Homebrew

### DOF Integration
- Existing `/dof:scaffold-go` becomes thin wrapper around `archway new go`
- Archway configured as MCP server in Claude Code
- When Claude Code calls archway via MCP, Claude does the LLM reasoning

## v1 Scope (MVP)

- **Go provider only** (full 66-template scaffold)
- **Brownfield analysis** (deterministic + LLM-powered)
- **MCP server** + **standalone CLI**
- **LLM integration** (Ollama + OpenAI-compatible)
- **DOF wrapper** update

## Planned Phases

### Phase 1: Project Foundation
- Cobra CLI, command structure, config system
- Language auto-detection
- Provider interface

### Phase 2: Template Engine + Go Scaffold
- Template renderer (text/template + embed.FS)
- Port all 66 templates
- TUI wizard (Huh)
- Post-scaffold hooks

### Phase 3: Deterministic Go Analyzer
- go/packages + inspector AST analysis
- Architecture, framework, convention detection
- Structure validation
- JSON + human-readable output

### Phase 4: LLM Integration
- Provider abstraction
- Ollama + OpenAI auto-detect
- ADR generation, invariant extraction

### Phase 5: MCP Server
- Expose tools via MCP SDK
- Claude Code + Cursor configuration

### Phase 6: DOF Integration + Polish
- Update DOF scaffold-go wrapper
- README, docs, GoReleaser

## Post-MVP (v2)
- PHP + Node/TypeScript providers
- Migration wizard (dst-based transformations)
- Architecture enforcement (`archway check` in CI)
- Python, Rust, Java/Kotlin, Ruby providers

## Repo
- `github.com/dcsg/archway` (private, open-source later)

## Research Files
The following research was conducted and can be found at:
- `/tmp/archway-research/go-cli-ecosystem.md` — Go scaffolding, architecture enforcement, brownfield analysis, refactoring, plugins (1,800+ lines)
- `/tmp/archway-research/llm-integration.md` — LLM integration approaches: OpenAI API, MCP, plugins, exec-based, local models (1,900+ lines)
- `/tmp/archway-research/terraform-model.md` — Terraform provider architecture, plugin systems, distribution (2,200+ lines)
