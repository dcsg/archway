# Plan: Archway v1.1 to v1.4

## Overview
**Task:** Implement the unified v1.1-v1.4 roadmap — proxy rules engine, new architectures + capability expansion, design companion, token compaction + decision gates.
**Total Phases:** 20
**Estimated Cost:** ~$12.50
**Created:** 2026-03-10

## Progress

| Phase | Status | Updated |
|-------|--------|---------|
| 1     | done   | 2026-03-10 |
| 2     | done   | 2026-03-10 |
| 3     | done   | 2026-03-10 |
| 4     | done   | 2026-03-10 |
| 5     | done   | 2026-03-10 |
| 6     | done   | 2026-03-10 |
| 7     | done   | 2026-03-10 |
| 8     | done   | 2026-03-10 |
| 9     | done   | 2026-03-10 |
| 10    | done   | 2026-03-10 |
| 11    | done   | 2026-03-11 |
| 12    | done   | 2026-03-11 |
| 13    | done   | 2026-03-11 |
| 14    | done   | 2026-03-11 |
| 15    | done   | 2026-03-11 |
| 16    | done   | 2026-03-11 |
| 17    | done   | 2026-03-11 |
| 18    | done   | 2026-03-11 |
| 19    | done   | 2026-03-11 |
| 20    | done   | 2026-03-11 |

**IMPORTANT:** Update this table as phases complete. This table is the persistent state that survives context compaction.

## Version Boundaries

| Version | Phases | Theme |
|---------|--------|-------|
| v1.1    | 1-5    | Proxy Rules + Pre-commit (Enforce) |
| v1.2    | 6-12   | New Architectures + Capability Expansion (Compose) |
| v1.3    | 13-16  | Design Companion + Capability Catalog (Guide) |
| v1.4    | 17-20  | Token Compaction + Decision Gates (Guide + Enforce) |

## Model Assignment

| Phase | Task | Model | Reasoning | Est. Cost |
|-------|------|-------|-----------|-----------|
| 1 | Rule engine core (parser, grep, scope) | opus | Novel package, architecture decisions | ~$0.80 |
| 2 | AST engine routing + checker integration | sonnet | Integration of existing code | ~$0.08 |
| 3 | CLI flags + reporting | sonnet | Straightforward CLI work | ~$0.08 |
| 4 | Auto-generate rules from archway.yaml | opus | Novel feature, closes Guide→Enforce loop | ~$0.80 |
| 5 | v1.1 tests | sonnet | Test scaffolding, well-defined patterns | ~$0.08 |
| 6 | Layered architecture | sonnet | Template work, clear structure | ~$0.08 |
| 7 | Clean architecture | sonnet | Template work, clear structure | ~$0.08 |
| 8 | Capabilities: transport + data (6 caps) | sonnet | Template work, repeatable pattern | ~$0.08 |
| 9 | Capabilities: patterns + devex (6 caps) | sonnet | Template work, repeatable pattern | ~$0.08 |
| 10 | Capabilities: frontend Go (3 caps) | opus | Novel territory, templ/HTMX patterns | ~$0.80 |
| 11 | Suggestions + compatibility matrix | sonnet | Rule updates, matrix expansion | ~$0.08 |
| 12 | v1.2 tests + docs | sonnet | Test scaffolding + doc updates | ~$0.08 |
| 13 | Capability catalog content generation | opus | Core guide enhancement, content design | ~$0.80 |
| 14 | Interaction warnings + suggestions in guide | sonnet | Uses existing suggestion engine | ~$0.08 |
| 15 | Catalog-only mode + rule summaries | sonnet | CLI flag + content integration | ~$0.08 |
| 16 | v1.3 tests | sonnet | Test scaffolding | ~$0.08 |
| 17 | Split guide output (index + categories) | opus | Architectural refactor of guide | ~$0.80 |
| 18 | Decision gates schema + archway decide CLI | opus | Novel feature, UX decisions | ~$0.80 |
| 19 | archway check --decisions | sonnet | Integration with existing checker | ~$0.08 |
| 20 | v1.4 tests | sonnet | Test scaffolding | ~$0.08 |

## Execution Strategy

| Phase | Depends On | Parallel With |
|-------|-----------|---------------|
| 1     | None      | -             |
| 2     | 1         | -             |
| 3     | 1, 2      | -             |
| 4     | 1         | 2, 3          |
| 5     | 1, 2, 3, 4| -             |
| 6     | None      | 7             |
| 7     | None      | 6             |
| 8     | None      | 6, 7, 9       |
| 9     | None      | 6, 7, 8       |
| 10    | None      | 6, 7, 8, 9    |
| 11    | 6, 7, 8, 9, 10 | -         |
| 12    | 6, 7, 8, 9, 10, 11 | -     |
| 13    | 12        | -             |
| 14    | 13        | -             |
| 15    | 13, 14    | -             |
| 16    | 13, 14, 15| -             |
| 17    | 16        | -             |
| 18    | 16        | 17            |
| 19    | 18        | -             |
| 20    | 17, 18, 19| -             |

---

## ═══════════════════════════════════════════
## v1.1 — PROXY RULES + PRE-COMMIT
## ═══════════════════════════════════════════

---

## Phase 1: Rule Engine Core

**Objective:** Build the proxy rule engine — YAML parser, grep engine, scope matching, and startup validation.
**Model:** `opus`
**Max Iterations:** 15
**Completion Promise:** `RULE ENGINE CORE DONE`
**Dependencies:** None

**Prompt:**
```
Implement the proxy rule engine for Archway per Design 008. This is a new package: `internal/rules/`.

Read these files first for context:
- `docs/soul.md` — project context
- `internal/checker/checker.go` — existing checker (you'll integrate with this later)
- `internal/checker/antipatterns.go` — existing anti-pattern detectors
- `internal/config/config.go` — config types

### What to Build

Create `internal/rules/` package with these files:

#### 1. `internal/rules/rule.go` — Rule types

```go
type Rule struct {
    ID          string   `yaml:"id"`
    Engine      string   `yaml:"engine"`      // "grep" or "ast"
    Description string   `yaml:"description"`
    Severity    string   `yaml:"severity"`     // "error" or "warning"
    Ref         string   `yaml:"ref,omitempty"` // back-reference to doc
    // Grep engine fields
    Pattern        string `yaml:"pattern,omitempty"`
    MustContain    string `yaml:"must-contain,omitempty"`
    MustNotContain string `yaml:"must-not-contain,omitempty"`
    FileMustContain string `yaml:"file-must-contain,omitempty"`
    // AST engine fields
    Detector string `yaml:"detector,omitempty"` // name of built-in detector
    // Scope
    Scope   []string `yaml:"scope"`
    Exclude []string `yaml:"exclude,omitempty"`
}

type RuleViolation struct {
    RuleID      string
    Engine      string
    Description string
    Severity    string
    Ref         string
    File        string
    Line        int
    Match       string // the matched line content
}

type RuleStatus struct {
    Rule   Rule
    Status string // "valid", "invalid", "stale"
    Error  string // reason if invalid
}
```

#### 2. `internal/rules/loader.go` — YAML loading + validation

```go
// LoadRules reads all .yaml files from a directory and returns parsed rules + statuses
func LoadRules(rulesDir string) ([]Rule, []RuleStatus, error)

// ValidateRule checks a single rule for correctness
func ValidateRule(r Rule, projectRoot string) RuleStatus
```

Validation rules:
- `id` required, non-empty
- `engine` must be "grep" or "ast"
- `severity` must be "error" or "warning" (default to "error")
- For grep engine: at least one of `pattern` or `file-must-contain` required
- For ast engine: `detector` required
- `scope` required, non-empty
- If scope globs match 0 files → status is "stale" (not invalid)
- Malformed YAML or missing required fields → status is "invalid"

Use `filepath.Glob()` or `doublestar` for glob matching against project files.

#### 3. `internal/rules/grep.go` — Grep engine

```go
// RunGrep executes a grep-engine rule against files matching its scope
func RunGrep(rule Rule, projectRoot string) ([]RuleViolation, error)
```

Logic:
1. Expand scope globs to get file list (respecting exclude patterns)
2. For each file:
   a. If `file-must-contain` is set: read entire file, check if regex matches anywhere. If NOT found, report violation for the file (line 0).
   b. If `pattern` is set: scan line by line with regex
   c. For each line matching `pattern`:
      - If `must-contain` set: check line also matches must-contain. If NOT → violation
      - If `must-not-contain` set: check line does NOT match must-not-contain. If it DOES → violation
      - If neither must-contain nor must-not-contain: the pattern match itself IS the violation (used for "file must not contain this pattern")
3. Build RuleViolation with file, line number, matched content

Use Go's `regexp` package. Compile patterns once per rule, not per file.

#### 4. `internal/rules/scope.go` — Scope matching

```go
// ExpandScope returns all files matching scope globs minus exclude globs
func ExpandScope(scope, exclude []string, projectRoot string) ([]string, error)
```

- Walk the project directory
- Match each file path against scope globs (use `filepath.Match` or `doublestar.Match`)
- Exclude files matching exclude globs
- Return sorted list of matching file paths
- Skip `.git/`, `vendor/`, `node_modules/` directories automatically

#### 5. `internal/rules/engine.go` — Engine dispatcher

```go
// RunRules loads rules from directory, validates them, runs valid ones, returns results
func RunRules(rulesDir, projectRoot string) (*RunResult, error)

