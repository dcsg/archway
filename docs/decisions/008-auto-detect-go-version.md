# ADR-008: Auto-detect language version at scaffold time

**Status:** accepted
**Date:** 2026-03-11

---

## Context

Architecture manifests hardcoded a specific language version (e.g. `default: "1.23"` for Go) as the default for version template variables. Every time a language released a new version, all manifests needed manual updates — a maintenance burden that adds no value and scales poorly as archway becomes multi-language.

## Decision

Replace hardcoded version defaults with `default: "auto"` in architecture manifests. At scaffold time, `resolveAutoVersion()` detects the user's installed language version by dispatching on the variable name:

| Variable | Detection | Format |
|----------|-----------|--------|
| `GoVersion` | `go env GOVERSION` (fallback: `runtime.Version()`) | major.minor (e.g. `1.26`) |
| `NodeVersion` | `node --version` | major (e.g. `22`) |
| `RustVersion` | `rustc --version` | major.minor (e.g. `1.82`) |
| `PythonVersion` | `python3 --version` | major.minor (e.g. `3.12`) |

Users can still override with `--set GoVersion=1.24` (or any version variable) for specific version pinning.

New languages are added by:
1. Adding a detector function
2. Mapping the variable name in `versionVariableToLanguage`

## Rationale

The user's installed language version is the most sensible default — it's what they'll compile with. Hardcoding creates unnecessary coupling between archway releases and language releases. A generic mechanism avoids duplicating this pattern per language.

## Alternatives Considered

### Keep hardcoding the latest stable version per language
- **Pros:** Simple, no runtime detection needed
- **Cons:** Requires template updates on every release of every language, scales badly with polyglot support

### Fetch latest version from language-specific APIs (go.dev, nodejs.org, etc.)
- **Pros:** Always up-to-date regardless of local install
- **Cons:** Requires network access during scaffold, fragile, slower, one API per language

### Single `LanguageVersion` variable name
- **Pros:** Simpler mapping
- **Cons:** Breaking change for existing `GoVersion` templates, less explicit in multi-language projects

## Consequences

### Positive
- Zero maintenance when any language releases a new version
- Scaffolded projects always match the user's local toolchain
- Scales to any number of languages with minimal code
- Still overridable via `--set` flags

### Negative / Trade-offs
- Requires the language toolchain to be installed for detection (reasonable — archway targets developers using that language)
- If the binary is missing, returns empty string (user must provide the version explicitly)

---

*Captured by keel:adr — 2026-03-11*
