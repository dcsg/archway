# Plan: Archway v1.0 Release

## Overview
**Task:** Ship Archway v1.0 — polish CLI, implement `archway guide` (PRD-002), fix all known issues, first GitHub release via GoReleaser + Homebrew
**Total Phases:** 8
**Created:** 2026-03-09

## Progress

| Phase | Status | Updated |
|-------|--------|---------|
| 1     | done   | 2026-03-09 |
| 2     | done   | 2026-03-09 |
| 3     | done   | 2026-03-09 |
| 4     | done   | 2026-03-09 |
| 5     | done   | 2026-03-09 |
| 6     | done   | 2026-03-09 |
| 7     | done   | 2026-03-09 |
| 8     | done   | 2026-03-09 |

**IMPORTANT:** Update this table as phases complete. This table is the persistent state that survives context compaction.

## Model Assignment

| Phase | Task | Model | Reasoning | Est. Cost |
|-------|------|-------|-----------|-----------|
| 1 | Fix provider metadata + unused flags | sonnet | Simple code changes but need to verify side effects | ~$0.08 |
| 2 | Merge/resolve template worktree | sonnet | File decisions + moves, moderate complexity | ~$0.08 |
| 3 | Fix ghost capabilities in docs | sonnet | Cross-file doc consistency | ~$0.08 |
| 4 | `archway guide` — core engine | opus | Novel feature, architecture decisions, multi-tool output | ~$0.80 |
| 5 | `archway guide` — pattern extraction | opus | Template parsing, code pattern generation | ~$0.80 |
| 6 | CLI integration tests | sonnet | Test scaffolding, well-defined patterns | ~$0.08 |
| 7 | Website + docs update | sonnet | Content updates, capability count fixes | ~$0.08 |
| 8 | GoReleaser + first release | sonnet | Config validation, tagging, release | ~$0.08 |

## Execution Strategy

| Phase | Depends On | Parallel With |
|-------|-----------|---------------|
| 1     | None      | 2, 3          |
| 2     | None      | 1, 3          |
| 3     | None      | 1, 2          |
| 4     | 1, 2      | -             |
| 5     | 4         | -             |
| 6     | 1, 2      | 4             |
| 7     | 4, 5      | -             |
| 8     | ALL       | -             |

---

## Phase 1: CLI Cleanup

**Objective:** Fix provider architecture count, remove unused flags, clean up dead code.
**Model:** `sonnet`
**Max Iterations:** 5
**Completion Promise:** `CLI CLEANUP DONE`
**Dependencies:** None

**Prompt:**
```
Fix the following issues in the Archway CLI. Read each file before modifying.

1. **Fix SupportedArchitectures** in `providers/golang/provider.go` (around line 150):
   - Currently claims: `[]string{"hexagonal", "clean", "ddd", "layered", "flat"}`
   - Only 2 architectures exist on disk: `providers/golang/templates/architectures/hexagonal/` and `flat/`
   - Change to: `[]string{"hexagonal", "flat"}`
   - Search the entire codebase for any other hardcoded architecture lists that need updating

2. **Remove unused CLI flags** in `internal/cli/root.go`:
   - `--verbose` (line 40) — declared but never read anywhere
   - `--config` (line 39) — declared but never read anywhere
   - `--no-color` (line 41) — declared but never propagated
   - Remove these flags AND the `globalOptions` struct fields for them
   - Search for any code that references `opts.Verbose`, `opts.ConfigPath`, or `opts.NoColor` — if any exists, either implement it or remove the reference
   - Keep `--output` flag — it IS used

3. **Check Migrate method** in `providers/golang/provider.go`:
   - If `Migrate()` returns `ErrNotImplemented`, that's fine — leave it for v2
   - But ensure no CLI command references migrate

4. Run `go build ./...` and `go test ./...` to verify nothing breaks.

When complete, output: CLI CLEANUP DONE
```

---

## Phase 2: Template Worktree Resolution

**Objective:** Resolve the worktree state — merge or discard WIP api/cli templates, ensure template structure is clean.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `TEMPLATES RESOLVED`
**Dependencies:** None