type RunResult struct {
    Violations []RuleViolation
    Statuses   []RuleStatus   // all rules with their validation status
    Duration   time.Duration
}

// Convenience methods
func (r *RunResult) ErrorCount() int    // violations with severity "error"
func (r *RunResult) WarningCount() int  // violations with severity "warning"
func (r *RunResult) ValidRuleCount() int
func (r *RunResult) InvalidRuleCount() int
func (r *RunResult) StaleRuleCount() int
```

### Important Design Decisions

- Rules directory: `.archway/rules/` (hardcoded for now)
- Grep engine only in this phase (AST routing in Phase 2)
- For AST engine rules, return an error "ast engine not yet implemented" — Phase 2 handles this
- Skip binary files when scanning (detect by checking first 512 bytes for null bytes)
- Maximum file size: 1MB (skip larger files with a warning)
- Regex compilation errors should make the rule "invalid", not crash

### Tests

Create `internal/rules/loader_test.go`, `internal/rules/grep_test.go`, `internal/rules/scope_test.go`:

- Test YAML loading with valid/invalid/malformed files
- Test grep engine with pattern, must-contain, must-not-contain, file-must-contain combinations
- Test scope expansion with globs and excludes
- Test stale detection (scope matches 0 files)
- Test binary file skipping
- Use `t.TempDir()` for test fixtures

Create test fixtures in `internal/rules/testdata/`:
- `valid-grep-rule.yaml` — well-formed grep rule
- `invalid-no-engine.yaml` — missing engine field
- `invalid-no-scope.yaml` — missing scope
- Sample Go source files for grep testing

Run `go build ./...` and `go test ./internal/rules/...` when done.

When complete, output: RULE ENGINE CORE DONE
```

---

## Phase 2: AST Engine Routing + Checker Integration

**Objective:** Connect proxy rule AST engine to existing 11 detectors, and integrate RunResult into CheckResult.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `AST ENGINE INTEGRATED`
**Dependencies:** Phase 1

**Prompt:**
```
Integrate the proxy rule engine (internal/rules/) with the existing checker (internal/checker/).

Read these files first:
- `internal/rules/engine.go` — the rule engine from Phase 1
- `internal/rules/rule.go` — rule types
- `internal/checker/checker.go` — existing checker
- `internal/checker/antipatterns.go` — the 11 detectors

### Tasks

#### 1. AST Engine Implementation (`internal/rules/ast.go`)

Create the AST engine that routes to existing detectors:

```go
// RunAST executes an ast-engine rule by delegating to a named built-in detector
func RunAST(rule Rule, projectRoot string, packages []*packages.Package) ([]RuleViolation, error)
```

Map detector names to existing functions:
- "global-mutable-state" → detectGlobalMutableState
- "init-abuse" → detectInitAbuse
- "naked-goroutines" → detectNakedGoroutines
- "swallowed-errors" → detectSwallowedErrors
- "context-background" → detectContextBackground
- "sql-concatenation" → detectSQLConcatenation
- "uuid-v4-as-key" → detectUUIDv4AsKey
- "fat-handlers" → detectFatHandlers
- "god-packages" → detectGodPackages
- "domain-importing-adapters" → detectDomainImportingAdapters
- "mvc-in-hexagonal" → detectMVCInHexagonal

This requires making the detector functions accessible from outside the checker package. Options:
- Export them (rename to Detect...)
- Create a detector registry in the checker package
- Move detectors to internal/rules/ (least preferred — keep existing code stable)

Recommended: Create a registry in checker:
```go
// internal/checker/registry.go
var DetectorRegistry = map[string]DetectorFunc{
    "global-mutable-state": detectGlobalMutableState,
    // ...
}
type DetectorFunc func(*ast.File, *token.FileSet, string) []AntiPattern
```

The AST engine converts AntiPattern results to RuleViolation format.

For unknown detector names → rule status "invalid" with error "unknown detector: {name}".

#### 2. Integrate into Check() (`internal/checker/checker.go`)

Modify `Check()` to also run proxy rules:

```go
func Check(cfg *config.ArchwayConfig, projectPath string) (*CheckResult, error) {
    // ... existing checks ...

    // NEW: Run proxy rules if .archway/rules/ exists
    rulesDir := filepath.Join(projectPath, ".archway", "rules")
    if dirExists(rulesDir) {
        runResult, err := rules.RunRules(rulesDir, projectPath)
        if err != nil {
            // log warning, don't fail the whole check
        }
        result.ProxyRuleResult = runResult
    }

    return result, nil
}
```

Add to CheckResult:
```go
type CheckResult struct {
    // ... existing fields ...
    ProxyRuleResult *rules.RunResult // nil if no rules directory
}

// Update TotalViolations() to include proxy rule errors
// Update Passed() to account for proxy rule errors
```

#### 3. Update Output Formatters

Update `internal/cli/check.go` terminal output to show proxy rule results:

```
PROXY RULES (12 checked, 1 invalid, 1 stale)
  ✗ internal/handler/order.go:4 — Handler imports database/sql directly
    Rule: INV-001-R1 (grep) | Ref: docs/invariants/001-layer-separation.md
  ⚠ internal/service/order.go:15 — SQL query without tenant_id
    Rule: INV-003-R2 (grep) | Ref: docs/invariants/003-tenant-isolation.md
  ⊘ INVALID: SEC-002-R1.yaml — unknown engine "semgreppp"
  ⊘ STALE: INV-007-R3.yaml — scope matches 0 files
```

Update JSON output to include proxy rule violations.

Run `go build ./...` and `go test ./...` when done.

When complete, output: AST ENGINE INTEGRATED
```

---

## Phase 3: CLI Flags + Reporting

**Objective:** Add --proxy-rules, --staged, --rule flags and severity-based exit codes.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `CLI FLAGS DONE`
**Dependencies:** Phase 1, Phase 2

**Prompt:**
```
Add CLI flags to `archway check` for proxy rule control.

Read these files first:
- `internal/cli/check.go` — existing check command
- `internal/rules/engine.go` — rule engine
- `internal/checker/checker.go` — checker integration from Phase 2

### Tasks

#### 1. New CLI Flags

Add to `archway check`:

```go
--proxy-rules    // Run ONLY proxy rules (skip built-in detectors)
--detectors      // Run ONLY built-in detectors (skip proxy rules)
--rule ID        // Run a single proxy rule by ID
--staged         // Only check files in git staging area
```

Default (no flags): run both built-in detectors AND proxy rules.

#### 2. --staged Implementation

When `--staged` is set:
1. Run `git diff --cached --name-only --diff-filter=ACM` to get staged files
2. Filter the file list — only check staged files
3. Pass this filtered list to both the checker and rule engine
4. This requires modifying the scope expansion to accept an optional file filter

Add to `internal/rules/scope.go`:
```go
func ExpandScopeFiltered(scope, exclude []string, projectRoot string, allowedFiles []string) ([]string, error)
```

If `allowedFiles` is nil, behave as before (all files). If set, intersect with scope results.

For the built-in checker: add an optional file filter to `Check()`:
```go
func Check(cfg *config.ArchwayConfig, projectPath string, opts ...CheckOption) (*CheckResult, error)

type CheckOption func(*checkOptions)
func WithFileFilter(files []string) CheckOption
func WithProxyRulesOnly() CheckOption
func WithDetectorsOnly() CheckOption
func WithSingleRule(id string) CheckOption
```

#### 3. Exit Codes

- Exit 0: no error-severity violations (warnings are OK)
- Exit 1: at least one error-severity violation
- Invalid/stale rules: report but don't affect exit code

#### 4. --staged Pre-commit Script

When `--staged` is used, print a helpful message at the end:

```
Tip: Add to .git/hooks/pre-commit:
  #!/bin/sh
  archway check --staged
```

Run `go build ./...` and `go test ./...` when done.

When complete, output: CLI FLAGS DONE
```

---

## Phase 4: Auto-Generate Rules from archway.yaml

**Objective:** Make `archway guide` generate proxy rules in `.archway/rules/` based on declared architecture and capabilities.
**Model:** `opus`
**Max Iterations:** 12
**Completion Promise:** `RULE GENERATION DONE`
**Dependencies:** Phase 1

**Prompt:**
```
Extend `archway guide` to auto-generate proxy rules from the project's archway.yaml. This closes the Guide → Enforce loop: the same source of truth (archway.yaml) generates both AI instructions AND enforceable rules.

Read these files first:
- `internal/guide/guide.go` — current guide generation
- `internal/guide/targets.go` — output targets
- `internal/rules/rule.go` — rule types from Phase 1
- `internal/config/config.go` — ArchwayConfig struct
- `providers/golang/templates/architectures/hexagonal/manifest.yaml` — hexagonal components

### What to Build

#### 1. Rule Generator (`internal/guide/rules.go`)

```go
// GenerateRules creates proxy rule YAML files based on the project's architecture and capabilities
func GenerateRules(cfg *config.ArchwayConfig, outputDir string) ([]GeneratedRule, error)

