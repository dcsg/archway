# Plan: Archway v1 — Architecture-Aware Service Composer Refactor

## Overview
**Task:** Refactor Archway from "Terraform for Code Architecture" into "Architecture-aware service composer." Kill MCP/LLM/Keel overlap, implement composition model (architecture + capabilities + language), smart wizard with suggestions, `archway check` command, named components in archway.yaml.
**Total Phases:** 7
**Created:** 2026-03-08

## Progress

| Phase | Status | Updated |
|-------|--------|---------|
| 1     | done   | 2026-03-08 |
| 2     | done   | 2026-03-08 |
| 3     | done   | 2026-03-08 |
| 4     | done   | 2026-03-08 |
| 5     | done   | 2026-03-08 |
| 6     | done   | 2026-03-08 |
| 7     | done   | 2026-03-08 |

**IMPORTANT:** Update this table as phases complete. This table is the persistent state that survives context compaction.

## Model Assignment

| Phase | Task | Model | Reasoning | Est. Cost |
|-------|------|-------|-----------|-----------|
| 1 | Kill MCP + LLM + Keel overlap | Sonnet | Straightforward deletions + minor edits, but need care with imports | $0.08 |
| 2 | Restructure archway.yaml schema (named components + capabilities) | Sonnet | Schema design + Go struct changes + tests | $0.08 |
| 3 | Restructure templates into capability modules | Sonnet | File reorganization + manifest design, moderate complexity | $0.08 |
| 4 | Build capability composition renderer | Opus | Core architecture work — composing multiple capability manifests into a single scaffold | $0.80 |
| 5 | Build smart wizard with suggestion engine | Opus | Complex UX logic — contextual suggestions, multi-step composition, conditional flows | $0.80 |
| 6 | Build `archway check` command | Opus | New command combining analyzer + rules validation + violation reporting | $0.80 |
| 7 | Tests + cleanup + docs | Sonnet | CLI tests, integration tests, update README/PRD/soul | $0.08 |

## Execution Strategy

| Phase | Depends On | Parallel With |
|-------|-----------|---------------|
| 1     | None      | 2             |
| 2     | None      | 1             |
| 3     | 2         | -             |
| 4     | 2, 3      | -             |
| 5     | 4         | 6             |
| 6     | 2         | 5             |
| 7     | 1-6       | -             |

**Waves:**
- Wave 1: Phase 1 + Phase 2 (parallel — independent concerns)
- Wave 2: Phase 3 (depends on schema from Phase 2)
- Wave 3: Phase 4 (depends on templates from Phase 3)
- Wave 4: Phase 5 + Phase 6 (parallel — wizard and check are independent)
- Wave 5: Phase 7 (final cleanup after all features land)

---

## Phase 1: Kill MCP, LLM, and Keel Overlap

**Objective:** Remove MCP server, LLM integration, and Keel-overlapping template files to simplify the codebase and clarify scope.
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `PHASE 1 COMPLETE`
**Dependencies:** None

**Prompt:**
```
Remove MCP server, LLM integration, and Keel overlap from the Archway codebase.

## Context
Read docs/soul.md and the memory files at ~/.claude/projects/-Users-danielgomes-go-src-github-com-dcsg-archway/memory/decisions.md for why we're making these changes.

## Steps

### 1. Kill MCP Server
- Delete the entire directory: internal/mcp/ (server.go, tools.go, resources.go, server_test.go)
- Delete: internal/cli/mcp.go
- In internal/cli/root.go: remove the newMCPCommand(opts) call from cmd.AddCommand()
- In go.mod: remove github.com/modelcontextprotocol/go-sdk dependency

### 2. Kill LLM Integration
- Delete the entire directory: internal/llm/ (provider.go, openai.go, noop.go, analysis.go, detect.go, llm_test.go)
- Delete: internal/cli/configure.go
- In internal/cli/root.go: remove the newConfigureCommand(opts) call from cmd.AddCommand()
- In internal/cli/analyze.go:
  - Remove the import of "github.com/dcsg/archway/internal/llm"
  - Remove the --no-llm flag definition
  - Remove the LLM enhancement block (the section that calls llm.DetectProviderWithInfo, llm.GenerateADRs, llm.ExtractInvariants, llm.SemanticAssessment)
  - Simplify AnalyzeRequest to just: provider.AnalyzeRequest{Path: opts.Path}
  - Remove IncludeLLM from the AnalyzeRequest if it exists
- In internal/provider/provider.go: remove IncludeLLM field from AnalyzeRequest struct, remove LLM-related fields from AnalyzeResponse if any
- In go.mod: remove github.com/sashabaranov/go-openai dependency

### 3. Kill Keel Overlap in Templates
- Delete: providers/golang/templates/api/files/.claude/ (entire directory — CLAUDE.md.tmpl, soul.md.tmpl, invariants.md.tmpl, decisions/*.tmpl)
- Delete: providers/golang/templates/cli/files/.claude/ (entire directory — CLAUDE.md.tmpl, soul.md.tmpl)
- In providers/golang/templates/api/files/.pre-commit-config.yaml.tmpl: keep this, it's not Keel overlap

### 4. Clean up dependencies
- Run: go mod tidy
- Run: go build ./...
- Run: go test ./...
- Run: go vet ./...

### 5. Update docs
- In docs/soul.md: remove mention of MCP Server mode and LLM features from "What This Is"
- Update the tagline to "Architecture-aware service composer"

All tests must pass. When complete, output: PHASE 1 COMPLETE
```

