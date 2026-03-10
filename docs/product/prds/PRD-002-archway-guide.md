# PRD-002: AI-Native Architecture Guidance — `archway guide`

**Status:** draft
**Date:** 2026-03-09

---

## Problem

AI agents (Claude Code, Cursor, Copilot, Windsurf) write code without architectural awareness. They don't know the project uses hexagonal architecture, which layer a new file belongs in, or what patterns to follow when adding an endpoint. The result: AI-generated code violates architecture conventions, and developers spend time correcting it — or worse, the violations ship.

Today's workflow is reactive: the agent writes code → `archway check` catches violations → the developer fixes them. This is backwards. The agent should know the architecture **before** it writes a single line.

No tool in the market solves this. Architecture tools output to terminals for humans. AI agents can't read terminal output proactively — they need rules and context injected into their session before they start coding.

## Users

- **Developers using AI agents** — anyone using Claude Code, Cursor, Copilot, or Windsurf to write code in a project with an `archway.yaml`
- **Tech leads** — want assurance that AI agents follow the team's architecture conventions
- **Teams adopting AI-assisted development** — need architectural guardrails that work with AI, not against it

## Goals

- AI agents write architecturally correct code **from the first line** — no reactive correction loop
- Every AI tool works — no vendor lock-in, no MCP dependency, no protocol requirements
- Templates become living patterns — not just scaffolding artifacts but reference material AI agents use to extend the codebase
- Zero manual maintenance — when `archway.yaml` or templates change, agent instructions update automatically

## Non-Goals

- Not an MCP server — agents don't call Archway, Archway feeds agents
- Not a linter — `archway check` handles enforcement; `archway guide` handles prevention
- Not documentation — output is optimized for AI consumption (concise, structured, actionable), not human reading
- Not a one-time generation — must stay in sync with `archway.yaml` and templates

## Requirements

### Must Have

- `archway guide` command that reads `archway.yaml` + templates and generates AI agent instruction files
- Output targets:
  - `.claude/rules/archway.md` (Claude Code — rules are auto-loaded per session)
  - `.cursorrules` (Cursor — append to existing, use sentinels)
  - `.github/copilot-instructions.md` (GitHub Copilot — append to existing)
  - `.windsurfrules` (Windsurf — append to existing)
- Generated content includes:
  - Architecture type and layer rules (what imports what, what goes where)
  - Directory structure with purpose of each directory
  - Code patterns extracted from templates (how to add an endpoint, a repository, a domain entity)
  - Capability inventory (what's in the project and where it lives)
  - Anti-patterns to avoid (from the architecture's known violations)
- Sentinel-based merging (`<!-- archway:start -->` / `<!-- archway:end -->`) — never overwrite user content
- `archway guide` runs automatically as a post-scaffold hook after `archway new`
- Idempotent — running `archway guide` twice produces the same output

### Should Have

- `--target` flag to generate for a specific tool only (`--target=claude`, `--target=cursor`)
- `--watch` mode that regenerates when `archway.yaml` changes
- Pattern extraction from templates — automatically converts `.tmpl` files into "follow this pattern" instructions
- Capability-aware guidance — if the project has `mysql` capability, include DB pattern guidance; if `grpc`, include proto/handler patterns

### Won't Have (v1)

- MCP server integration (separate feature, may come later)
- Custom rule authoring (use archway.yaml + templates as the source of truth)
- Per-file or per-directory scoped guidance (all guidance is project-wide)
- Auto-detection of which AI tools are in use (user runs `archway guide` or specifies `--target`)

## User Stories

**As a** developer using Claude Code, **I want** my AI agent to know my project's hexagonal architecture **so that** it places new code in the correct layer without me having to explain it every session.

**As a** tech lead, **I want** `archway guide` to run after `archway new` **so that** every new project is immediately AI-agent-ready with the right architectural context.

**As a** developer adding a new capability, **I want** `archway guide` to regenerate **so that** AI agents know about the new capability's patterns and file locations.

**As a** team using multiple AI tools, **I want** one command that generates instructions for all tools **so that** every team member gets the same architectural guardrails regardless of their editor.

## Acceptance Criteria

- [ ] `archway guide` generates `.claude/rules/archway.md` from `archway.yaml` + templates
- [ ] Generated rules include layer boundaries, directory structure, and code patterns
- [ ] Generated rules include anti-patterns derived from the architecture type
- [ ] Running `archway guide` twice produces identical output (idempotent)
- [ ] Sentinel-based merging preserves user content in `.cursorrules` and other shared files
- [ ] `archway new` runs `archway guide` automatically after scaffolding
- [ ] An AI agent (Claude Code) using the generated rules places a new HTTP handler in the correct directory and follows the correct pattern without additional prompting
- [ ] Works for hexagonal and flat architectures (v1); extensible to all 7

## Technical Notes

- Pure Go implementation — reads `archway.yaml` via Viper, reads templates from embedded FS
- Template pattern extraction: parse `.tmpl` files, strip template directives, present as "follow this pattern"
- ~300-500 LOC estimated
- No Rust engine dependency — this is a Go-side feature
- Output format is Markdown — universally understood by all AI agents
- Sentinel comments (`<!-- archway:start -->` / `<!-- archway:end -->`) prevent overwriting user content, same pattern as Keel uses for CLAUDE.md

## Open Questions

- Should `archway guide` also generate a `CLAUDE.md` section (root-level), or only `.claude/rules/`?
- Should the generated guidance include example code inline, or reference template file paths?
- How granular should capability-specific guidance be? (e.g., full MySQL migration patterns vs. just "migrations live in migrations/")
- Should `archway guide --watch` be a v1 feature or deferred?

---

*Written by keel:prd — 2026-03-09*