type GeneratedRule struct {
    Filename string
    Rule     rules.Rule
}
```

#### 2. Architecture-Based Rules

For **hexagonal** architecture, generate:

```yaml
# .archway/rules/arch-domain-isolation.yaml
id: arch-domain-isolation
engine: grep
description: "Domain layer must not import adapter or infrastructure packages"
severity: error
ref: archway.yaml
pattern: '"[^"]*/(adapter|infrastructure|platform)/[^"]*"'
scope:
  - "domain/**/*.go"
exclude:
  - "*_test.go"
```

```yaml
# .archway/rules/arch-port-direction.yaml
id: arch-port-direction
engine: grep
description: "Port layer must not import adapter or service packages"
severity: error
ref: archway.yaml
pattern: '"[^"]*/(adapter|service)/[^"]*"'
scope:
  - "port/**/*.go"
exclude:
  - "*_test.go"
```

For **flat** architecture: no layer rules generated (no restrictions).

Generate rules dynamically from `archway.yaml` components:
- For each component, look at what it may NOT depend on (everything not in `may_depend_on`)
- Generate a grep rule that catches imports from forbidden layers
- Use the module path from archway.yaml to build accurate import patterns

#### 3. Capability-Based Rules

For capabilities, generate rules based on known best practices:

**If has `mysql` or `postgres`:**
```yaml
id: cap-sql-parameterized
engine: grep
description: "SQL queries must use parameterized queries, not string concatenation"
severity: error
pattern: '(fmt\.Sprintf|"+.*SELECT|"+.*INSERT|"+.*UPDATE|"+.*DELETE)'
scope:
  - "**/*.go"
exclude:
  - "*_test.go"
  - "vendor/**"
```

**If has `http-api`:**
```yaml
id: cap-handler-context
engine: grep
description: "HTTP handlers should use request context, not context.Background()"
severity: warning
pattern: "context\\.Background\\(\\)"
scope:
  - "adapter/httphandler/**/*.go"
  - "internal/handler/**/*.go"
exclude:
  - "*_test.go"
```

**If has `auth-jwt`:**
```yaml
id: cap-auth-check
engine: grep
description: "Handler files should reference auth middleware or JWT validation"
severity: warning
file-must-contain: "(middleware|auth|jwt|token)"
scope:
  - "adapter/httphandler/**/*.go"
  - "internal/handler/**/*.go"
exclude:
  - "*_test.go"
  - "**/router.go"
```

**If has `observability`:**
```yaml
id: cap-tracing-context
engine: grep
description: "Functions making external calls should pass context for tracing"
severity: warning
pattern: "(http\\.Get|http\\.Post|sql\\.Query|sql\\.Exec)"
must-not-contain: "ctx"
scope:
  - "**/*.go"
exclude:
  - "*_test.go"
  - "vendor/**"
```

Define a map of capability → rule templates. Start with rules for these capabilities:
- mysql/postgres (SQL safety)
- http-api (handler patterns)
- auth-jwt (auth enforcement)
- observability (context propagation)
- grpc (protobuf imports)
- kafka-consumer (error handling patterns)

#### 4. Integration with `archway guide`

Modify `GenerateFromConfig()` to also generate rules:

```go
func GenerateFromConfig(projectDir string, cfg *config.ArchwayConfig, target string, templateFS ...fs.FS) error {
    // ... existing guide generation ...

    // NEW: Generate proxy rules
    rulesDir := filepath.Join(projectDir, ".archway", "rules")
    os.MkdirAll(rulesDir, 0755)
    generated, err := GenerateRules(cfg, rulesDir)
    // Report what was generated
}
```

Rules should be idempotent — regenerating overwrites existing auto-generated rules. Add a header comment to each generated file:
```yaml
# Auto-generated by `archway guide` from archway.yaml
# Regenerate with: archway guide
# To customize: copy this file, change the id, and modify as needed
```

#### 5. Don't Overwrite User Rules

Auto-generated rules use the `arch-` and `cap-` prefixes. User/AI-authored rules use other prefixes (e.g., `INV-`, `ADR-`, `SEC-`). The generator should:
- Only write files matching `arch-*.yaml` and `cap-*.yaml`
- Never touch files with other prefixes
- Delete stale auto-generated rules (e.g., if a capability is removed from archway.yaml)

### Tests

Create `internal/guide/rules_test.go`:
- Test hexagonal generates domain-isolation + port-direction rules
- Test flat generates no layer rules
- Test mysql capability generates SQL parameterized rule
- Test http-api capability generates handler context rule
- Test idempotency (run twice, same result)
- Test user rules are not overwritten

Run `go build ./...` and `go test ./...` when done.

When complete, output: RULE GENERATION DONE
```

---

## Phase 5: v1.1 Tests + Integration

**Objective:** End-to-end tests for the full proxy rules pipeline — load, validate, run, report.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `V1.1 TESTS PASSING`
**Dependencies:** Phase 1, 2, 3, 4

**Prompt:**
```
Write comprehensive tests for the v1.1 proxy rules feature.

Read these files first:
- `internal/rules/` — all rule engine files from Phases 1-2
- `internal/guide/rules.go` — auto-generation from Phase 4
- `internal/cli/check.go` — CLI integration from Phase 3
- `internal/checker/antipatterns_test.go` — existing test patterns to follow
- `internal/cli/cli_test.go` — existing CLI test patterns

### Tests to Write

#### 1. Integration Test: Full Pipeline (`internal/rules/integration_test.go`)

```go
func TestFullPipeline(t *testing.T) {
    // 1. Create temp project with archway.yaml + Go source files
    // 2. Create .archway/rules/ with grep rules
    // 3. Run RunRules()
    // 4. Assert violations found in expected files/lines
    // 5. Assert no violations in clean files
}

func TestFullPipelineWithAST(t *testing.T) {
    // Same but with engine: ast rules referencing built-in detectors
}

func TestMixedEngines(t *testing.T) {
    // Mix grep and ast rules, verify both run and results merge
}
```

#### 2. CLI Integration Test (`internal/cli/e2e_test.go` — extend existing)

Add test cases:
```go
func TestCheckWithProxyRules(t *testing.T) {
    // Scaffold project → add .archway/rules/ → run archway check → verify output
}

func TestCheckProxyRulesOnly(t *testing.T) {
    // --proxy-rules flag → only rule violations, no anti-patterns
}

func TestCheckDetectorsOnly(t *testing.T) {
    // --detectors flag → only anti-patterns, no rule violations
}

func TestCheckStaged(t *testing.T) {
    // Init git repo → stage files → run --staged → only staged files checked
}
```

#### 3. Rule Generation Tests (`internal/guide/rules_test.go` — extend from Phase 4)

```go
func TestGenerateAndCheck(t *testing.T) {
    // 1. Create project with archway.yaml (hexagonal + mysql + http-api)
    // 2. Run GenerateRules()
    // 3. Add violating Go source files
    // 4. Run rules.RunRules()
    // 5. Assert violations caught by generated rules
}

func TestGeneratePreservesUserRules(t *testing.T) {
    // 1. Create user rule (INV-001-R1.yaml)
    // 2. Run GenerateRules()
    // 3. Assert user rule still exists unchanged
    // 4. Assert auto-generated rules also exist
}
```

#### 4. Edge Cases

- Empty rules directory → no violations, no errors
- Rules directory doesn't exist → skip silently
- Rule with regex compilation error → marked invalid, other rules still run
- Very large file (>1MB) → skipped with warning
- Binary file → skipped silently
- Rule matching 0 scope files → marked stale

Run `go test ./...` to verify everything passes.

When complete, output: V1.1 TESTS PASSING
```

---

## ═══════════════════════════════════════════
## v1.2 — NEW ARCHITECTURES + CAPABILITIES
## ═══════════════════════════════════════════

---

## Phase 6: Layered Architecture

**Objective:** Add the layered architecture pattern (handler → service → repository) — the most common Go project structure.
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `LAYERED ARCH DONE`
**Dependencies:** None

**Prompt:**
```
Add the "layered" architecture to Archway. This is the most common Go project structure: handler → service → repository.

Read these files first:
- `providers/golang/templates/architectures/hexagonal/manifest.yaml` — example manifest
- `providers/golang/templates/architectures/hexagonal/files/` — example template structure
- `providers/golang/templates/architectures/flat/manifest.yaml` — simpler example
- `internal/guide/guide.go` — guide content generation (you'll add layered support)
- `internal/checker/antipatterns.go` — anti-pattern detectors (some are architecture-aware)
- `providers/golang/provider.go` — provider info (SupportedArchitectures list)

### Layered Architecture Structure

```
cmd/{servicename}/
  main.go              # Wiring + startup
internal/
  handler/             # HTTP/gRPC handlers (transport layer)
    order_handler.go
    router.go
  service/             # Business logic (application layer)
    order_service.go
  repository/          # Data access (persistence layer)
    order_repository.go
    order_repository_interface.go  # Interface in same package
  model/               # Domain models (shared across layers)
    order.go