---

## Phase 2: Restructure archway.yaml Schema

**Objective:** Add named components and capabilities to the archway.yaml schema, replacing raw package-glob dependency rules with a cleaner component-based model.
**Model:** `sonnet`
**Max Iterations:** 15
**Completion Promise:** `PHASE 2 COMPLETE`
**Dependencies:** None

**Prompt:**
```
Redesign the archway.yaml schema to support named components and capabilities.

## Context
Read docs/soul.md, docs/decisions/*, and ~/.claude/projects/-Users-danielgomes-go-src-github-com-dcsg-archway/memory/competitor-research.md (the go-arch-lint section on named components).

## Current Schema
Read internal/config/archway_yaml.go to understand the current schema.

## New Schema Design

Update ArchwayConfig to:

```go
type ArchwayConfig struct {
    Language     string            `yaml:"language"`
    Architecture string            `yaml:"architecture"`     // hexagonal, clean, layered, flat
    Capabilities []string          `yaml:"capabilities"`     // http-api, grpc, kafka-consumer, mysql, etc.
    Components   []Component       `yaml:"components"`       // Named components (replaces raw Dependencies)
    Rules        RulesConfig       `yaml:"rules"`
    Extends      []string          `yaml:"extends,omitempty"`
}

type Component struct {
    Name        string   `yaml:"name"`           // "domain", "ports", "service", "adapters"
    In          []string `yaml:"in"`             // Package globs: ["domain/**"]
    MayDependOn []string `yaml:"may_depend_on"`  // ["domain", "ports"]
}

type RulesConfig struct {
    Naming    []NamingRule    `yaml:"naming,omitempty"`
    Structure StructureConfig `yaml:"structure"`
    Functions FunctionRules   `yaml:"functions"`
}
```

Key changes:
- Remove Dependencies from RulesConfig (replaced by Components at top level)
- Add Capabilities field (list of enabled capabilities)
- Add Components field (named components with dependency rules)
- Keep Naming, Structure, Functions in RulesConfig

## Steps

### 1. Update the Go structs in internal/config/archway_yaml.go
- Add Component struct
- Add Capabilities field to ArchwayConfig
- Remove DependencyRule struct (replaced by Component)
- Remove Dependencies from RulesConfig
- Update DefaultArchwayConfig() to use the new schema:
  ```yaml
  components:
    - name: domain
      in: ["domain/**"]
      may_depend_on: []
    - name: ports
      in: ["port/**"]
      may_depend_on: [domain]
    - name: service
      in: ["service/**"]
      may_depend_on: [domain, ports]
    - name: adapters
      in: ["adapter/**"]
      may_depend_on: [ports, domain]
    - name: platform
      in: ["platform/**"]
      may_depend_on: []
  ```

### 2. Update the analyzer graph
- In internal/analyzer/graph/graph.go: update LayerViolations() to accept []Component instead of []DependencyRule
- The logic is the same (map packages to components via globs, check allowed dependencies), just different input type
- Component.In replaces DependencyRule.Packages
- Component.Name replaces DependencyRule.Layer
- Component.MayDependOn replaces DependencyRule.MayDependOn

### 3. Update the analyzer
- In internal/analyzer/analyzer.go: update any references to the old DependencyRule type
- In providers/golang/provider.go: update AnalyzeRequest/Response if they reference dependency rules

### 4. Update validation
- In ValidateArchwayYAML(): validate component names are unique, no self-references in MayDependOn

### 5. Update testdata
- In internal/config/testdata/archway.yaml: update to new schema format

### 6. Update tests
- Update all tests that reference the old schema
- Add tests for the new Component type validation

### 7. Verify
- go build ./...
- go test ./...
- go vet ./...

When complete, output: PHASE 2 COMPLETE
```

---

## Phase 3: Restructure Templates Into Capability Modules

