# PRD: Archway v1 — Terraform for Code Architecture

**Author:** Daniel Gomes
**Status:** Draft
**Created:** 2026-02-14
**Last Updated:** 2026-02-14
**Source of Truth:** `.planning/CONTEXT.md`

---

## 1. Problem Statement

### What problem are we solving?

Go developers face three recurring pain points that no single tool addresses:

1. **Scaffolding friction:** Starting a new Go service with production-grade patterns (hexagonal architecture, CQRS, observability, auth, messaging) requires manually wiring dozens of files. Existing tools like go-blueprint cover basics but lack opinionated, battle-tested templates for real-world services. And once scaffolded, there's no record of the *desired* architecture — drift begins immediately.

2. **Brownfield blindness:** Teams inherit or grow codebases without documented architecture. There's no tool that can analyze an existing Go project and tell you "this is hexagonal architecture using Chi, with structured logging via slog, and these dependency violations." Detection is manual and error-prone.

3. **No plan/apply model for code architecture:** Infrastructure has Terraform (declare desired state, plan changes, apply safely). Code architecture has nothing equivalent. Developers must manually figure out what needs to change and make ad-hoc modifications with no safety net.

### Who has this problem?

- **Individual Go developers** starting new services who want production-ready scaffolds without copy-pasting from old projects
- **Tech leads** who need to understand existing codebases quickly and define desired architecture
- **Teams** adopting hexagonal/clean/DDD patterns who want a safe, step-by-step migration path
- **AI-powered workflows** (Claude Code, Cursor) that need structured code analysis tools via MCP

---

## 2. Core Mental Model: Terraform for Code Architecture

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

# Scaffold new (greenfield)
archway new                     # TUI wizard, generates project + archway.yaml

# Understand current state
archway analyze                 # scans codebase, produces analysis (like terraform show)

# Compare desired vs actual (v2)
archway plan                    # shows what would change (like terraform plan)

# Execute changes (v2)
archway apply                   # applies migrations step by step (like terraform apply)

# Validate ongoing compliance (v2)
archway check                   # CI-friendly validation (like terraform validate)
```

---

## 3. Goals & Non-Goals

### Goals

1. **Scaffold production-ready Go services** from an interactive TUI wizard with 66+ templates covering hexagonal architecture, CQRS, multiple transports (HTTP, gRPC, Kafka), data stores, auth, email, and observability — generating `archway.yaml` as part of the scaffold so desired state is set from day one
2. **Detect existing architecture** by analyzing Go source code — identify architecture pattern (hexagonal, clean, DDD, flat), framework, conventions, and dependency structure
3. **Define a clear extensibility model** with three layers: templates (data, not code), rules/presets (shareable YAML), and language analyzers (embedded code)
4. **Expose analysis as MCP tools** so AI clients can leverage Archway's deterministic analysis with their own LLM reasoning
5. **Integrate optional LLM capabilities** for semantic analysis, ADR generation, and invariant extraction — with auto-detection chain: Ollama local -> env API keys -> config -> prompt user
6. **Ship as a single binary** with embedded templates and providers, distributed via GoReleaser and Homebrew
7. **Update DOF integration** so `/dof:scaffold-go` becomes a thin wrapper around `archway new go`

### Non-Goals (v1)

- **Multi-language support:** v1 is Go-only. PHP, Node, Python, Rust, Java/Kotlin, Ruby providers are post-MVP
- **gRPC plugin system:** Embedded providers only. hashicorp/go-plugin extraction happens post-MVP when community demand exists
- **Migration execution:** `archway plan` and `archway apply` are v2. v1 focuses on analysis and scaffolding, not code transformation
- **Architecture enforcement in CI:** `archway check` with JSON/SARIF output is v2. v1 produces analysis reports but doesn't gate CI
- **Custom template registry:** No hosted template marketplace. Templates are embedded or loaded from local paths/git repos
- **IDE plugins:** No VSCode/GoLand extensions. CLI and MCP server only
- **Web UI:** No dashboard or web interface

---

## 4. Extensibility Model

Three layers, three extensibility approaches:

### Layer 1: Templates (data, not code) — anyone can create

- A template is a directory of files + `manifest.yaml` + `wizard.yaml`
- No Go knowledge needed — just text files with `{{.Variable}}` placeholders
- Distribution: embedded in binary (official), git repos, local directories
- Like Cookiecutter templates but with a standard manifest

### Layer 2: Rules & Presets (YAML config) — shareable

- Users declare desired state in `archway.yaml`
- Dependency rules, naming conventions, structure requirements, function limits
- Shareable presets like ESLint configs: `archway/go-hexagonal-strict`, `archway/php-symfony`
- Users can create and share presets as git repos
- `extends:` key for composing presets

### Layer 3: Language Analyzers (code, embedded) — one per language

- Each language has an analyzer that understands its AST, types, imports
- Go: native `go/ast` + `go/packages` (fast, no deps)
- PHP/Node/Python (post-MVP): exec bridge to language-native tools
- All compiled into one binary — no separate plugin binaries

### archway.yaml Specification

```yaml
language: go
architecture: hexagonal

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