```

### Layer Rules

```
handler → service → repository → (external)
model ← (shared, all layers can import)
```

- handler imports service (never repository directly)
- service imports repository interfaces
- repository implements interfaces, accesses DB
- model is shared (entities, value objects, DTOs)
- handler NEVER imports repository
- repository NEVER imports handler or service

### Tasks

#### 1. Create Architecture Templates

Create `providers/golang/templates/architectures/layered/`:

**manifest.yaml:**
```yaml
name: layered
description: "Layered architecture with handler, service, and repository layers"
variables:
  - name: ServiceName
    type: string
    required: true
  - name: ModulePath
    type: string
    required: true
  - name: GoVersion
    type: string
    default: "1.24"
components:
  - name: handler
    in: ["internal/handler/**"]
    may_depend_on: [service, model]
  - name: service
    in: ["internal/service/**"]
    may_depend_on: [repository, model]
  - name: repository
    in: ["internal/repository/**"]
    may_depend_on: [model]
  - name: model
    in: ["internal/model/**"]
    may_depend_on: []
hooks:
  - "go mod tidy"
  - "gofmt -w ."
```

**Template files** in `files/`:
- `cmd/__ServiceName__/main.go.tmpl` — main with wiring
- `internal/handler/router.go.tmpl` — HTTP router setup
- `internal/handler/health.go.tmpl` — health endpoint
- `internal/service/.gitkeep.tmpl` — placeholder
- `internal/repository/.gitkeep.tmpl` — placeholder
- `internal/model/.gitkeep.tmpl` — placeholder
- `go.mod.tmpl` — module file
- `archway.yaml.tmpl` — generated config

Follow the patterns from hexagonal templates — use the same template variables ({{ .ServiceName }}, {{ .ModulePath }}, etc.) and partial system.

#### 2. Update Guide Output

In `internal/guide/guide.go`, add layered architecture support to `buildContent()`:
- Layer descriptions for handler/service/repository/model
- Dependency direction (handler → service → repository)
- NEVER rules (handler never imports repository)
- Adding code guidelines (new endpoint, new service, new repo)

#### 3. Update Provider

In `providers/golang/provider.go`:
- Add "layered" to SupportedArchitectures
- Handle "layered" in Scaffold() (should work automatically via composition engine)

#### 4. Auto-Generated Rules

The rule generator from Phase 4 should handle layered automatically (it reads components from archway.yaml). Verify that generating rules for a layered project produces correct layer isolation rules.

#### 5. Tests

- Test scaffold: `archway new test-svc --arch layered --no-wizard` produces correct structure
- Test check: layered project with no violations passes
- Test check: handler importing repository is caught
- Test guide: layered output contains correct layer rules

Run `go build ./...` and `go test ./...` when done.

When complete, output: LAYERED ARCH DONE
```

---

## Phase 7: Clean Architecture

**Objective:** Add the clean architecture pattern (entity → usecase → interface → infrastructure).
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `CLEAN ARCH DONE`
**Dependencies:** None

**Prompt:**
```
Add the "clean" architecture to Archway. This follows Uncle Bob's Clean Architecture.

Read these files first:
- `providers/golang/templates/architectures/layered/manifest.yaml` — from Phase 6
- `providers/golang/templates/architectures/hexagonal/` — reference implementation
- `internal/guide/guide.go` — guide generation
- `providers/golang/provider.go` — provider registration

### Clean Architecture Structure

```
cmd/{servicename}/
  main.go
internal/
  entity/              # Enterprise business rules (innermost)
    order.go
    order_id.go
  usecase/             # Application business rules
    place_order.go
    get_order.go
  interface/           # Interface adapters
    handler/           # HTTP/gRPC handlers
      order_handler.go
    presenter/         # Response formatting
      order_presenter.go
    gateway/           # Repository interfaces + DTOs
      order_gateway.go
  infrastructure/      # Frameworks & drivers (outermost)
    persistence/       # Database implementations
      mysql_order_repo.go
    web/               # Web framework setup
      router.go
    config/            # Configuration loading
      config.go