**Prompt:**
```
The git status shows major template changes:
- `providers/golang/templates/go-hexagonal/` DELETED
- `providers/golang/templates/go-minimal/` DELETED
- `providers/golang/templates/api/` NEW (untracked)
- `providers/golang/templates/cli/` NEW (untracked)
- `providers/golang/templates/wizard.yaml` NEW (untracked)

There's also a worktree at `.claude/worktrees/agent-af91544b/` with WIP template code.

Tasks:

1. **Understand current template state:**
   - Read `providers/golang/templates/wizard.yaml` (the new provider-level wizard)
   - Read the `api/` and `cli/` template directories — what do they contain?
   - Read the architectures/ and capabilities/ directories
   - Understand how the scaffold engine (`internal/scaffold/`) loads these

2. **Verify the scaffold pipeline works:**
   - Check `providers/golang/provider.go` — how does `Scaffold()` load templates?
   - Check `internal/scaffold/renderer.go` — how does it resolve template paths?
   - Check `internal/scaffold/composer.go` — how does composition work?
   - Check `internal/scaffold/wizard.go` — how does the wizard load wizard.yaml?

3. **Ensure `archway new` works end-to-end:**
   - Templates must be loadable from the embedded FS
   - Wizard must present architecture choices correctly
   - Scaffolding must produce valid Go projects
   - Run `go build ./...` and `go test ./...`

4. **Clean up worktree if no longer needed:**
   - If the api/cli templates are the intended v1.0 state, ensure they're properly structured
   - If the worktree at `.claude/worktrees/` is stale, note it for deletion

Do NOT commit anything. Just ensure the template structure is consistent and the scaffold pipeline works.

When complete, output: TEMPLATES RESOLVED
```

---

## Phase 3: Ghost Capability Cleanup

**Objective:** Remove or fix all references to capabilities that don't exist (feature-flags, and any others).
**Model:** `sonnet`
**Max Iterations:** 5
**Completion Promise:** `GHOST CAPS FIXED`
**Dependencies:** None

**Prompt:**
```
Search the entire codebase for capabilities that are documented/referenced but don't actually exist in `providers/golang/templates/capabilities/`.

1. **List all actual capabilities** in `providers/golang/templates/capabilities/` — get the definitive list from disk.

2. **Search for "feature-flag" and "feature_flag"** across:
   - `docs/` — any documentation referencing it
   - `website/src/content/docs/` — any website docs
   - `providers/golang/` — any code referencing it
   - `internal/` — any code referencing it
   - `README.md`

3. **For each ghost capability found:**
   - If it's in docs/website: remove the reference or mark as "planned"
   - If it's in code: remove or guard with a check

4. **Verify capability counts** are consistent:
   - Count actual capabilities on disk
   - Update any docs/code that claims a specific number (e.g., "36 capabilities" or "38 capabilities")
   - Update README.md, website homepage, docs

5. **Check category counts** — the docs may claim "8 categories" or "9 categories". Verify against actual capability organization.

When complete, output: GHOST CAPS FIXED
```

---

## Phase 4: `archway guide` — Core Engine

**Objective:** Implement the `archway guide` command that generates AI agent instruction files from archway.yaml + templates.
**Model:** `opus`
**Max Iterations:** 15
**Completion Promise:** `GUIDE CORE DONE`
**Dependencies:** Phase 1, Phase 2

**Prompt:**
```
Implement the `archway guide` command per PRD-002 (`docs/product/prds/PRD-002-archway-guide.md`). Read the PRD first.

This is a NEW command that reads `archway.yaml` + template structure and generates AI agent instruction files.

### Architecture

Create these files:
- `internal/cli/guide.go` — CLI command definition
- `internal/guide/guide.go` — core engine
- `internal/guide/targets.go` — output target definitions (Claude, Cursor, Copilot, Windsurf)
- `internal/guide/renderer.go` — renders architecture-specific content

### CLI Command (`internal/cli/guide.go`)

```go
func newGuideCommand(opts *globalOptions) *cobra.Command {
    var target string
    cmd := &cobra.Command{
        Use:   "guide",
        Short: "Generate AI agent instructions from your architecture",
        Long:  "Reads archway.yaml and generates architecture-aware instructions for AI tools (Claude Code, Cursor, Copilot, Windsurf).",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Find and read archway.yaml in current directory
            // 2. Load architecture info (layer structure, capabilities)
            // 3. Generate instruction content
            // 4. Write to target files
        },
    }
    cmd.Flags().StringVar(&target, "target", "all", "Target tool: all|claude|cursor|copilot|windsurf")
    return cmd
}
```

Register it in `newRootCommand()` in `internal/cli/root.go`.

### Core Engine (`internal/guide/guide.go`)

The engine should:

1. **Read archway.yaml** using the existing config package (`internal/config/`):
   - Architecture type (hexagonal, flat)
   - Capabilities list
   - Module path
   - Service name

2. **Build architecture context**:
   - For hexagonal: layer rules (domain → port → service → adapter), what each layer contains, import restrictions
   - For flat: single package structure, what goes where
   - Per-capability guidance based on which capabilities are enabled

3. **Generate markdown content** with sections:
   - `## Architecture: {type}` — layer rules, directory structure, dependency direction
   - `## Layer Rules` — what can import what, what NEVER imports what
   - `## Directory Structure` — purpose of each directory
   - `## Patterns` — how to add common things (new endpoint, new repository, new domain entity)
   - `## Capabilities` — what's installed and where each capability's code lives
   - `## Anti-patterns` — things to avoid for this architecture

