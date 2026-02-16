# ADR-004: Terraform Mental Model for Code Architecture

**Status:** accepted
**Date:** 2026-02-14
**Deciders:** Daniel Gomes

## Context

Archway needs a coherent user experience that ties together scaffolding, analysis, migration, and governance. Without a unifying model, these become disconnected features. We need a mental model that users already understand.

## Constraints

- Must accommodate both greenfield (scaffold) and brownfield (analyze/migrate) workflows
- Must support declarative desired state (not just imperative commands)
- v1 delivers init/new/analyze; v2 adds plan/apply/check
- The model must make sense even with only v1 commands

## Options Considered

### Option A: Terraform model (init → analyze → plan → apply → check) — CHOSEN

Map Archway concepts directly to Terraform's workflow:

| Terraform | Archway | Phase |
|-----------|---------|-------|
| `terraform init` | `archway init` / `archway new` | Declare desired state |
| `terraform show` | `archway analyze` | Understand current state |
| `terraform plan` | `archway plan` | Diff desired vs actual |
| `terraform apply` | `archway apply` | Execute changes safely |
| `terraform validate` | `archway check` | Ongoing compliance |

**Pros:**
- Millions of developers know the Terraform workflow
- Natural narrative: "declare what you want, see what you have, plan changes, apply safely"
- Declarative config (archway.yaml) is the single source of truth
- Each command has a clear, single responsibility
- Greenfield and brownfield converge: both produce archway.yaml as desired state

**Cons:**
- Not a 1:1 mapping (code architecture != infrastructure)
- Might set expectations for features not yet built (plan/apply in v2)
- Terraform's state management complexity doesn't apply

### Option B: Linter model (analyze → fix)

Like golangci-lint: scan code, report issues, optionally auto-fix.

**Pros:**
- Very familiar to Go developers
- Simple two-step: detect problems, fix problems

**Cons:**
- No concept of "desired state" — just rules
- No migration planning
- No scaffolding story
- Reactive (find violations) not proactive (declare architecture)

### Option C: Generator model (scaffold → validate)

Like Yeoman/go-blueprint: generate code, then lint it.

**Pros:**
- Simple mental model
- Good for greenfield

**Cons:**
- No brownfield story
- No migration path
- Scaffolding and governance feel disconnected
- No "desired state" concept

### Option D: IDE model (analyze → refactor)

Like GoLand: understand code, apply refactorings.

**Pros:**
- Rich analysis capabilities
- Interactive refactoring

**Cons:**
- Requires IDE integration (not CLI-first)
- No declarative config
- No CI story

## Decision

**Option A: Terraform mental model.** The init/analyze/plan/apply/check workflow provides a complete narrative from project creation to ongoing governance. The `archway.yaml` file serves as the declarative desired state, analogous to HCL. Even with only v1 commands (init, new, analyze), the model makes sense — users declare desired state and understand current state, setting the foundation for v2's plan/apply/check.

## Consequences

- `archway.yaml` is the central artifact — all commands read or write it
- `archway new` generates archway.yaml alongside code (desired state from day one)
- `archway init` creates archway.yaml for brownfield projects
- `archway analyze` produces structured output comparable to archway.yaml format
- v2 commands (plan/apply/check) are natural extensions, not bolted-on features
- Documentation and marketing can use "Terraform for Code Architecture" positioning
- Must be careful not to over-promise v2 features in v1 messaging