```

### Layer Rules

```
entity (innermost) → no imports from other layers
usecase → entity only
interface → usecase, entity
infrastructure → interface, usecase, entity
```

### Tasks

#### 1. Create Architecture Templates

Create `providers/golang/templates/architectures/clean/`:

**manifest.yaml** with components:
- entity: `internal/entity/**`, may_depend_on: []
- usecase: `internal/usecase/**`, may_depend_on: [entity]
- interface: `internal/interface/**`, may_depend_on: [usecase, entity]
- infrastructure: `internal/infrastructure/**`, may_depend_on: [interface, usecase, entity]

**Template files** following same patterns as hexagonal.

#### 2. Update Guide Output

Add clean architecture support to guide content:
- Layer descriptions (entity, usecase, interface, infrastructure)
- Dependency direction (infrastructure → interface → usecase → entity)
- Adding code guidelines specific to clean architecture
- Anti-patterns (entity importing infrastructure, etc.)

#### 3. Update Provider

Add "clean" to SupportedArchitectures in `providers/golang/provider.go`.

#### 4. Tests

- Scaffold test
- Check test (violation + clean)
- Guide output test

Run `go build ./...` and `go test ./...` when done.

When complete, output: CLEAN ARCH DONE
```

---

## Phase 8: New Capabilities — Transport + Data

**Objective:** Add 6 new capabilities: graphql, sse, mongodb, sqlite, s3, dynamodb.
**Model:** `sonnet`
**Max Iterations:** 12
**Completion Promise:** `TRANSPORT DATA CAPS DONE`
**Dependencies:** None

**Prompt:**
```
Add 6 new Go capabilities to Archway: graphql, sse, mongodb, sqlite, s3, dynamodb.

Read these files first for the pattern to follow:
- `providers/golang/templates/capabilities/http-api/capability.yaml` — manifest example
- `providers/golang/templates/capabilities/http-api/files/` — template files example
- `providers/golang/templates/capabilities/mysql/capability.yaml` — data capability example
- `providers/golang/templates/capabilities/mysql/files/` — data template example
- `internal/scaffold/suggestions.go` — suggestion rules (you'll update in Phase 11)
- `internal/guide/guide.go` — capabilityDirMap (you'll update)

### For Each Capability, Create:

1. `providers/golang/templates/capabilities/{name}/capability.yaml` — manifest with requires, suggests, conflicts
2. `providers/golang/templates/capabilities/{name}/files/` — template files
3. Update `capabilityDirMap` in `internal/guide/guide.go`

### Capability Specifications

#### 1. graphql (gqlgen)
```yaml
name: graphql
description: "GraphQL API with gqlgen code generation"
requires: []
suggests: [auth-jwt, observability, health, docker]
conflicts: []
```
Files:
- `adapter/graphql/resolver.go.tmpl` — base resolver
- `adapter/graphql/schema.graphqls.tmpl` — base schema
- `adapter/graphql/gqlgen.yml.tmpl` — gqlgen config
- `tools.go.tmpl` — go generate directive

#### 2. sse (Server-Sent Events)
```yaml
name: sse
description: "Server-Sent Events for real-time streaming"
requires: [http-api]
suggests: [health, observability]
conflicts: []
```
Files:
- `adapter/httphandler/sse.go.tmpl` — SSE handler with flusher

#### 3. mongodb
```yaml
name: mongodb
description: "MongoDB connection and repository pattern using official Go driver"
requires: []
suggests: [docker, health, observability, config]
conflicts: []
```
Files:
- `adapter/mongorepo/connection.go.tmpl` — connection + client setup
- `adapter/mongorepo/repository.go.tmpl` — base repository with CRUD

#### 4. sqlite
```yaml
name: sqlite
description: "SQLite database with modernc.org/sqlite (pure Go, no CGO)"
requires: []
suggests: [migrations, config]
conflicts: []
```
Files:
- `adapter/sqliterepo/connection.go.tmpl` — connection setup
- `adapter/sqliterepo/repository.go.tmpl` — base repository

#### 5. s3 (AWS S3 / compatible)
```yaml
name: s3
description: "AWS S3 object storage client"
requires: [config]
suggests: [docker, observability]
conflicts: []
```
Files:
- `adapter/s3client/client.go.tmpl` — S3 client with upload/download/delete
- `adapter/s3client/config.go.tmpl` — S3 configuration

#### 6. dynamodb
```yaml
name: dynamodb
description: "AWS DynamoDB document database client"
requires: [config]
suggests: [docker, observability]
conflicts: []
```
Files:
- `adapter/dynamorepo/client.go.tmpl` — DynamoDB client
- `adapter/dynamorepo/repository.go.tmpl` — base repository with PutItem/GetItem/Query

### Template Guidelines

- Use `{{ .ServiceName }}`, `{{ .ModulePath }}` variables
- Follow hexagonal patterns for adapter placement
- Include proper error handling (wrapped errors with context)
- Include context propagation
- Add shutdown/cleanup hooks where appropriate
- Use constructor pattern (NewXxxRepository, NewXxxClient)
- Define interfaces in port/ or at the top of the adapter file (for flat arch)

### Update capabilityDirMap

In `internal/guide/guide.go`, add entries:
```go
"graphql":  "adapter/graphql/"
"sse":      "adapter/httphandler/"
"mongodb":  "adapter/mongorepo/"
"sqlite":   "adapter/sqliterepo/"
"s3":       "adapter/s3client/"
"dynamodb": "adapter/dynamorepo/"
```

Run `go build ./...` and `go test ./...` when done.

When complete, output: TRANSPORT DATA CAPS DONE
```

---

## Phase 9: New Capabilities — Patterns + DevEx

**Objective:** Add 6 new capabilities: saga, feature-flags, multi-tenancy, ci-gitlab, makefile, devcontainer.
**Model:** `sonnet`
**Max Iterations:** 12
**Completion Promise:** `PATTERNS DEVEX CAPS DONE`
**Dependencies:** None

**Prompt:**
```
Add 6 new capabilities: saga, feature-flags, multi-tenancy, ci-gitlab, makefile, devcontainer.

Read the same reference files as Phase 8 for the pattern to follow.

### Capability Specifications

#### 1. saga
```yaml
name: saga
description: "Saga orchestrator for distributed transactions"
requires: [event-bus]
suggests: [observability, outbox]
conflicts: []
```
Files:
- `service/saga/orchestrator.go.tmpl` — saga orchestrator with step execution + compensation
- `service/saga/step.go.tmpl` — step interface (Execute + Compensate)

#### 2. feature-flags
```yaml
name: feature-flags
description: "Feature flag evaluation with local config fallback"
requires: [config]
suggests: [observability]
conflicts: []
```
Files:
- `platform/featureflags/evaluator.go.tmpl` — flag evaluator interface + local implementation
- `platform/featureflags/flags.go.tmpl` — flag definitions

#### 3. multi-tenancy
```yaml
name: multi-tenancy
description: "Row-level tenant isolation with tenant context middleware"
requires: [http-api, config]
suggests: [auth-jwt, observability]
conflicts: []
```
Files:
- `adapter/httphandler/middleware/tenant.go.tmpl` — tenant extraction middleware
- `platform/tenant/context.go.tmpl` — tenant context helpers (set/get from context)
- `platform/tenant/resolver.go.tmpl` — tenant resolution interface

#### 4. ci-gitlab
```yaml
name: ci-gitlab
description: "GitLab CI/CD pipeline configuration"
requires: []
suggests: [docker, linting, testing]
conflicts: [ci-github]
```
Files:
- `.gitlab-ci.yml.tmpl` — multi-stage pipeline (lint, test, build, deploy)

#### 5. makefile
```yaml
name: makefile
description: "Makefile with common development targets"
requires: []
suggests: []
conflicts: []
```
Files:
- `Makefile.tmpl` — targets: build, test, lint, run, docker-build, clean, help

#### 6. devcontainer
```yaml
name: devcontainer
description: "VS Code Dev Container configuration for consistent development environments"
requires: []
suggests: [docker]
conflicts: []
```
Files:
- `.devcontainer/devcontainer.json.tmpl` — container config with Go extensions
- `.devcontainer/Dockerfile.tmpl` — Go dev image

### Update capabilityDirMap

```go
"saga":          "service/saga/"
"feature-flags": "platform/featureflags/"
"multi-tenancy": "adapter/httphandler/middleware/, platform/tenant/"
"ci-gitlab":     ".gitlab-ci.yml"
"makefile":      "Makefile"
"devcontainer":  ".devcontainer/"
```

Run `go build ./...` and `go test ./...` when done.

When complete, output: PATTERNS DEVEX CAPS DONE
```

---

## Phase 10: New Capabilities — Frontend Go

**Objective:** Add 3 frontend Go capabilities: templ, htmx, static-assets.
**Model:** `opus`
**Max Iterations:** 10
**Completion Promise:** `FRONTEND GO CAPS DONE`
**Dependencies:** None

**Prompt:**
```
Add 3 full-stack Go capabilities: templ, htmx, static-assets. These enable Go developers to build server-rendered web applications with modern patterns.

Read the same reference files as Phase 8 for the capability pattern.

### Capability Specifications

#### 1. templ (Go HTML templating)
```yaml
name: templ
description: "Type-safe Go HTML templating with templ — server-rendered components"
requires: [http-api]
suggests: [htmx, static-assets]
conflicts: []
```
Files:
- `adapter/httphandler/views/layout.templ.tmpl` — base HTML layout with head/body
- `adapter/httphandler/views/home.templ.tmpl` — example page component
- `adapter/httphandler/views/components/header.templ.tmpl` — reusable header
- `tools.go.tmpl` — templ generate directive (or update existing)
- Update main.go partial to serve templ handlers

The templates should show the templ pattern:
```go
// home.templ
package views

templ Home(title string) {
    @Layout(title) {
        <main class="container">
            <h1>{ title }</h1>
        </main>
    }
}
```

#### 2. htmx (hypermedia-driven interactions)
```yaml
name: htmx
description: "HTMX integration for server-driven interactivity without JavaScript"
requires: [http-api, templ]
suggests: [static-assets]
conflicts: []
```
Files:
- `adapter/httphandler/htmx.go.tmpl` — HTMX middleware (HX-Request detection, partial vs full page)
- `adapter/httphandler/views/partials/counter.templ.tmpl` — example HTMX partial

The HTMX middleware should:
- Detect `HX-Request` header
- Return partial templ component for HTMX requests
- Return full page for regular requests
- Set appropriate `HX-Trigger`, `HX-Push-Url` headers

#### 3. static-assets
```yaml
name: static-assets
description: "Static file serving with embedded assets and cache headers"
requires: [http-api]
suggests: [templ]
conflicts: []
```
Files:
- `adapter/httphandler/static.go.tmpl` — static file handler with embed.FS + cache headers
- `static/css/style.css.tmpl` — minimal CSS
- `static/js/.gitkeep.tmpl` — JS directory placeholder

### Update capabilityDirMap

```go
"templ":          "adapter/httphandler/views/"
"htmx":           "adapter/httphandler/"
"static-assets":  "static/, adapter/httphandler/"
```

### Important

These capabilities are novel for Archway — no existing frontend Go patterns to reference. Research templ and HTMX best practices. The templates should be production-quality examples that teach the pattern, not just stubs.

Run `go build ./...` and `go test ./...` when done.

When complete, output: FRONTEND GO CAPS DONE
```

---

## Phase 11: Suggestion Rules + Compatibility Matrix

**Objective:** Update suggestion rules and capability warnings for all new capabilities and architectures.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `SUGGESTIONS UPDATED`
**Dependencies:** Phase 6, 7, 8, 9, 10

**Prompt:**
```
Update the suggestion engine and capability warnings for the expanded capability set.

Read these files first:
- `internal/scaffold/suggestions.go` — existing 18 rules + 9 warnings
- `providers/golang/matrix.go` — compatibility matrix
- All new capability.yaml files from Phases 8-10

### Tasks

#### 1. New Suggestion Rules

Add rules for the new capabilities:

```go
// Transport
{IfAny: []string{"graphql"}, Missing: "auth-jwt", Reason: "GraphQL APIs need authentication"},
{IfAny: []string{"graphql"}, Missing: "observability", Reason: "GraphQL resolvers benefit from tracing"},

// Data
{IfAny: []string{"mongodb", "dynamodb"}, Missing: "health", Reason: "Database connections need health checks"},
{IfAny: []string{"mongodb", "sqlite", "dynamodb"}, Missing: "config", Reason: "Database connections need configuration"},

// Patterns
{IfAny: []string{"saga"}, Missing: "observability", Reason: "Distributed transactions need tracing"},
{IfAny: []string{"multi-tenancy"}, Missing: "auth-jwt", Reason: "Tenant isolation requires authentication"},
{IfAny: []string{"feature-flags"}, Missing: "observability", Reason: "Track feature flag usage"},

// Frontend
{IfAny: []string{"templ"}, Missing: "htmx", Reason: "templ + HTMX is the standard Go full-stack pattern"},
{IfAny: []string{"templ"}, Missing: "static-assets", Reason: "Server-rendered apps need CSS/JS serving"},
{IfAny: []string{"htmx"}, Missing: "static-assets", Reason: "HTMX needs the htmx.js library served"},
```

#### 2. New Capability Warnings

```go
// multi-tenancy without proper data isolation
{IfAll: []string{"multi-tenancy"}, Without: "auth-jwt", Severity: "error",
 Message: "Multi-tenancy without authentication risks tenant data leaks"},

// saga without observability
{IfAll: []string{"saga"}, Without: "observability", Severity: "warning",
 Message: "Sagas without tracing make distributed debugging very difficult"},
```

#### 3. Update Compatibility Matrix

In `providers/golang/matrix.go`, add all new capabilities to the compatibility matrix. Ensure:
- New capabilities work with both hexagonal AND layered AND clean architectures
- Template file paths adjust per architecture (e.g., layered uses `internal/handler/` not `adapter/httphandler/`)

#### 4. Verify

- Run all existing suggestion tests
- Add tests for new suggestion rules
- Test that `archway new` wizard shows new capabilities in selection
- Test that suggestions fire correctly for new capabilities

Run `go build ./...` and `go test ./...` when done.

When complete, output: SUGGESTIONS UPDATED
```

---

## Phase 12: v1.2 Tests + Documentation

**Objective:** Test all new architectures and capabilities, update docs and website.
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `V1.2 DOCS DONE`
**Dependencies:** Phase 6, 7, 8, 9, 10, 11

**Prompt:**
```
Write tests for new architectures + capabilities and update all documentation.

### Tests

#### Architecture Tests

For each new architecture (layered, clean):
1. Scaffold test: `archway new test-svc --arch {arch} --no-wizard` → verify directory structure
2. Check test: clean project passes, violation is caught
3. Guide test: output contains correct layer rules
4. Composition test: architecture + capabilities compose correctly

#### Capability Tests

For each new capability batch:
1. Manifest loading: each capability.yaml parses correctly
2. Scaffold: capability renders template files to expected locations
3. Composition: capability requirements are resolved
4. Conflicts: conflicting capabilities are detected (ci-github vs ci-gitlab)

### Documentation Updates

#### 1. README.md
- Update architecture count (2 → 4)
- Update capability count to new total
- Add layered and clean to architecture list
- Add new capability categories

#### 2. Website (`website/src/content/docs/`)
- Update `concepts/capabilities.mdx` — add new capabilities with descriptions
- Update `reference/capabilities-matrix.mdx` — add new rows
- Update `index.mdx` — update counts
- Add architecture pages for layered and clean if needed

#### 3. Product Spec (`docs/product/spec.md`)
- Update architecture table: layered and clean → "shipped"
- Update capability count

#### 4. Verify
- `cd website && npm run build` — website builds without errors
- `go test ./...` — all tests pass

Do NOT change the tagline. The tagline is LOCKED.

When complete, output: V1.2 DOCS DONE
```

---

## ═══════════════════════════════════════════
## v1.3 — DESIGN COMPANION
## ═══════════════════════════════════════════

---

## Phase 13: Capability Catalog Content Generation

**Objective:** Generate a capability catalog section in guide output — descriptions, when-to-use guidance, and capability relationships.
**Model:** `opus`
**Max Iterations:** 12
**Completion Promise:** `CATALOG CONTENT DONE`
**Dependencies:** Phase 12

**Prompt:**
```
Add a capability catalog to `archway guide` output. This transforms the guide from static architecture rules into a design companion that helps AI agents make capability decisions.

Read these files first:
- `internal/guide/guide.go` — current guide generation
- `internal/scaffold/suggestions.go` — suggestion rules (reuse for catalog)
- `internal/scaffold/composer.go` — capability manifests (source of capability metadata)
- All `capability.yaml` files — what metadata is available

Design 004 was validated by experiments: E1 (+127% capability implementation), E2 (+53% improvement).

### What to Build

#### 1. Capability Catalog Data (`internal/guide/catalog.go`)

```go
type CatalogEntry struct {
    Name         string
    Category     string   // "transport", "data", "resilience", "patterns", "security", "observability", "infrastructure", "quality", "frontend"
    Description  string   // from capability.yaml
    WhenToUse    string   // guidance on when this capability is appropriate
    Installed    bool     // is this capability in the project's archway.yaml?
    Suggests     []string // from capability.yaml
    Interactions []string // warnings when combined with other capabilities
}

// BuildCatalog reads all capability manifests and returns organized catalog
func BuildCatalog(templateFS fs.FS, installedCaps []string) ([]CatalogEntry, error)
```

Category assignment: map each capability to a category. Hardcode the mapping:
```go
var capabilityCategories = map[string]string{
    "http-api": "transport", "grpc": "transport", "graphql": "transport",
    "websocket": "transport", "kafka-consumer": "transport", "sse": "transport",
    "mysql": "data", "postgres": "data", "redis": "data",
    "mongodb": "data", "sqlite": "data", "dynamodb": "data", "s3": "data",
    "repository": "data",
    // ... etc for all capabilities
}
```

WhenToUse: hardcode per capability (this is design knowledge, not derivable from manifests):
```go
var whenToUse = map[string]string{
    "http-api":     "REST APIs, webhooks, web applications",
    "grpc":         "Internal microservice communication, streaming, high-performance RPC",
    "graphql":      "Client-facing APIs with complex querying needs, mobile apps",
    "circuit-breaker": "Calls to external services that may fail or slow down",
    "saga":         "Multi-service transactions that need rollback capability",
    "multi-tenancy": "SaaS applications serving multiple organizations",
    // ... etc
}
```

#### 2. Catalog in Guide Output

Add new section to `buildContent()` AFTER the existing capabilities section:

```markdown
## Capability Catalog

### Installed
| Capability | Category | Purpose |
|-----------|----------|---------|
| http-api | transport | REST APIs, webhooks, web applications |
| mysql | data | Relational data with SQL queries |

### Available (not installed)
| Capability | Category | When to Consider |
|-----------|----------|-----------------|
| circuit-breaker | resilience | You call external services that may fail |
| saga | patterns | You need multi-service transactions |

### Suggestions for This Project
Based on your installed capabilities:
- Consider **rate-limiting** — you have http-api (public APIs benefit from rate limiting)
- Consider **circuit-breaker** — you have http-client (external calls need resilience)
```

The suggestions section reuses logic from `internal/scaffold/suggestions.go` — call `ComputeSuggestions()` with installed capabilities and format the results.

#### 3. Content Sizing

Respect INV-001 (rules file sizing). The catalog must not blow up the guide file. Strategy:
- Installed capabilities: full detail (name, category, purpose, directory)
- Available capabilities: one-liner per capability
- Suggestions: top 5 only
- Total catalog section: target 300-500 tokens

### Tests

Create `internal/guide/catalog_test.go`:
- Test BuildCatalog with known capability set
- Test installed vs available filtering
- Test suggestion integration
- Test output token count stays within budget

Run `go build ./...` and `go test ./...` when done.

When complete, output: CATALOG CONTENT DONE
```

---

## Phase 14: Interaction Warnings + Suggestions in Guide

**Objective:** Add interaction warnings (dangerous combinations) and smart suggestions to guide output.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `WARNINGS IN GUIDE DONE`
**Dependencies:** Phase 13

**Prompt:**
```
Add interaction warnings and contextual suggestions to the guide output.

Read these files first:
- `internal/guide/catalog.go` — catalog from Phase 13
- `internal/guide/guide.go` — current content generation
- `internal/scaffold/suggestions.go` — existing warning rules

### What to Build

#### 1. Interaction Warnings Section

Add to guide output after the catalog:

```markdown
## Interaction Warnings

⚠ **http-api + no rate-limiting** — Public APIs without rate limiting are vulnerable to abuse
⚠ **kafka-consumer + no circuit-breaker** — Consumer failures can cascade without circuit breaking
⚠ **multi-tenancy + no auth-jwt** — CRITICAL: Tenant isolation without authentication risks data leaks
```

Source warnings from:
- Existing capability warnings in `suggestions.go`
- New warnings from Phase 11
- Filter to only show warnings relevant to installed capabilities

Severity levels:
- CRITICAL: data safety, security (shown first, prefixed with "CRITICAL:")
- WARNING: resilience, operational (shown after)

#### 2. Contextual Suggestions Section

After warnings, add project-specific suggestions:

```markdown
## Architecture Suggestions

Based on your current capabilities, consider:

1. **Add circuit-breaker** — You have http-client but no resilience patterns. External service failures will cascade.
2. **Add observability** — You have 3 data sources but no tracing. Debugging distributed queries will be difficult.
3. **Add health** — You have mysql + redis but no health checks. Kubernetes won't know if your service is ready.
```

Logic: run `ComputeSuggestions()`, filter to top 5 most impactful, format with contextual explanation.

#### 3. INV-001 Compliance

Interaction warnings + suggestions add ~200-300 tokens. Verify the total guide output still complies with INV-001. If it exceeds limits, truncate suggestions (show top 3 instead of 5).

### Tests

- Test warnings shown for known dangerous combinations
- Test no warnings for safe combinations
- Test suggestions reflect actual missing capabilities
- Test token count compliance

Run `go build ./...` and `go test ./...` when done.

When complete, output: WARNINGS IN GUIDE DONE
```

---

## Phase 15: Catalog-Only Mode + Rule Summaries

**Objective:** Add --catalog-only flag and include proxy rule summaries in guide output.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `CATALOG MODE DONE`
**Dependencies:** Phase 13, 14

**Prompt:**
```
Add --catalog-only mode and proxy rule summaries to archway guide.

Read these files first:
- `internal/cli/guide.go` — CLI command
- `internal/guide/guide.go` — guide generation
- `internal/rules/loader.go` — rule loading (for summaries)

### Tasks

#### 1. --catalog-only Flag

Add `--catalog-only` flag to `archway guide`:
- When set: output ONLY the capability catalog + warnings + suggestions sections
- Skip: architecture rules, layer rules, dependency direction, adding code guidelines
- Use case: pre-archway projects that want design guidance without full architecture rules
- Works without archway.yaml — reads capability manifests directly from embedded FS
- Still outputs to all targets (or specified --target)

#### 2. Proxy Rule Summaries in Guide

When `.archway/rules/` exists, add a section to guide output:

```markdown
## Active Rules

{n} proxy rules enforced by `archway check`:

| Rule | Engine | Severity | Scope |
|------|--------|----------|-------|
| arch-domain-isolation | grep | error | domain/**/*.go |
| cap-sql-parameterized | grep | error | **/*.go |
| INV-003-R1 | grep | error | internal/repository/**/*.go |