### Output Targets (`internal/guide/targets.go`)

Write to four targets with sentinel-based merging:

1. **Claude Code**: `.claude/rules/archway.md`
   - Write directly (rules are auto-loaded per session)
   - Add `<!-- archway:generated -->` header

2. **Cursor**: `.cursorrules`
   - Sentinel merge: `<!-- archway:start -->` / `<!-- archway:end -->`
   - If file doesn't exist, create with just the archway block
   - If file exists without sentinels, append
   - If file exists with sentinels, replace between them

3. **GitHub Copilot**: `.github/copilot-instructions.md`
   - Same sentinel merge pattern
   - Create `.github/` directory if needed

4. **Windsurf**: `.windsurfrules`
   - Same sentinel merge pattern

### Sentinel Merge Logic

```go
func mergeWithSentinels(existingContent, newContent, startSentinel, endSentinel string) string {
    // If no existing content → return newContent wrapped in sentinels
    // If existing has sentinels → replace between them
    // If existing has no sentinels → append sentinels + newContent
}
```

### Content Generation

For a hexagonal project with http-api + mysql + docker capabilities, the generated content should look like:

```markdown
# Archway — Architecture Guide

> Auto-generated by `archway guide`. Do not edit manually.
> Regenerate with: `archway guide`

## Architecture: Hexagonal

This project follows hexagonal (ports & adapters) architecture.

### Layer Rules
- **domain/** — Business logic, entities, value objects. NO external imports.
- **port/** — Interfaces (inbound + outbound). Depends only on domain.
- **service/** — Application services. Orchestrates domain via ports.
- **adapter/** — Implementations (HTTP handlers, DB repos, external clients). Implements port interfaces.

### Dependency Direction
Dependencies flow INWARD only:
  adapter → service → port → domain

NEVER:
- domain importing from adapter, service, or port
- port importing from adapter or service
- service importing from adapter

### Adding Code

**New HTTP endpoint:**
- Handler goes in `adapter/httphandler/`
- Register route in `adapter/httphandler/router.go`
- Business logic goes in `service/`, called via port interface

**New database repository:**
- Interface goes in `port/outbound.go`
- Implementation goes in `adapter/mysqlrepo/`

**New domain entity:**
- File goes in `domain/`
- No imports from other layers allowed

**New external service client:**
- Interface goes in `port/outbound.go`
- Implementation goes in `adapter/` or `platform/httpclient/`

### Capabilities
- **http-api** — Chi router in `adapter/httphandler/`, middleware in same package
- **mysql** — Connection in `adapter/mysqlrepo/`
- **docker** — `Dockerfile` and `docker-compose.yml` at project root

### Anti-patterns to Avoid
- Don't put business logic in HTTP handlers (adapter layer)
- Don't import adapter packages from domain
- Don't use global mutable state
- Don't skip error wrapping — always add context with fmt.Errorf
```

### Integration with `archway new`

After scaffolding completes in `providers/golang/provider.go` `Scaffold()`, call `guide.Generate()` automatically so every new project starts with AI agent instructions.

### Tests

Create `internal/guide/guide_test.go`:
- Test hexagonal content generation
- Test flat content generation
- Test sentinel merging (no existing, existing without sentinels, existing with sentinels)
- Test each output target path