---

## 5. User Stories

### US-1: Scaffold a new Go service

**As a** Go developer starting a new microservice,
**I want to** run an interactive wizard that asks me about my service type, transports, data stores, and preferences,
**So that** I get a complete, production-ready project scaffold with `archway.yaml` defining desired state from day one.

**Acceptance Criteria:**
- Running `archway new` launches a TUI wizard (Bubbletea/Huh)
- Wizard covers: service name, module path, service type, HTTP framework (Chi), transports (HTTP/gRPC/Kafka), data stores (Postgres/Redis/MongoDB), auth (JWT/OAuth), email provider, observability (OTel/slog)
- Generated project compiles and passes `go vet` immediately
- Generated project includes `archway.yaml` with rules matching the chosen architecture
- Post-scaffold hooks run: `go mod init`, `go mod tidy`, `git init`, optional pre-commit install
- Non-interactive mode supported: `archway new --name myservice --transport http --db postgres --framework chi`
- Generated code follows hexagonal architecture with CQRS patterns

### US-2: Initialize desired architecture for an existing project

**As a** tech lead adopting Archway on an existing codebase,
**I want to** run `archway init` to generate an `archway.yaml` that describes my desired architecture,
**So that** I have a declarative config defining what "good" looks like for my project.

**Acceptance Criteria:**
- `archway init` launches an interactive wizard asking about desired architecture
- Can optionally run `archway analyze` first to pre-populate based on current state
- Generates `archway.yaml` with dependency rules, naming conventions, structure requirements
- Supports `extends:` to inherit from shareable presets (e.g., `archway/go-hexagonal-strict`)
- Config is human-readable and version-controllable

### US-3: Analyze an existing Go codebase

**As a** tech lead inheriting a Go codebase,
**I want to** run `archway analyze` and get a structured report of the project's architecture, frameworks, conventions, and patterns,
**So that** I can understand the codebase without reading every file.

**Acceptance Criteria:**
- Auto-detects language from manifest files (go.mod)
- Detects architecture pattern: hexagonal, clean, DDD, layered, flat (with confidence score)
- Detects framework: Chi, Gin, Echo, Fiber, stdlib, gRPC
- Detects conventions: error handling (sentinel/typed/wrapped), logging (slog/zap/zerolog), config (viper/koanf/env), testing patterns (table-driven/BDD)
- Builds dependency graph and identifies layer relationships
- Outputs in human-readable (terminal), JSON, and Markdown formats
- Deterministic layer works without LLM; LLM layer enhances with semantic analysis when available

### US-4: Use Archway from AI clients via MCP

**As an** AI-powered development tool (Claude Code, Cursor),
**I want to** call Archway's analysis tools via MCP protocol,
**So that** I can combine deterministic code analysis with LLM reasoning.

**Acceptance Criteria:**
- `archway mcp serve` starts an MCP server (stdio transport)
- Exposes tools: `analyze_codebase`, `detect_architecture`, `list_templates`, `scaffold_project`
- Exposes resources: current codebase analysis results, archway.yaml config
- Compatible with Claude Code MCP configuration
- Host LLM performs reasoning; Archway provides structured, deterministic data