Run `archway check` to validate. Run `archway check --staged` as pre-commit hook.
```

Load rules via `rules.LoadRules()`, extract metadata, format as table. Don't show stale/invalid rules in guide (they show in check output instead).

#### 3. Tests

- Test --catalog-only produces catalog without architecture sections
- Test --catalog-only works without archway.yaml
- Test rule summaries appear when .archway/rules/ exists
- Test rule summaries are absent when no rules directory

Run `go build ./...` and `go test ./...` when done.

When complete, output: CATALOG MODE DONE
```

---

## Phase 16: v1.3 Tests

**Objective:** Integration tests for the design companion features.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `V1.3 TESTS PASSING`
**Dependencies:** Phase 13, 14, 15

**Prompt:**
```
Write integration tests for all v1.3 design companion features.

### Tests

#### 1. Full Guide Output Test

```go
func TestGuideWithCatalog(t *testing.T) {
    // 1. Create archway.yaml with hexagonal + http-api + mysql + auth-jwt
    // 2. Create .archway/rules/ with 2 proxy rules
    // 3. Run GenerateFromConfig()
    // 4. Read generated .claude/rules/archway.md
    // 5. Assert contains: architecture rules, catalog, warnings, suggestions, rule summaries
    // 6. Assert catalog shows http-api/mysql/auth-jwt as installed
    // 7. Assert catalog shows available capabilities
    // 8. Assert warnings section present
    // 9. Assert rule summary table present
}
```

#### 2. Catalog-Only Test

```go
func TestCatalogOnlyMode(t *testing.T) {
    // 1. No archway.yaml needed
    // 2. Run with --catalog-only equivalent
    // 3. Assert contains: catalog, suggestions
    // 4. Assert does NOT contain: architecture rules, layer rules
}
```

#### 3. INV-001 Compliance Test

```go
func TestGuideTokenCompliance(t *testing.T) {
    // 1. Generate guide with maximum capabilities (all installed)
    // 2. Count approximate tokens (words * 1.3)
    // 3. Assert within INV-001 bounds
    // If over budget, this test forces content trimming
}
```

#### 4. CLI Integration

```go
func TestGuideCLICatalogOnly(t *testing.T) {
    // Run archway guide --catalog-only --target claude
    // Verify .claude/rules/archway.md created with catalog content
}
```

Run `go test ./...` when done.

When complete, output: V1.3 TESTS PASSING
```