**Objective:** Reorganize the monolithic api template into composable capability modules that can be mixed and matched.
**Model:** `sonnet`
**Max Iterations:** 20
**Completion Promise:** `PHASE 3 COMPLETE`
**Dependencies:** Phase 2

**Prompt:**
```
Restructure the Go provider templates from monolithic templates into composable capability modules.

## Context
Read docs/soul.md and ~/.claude/projects/-Users-danielgomes-go-src-github-com-dcsg-archway/memory/decisions.md for the composition model decision.

## Current Structure
Read providers/golang/templates/api/ to understand the current template layout:
- manifest.yaml defines all variables (HasHTTP, HasGRPC, HasKafka, etc.)
- wizard.yaml defines the interactive form
- files/ contains all template files with {{if .HasHTTP}} conditionals

## New Structure

Reorganize into:

```
providers/golang/templates/
  wizard.yaml                    # Provider-level wizard (kept as-is, routes to architecture)
  architectures/
    hexagonal/
      manifest.yaml              # Base variables for hexagonal: ServiceName, ModulePath
      files/                     # Core hexagonal structure
        cmd/__ServiceName__/main.go.tmpl
        domain/errors.go.tmpl
        domain/clock.go.tmpl
        port/inbound.go.tmpl
        port/outbound.go.tmpl
        config/config.go.tmpl
        config/config.yaml.example.tmpl
        go.mod.tmpl
        Makefile.tmpl
        Dockerfile.tmpl
        README.md.tmpl
        .gitignore.tmpl
        .editorconfig.tmpl
    flat/
      manifest.yaml
      files/
        main.go.tmpl
        go.mod.tmpl
  capabilities/
    http-api/
      manifest.yaml              # Variables: (none extra, uses base)
      files/
        adapter/httphandler/router.go.tmpl
        adapter/httphandler/handler.go.tmpl
        adapter/httphandler/middleware.go.tmpl
        adapter/httphandler/response.go.tmpl
        adapter/httphandler/pagination.go.tmpl
        api/openapi.yaml.tmpl
    grpc/
      manifest.yaml
      files/
        adapter/grpchandler/server.go.tmpl
        adapter/grpchandler/interceptors.go.tmpl
        proto/buf.gen.yaml.tmpl
        proto/buf.yaml.tmpl
        proto/v1/service.proto.tmpl
    kafka-consumer/
      manifest.yaml
      files/
        adapter/kafkahandler/consumer.go.tmpl
        adapter/kafkahandler/handler.go.tmpl
    mysql/
      manifest.yaml
      files/
        adapter/mysqlrepo/connection.go.tmpl
    redis/
      manifest.yaml
      files/
        adapter/redisrepo/connection.go.tmpl
    auth-jwt/
      manifest.yaml
      files/
        adapter/httphandler/middleware_auth.go.tmpl
    rate-limiting/
      manifest.yaml
      files/
        adapter/httphandler/middleware_ratelimit.go.tmpl
    observability/
      manifest.yaml
      files/
        platform/observability/logger.go.tmpl
        platform/observability/otel.go.tmpl
        platform/observability/redact.go.tmpl
        platform/lifecycle/app.go.tmpl
        .air.toml.tmpl
    email-gateway/
      manifest.yaml
      files/
        adapter/emailgateway/gateway.go.tmpl
    http-client/
      manifest.yaml
      files/
        platform/httpclient/client.go.tmpl
    testing/
      manifest.yaml
      files/
        service/example_test.go.tmpl
        internal/testutil/helpers.go.tmpl
    ci-github/
      manifest.yaml
      files/
        .github/ISSUE_TEMPLATE/bug_report.md.tmpl
        .github/ISSUE_TEMPLATE/feature_request.md.tmpl
        .github/pull_request_template.md.tmpl
    docker/
      manifest.yaml
      files/
        docker-compose.yml.tmpl
        .env.example.tmpl
    linting/
      manifest.yaml
      files/
        .golangci.yaml.tmpl
    pre-commit/
      manifest.yaml
      files/
        .pre-commit-config.yaml.tmpl
  cli/                           # Keep CLI template as a separate "architecture" or simple template
    manifest.yaml
    wizard.yaml
    files/
      main.go.tmpl
      go.mod.tmpl