### US-5: Configure LLM for enhanced analysis

**As a** developer wanting deeper code insights,
**I want to** configure an LLM provider (local or cloud) for semantic analysis features,
**So that** Archway can generate ADRs, extract invariants, and provide richer analysis.

**Acceptance Criteria:**
- Auto-detection chain: Ollama (localhost:11434) -> `OPENAI_API_KEY` env var -> `~/.archway/config.yaml` -> interactive prompt
- `archway configure llm` runs interactive LLM setup
- Supports: OpenAI, Groq, Together, Ollama (any OpenAI-compatible API via base URL)
- All LLM features degrade gracefully — deterministic analysis always works without LLM
- LLM-powered features: ADR generation from codebase analysis, invariant extraction from validation/test patterns, semantic architecture assessment

### US-6: Detect project language automatically

**As a** developer running Archway on any project,
**I want** the tool to auto-detect the project language from manifest files,
**So that** I don't need to specify the language manually.

**Acceptance Criteria:**
- Detects Go from `go.mod`
- (Post-v1) Detects PHP from `composer.json`, Node from `package.json`, Python from `pyproject.toml`/`requirements.txt`, Rust from `Cargo.toml`, Java from `pom.xml`/`build.gradle`
- Falls back to asking user when ambiguous or unknown
- Detection result cached for session

### US-7: Use DOF scaffold-go with Archway

**As a** Claude Code user with the DOF skill,
**I want** `/dof:scaffold-go` to use Archway under the hood,
**So that** I get the same quality scaffold whether I use the CLI or the DOF skill.

**Acceptance Criteria:**
- `/dof:scaffold-go` becomes a thin wrapper that calls `archway new go`
- Archway configured as MCP server in Claude Code
- When Claude Code calls archway via MCP, Claude does the LLM reasoning
- Standalone CLI and DOF skill produce identical output

---

## 6. Requirements

### 6.1 Functional Requirements

#### Must Have (P0)

| ID | Requirement | User Story |
|----|-------------|------------|
| F-01 | CLI with Cobra: `archway new`, `archway init`, `archway analyze`, `archway configure`, `archway version`, `archway mcp serve` | All |
| F-02 | TUI wizard for `archway new` using Bubbletea + Huh forms | US-1 |
| F-03 | Template engine using text/template + embed.FS with manifest.yaml + wizard.yaml per template | US-1 |
| F-04 | Go provider implementing `Scaffold()`, `Analyze()`, `GetInfo()` | US-1, US-3 |
| F-05 | Language auto-detection from manifest files | US-6 |
| F-06 | Deterministic Go analysis: go/packages + ast/inspector for architecture, framework, and convention detection | US-3 |
| F-07 | Dependency graph construction and layer relationship mapping | US-3 |
| F-08 | Architecture pattern detection: hexagonal, clean, DDD, layered, flat | US-3 |
| F-09 | Framework detection: Chi, Gin, Echo, Fiber, stdlib, gRPC | US-3 |
| F-10 | Convention detection: error handling, logging, config, testing patterns | US-3 |
| F-11 | `archway.yaml` config format with dependency rules, naming, structure, functions | US-2 |
| F-12 | `archway init` interactive wizard to generate archway.yaml | US-2 |
| F-13 | `archway new` generates archway.yaml alongside project scaffold | US-1 |
| F-14 | Output formats: terminal (human-readable), JSON, Markdown | US-3 |
| F-15 | Post-scaffold hooks: `go mod init`, `go mod tidy`, `git init` | US-1 |
| F-16 | Non-interactive mode for all commands (flags-based) | US-1, US-3 |
| F-17 | Config system via Viper: YAML config + env vars + flags | US-5 |
| F-18 | Confidence scores on architecture detection results | US-3 |

#### Should Have (P1)