---

## ═══════════════════════════════════════════
## v1.4 — TOKEN COMPACTION + DECISION GATES
## ═══════════════════════════════════════════

---

## Phase 17: Split Guide Output

**Objective:** Refactor guide to emit split files — always-loaded index + path-scoped category files.
**Model:** `opus`
**Max Iterations:** 12
**Completion Promise:** `GUIDE SPLIT DONE`
**Dependencies:** Phase 16

**Prompt:**
```
Refactor `archway guide` to emit split files instead of one monolithic guide. This implements Design 006 (token compaction).

Read these files first:
- `internal/guide/guide.go` — current monolithic generation
- `internal/guide/targets.go` — output targets
- `internal/guide/catalog.go` — catalog from v1.3
- `docs/invariants/001-rules-file-sizing.md` — INV-001 constraints

### Architecture

**Before (v1.0-v1.3):** One file `.claude/rules/archway.md` with everything.
**After (v1.4):** Multiple files in `.claude/rules/`:

```
.claude/rules/
  archway-index.md      # Always loaded (globs: ["**/*"])
  archway-http.md       # Loaded when editing HTTP handlers
  archway-data.md       # Loaded when editing repositories
  archway-security.md   # Loaded when editing auth/middleware
  archway-resilience.md # Loaded when editing resilience patterns
  archway-patterns.md   # Loaded when editing domain/events
  archway-quality.md    # Loaded when editing tests
  archway-frontend.md   # Loaded when editing views/templates
```

### Index File (~800 tokens)

```markdown
---
globs: ["**/*"]
---
<!-- archway:generated -->
# Archway — {architecture} Architecture

> Auto-generated by `archway guide`. Re-run to update.

## Architecture: {type}
{2-3 sentence summary}

## Layer Rules (summary)
{one-line per layer with dependency direction}

## Active Capabilities
{bullet list: name — one-liner}

## Active Rules
{count} proxy rules enforced. Run `archway check` to validate.

## Decision Status
{resolved/pending counts if Design 005 is implemented}

## Critical Warnings
{only CRITICAL severity warnings}
```

### Category Files (~500-800 tokens each)

Each category file:
```markdown
---
globs: ["internal/handler/**", "adapter/httphandler/**", "internal/middleware/**"]
---
<!-- archway:generated -->
# Archway — HTTP & Transport

## Installed Capabilities
- **http-api** — Chi router, middleware pattern
- **graphql** — gqlgen, resolver pattern

## Adding Code
{how to add endpoints, middleware, etc.}

## Patterns
{extracted from templates}

## Suggestions
{missing capabilities relevant to this category}

## Warnings
{interaction warnings relevant to this category}
```

### Category Assignment

Map capabilities to categories:
```go
var categoryMap = map[string][]string{
    "http":       {"http-api", "grpc", "graphql", "websocket", "sse"},
    "data":       {"mysql", "postgres", "redis", "mongodb", "sqlite", "dynamodb", "s3", "repository"},
    "security":   {"auth-jwt", "cors", "multi-tenancy"},
    "resilience": {"circuit-breaker", "retry", "rate-limiting", "health"},
    "patterns":   {"cqrs", "event-bus", "outbox", "saga", "scheduler", "worker"},
    "quality":    {"testing", "linting", "pre-commit"},
    "frontend":   {"templ", "htmx", "static-assets"},
}
```

Glob mapping per category:
```go
var categoryGlobs = map[string][]string{
    "http":       {"internal/handler/**", "adapter/httphandler/**", "adapter/grpchandler/**", "adapter/graphql/**", "proto/**"},
    "data":       {"internal/repository/**", "adapter/*repo/**", "adapter/s3client/**", "migrations/**"},
    "security":   {"internal/middleware/**", "internal/auth/**", "adapter/httphandler/middleware/**", "platform/tenant/**"},
    "resilience": {"platform/resilience/**", "internal/circuit/**"},
    "patterns":   {"domain/**", "internal/event/**", "service/saga/**"},
    "quality":    {"**/*_test.go", "testdata/**"},
    "frontend":   {"adapter/httphandler/views/**", "static/**", "internal/views/**"},
}
```

### Implementation

#### 1. Modify `GenerateFromConfig()`

When target is "claude":
- Generate split files (index + category files for installed capabilities)
- Delete stale category files (capability removed from archway.yaml)
- Only generate category files that have at least one installed capability

When target is "cursor", "copilot", "windsurf":
- Keep monolithic output (these tools don't support path-scoped loading)
- But apply token budget from INV-001

#### 2. Content Distribution

Split existing content across files:
- Architecture summary → index
- Full layer rules → index (short) + relevant category files (detailed)
- Capability catalog → distributed to category files
- Warnings → CRITICAL to index, category-specific to category files
- Suggestions → category files
- Patterns → category files
- Rule summaries → index (table), category files (relevant rules)

#### 3. INV-001 Validation

After generating all files, validate each against INV-001:
- Count instructions (15-25 max)
- Estimate tokens (~words * 1.3, target 500-1500)
- If over budget: trim suggestions, collapse patterns, shorten descriptions
- Log warning if any file exceeds bounds

### Tests

- Test split generates correct file count for a project with http-api + mysql + auth-jwt
- Test index contains summary but not full details
- Test category files have correct globs frontmatter
- Test monolithic output still works for non-Claude targets
- Test stale file cleanup
- Test INV-001 compliance per file
- Test empty category (no installed caps) → file not generated

Run `go build ./...` and `go test ./...` when done.

When complete, output: GUIDE SPLIT DONE
```

---

## Phase 18: Decision Gates Schema + archway decide CLI

**Objective:** Add decisions section to archway.yaml and implement `archway decide` interactive CLI.
**Model:** `opus`
**Max Iterations:** 12
**Completion Promise:** `DECISION GATES DONE`
**Dependencies:** Phase 16

**Prompt:**
```
Implement decision gates per Design 005. Add a `decisions` section to archway.yaml and an `archway decide` command.

Read these files first:
- `internal/config/config.go` — ArchwayConfig struct
- `internal/cli/root.go` — command registration
- `internal/scaffold/suggestions.go` — wizard TUI patterns (reuse for decide)

Design 005 validated by experiments: E4b (6/6 compliance), E8 (debrief filter).

### What to Build

#### 1. Config Schema Extension (`internal/config/config.go`)

Add to ArchwayConfig:
```go
type ArchwayConfig struct {
    // ... existing fields ...
    Decisions []Decision `yaml:"decisions,omitempty"`
}