```

## Steps

### 1. Create the directory structure
- Create providers/golang/templates/architectures/hexagonal/
- Create providers/golang/templates/capabilities/{http-api,grpc,kafka-consumer,mysql,redis,auth-jwt,rate-limiting,observability,email-gateway,http-client,testing,ci-github,docker,linting,pre-commit}/

### 2. Move files from api/files/ into the correct capability directories
- Core hexagonal files (domain/, port/, config/, cmd/, Makefile, Dockerfile, README, .gitignore, .editorconfig, go.mod) → architectures/hexagonal/files/
- HTTP handler files → capabilities/http-api/files/
- gRPC files → capabilities/grpc/files/
- Kafka files → capabilities/kafka-consumer/files/
- MySQL files → capabilities/mysql/files/
- Redis files → capabilities/redis/files/
- Auth JWT file → capabilities/auth-jwt/files/
- Rate limiting file → capabilities/rate-limiting/files/
- Observability files → capabilities/observability/files/
- Email gateway → capabilities/email-gateway/files/
- HTTP client → capabilities/http-client/files/
- Test files → capabilities/testing/files/
- CI files → capabilities/ci-github/files/
- Docker compose + env → capabilities/docker/files/
- Golangci config → capabilities/linting/files/
- Pre-commit config → capabilities/pre-commit/files/

### 3. Create manifest.yaml for each capability
Each manifest.yaml should declare:
```yaml
name: http-api
description: "HTTP API with Chi router, middleware, and OpenAPI spec"
variables: []  # Additional variables this capability needs (most use base vars)
suggests:      # Capabilities this one suggests
  - rate-limiting
  - auth-jwt
  - observability
requires: []   # Capabilities this one requires (e.g., auth-jwt requires http-api)
```

### 4. Create architecture manifests
Each architecture manifest.yaml declares:
```yaml
name: hexagonal
description: "Hexagonal architecture with domain, ports, adapters, and platform layers"
variables:
  - name: ServiceName
    type: string
    required: true
  - name: ModulePath
    type: string
    required: true
components:
  - name: domain
    in: ["domain/**"]
    may_depend_on: []
  - name: ports
    in: ["port/**"]
    may_depend_on: [domain]
  - name: service
    in: ["service/**"]
    may_depend_on: [domain, ports]
  - name: adapters
    in: ["adapter/**"]
    may_depend_on: [ports, domain]
  - name: platform
    in: ["platform/**"]
    may_depend_on: []
```

### 5. Remove old api/ and cli/ template directories
- Delete providers/golang/templates/api/ (already moved)
- Keep providers/golang/templates/cli/ as-is for now (or move to architectures/flat/ if appropriate)

### 6. Update providers/golang/templates.go
- Update the embed directive to point to the new structure
- Ensure GetTemplateFS() returns the right FS

### 7. Verify the structure
- ls -R the new structure to confirm all files are in place
- go build ./... (should still compile even if renderer isn't updated yet)

Do NOT update the renderer or CLI in this phase — that's Phase 4.

When complete, output: PHASE 3 COMPLETE
```

---

## Phase 4: Build Capability Composition Renderer

**Objective:** Update the scaffold renderer to compose multiple capability modules into a single project output.
**Model:** `opus`
**Max Iterations:** 25
**Completion Promise:** `PHASE 4 COMPLETE`
**Dependencies:** Phase 2, Phase 3

**Prompt:**
```
Build the capability composition renderer that overlays architecture + capabilities into a single scaffold output.

## Context
Read docs/soul.md, the new template structure from Phase 3, and the current renderer at internal/scaffold/renderer.go.

## Current Renderer
The current renderer:
1. Reads a single manifest.yaml
2. Walks a single files/ directory
3. Renders each .tmpl file with variables
4. Skips empty files (conditional exclusion)

## New Renderer Design

The composition renderer must:
1. Load the **architecture** manifest (e.g., architectures/hexagonal/manifest.yaml) — get base variables and components
2. Load each **capability** manifest — get additional variables, suggestions, requirements
3. **Merge variables** — architecture base + all capability variables (no conflicts allowed)
4. **Merge files** — overlay capability files on top of architecture files
5. **Handle main.go composition** — the tricky part: main.go needs imports and wiring from each capability
6. **Handle port/outbound.go composition** — outbound ports need interfaces from each capability (MySQL repo, Redis cache, etc.)

### The main.go Problem

Each capability needs to inject:
- Import lines into main.go
- Initialization code (e.g., `db := mysqlrepo.NewConnection(cfg)`)
- Wiring code (e.g., `svc := service.New(db, cache, logger)`)
- Shutdown/cleanup code

**Solution: Template partials**

Each capability can provide a `_partials/` directory with injectable snippets:

```
capabilities/mysql/
  _partials/
    main_imports.go.tmpl      # import "{{ .ModulePath }}/adapter/mysqlrepo"
    main_init.go.tmpl         # db, err := mysqlrepo.NewConnection(cfg.Database)
    main_wire.go.tmpl         # (provides db to service constructor)
    main_cleanup.go.tmpl      # defer db.Close()
  files/
    adapter/mysqlrepo/connection.go.tmpl