| ID | Requirement | User Story |
|----|-------------|------------|
| F-19 | MCP server mode: `archway mcp serve` with stdio transport | US-4 |
| F-20 | MCP tools: `analyze_codebase`, `detect_architecture`, `list_templates`, `scaffold_project` | US-4 |
| F-21 | LLM integration via sashabaranov/go-openai (OpenAI-compatible) | US-5 |
| F-22 | Ollama auto-detection and local model support | US-5 |
| F-23 | LLM auto-detection chain: Ollama -> env keys -> config -> prompt | US-5 |
| F-24 | `archway configure llm` interactive setup | US-5 |
| F-25 | Shareable presets via `extends:` in archway.yaml | US-2 |
| F-26 | DOF integration: `/dof:scaffold-go` wraps `archway new go` | US-7 |

#### Could Have (P2)

| ID | Requirement | User Story |
|----|-------------|------------|
| F-27 | LLM-powered ADR generation from codebase analysis | US-5 |
| F-28 | LLM-powered invariant extraction from validation/test patterns | US-5 |
| F-29 | LLM-powered semantic architecture assessment | US-5 |
| F-30 | `archway analyze --init` to pre-populate archway.yaml from current state | US-2 |
| F-31 | Template loading from git repos and local directories (beyond embedded) | US-1 |

### 6.2 Non-Functional Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| NF-01 | **Single binary distribution** — no runtime dependencies except optional Ollama for LLM features | All platforms |
| NF-02 | **Analysis performance** — analyze a 50-file Go project in under 5 seconds (deterministic layer) | Macbook M-series |
| NF-03 | **Scaffold performance** — generate a full project in under 2 seconds | Any platform |
| NF-04 | **Binary size** — under 30MB with all embedded templates | Release build |
| NF-05 | **Cross-platform** — Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64) | GoReleaser |
| NF-06 | **Go version** — requires Go 1.22+ | Build requirement |
| NF-07 | **Graceful degradation** — all features work without LLM; LLM adds optional enhancement | Core requirement |
| NF-08 | **API key security** — never log or display API keys; store in `~/.archway/config.yaml` with 0600 permissions | Security |
| NF-09 | **Code privacy** — warn user before sending code to cloud LLM; default to local when available | Privacy |
| NF-10 | **Test coverage** — minimum 80% coverage on core packages (analyzer, scaffolder, config) | CI gate |

---

## 7. Architecture Overview

### Provider Interface

```go
type LanguageProvider interface {
    Scaffold(ctx context.Context, req ScaffoldRequest) (*ScaffoldResponse, error)
    Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error)
    Migrate(ctx context.Context, req MigrateRequest) (*MigrateResponse, error)  // v2
    GetInfo(ctx context.Context) (*ProviderInfo, error)
}
```

### High-Level Structure

```
archway/
  cmd/archway/           # CLI entry point (Cobra)
  internal/
    cli/                 # Command definitions
    config/              # Viper-based configuration
    provider/            # Provider interface + registry
    scaffold/            # Template engine + rendering
    analyzer/            # Deterministic code analysis
      detector/          # Language, architecture, framework detection
      graph/             # Dependency graph construction
    llm/                 # LLM provider abstraction
    mcp/                 # MCP server implementation
    output/              # Formatters (terminal, JSON, Markdown)
  providers/
    golang/              # Go language provider
      templates/         # 66+ embedded templates (manifest.yaml + wizard.yaml each)
      analyzer/          # Go-specific AST analysis
      scaffolder/        # Go-specific scaffold logic
```

### Dual-Mode LLM Strategy

```
Standalone CLI:
  archway analyze → Deterministic analysis + optional LLM enhancement
  LLM chain: Ollama local → env API keys → config → prompt user

MCP Server:
  archway mcp serve → Expose deterministic tools only
  Host LLM (Claude, Cursor) does all reasoning
  Archway provides structured data, not LLM opinions
```

---

## 8. Phased Delivery

### Phase 1: Project Foundation

- Cobra CLI skeleton: `new`, `init`, `analyze`, `configure`, `version`, `mcp serve`
- Viper config system (`~/.archway/config.yaml`)
- Language auto-detection from manifest files
- Provider interface definition
- Go provider registration (embedded)
- `archway.yaml` config format parser