Run `go build ./...` and `go test ./...` when done.

When complete, output: GUIDE CORE DONE
```

---

## Phase 5: `archway guide` — Pattern Extraction

**Objective:** Extract code patterns from actual template files so `archway guide` output includes real project-specific patterns.
**Model:** `opus`
**Max Iterations:** 10
**Completion Promise:** `PATTERN EXTRACTION DONE`
**Dependencies:** Phase 4

**Prompt:**
```
Enhance `archway guide` to extract real code patterns from the project's template files, not just static text.

Read PRD-002 (`docs/product/prds/PRD-002-archway-guide.md`) for context on pattern extraction.

### What to Build

1. **Template Pattern Extractor** (`internal/guide/patterns.go`):
   - Given an architecture type and list of capabilities, read the corresponding template files from the embedded FS
   - Strip Go template directives (`{{ }}`) from `.tmpl` files
   - Extract representative code patterns (handler functions, repository implementations, domain entities)
   - Format as "follow this pattern" guidance

2. **Capability-Aware Patterns**:
   - If project has `http-api` → extract handler pattern from `adapter/httphandler/handler.go.tmpl`
   - If project has `mysql` → extract repo pattern from `adapter/mysqlrepo/connection.go.tmpl`
   - If project has `grpc` → extract gRPC handler pattern
   - If project has `kafka-consumer` → extract consumer pattern
   - Map capability names to their template files

3. **Smart Stripping**:
   ```go
   func stripTemplateDirectives(content string) string {
       // Remove {{ .ServiceName }} → leave as placeholder or replace with example
       // Remove {{ if .HasCapability "x" }} blocks
       // Remove {{ range }} blocks
       // Keep the actual Go code structure
   }
   ```

4. **Integration**:
   - Add extracted patterns to the `## Patterns` section of guide output
   - Patterns should be real code from templates, not hand-written examples
   - This makes the guide output project-specific — a hexagonal project with grpc gets different patterns than one with http-api

### Tests

Add tests to `internal/guide/guide_test.go`:
- Test pattern extraction from a template file
- Test template directive stripping
- Test capability-to-template mapping

Run `go build ./...` and `go test ./...` when done.

When complete, output: PATTERN EXTRACTION DONE
```

---

## Phase 6: CLI Integration Tests

**Objective:** Add integration tests for all CLI commands to ensure they work end-to-end.
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `CLI TESTS PASSING`
**Dependencies:** Phase 1, Phase 2

**Prompt:**
```
Add CLI integration tests for Archway. There are currently ZERO tests in `internal/cli/`.

### Test File: `internal/cli/cli_test.go`

Use Go's standard testing + `os/exec` or test the cobra commands directly. Follow the project's existing test patterns (check `internal/scaffold/wizard_test.go` and `internal/scaffold/tui_test.go` for style).

### Tests to Write

1. **`archway new` (non-interactive):**
   - `archway new --name test-svc --arch hexagonal --cap http-api --no-wizard`
   - Verify: output directory created, expected files exist (go.mod, main.go, archway.yaml, domain/, adapter/, etc.)
   - Verify: go.mod has correct module path
   - Verify: archway.yaml has correct architecture and capabilities
   - Clean up temp directory after test

2. **`archway new` with flat architecture:**
   - `archway new --name test-flat --arch flat --no-wizard`
   - Verify: simpler structure (main.go, go.mod, archway.yaml)

3. **`archway new` with invalid architecture:**
   - `archway new --name test-bad --arch nonexistent --no-wizard`
   - Verify: returns error with helpful message

4. **`archway check` on a clean project:**
   - Scaffold a hexagonal project first
   - Run `archway check` on it
   - Verify: exits 0, no violations

5. **`archway analyze` on a scaffolded project:**
   - Scaffold a hexagonal project
   - Run `archway analyze`
   - Verify: detects hexagonal architecture

6. **`archway init`:**
   - Run in a temp directory with some Go files
   - Verify: creates archway.yaml

7. **`archway guide`:**
   - Scaffold a project, then run `archway guide`
   - Verify: `.claude/rules/archway.md` created
   - Verify: `.cursorrules` created
   - Verify: content contains architecture layer rules

8. **`archway version`:**
   - Verify: outputs version string without error

### Test Helpers