```

The architecture's main.go.tmpl collects all partials:

```go
package main

import (
    "{{ .ModulePath }}/config"
    {{ range .Partials.main_imports }}{{ . }}{{ end }}
)

func main() {
    cfg := config.Load()

    {{ range .Partials.main_init }}{{ . }}{{ end }}

    {{ range .Partials.main_cleanup }}defer {{ . }}{{ end }}

    // ... start server
}
```

### Implementation Steps

### 1. Update internal/scaffold/manifest.go
Add a CapabilityManifest struct:
```go
type CapabilityManifest struct {
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Variables   []Variable `yaml:"variables"`
    Suggests    []string `yaml:"suggests"`
    Requires    []string `yaml:"requires"`
}
```

Add an ArchitectureManifest struct:
```go
type ArchitectureManifest struct {
    Name        string      `yaml:"name"`
    Description string      `yaml:"description"`
    Variables   []Variable  `yaml:"variables"`
    Components  []Component `yaml:"components"`  // from config package
}
```

Add a CompositionPlan struct:
```go
type CompositionPlan struct {
    Architecture     ArchitectureManifest
    Capabilities     []CapabilityManifest
    MergedVariables  map[string]interface{}
    Partials         map[string][]string  // partial_name → rendered snippets
}
```

### 2. Create internal/scaffold/composer.go
New file that handles the composition:
```go
func ComposeProject(templateFS fs.FS, architecture string, capabilities []string, vars map[string]interface{}) (*CompositionPlan, error)
```

This function:
1. Reads architectures/{name}/manifest.yaml
2. For each capability, reads capabilities/{name}/manifest.yaml
3. Validates requirements (e.g., auth-jwt requires http-api)
4. Merges variables (architecture defaults + capability defaults + user values)
5. Collects partials from each capability's _partials/ directory
6. Returns a CompositionPlan

### 3. Update internal/scaffold/renderer.go
Add a new method:
```go
func (r *Renderer) RenderComposition(plan *CompositionPlan, outputDir string) (*RenderResult, error)
```

This method:
1. Renders architecture files/ with merged variables + partials
2. For each capability, renders its files/ with the same variables
3. Overlays results (capability files go into the same output directory)
4. Returns list of files created

### 4. Update providers/golang/provider.go
Update Scaffold() to use the new composition renderer:
- If the request includes Architecture + Capabilities → use ComposeProject + RenderComposition
- Backward-compatible: if request has TemplateName, fall back to old renderer

### 5. Generate archway.yaml
After scaffolding, generate archway.yaml using the architecture's components and selected capabilities:
```yaml
language: go
architecture: hexagonal
capabilities: [http-api, mysql, observability, auth-jwt]
components:
  - name: domain
    in: ["domain/**"]
    may_depend_on: []
  # ... from architecture manifest
```

### 6. Tests
- Test CompositionPlan creation with various capability combos
- Test variable merging (no conflicts)
- Test requirement validation (auth-jwt without http-api → error)
- Test partial collection and rendering
- Test file overlay (capability files land in correct directories)

go build ./... && go test ./... && go vet ./...

When complete, output: PHASE 4 COMPLETE
```

---

## Phase 5: Build Smart Wizard With Suggestion Engine

**Objective:** Build the interactive wizard that walks users through architecture + capability selection, with contextual suggestions for things they might have forgotten.
**Model:** `opus`
**Max Iterations:** 20
**Completion Promise:** `PHASE 5 COMPLETE`
**Dependencies:** Phase 4

**Prompt:**
```
Build the smart wizard with contextual capability suggestions for archway new.

## Context
Read docs/soul.md and ~/.claude/projects/-Users-danielgomes-go-src-github-com-dcsg-archway/memory/decisions.md. The wizard suggests capabilities users might forget — suggest-only, user confirms each.

## Current Wizard
Read internal/scaffold/tui.go, internal/scaffold/wizard.go, and internal/cli/new.go to understand the current wizard flow.

## New Wizard Flow

### Step 1: Project Basics
```
Service name: [my-service]
Module path:  [github.com/org/my-service]
Output dir:   [.]
```

### Step 2: Architecture
```
What architecture pattern?
  → hexagonal (recommended for APIs and services)
  → flat (simple scripts and CLIs)