### Phase 2: Template Engine + Go Scaffold

- Template renderer (text/template + embed.FS)
- Template format: directory + `manifest.yaml` + `wizard.yaml`
- Port all 66 templates from existing DOF scaffold-go skill
- TUI wizard with Bubbletea + Huh
- Post-scaffold hooks (go mod init, git init)
- `archway new` generates `archway.yaml` alongside scaffold
- Non-interactive mode (flags)

### Phase 3: Deterministic Go Analyzer

- go/packages + inspector AST analysis pipeline
- Architecture pattern detection (hexagonal, clean, DDD, flat) with confidence scores
- Framework and convention detection
- Dependency graph construction
- Output formatters: terminal, JSON, Markdown
- `archway init` wizard to generate archway.yaml

### Phase 4: LLM Integration

- LLM provider abstraction (Provider interface)
- OpenAI-compatible client (sashabaranov/go-openai)
- Ollama auto-detection
- Auto-detection chain
- `archway configure llm` interactive setup
- LLM-enhanced analysis (ADR generation, invariant extraction)

### Phase 5: MCP Server

- MCP server via modelcontextprotocol/go-sdk
- Tool definitions: analyze, detect, list-templates, scaffold
- Resource definitions: codebase analysis, archway.yaml config
- Claude Code + Cursor configuration docs

### Phase 6: DOF Integration + Polish

- Update `/dof:scaffold-go` to wrap `archway new go`
- GoReleaser configuration
- Homebrew formula
- README, documentation

---

## 9. v2 Roadmap (Post-MVP)

Features explicitly deferred from v1, per CONTEXT.md:

| Feature | Description |
|---------|-------------|
| `archway plan` | Compare current codebase against desired state in archway.yaml, generate migration plan |
| `archway apply` | Execute migration plan step by step, each step as a git commit (fully reversible), using dst for comment-preserving Go transformations |
| `archway check` | CI-friendly architecture validation with JSON/SARIF output, exit codes, gradual adoption thresholds, `archway:ignore` comments |
| PHP provider | Language analyzer + templates for PHP (Symfony, Laravel) |
| Node/TypeScript provider | Language analyzer + templates |
| Python, Rust, Java/Kotlin, Ruby providers | Additional language support |
| Shareable preset registry | Discover and install presets from git repos |
| HTML reports | Stakeholder-friendly output |
| Dependency visualization | `archway graph` with DOT format output |

---

## 10. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Time to scaffold** | < 3 minutes from `archway new` to `go run` | Manual testing |
| **Analysis accuracy** | > 85% correct architecture detection on known codebases | Test against 10+ open-source Go projects |
| **Zero-config start** | User can run `archway analyze` with no config on any Go project | Manual testing |
| **LLM graceful degradation** | 100% of features work without LLM configured | Automated test suite |
| **Binary portability** | Works on Linux/macOS/Windows amd64+arm64 | GoReleaser matrix |
| **Template quality** | Generated code passes `go vet`, `golangci-lint`, and compiles | CI on generated output |
| **archway.yaml generated** | Every `archway new` produces a valid archway.yaml | Automated test |

---

## 11. Open Questions

1. **Template source for Go scaffold:** The existing 66 templates live at `~/dotfiles/claude/templates/go-service/`. Do we port them as-is or refactor during embedding into manifest.yaml + wizard.yaml format?
2. **Preset distribution:** How do users discover and install shareable presets? Git clone? `archway preset add <url>`?
3. **MCP server transport:** stdio only (simpler) or also HTTP/SSE (broader compatibility)?
4. **LLM token budget:** Should `archway analyze` have a configurable max-token budget for LLM calls to control costs?
5. **archway.yaml bootstrap:** When running `archway init` on a brownfield project, should it auto-run `archway analyze` to pre-populate rules based on current state?

---

*This PRD covers Archway v1 (MVP). For v2 features (plan/apply, check, multi-language providers, migration execution), see v2 roadmap above.*