type Decision struct {
    Topic       string `yaml:"topic"`       // e.g., "authentication-strategy"
    Tier        int    `yaml:"tier"`         // 1 (foundational), 2 (infrastructure), 3 (design)
    Status      string `yaml:"status"`       // "undecided", "decided"
    Choice      string `yaml:"choice"`       // e.g., "jwt-with-refresh-tokens"
    Rationale   string `yaml:"rationale"`    // why this choice
    DecidedBy   string `yaml:"decided-by"`   // who decided
    DecidedAt   string `yaml:"decided-at"`   // when
}
```

#### 2. Decision Gap Catalog (`internal/guide/decisions.go`)

Define the standard decision gaps:

**Tier 1 — Foundational (must decide before writing code):**
- architecture-pattern: "How is code organized?" (hexagonal, layered, clean, flat)
- authentication-strategy: "How are users/services authenticated?" (JWT, OAuth2, API keys, gateway)
- data-model-ownership: "Who owns the data?" (service-owned DB, shared DB, event-sourced)
- tenant-isolation: "How are tenants separated?" (row-level, schema-per-tenant, DB-per-tenant, N/A)

**Tier 2 — Infrastructure (before production):**
- failure-strategy: "What happens when external calls fail?" (fail-fast, retry, queue, degrade)
- migration-strategy: "How are DB migrations managed?" (golang-migrate, goose, manual)
- deployment-model: "Where does this run?" (Kubernetes, ECS, VM, serverless)
- observability-stack: "How is the system observed?" (OpenTelemetry, Datadog, CloudWatch)

**Tier 3 — Design (per-feature):**
- sync-vs-async: "Should this operation be synchronous or asynchronous?"
- consistency-model: "Strong or eventual consistency?"
- error-exposure: "How much error detail reaches the client?"

Auto-populate: some decisions can be inferred from archway.yaml:
- If `architecture: hexagonal` → architecture-pattern is decided
- If `auth-jwt` in capabilities → authentication-strategy is decided
- If `mysql` in capabilities → data-model-ownership is partially decided

#### 3. `archway decide` Command (`internal/cli/decide.go`)

Interactive CLI for resolving decisions:

```bash
archway decide                          # Interactive — show all undecided, pick one
archway decide authentication-strategy  # Resolve a specific decision
archway decide --list                   # List all decisions with status
```

Interactive flow:
1. Show undecided decisions grouped by tier
2. User selects one
3. Show options (from the gap catalog) + "other"
4. User picks choice
5. Ask for rationale (optional, one sentence)
6. Update archway.yaml
7. Optionally generate an ADR stub

Use `charmbracelet/huh` for TUI forms (same as wizard).

#### 4. Decision Gates in Guide Output

Add to the guide index file:

```markdown
## Decision Status

**Tier 1 (Foundational):**
- ✓ architecture-pattern: hexagonal
- ✓ authentication-strategy: JWT with refresh tokens
- ✗ tenant-isolation: UNDECIDED — resolve before writing data access code
- ✗ data-model-ownership: UNDECIDED

**Action:** Run `archway decide` to resolve open decisions.
**Rule:** Do NOT implement features that depend on undecided Tier 1 topics.
```

#### 5. Auto-Populate on Init/New

When `archway new` or `archway init` creates archway.yaml, auto-populate the decisions section:
- Set decided decisions based on capabilities/architecture
- Set undecided for everything else
- Tier 1 decisions are always included
- Tier 2 included if relevant capabilities are present
- Tier 3 not included by default (added per-feature)

### Tests

- Test Decision schema parsing from archway.yaml
- Test auto-population from capabilities
- Test decide command updates archway.yaml correctly
- Test guide output shows decision status
- Test tier filtering

Run `go build ./...` and `go test ./...` when done.

When complete, output: DECISION GATES DONE
```

---

## Phase 19: archway check --decisions

**Objective:** Add --decisions flag to archway check that validates decision gate completeness.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `DECISION CHECK DONE`
**Dependencies:** Phase 18

**Prompt:**
```
Add decision gate validation to `archway check`.

Read these files first:
- `internal/checker/checker.go` — existing check flow
- `internal/config/config.go` — Decision type from Phase 18
- `internal/cli/check.go` — CLI command

### Tasks

#### 1. Decision Checker (`internal/checker/decisions.go`)

```go
type DecisionViolation struct {
    Topic    string
    Tier     int
    Message  string
    Severity string // "error" for Tier 1, "warning" for Tier 2-3
}

func CheckDecisions(decisions []config.Decision) []DecisionViolation
```

Logic:
- Tier 1 undecided → severity "error" (blocks)
- Tier 2 undecided → severity "warning" (reports)
- Tier 3 undecided → ignored (per-feature, not project-level)

#### 2. CLI Integration

Add `--decisions` flag to `archway check`:
- When set: ONLY check decisions (skip detectors + proxy rules)
- Default: decisions are included in the standard check (if decisions section exists in archway.yaml)
- Tier 1 undecided → exit code 1

#### 3. Output

```
DECISION GATES
  ✓ architecture-pattern: hexagonal (Tier 1)
  ✓ authentication-strategy: JWT (Tier 1)
  ✗ tenant-isolation: UNDECIDED (Tier 1) — resolve before writing data access code
  ⚠ failure-strategy: UNDECIDED (Tier 2) — resolve before production

  Tier 1: 2/4 decided (2 blocking)
  Tier 2: 1/4 decided (3 warnings)
```

#### 4. Pre-commit Integration

When `archway check --staged` runs, also check Tier 1 decisions. If any Tier 1 is undecided, block the commit with:
```
BLOCKED: Undecided Tier 1 decisions. Run `archway decide` to resolve:
  - tenant-isolation
  - data-model-ownership
```

### Tests

- Test all Tier 1 decided → pass
- Test Tier 1 undecided → error
- Test Tier 2 undecided → warning (no error)
- Test --decisions flag runs only decision checks
- Test integration with standard check flow

Run `go build ./...` and `go test ./...` when done.

When complete, output: DECISION CHECK DONE
```

---

## Phase 20: v1.4 Tests + Final Integration

**Objective:** End-to-end tests for split guide + decision gates, final integration verification.
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `V1.4 COMPLETE`
**Dependencies:** Phase 17, 18, 19

**Prompt:**
```
Final integration tests for v1.4 and verification that the full v1.1-v1.4 pipeline works end-to-end.

### v1.4 Integration Tests

#### 1. Split Guide + Decisions Combined

```go
func TestFullGuideWithDecisions(t *testing.T) {
    // 1. Create archway.yaml: hexagonal + http-api + mysql + auth-jwt
    //    With decisions: architecture-pattern decided, tenant-isolation undecided
    // 2. Create .archway/rules/ with proxy rules
    // 3. Run archway guide --target claude
    // 4. Verify split files generated:
    //    - archway-index.md with decision status + critical warnings
    //    - archway-http.md with http capabilities
    //    - archway-data.md with mysql capabilities
    //    - archway-security.md with auth-jwt
    // 5. Verify each file has correct globs: frontmatter
    // 6. Verify each file complies with INV-001
    // 7. Verify monolithic output still works for --target cursor
}
```

#### 2. Decision Gate Workflow

```go
func TestDecisionWorkflow(t *testing.T) {
    // 1. archway new → auto-populates decisions
    // 2. archway check --decisions → reports undecided Tier 1
    // 3. Manually resolve decision in archway.yaml
    // 4. archway check --decisions → passes
    // 5. archway guide → shows updated decision status in index
}
```

### Full Pipeline Test (v1.1 → v1.4)

```go
func TestFullPipeline(t *testing.T) {
    // 1. archway new my-svc --arch hexagonal --cap http-api,mysql,auth-jwt --no-wizard
    // 2. archway guide → generates split files + proxy rules
    // 3. archway check → runs detectors + proxy rules + decisions
    // 4. Add violating code (domain importing adapter)
    // 5. archway check → catches violation via proxy rule
    // 6. archway check --staged → works with git staging
    // 7. archway guide --catalog-only → works
    // 8. archway decide --list → shows decisions
    // 9. Verify all outputs (terminal, JSON) are correct
}
```

### Documentation Updates

Update for v1.4:
- README: mention split guide, decision gates
- Website: add decision gates page
- CLI help text: verify all new flags documented

### Final Verification

```bash
go build ./...
go test ./...
go vet ./...
golangci-lint run
```

When complete, output: V1.4 COMPLETE
```

---

## Known Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| 15 new capabilities is a LOT of template work | High | Batch into 3 phases (8-10), parallelize |
| templ/HTMX templates are novel — no existing patterns | Medium | Phase 10 uses Opus, research during implementation |
| Split guide output may break existing .cursorrules users | Medium | Phase 17 keeps monolithic for non-Claude targets |
| Decision gate auto-population may be wrong | Low | Conservative defaults — only infer from explicit capabilities |
| INV-001 compliance at scale (many capabilities) | Medium | Phase 17 includes validation, test enforces budget |
| Capability category mapping may need adjustment | Low | Hardcoded map, easy to update |
| Auto-generated proxy rules may produce false positives | Medium | Users can delete/modify generated rules |
| Clean architecture's `internal/interface/` may conflict with Go keyword | Medium | Phase 7 may need alternative naming (e.g., `interfaces/` or `gateway/`) |

---

## Post-v1.4 Notes

After v1.4 completes:
- Update Obsidian roadmap with completion status
- Tag releases: v1.1.0, v1.2.0, v1.3.0, v1.4.0
- Consider which v1.5 items to prioritize (archway diff, archway add, archway health)
- Begin Keel design session using `10 - Projects/Keel/prompts/design-keel-context-injection.md`
- Begin DT-001 design session for TypeScript provider (v2.0 prep)