```
(List comes from scanning architectures/ directory)

### Step 3: Capabilities (Multi-Select)
```
What does your service need? (space to toggle, enter to confirm)

  Transport:
    [x] HTTP API          Chi router, middleware, OpenAPI spec
    [ ] gRPC              Protocol buffers, interceptors

  Data:
    [x] MySQL             Connection pool, health check
    [ ] Redis             Cache layer, connection pool

  Messaging:
    [ ] Kafka Consumer    Consumer group, handler pattern

  Resilience:
    [ ] Rate Limiting     Token bucket middleware

  Auth:
    [ ] JWT Auth          JWT middleware, token validation

  Observability:
    [x] Observability     Structured logging, OpenTelemetry, health check

  Quality:
    [x] Testing           Example tests, test helpers
    [x] Linting           golangci-lint config
    [ ] Pre-commit        Pre-commit hooks

  Infrastructure:
    [x] Docker            Dockerfile, docker-compose
    [ ] CI (GitHub)       GitHub Actions workflows

  Integration:
    [ ] Email Gateway     Email sending adapter
    [ ] HTTP Client       Outbound HTTP client with retry
```

### Step 4: Suggestions (The Differentiator)
After user confirms capabilities, analyze for gaps:

```go
type Suggestion struct {
    Capability  string
    Reason      string
    TriggerRule string  // e.g., "http-api selected but no rate-limiting"
}

var suggestionRules = []SuggestionRule{
    {If: has("http-api"), Missing: "rate-limiting", Reason: "HTTP APIs benefit from rate limiting to prevent abuse"},
    {If: has("http-api"), Missing: "auth-jwt", Reason: "HTTP APIs typically need authentication"},
    {If: has("http-api"), Missing: "observability", Reason: "APIs need health checks and structured logging for production"},
    {If: hasAny("mysql", "redis"), Missing: "observability", Reason: "Database connections need health checks and monitoring"},
    {If: has("kafka-consumer"), Missing: "observability", Reason: "Message consumers need logging and metrics"},
    {If: not(hasAny("ci-github")), Always: true, Reason: "CI/CD catches issues before they reach production"},
    {If: not(has("docker")), And: hasAny("mysql", "redis", "kafka-consumer"), Reason: "docker-compose simplifies local development with external dependencies"},
    {If: has("http-api"), Missing: "testing", Reason: "APIs need handler tests for reliability"},
    {If: has("grpc"), Missing: "observability", Reason: "gRPC services need interceptor-based observability"},
}
```

Display:
```
💡 Based on your selections, you might also want:

  [ ] Rate Limiting     — HTTP APIs benefit from rate limiting to prevent abuse
  [ ] JWT Auth          — HTTP APIs typically need authentication
  [ ] CI (GitHub)       — CI/CD catches issues before they reach production

  (space to add, enter to skip all)
```

### Step 5: Confirmation
```
Ready to scaffold:

  Project:        my-service
  Module:         github.com/org/my-service
  Architecture:   hexagonal
  Capabilities:   http-api, mysql, observability, testing, linting, docker
  Output:         ./my-service

  Equivalent command:
  archway new my-service --arch hexagonal --cap http-api,mysql,observability,testing,linting,docker

  Proceed? [Y/n]
```

## Implementation Steps

### 1. Create internal/scaffold/suggestions.go
- Define SuggestionRule struct with If/Missing/Reason fields
- Define the suggestion rules table
- Function: `ComputeSuggestions(selected []string) []Suggestion`
- Unit tests for all suggestion rules

### 2. Update internal/scaffold/tui.go
- New function: RunCompositionWizard() that implements Steps 1-5
- Use charmbracelet/huh for all form elements
- Step 3 uses huh.MultiSelect grouped by category
- Step 4 shows suggestions only if any exist
- Step 5 shows summary + non-interactive equivalent

### 3. Update internal/cli/new.go
- Replace the current wizard flow with RunCompositionWizard()
- Support --arch flag (architecture selection)
- Support --cap flag (comma-separated capabilities)
- Support --set flag (variable overrides, kept from current)
- If --arch and --cap provided, skip wizard (non-interactive mode)
- After wizard, call the composition renderer from Phase 4

### 4. Add --no-wizard flag
- Skip all interactive prompts, use flags only
- Require --arch and --cap when --no-wizard is set

### 5. Print non-interactive equivalent
After wizard completes, print the equivalent CLI command (from go-blueprint research).

### 6. Tests
- Test suggestion engine rules
- Test wizard state transitions
- Test non-interactive mode with flags
- Test that all capabilities from the directory are listed

go build ./... && go test ./... && go vet ./...

When complete, output: PHASE 5 COMPLETE
```

---

## Phase 6: Build `archway check` Command

**Objective:** Implement the `archway check` command that validates an existing project against its archway.yaml rules and reports violations.
**Model:** `opus`
**Max Iterations:** 20
**Completion Promise:** `PHASE 6 COMPLETE`
**Dependencies:** Phase 2

