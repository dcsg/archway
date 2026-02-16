# ADR-001: Embedded Language Providers Over Plugin System

**Status:** accepted
**Date:** 2026-02-14
**Deciders:** Daniel Gomes

## Context

Archway needs a language provider architecture to support multiple programming languages (Go first, then PHP, Node, Python, etc.). Each provider implements `Scaffold()`, `Analyze()`, `Migrate()`, `GetInfo()`. The question is how providers are packaged and loaded.

## Constraints

- MVP is Go-only; multi-language is post-MVP
- Single binary distribution is a hard requirement (NF-01)
- Must work offline, no network dependencies for core features
- Small team, solo maintainer initially

## Options Considered

### Option A: Embedded providers (Go packages compiled into binary) — CHOSEN

Providers are Go packages imported directly:
```go
import goprovider "archway/providers/golang"
var providers = map[string]Provider{"go": &goprovider.Provider{}}
```

**Pros:**
- Single binary, zero setup
- No version compatibility issues
- Simple debugging (single process)
- Fast — direct function calls, no serialization
- Easy to test (standard Go test tooling)

**Cons:**
- Adding a language requires recompiling binary
- Binary grows with each provider
- Non-Go languages need exec bridges (e.g., PHP parser via subprocess)

### Option B: gRPC plugins (hashicorp/go-plugin)

Each provider is a separate binary, communication via gRPC over loopback.

**Pros:**
- Language-agnostic plugins (PHP provider in PHP)
- Crash isolation (plugin crash doesn't kill host)
- Independent versioning and updates
- Proven at scale (Terraform, Vault)

**Cons:**
- Significant complexity (protobuf, gRPC, handshake, versioning)
- Multiple binaries to distribute
- RPC overhead (10-100x slower than direct calls)
- Debugging across processes is harder
- Overkill when there's only 1 provider (Go)

### Option C: WASM plugins (knqyf263/go-plugin)

Providers compiled to WebAssembly, loaded in-process via Wazero runtime.

**Pros:**
- Single-process execution
- Sandboxed (safe for untrusted plugins)
- Portable (single compilation target)

**Cons:**
- WASM ecosystem immature for Go (2026)
- Complex FFI for data exchange
- Performance overhead vs native
- Very few Go developers have WASM experience

### Option D: Native Go plugins (.so)

Dynamic shared libraries loaded at runtime.

**Pros:**
- In-process, fast

**Cons:**
- Linux/macOS only (no Windows)
- Exact Go version + dependency match required
- Fragile, widely considered deprecated
- golangci-lint abandoned this approach

## Decision

**Option A: Embedded providers.** For MVP with a single Go provider, the simplicity advantage is overwhelming. The provider interface is designed for future extraction — when community demand exists for 10+ languages, we can extract to Option B (gRPC plugins) without changing the provider interface contract.

## Consequences

- Binary includes all providers (acceptable for Go-only MVP)
- Non-Go language analysis requires exec bridges to language-native tools
- Provider interface must be clean enough to extract to gRPC later without breaking changes
- Post-MVP migration path: embedded → gRPC plugins when community demand exists
