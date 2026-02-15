# ADR-005: Three-Layer Extensibility Model

**Status:** accepted
**Date:** 2026-02-14
**Deciders:** Daniel Gomes

## Context

Archway must be extensible at multiple levels: template authors, rule authors, and language support contributors. Different levels of extensibility require different skill levels and mechanisms. The challenge is enabling broad contribution without requiring Go expertise for everything.

## Constraints

- Single binary distribution (no dynamic loading for MVP)
- Template creation should not require programming knowledge
- Rule presets should be shareable across teams (like ESLint configs)
- Language analyzers require deep knowledge of AST/types — restricted scope

## Options Considered

### Option A: Three-layer model (templates / presets / analyzers) — CHOSEN

**Layer 1 — Templates (data, not code):**
- Directory of files + manifest.yaml + wizard.yaml
- `{{.Variable}}` placeholders via text/template
- Anyone can create — no Go knowledge needed
- Distribution: embedded (official), git repos, local dirs

**Layer 2 — Rules & Presets (shareable YAML):**
- `archway.yaml` with dependency rules, naming, structure, functions
- `extends:` key composes presets (like ESLint configs)
- Shareable as git repos: `archway/go-hexagonal-strict`
- Teams can create org-specific presets

**Layer 3 — Language Analyzers (embedded Go code):**
- One per language, compiled into binary
- Go: native go/ast + go/packages
- Post-MVP others: exec bridge to language-native tools
- Highest barrier — requires Go expertise and PR to core

**Pros:**
- Clear skill gradient: YAML → YAML → Go code
- Templates accessible to widest audience
- Presets enable community-driven best practices
- Analyzers are protected from instability (compiled in, tested)
- Each layer has its own distribution story

**Cons:**
- Three different extension mechanisms to document
- Preset resolution (extends chain) adds complexity
- Embedded analyzers require core release for new languages

### Option B: Everything is a plugin (hashicorp/go-plugin)

Templates, rules, and analyzers all as gRPC plugins.

**Pros:**
- Uniform extension mechanism
- Community can contribute everything independently

**Cons:**
- Massive complexity for simple use cases (templates as gRPC plugins?)
- Multiple binaries to manage
- Overkill for YAML config sharing
- High barrier for template authors

### Option C: Everything embedded, no extensibility

All templates, rules, and analyzers compiled in. Users configure via flags only.

**Pros:**
- Simplest implementation
- No extension mechanism to support

**Cons:**
- No community contribution path
- No team-specific customization
- Every change requires core release
- Dead end for growth

### Option D: npm-style packages

Templates and rules as npm/Go modules with a central registry.

**Pros:**
- Familiar package management
- Versioning built-in

**Cons:**
- Requires registry infrastructure
- Heavy for YAML config sharing
- Go module system not designed for non-code assets

## Decision

**Option A: Three-layer model.** Each layer matches a natural skill level and contribution type. Templates are the lowest barrier (anyone can create a directory of files), presets are the community backbone (shareable YAML configs), and analyzers are the protected core (compiled Go code requiring expertise).

### Distribution Matrix

| Layer | Authoring | Distribution | Skill Required |
|-------|-----------|-------------|----------------|
| Templates | YAML + text files | embedded / git / local | None (text editing) |
| Presets | YAML config | git repos / embedded | Domain knowledge |
| Analyzers | Go code | compiled into binary | Go + AST expertise |

## Consequences

- Template format (manifest.yaml + wizard.yaml + files/) is a public contract — must be stable
- Preset resolution (`extends:` chain) must handle conflicts and ordering
- Analyzer interface must be designed for future extraction to plugins (ADR-001)
- Documentation needs three separate contribution guides
- Post-MVP: consider `archway template add <git-url>` and `archway preset add <git-url>` commands