**Prompt:**
```
Build the `archway check` command that validates a Go project against its archway.yaml rules.

## Context
Read docs/soul.md and ~/.claude/projects/-Users-danielgomes-go-src-github-com-dcsg-archway/memory/competitor-research.md (the arch-go, go-arch-lint, and ArchUnit sections). Key insights:
- arch-go: compliance + coverage metrics
- go-arch-lint: named components, technical debt tracking
- ArchUnit: precise violations with file/line, FreezingArchRule

## Existing Analysis
Read internal/analyzer/analyzer.go, internal/analyzer/graph/graph.go (LayerViolations), and internal/config/archway_yaml.go.

The analyzer already:
- Loads packages via go/packages
- Builds a dependency graph
- Detects layer violations via LayerViolations()
- Has StructureConfig (RequiredDirs, ForbiddenDirs) and FunctionRules (MaxLines, MaxParams)

But:
- Structure validation is NOT implemented (schema only)
- Function rule validation is NOT implemented (schema only)
- No CLI command wraps this
- No compliance/coverage metrics
- No formatted output

## archway check Output Format

```
$ archway check

Archway Check — my-service
═══════════════════════════════════════════════════════

Components:  5 defined, 4 covered (80% coverage)
Rules:       12 checked

DEPENDENCY VIOLATIONS (2)
  ✗ domain/order.go:15 imports adapter/mysqlrepo
    Component "domain" may not depend on "adapters"
  ✗ service/handler.go:8 imports platform/observability
    Component "service" may not depend on "platform"

STRUCTURE VIOLATIONS (1)
  ✗ Missing required directory: internal/port/

FUNCTION VIOLATIONS (3)
  ✗ adapter/httphandler/handler.go:ProcessOrder — 127 lines (max: 80)
  ✗ service/order.go:CreateOrder — 5 params (max: 4)
  ✗ adapter/mysqlrepo/connection.go:Query — 4 return values (max: 2)

NAMING VIOLATIONS (0)
  ✓ All naming rules pass

═══════════════════════════════════════════════════════
Result: FAIL — 6 violations found
  Compliance: 50% (6/12 rules passing)
  Coverage:   80% (4/5 components checked)
```

## Implementation Steps

### 1. Create internal/checker/checker.go
New package that orchestrates all checks:

```go
type CheckResult struct {
    DependencyViolations  []Violation
    StructureViolations   []Violation
    FunctionViolations    []Violation
    NamingViolations      []Violation
    ComponentsCovered     int
    ComponentsTotal       int
    RulesChecked          int
    RulesPassing          int
}

type Violation struct {
    Category string   // "dependency", "structure", "function", "naming"
    File     string   // file path
    Line     int      // line number (0 if N/A)
    Message  string   // human-readable description
    Rule     string   // which rule was violated
    Severity string   // "error" or "warning"
}

func Check(cfg *config.ArchwayConfig, path string) (*CheckResult, error)
```

### 2. Implement dependency checks
- Reuse existing graph.LayerViolations() but update it to work with the new Component type (from Phase 2)
- Add file + line number to violations (use go/packages position info)
- Calculate component coverage (which components have at least one package)

### 3. Implement structure checks
- For each RequiredDir: check if directory exists
- For each ForbiddenDir: check if directory exists (violation if it does)

### 4. Implement function checks
- Walk all Go files using go/ast
- For each function declaration:
  - Count lines (FuncDecl.End - FuncDecl.Pos)
  - Count params (FuncType.Params.NumFields())
  - Count return values (FuncType.Results.NumFields())
  - Compare against FunctionRules from archway.yaml
  - Record violations with file path and line number

### 5. Implement naming checks
- For each NamingRule: find matching packages/files
- Check patterns (MustEndWith, MustStartWith, Pattern regex)

### 6. Create internal/cli/check.go
Register `archway check` command:
```go
func newCheckCommand(opts *rootOptions) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "check",
        Short: "Validate project against archway.yaml rules",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Load archway.yaml
            // 2. Run checker.Check()
            // 3. Format and print results
            // 4. Exit 1 if violations found (for CI)
        },
    }
    cmd.Flags().StringVar(&path, "path", ".", "Project path")
    cmd.Flags().StringVar(&format, "output", "terminal", "Output format (terminal|json)")
    return cmd
}
```

### 7. Register in root.go
Add newCheckCommand(opts) to cmd.AddCommand()

### 8. JSON output
Support --output json for CI integration:
```json
{
  "result": "fail",
  "compliance": 0.5,
  "coverage": 0.8,
  "violations": [
    {"category": "dependency", "file": "domain/order.go", "line": 15, "message": "..."}
  ]
}
```

### 9. Tests
- Test with a project that has violations (use testdata fixtures)
- Test with a clean project (no violations)
- Test each violation type independently
- Test JSON output format
- Test exit code (0 for pass, 1 for fail)

### 10. Verify
go build ./... && go test ./... && go vet ./...

When complete, output: PHASE 6 COMPLETE
```