```go
func scaffoldTestProject(t *testing.T, name, arch string, caps ...string) string {
    t.Helper()
    dir := t.TempDir()
    // Run archway new in dir
    return dir
}
```

### Important
- Use `t.TempDir()` for all test directories (auto-cleanup)
- Use `t.Helper()` in helper functions
- Run `go test ./internal/cli/ -v` to verify all pass
- Run `go test ./...` to verify nothing else breaks

When complete, output: CLI TESTS PASSING
```

---

## Phase 7: Documentation + Website Update

**Objective:** Update all docs, website, and README to reflect v1.0 state with `archway guide` as 4th pillar.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `DOCS UPDATED`
**Dependencies:** Phase 4, Phase 5

**Prompt:**
```
Update all documentation to reflect Archway v1.0 with the 4th pillar (`archway guide`).

### Files to Update

1. **README.md:**
   - Add `archway guide` to the command list
   - Update capability count to match actual count from disk
   - Add "AI-Native" section explaining how `archway guide` feeds AI agents
   - Update the 3 pillars to 4 pillars (Guide → Compose → Analyze → Enforce)

2. **Website homepage** (`website/src/content/docs/index.mdx`):
   - Add Guide as 4th pillar card
   - Update capability count
   - Add "AI-Native" messaging — this is the differentiator

3. **Website docs** — check `website/src/content/docs/`:
   - Add a `guide.mdx` page explaining the command
   - Usage: `archway guide`, `archway guide --target claude`, etc.
   - Show example output
   - Explain which AI tools are supported

4. **CLI help text:**
   - Verify `archway guide --help` is clear and helpful
   - Verify `archway --help` lists all commands including guide

5. **docs/soul.md** — verify it already reflects 4 pillars (should be done from earlier)

6. **Capability count verification:**
   - Count actual capabilities in `providers/golang/templates/capabilities/`
   - Update ALL references to match: README, website, docs, code comments

Do NOT change the tagline. The tagline is LOCKED: "Architecture-aware service composer and enforcer"

Run the website build to verify no errors: `cd website && npm run build`

When complete, output: DOCS UPDATED
```

---

## Phase 8: GoReleaser + First Release

**Objective:** Validate GoReleaser config, tag v1.0.0, publish first release with Homebrew.
**Model:** `sonnet`
**Max Iterations:** 5
**Completion Promise:** `V1.0 RELEASED`
**Dependencies:** All previous phases

**Prompt:**
```
Prepare and execute the first Archway v1.0.0 release.

### Pre-Release Checklist

1. **Verify GoReleaser config** (`.goreleaser.yaml`):
   - Builds for linux/darwin/windows, amd64/arm64
   - Homebrew tap configured for `dcsg/homebrew-tap`
   - Changelog configured
   - Run `goreleaser check` to validate config

2. **Verify all tests pass:**
   ```bash
   go test ./...
   go vet ./...
   golangci-lint run
   ```

3. **Verify binary builds and works:**
   ```bash
   go build -o archway ./cmd/archway
   ./archway version
   ./archway --help
   ```

4. **Commit all changes:**
   - Stage all modified and new files
   - Commit message: "feat: archway v1.0.0 — guide, compose, analyze, enforce"
   - Do NOT push yet — wait for user confirmation

5. **Prepare release:**
   - Create git tag: `git tag -a v1.0.0 -m "Archway v1.0.0"`
   - Dry run: `goreleaser release --snapshot --clean`
   - Verify artifacts are created correctly

6. **Document release:**
   - Prepare release notes (changelog highlights)
   - List the 4 pillars
   - List supported architectures
   - List capability count
   - Mention AI agent guidance as key feature

DO NOT push the tag or run the actual release — just prepare everything and output what needs to be done. The user will trigger the actual release.

When complete, output: V1.0 RELEASED
```

---

## Known Risks

| Risk | Severity | Mitigation |
|---|---|---|
| Template worktree may have conflicts with main | Medium | Phase 2 resolves this before other work starts |
| `archway guide` pattern extraction may be complex for all template types | Medium | Start with hexagonal + http-api, add others iteratively |
| CLI tests may reveal bugs in scaffold pipeline | Low | Good — better to find now than after release |
| GoReleaser Homebrew tap may need manual setup | Low | One-time setup of dcsg/homebrew-tap repo |
