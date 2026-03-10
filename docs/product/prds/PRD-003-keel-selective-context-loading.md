# PRD-003: Selective Context Loading for Keel

**Status:** draft
**Date:** 2026-03-10

---

## Problem

Keel's `/keel:context` loads **all** ADRs, invariants, and product docs into the session context at startup. For small projects (3-5 ADRs) this is fine. For established projects with 15+ ADRs, 10+ invariants, and multiple PRDs, this consumes thousands of tokens of context window space — most of which is irrelevant to the current task.

The research is clear: more instructions = lower compliance (IFScale 2025: best model achieves 68.9% at 500 instructions). Loading irrelevant ADRs doesn't just waste tokens — it actively degrades the agent's ability to follow the relevant ones (Lost in the Middle effect, attention dilution).

This is a scaling problem. The more disciplined a team is about documenting decisions, the worse the context loading problem becomes — a perverse incentive against good governance.

## Users

- **Individual developers** using Keel on established projects with 10+ ADRs
- **Teams** with shared governance repos where ADR count grows over time
- **Enterprise adopters** where projects accumulate 50+ decisions over years

## Goals

- Reduce context token consumption by **50-70%** for projects with 10+ ADRs
- Maintain **100% loading of invariants** (hard constraints must never be skipped)
- Load only ADRs **relevant to the current task/branch** by default
- Provide escape hatch to load everything when needed (`--depth=full`)
- Zero configuration required — smart defaults that work out of the box

## Non-Goals

- Embedding-based semantic search (too complex for v1, requires vector DB)
- Per-file rules about which ADRs to load (manual curation doesn't scale)
- Modifying how Claude Code itself manages its context window
- Supporting non-Claude tools (Cursor, Copilot) — Keel is Claude Code only

## Requirements

### Must Have

- **Branch-aware ADR selection** — Match current git branch name against ADR filenames and titles. `feat/auth` loads `003-authentication-strategy.md` automatically.
- **Recency fallback** — When no branch match, load the 5 most recently modified ADRs
- **Always-load invariants** — All invariants load regardless of depth (they are safety constraints)
- **Depth flags** — `--depth=full` (everything), `--depth=focused` (smart selection), `--depth=minimal` (soul + current phase + invariants only)
- **Smart default** — Auto-select depth based on project size: `full` for <10 ADRs, `focused` for 10+
- **Transparency** — Show what was loaded and why: "Loaded 3/18 ADRs (matched branch: feat/auth)"

### Should Have

- **Keyword matching from active plan phase** — Current phase title/objective keywords matched against ADR content
- **ADR dependency tracking** — If ADR-007 references ADR-003, loading 007 should also load 003
- **On-demand loading hint** — When an ADR wasn't loaded but becomes relevant mid-session, suggest: "ADR-012 may be relevant — run `/keel:context --include 012`"
- **MEMORY.md integration** — Store which ADRs were most frequently loaded per branch pattern

### Won't Have (v1)

- Embedding/vector-based semantic matching
- Automatic ADR relevance scoring via LLM
- Cross-project ADR deduplication
- Real-time re-evaluation during session (load once at start)
- GUI/TUI for selecting ADRs

## User Stories

**As a** developer on an established project, **I want** Keel to load only the ADRs relevant to my current branch **so that** my context window has room for the actual code I'm working on.

**As a** team lead, **I want** invariants to always load **so that** hard constraints are never accidentally skipped regardless of context budget.

**As a** developer starting a quick task, **I want** minimal context loading **so that** I can get to work faster without wasting tokens on irrelevant governance docs.

**As a** developer doing architectural work, **I want** to load all context **so that** I have the full picture when making cross-cutting decisions.

## Acceptance Criteria

- [ ] `--depth=focused` loads <50% of tokens compared to `--depth=full` on a project with 15+ ADRs
- [ ] All invariants load at every depth level
- [ ] Branch name `feat/auth` correctly matches ADR files containing "auth" in filename or title
- [ ] When no branch matches, falls back to 5 most recently modified ADRs
- [ ] Output clearly states what was loaded and what was skipped
- [ ] `--depth=full` still loads everything (escape hatch)
- [ ] Default depth auto-adjusts based on ADR count (<10 → full, 10+ → focused)
- [ ] No breaking changes to existing `/keel:context` behavior for projects with <10 ADRs

## Technical Notes

- Implementation lives in the `/keel:context` skill definition
- Branch detection: `git branch --show-current`
- ADR matching: filename substring match + title line grep (fast, no embeddings needed)
- Recency: `ls -t` on ADR directory (modification time)
- Token estimation: count lines × ~4 tokens/line (rough but sufficient for budgeting)
- Consider a `context_budget:` field in `.keel/config.yaml` for teams that want explicit limits

## Open Questions

- Should `/keel:context` support `--include ADR-012,ADR-015` for manual additions to focused mode?
- Should there be a `context:` section in `.keel/config.yaml` for persistent depth preferences?
- Should we track ADR "hit rate" (how often each ADR is actually referenced in a session) to improve selection over time?
- How should this interact with the `PreCompact` hook — should we re-evaluate what to keep when context compacts?

---

*Written by keel:prd — 2026-03-10*