---

## Phase 7: Tests, Cleanup, and Documentation

**Objective:** Add CLI tests, integration tests, update all documentation, and clean up any remaining issues.
**Model:** `sonnet`
**Max Iterations:** 20
**Completion Promise:** `PHASE 7 COMPLETE`
**Dependencies:** Phases 1-6

**Prompt:**
```
Final phase: add comprehensive tests, clean up, update all documentation.

## Context
Read docs/soul.md and the progress table in this plan file.

## Steps

### 1. CLI Tests (currently 0% coverage)
Create internal/cli/cli_test.go with tests for:
- `archway new --name test --arch hexagonal --cap http-api --no-wizard --output /tmp/test` (non-interactive scaffold)
- `archway check --path ./testdata/clean-project` (no violations)
- `archway check --path ./testdata/violated-project` (violations detected, exit 1)
- `archway analyze --path ./testdata/hexagonal-project`
- `archway init --path ./testdata/existing-project --no-wizard`
- `archway version` (prints version)
- Unknown command (error)
- Help output contains all commands

### 2. Integration Test
Create internal/scaffold/integration_test.go:
- Scaffold a hexagonal project with http-api + mysql + observability + testing
- Verify output directory structure matches expected
- Verify go.mod exists and has correct module path
- Verify archway.yaml was generated with correct architecture + capabilities + components
- Verify all expected files exist
- Optionally: run `go build ./...` on the scaffolded project (if CI has Go installed)

### 3. Suggestion Engine Tests
Create internal/scaffold/suggestions_test.go:
- Test each suggestion rule individually
- Test that http-api suggests rate-limiting, auth-jwt, observability
- Test that mysql suggests observability
- Test that no suggestions when everything is selected
- Test edge cases (empty selection, single capability)

### 4. Update docs/soul.md
- Update "What This Is" to reflect the new identity
- Remove any references to MCP, LLM, configure command
- Add Capabilities and Components to the description
- Update Current State

### 5. Update docs/product/spec.md
- Add `archway check` to shipped features
- Update feature roadmap
- Update architecture description

### 6. Update README.md
- New tagline: "Architecture-aware service composer"
- Update CLI examples
- Show composition wizard flow
- Show `archway check` output
- Remove MCP and LLM references
- Add "Integrates with Keel" section

### 7. Update docs/decisions/
- Write ADR-007: Composition model (architecture + capabilities)
- Write ADR-008: Kill MCP and LLM in v1

### 8. Clean up test artifacts
- Add test-service/, test-seviceb/, testservice/ to .gitignore
- Add archway binary and bin/ to .gitignore
- Remove any leftover references to old template paths

### 9. Final verification
- go build ./...
- go test ./... -count=1
- go vet ./...
- golangci-lint run (if available)
- Verify all commands work: archway version, archway new --help, archway check --help, archway analyze --help

When complete, output: PHASE 7 COMPLETE
```

---

## Known Risks

### From Pre-Flight Review (2026-03-08)

**🔴 Critical (addressed in plan):**
1. **main.go fragment assembler needs topological sorting** — Each capability contributes import/init/shutdown fragments with declared init order (e.g., `mysql` before `http-handler`). Phase 4's `_partials/` approach must include an `order` field per capability and the renderer must topologically sort before assembly.
2. **Capabilities must be first-class entities** — Not boolean template variables. Each capability needs its own `capability.yaml` with `name`, `version`, `requires`, `conflicts`, `provides`. Phase 3's manifest.yaml per capability already addresses this, but must include requires/conflicts fields.

**🟡 Warnings (noted, some deferred):**
3. Architecture axis must be a real template selector, not just metadata — addressed in Phase 3 (architectures/ directory).
4. Provider-scaffold coupling — extract shared orchestrator in Phase 4 to avoid per-language duplication.
5. Go-specific defaults in `DefaultArchwayConfig()` — make provider-driven in Phase 2.
6. `archway check` scope — focus on architecture rules (cross-package deps, structure), NOT function-level rules that golangci-lint handles. Simplify FunctionRules or remove.
7. `extends` resolution — deferred to v1.5. Don't implement; remove or comment out.
8. Suggestion rules need compound predicates — implement with a simple DSL (hasAll/hasAny/not) in Phase 5.
9. Wizard + manifest variables should be co-located per capability — addressed by Phase 3 restructuring.
10. Third-party capability plugins — v2 feature. For v1, capabilities are embedded only. Design the registry interface but don't implement external loading.
