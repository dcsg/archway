# ADR-007: Polyglot Architecture — Go Orchestrator + Embedded Rust Analysis Engine

**Status:** accepted
**Date:** 2026-03-09
**Deciders:** Daniel Gomes

## Context

Archway's roadmap targets 7 architectures across multiple languages (Go, TypeScript, Python, Java). The current analysis engine uses `go/ast` and `go/packages`, which are Go-only. Scaling to N languages means either:

1. Reimplementing analyzers per language in Go (0% shareability for detection)
2. Using a universal parser that works across all languages

Research evaluated Go (with CGo tree-sitter), Rust, Zig, OCaml, TypeScript, and Python for the analysis engine. Key findings:

- **Tree-sitter** is the industry standard for cross-language parsing (100+ grammars)
- **Go + CGo tree-sitter**: ~3.9x slower parsing than native, breaks easy cross-compilation, ~69ns per FFI call adds up across millions of AST node traversals
- **Rust**: Native tree-sitter bindings (zero FFI overhead), `enum` types ideal for multi-language AST, proven by ast-grep and Semgrep's migration
- **Zig**: Best C interop (zero overhead to tree-sitter's C API) but pre-1.0, no ecosystem (no CLI frameworks, no YAML parsing, no TUI libraries). In a separate-binary architecture, Zig's C interop advantage disappears — Rust has the same zero-cost access plus a mature ecosystem.
- **Go CLI stack** (Cobra + Bubbletea + Viper) is unmatched and already built

No single language is best at everything. Go excels at CLI/TUI/config/templates. Rust excels at parsing/analysis/rule engines.

## Decision

Adopt a **polyglot architecture**: Go orchestrator + Rust analysis engine, communicating via **protobuf over stdin/stdout**. The Rust binary is embedded inside the Go binary via `//go:embed`.

```
┌─────────────────────────────────────────────┐
│  archway (Go)                               │
│  CLI, TUI, config, scaffold, templates      │
└──────────────┬──────────────────────────────┘
               │ stdin/stdout (protobuf, length-prefixed)
       ┌───────▼───────────────────────┐
       │  archway-engine (Rust)        │
       │  tree-sitter + rules engine   │
       │  analyze | check | detect     │
       └───────────────────────────────┘
```

**Go owns:** CLI (Cobra), TUI wizard (Bubbletea), config (Viper), template rendering, scaffold composition, archway.yaml generation, output formatting.

**Rust owns:** Code parsing (tree-sitter), import graph analysis, architecture detection, anti-pattern checking, rule engine, structural queries.

**Distribution:** Rust binary embedded in Go binary via `//go:embed`. Extracted to cache on first run. Single `archway` binary — users don't know it's polyglot. GoReleaser builds Rust engine per target platform first, then Go embeds the right one.

### Internal Protocol: Protobuf

The Go orchestrator and Rust engine communicate via **protobuf** over stdin/stdout with length-prefixed framing.

**Why protobuf over JSON/YAML/MessagePack:**

| Concern | Protobuf | JSON | MessagePack |
|---|---|---|---|
| Schema enforcement | `.proto` generates Go + Rust types — compile-time safety | Hand-maintained structs can silently drift | Same as JSON |
| Parse speed | 5-10x faster than JSON | Baseline | 2-3x faster |
| Streaming | Length-prefixed messages, natural incremental delivery | Needs NDJSON workaround | Possible but awkward |
| Debugging | `--debug` flag dumps as JSON | Human-readable by default | Binary, needs tooling |

**Protocol design:**

```protobuf
// engine.proto — the contract between Go and Rust

message AnalyzeRequest {
  string project_path = 1;
  string language = 2;
  string architecture = 3;
  repeated Rule rules = 4;
}

message EngineMessage {
  oneof payload {
    ProgressUpdate progress = 1;
    Finding finding = 2;
    AnalysisComplete complete = 3;
    EngineError error = 4;
  }
}

message ProgressUpdate {
  uint32 files_parsed = 1;
  uint32 total_files = 2;
}

message Finding {
  string file = 1;
  uint32 line = 2;
  string rule = 3;
  Severity severity = 4;
  string message = 5;
}

message AnalysisComplete {
  string architecture_detected = 1;
  float confidence = 2;
}
```

The `.proto` file generates matching types on both sides. A field rename or type change breaks compilation in both Go and Rust — mismatches are caught at build time, not runtime.

**Streaming:** The Rust engine sends length-prefixed `EngineMessage` packets as findings are discovered. Go reads and displays them incrementally — progress bars, real-time findings on large codebases.

**User-facing formats remain YAML:** archway.yaml, rule definitions, manifests — all human-readable YAML. Protobuf is purely internal, invisible to users.

## Consequences

### Positive

- **~85% shareability** across language providers. Engine and rule framework are 100% shared. Per-language tree-sitter queries are needed for language-specific anti-patterns (~10 query files per language).
- **Zero FFI overhead** for analysis — Rust calls tree-sitter natively
- **Keep the entire Go CLI stack** — no rewrite of wizard, renderer, config
- **Compile-time contract enforcement** — protobuf schema prevents Go/Rust struct drift
- **Streaming support** — progress and findings delivered incrementally
- **Proven model** — Semgrep (OCaml + C + Rust), Hugo (Go + embedded Dart Sass), Turborepo (Go + Rust)
- **Each component uses the best language for its job**

### Negative

- **Two build toolchains** — CI needs both Go and Rust, cross-compilation for Rust per GOOS/GOARCH. This is the highest-risk item — every release touches it.
- **Two languages for contributors** — CLI contributors need Go, engine contributors need Rust. Clear boundary mitigates this.
- **Protobuf toolchain** — adds `protoc` + `prost` (Rust) + `protoc-gen-go` to the build. One-time setup cost.
- **Embedded binary size** — adds ~5-10MB to the archway binary (total ~20-25MB)
- **Anti-pattern queries are per-language work** — 8 of 11 current detectors are Go-semantic. Each new language needs its own anti-pattern query set.

### Timeline

- **v1.x** — Pure Go. Ship with `go/ast`. No Rust yet.
- **v2.0** — Introduce Rust engine. Protobuf protocol + tree-sitter + YAML rules. Clean replacement of `go/ast` analysis (no fallback — clean cut).
- **v3.0** — TypeScript provider via new tree-sitter grammar + queries in the Rust engine. Go CLI unchanged.

## Alternatives Considered

| Alternative | Why rejected |
|---|---|
| Pure Go + CGo tree-sitter | 3.9x parse slowdown, breaks cross-compilation, no sum types for AST |
| Pure Rust rewrite | Loses Cobra + Bubbletea CLI stack, massive rewrite cost |
| Pure Zig | Pre-1.0, no ecosystem (CLI, YAML, TUI). C interop advantage irrelevant in separate-binary model |
| OCaml | Semgrep is leaving it for Rust, tiny contributor pool |
| Go shelling out to ast-grep | External dependency, less control over rule engine |
| Rust as C shared library (.so/.dylib) | Back to CGo cross-compile pain |
| JSON for internal protocol | No schema enforcement, Go/Rust structs can drift silently, slower parsing |
| MessagePack for internal protocol | No schema enforcement (same drift risk as JSON), not human-debuggable |
