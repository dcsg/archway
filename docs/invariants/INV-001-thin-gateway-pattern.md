# INV-001: Interactive code must be a thin gateway

All TUI, CLI prompt, and interactive code must be a thin presentation gateway. Business logic, validation rules, data transformation, and state computation must live in pure functions that are testable without a terminal, TTY, or interactive session.

## What This Means

- **Compliant:** `computeDerived(state)` is a pure function that computes boolean flags from state — tested with simple unit tests
- **Compliant:** `evaluateWhen(expr, values)` is a pure function that evaluates conditional expressions — tested with table-driven tests
- **Compliant:** `RunWizard()` calls `buildWizardGroups()` (testable) then `form.Run()` (thin gateway) — only 3 lines of untestable code
- **Violation:** Embedding capability conflict detection logic inside a `huh.Form` validation closure with no way to test it independently
- **Violation:** Computing derived state or making architectural decisions inside an interactive prompt handler

## Why

Code trapped inside interactive frameworks (`huh`, `bubbletea`, `cobra.RunE` with prompts) cannot be unit tested without mocking terminals. This creates blind spots where bugs hide. By keeping interactive code as a thin gateway — collect input, delegate to pure functions, render output — we guarantee that all business logic is covered by fast, deterministic tests. The gateway itself is trivially correct by inspection.

---

*Captured by keel:invariant — 2026-03-11*
