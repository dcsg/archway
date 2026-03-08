# ADR-006: Project Structure — internal/ + providers/ Layout

**Status:** accepted
**Date:** 2026-02-14
**Deciders:** Daniel Gomes

## Context

Archway needs a Go project structure that supports: CLI entry point, multiple internal packages (analyzer, scaffolder, config, LLM, MCP, output), language providers with embedded templates, and a clean public API surface. The structure must accommodate growth from 1 provider (Go) to many.

## Constraints

- Must follow Go conventions (cmd/, internal/, pkg/)
- Embedded templates (embed.FS) require specific directory placement
- Provider code must be isolatable for future plugin extraction (ADR-001)
- internal/ packages are not importable by external code (private)
- pkg/ packages are the public API (if any)

## Decision

```
archway/
  cmd/
    archway/                 # CLI entry point
      main.go

  internal/
    cli/                     # Cobra command definitions
      root.go
      new.go
      init.go
      analyze.go
      configure.go
      mcp.go
      version.go
    config/                  # Viper-based configuration
      config.go
      archway_yaml.go        # archway.yaml parser
    provider/                # Provider interface + registry
      provider.go            # LanguageProvider interface
      registry.go            # Provider discovery + registration
    scaffold/                # Template engine
      renderer.go            # text/template rendering
      hooks.go               # Post-scaffold hooks (go mod init, git init)
      wizard.go              # TUI wizard driven by wizard.yaml
    analyzer/                # Core analysis engine
      analyzer.go            # Analysis orchestration
      detector/              # Language, architecture, framework detection
        language.go
        architecture.go
        framework.go
        convention.go
      graph/                 # Dependency graph construction
        graph.go
        violations.go
    llm/                     # LLM provider abstraction
      provider.go            # LLM Provider interface
      openai.go              # OpenAI-compatible client
      detect.go              # Auto-detection chain
    mcp/                     # MCP server implementation
      server.go
      tools.go
      resources.go
    output/                  # Formatters
      terminal.go
      json.go
      markdown.go

  providers/
    golang/                  # Go language provider
      provider.go            # Implements LanguageProvider
      analyzer.go            # Go-specific AST analysis
      scaffolder.go          # Go-specific scaffold logic
      templates/             # 66+ embedded templates
        go-hexagonal/
          manifest.yaml
          wizard.yaml
          files/
            ...

  go.mod
  go.sum
  Makefile
  .goreleaser.yaml
```

## Rationale

### Why `internal/` for core packages?

Core packages (cli, config, analyzer, scaffold, llm, mcp, output) are implementation details. They should not be imported by external consumers. This gives us freedom to refactor without breaking external users.

### Why `providers/` at root (not `internal/providers/`)?

Providers are designed for future extraction to separate binaries (ADR-001). Keeping them at root level makes the extraction path clear: `providers/golang/` becomes its own module. Additionally, `providers/golang/templates/` needs embed.FS directives that are cleaner at a dedicated package level.

### Why not `pkg/`?

No public Go API is planned for v1. If we add one later (e.g., for programmatic embedding of Archway), it would go in `pkg/archway/`. Not creating it now avoids premature abstraction.

### Why `internal/cli/` separate from `cmd/archway/`?

`cmd/archway/main.go` is minimal (just calls CLI root). All command logic lives in `internal/cli/` so it can be tested without building the binary. Each command is a separate file for easy navigation.

## Consequences

- All core logic is in `internal/` — maximum refactoring freedom
- Provider code at `providers/` level signals intentional extraction boundary
- embed.FS directives in `providers/golang/templates/` keep template loading clean
- New languages add a new directory under `providers/` (e.g., `providers/php/`)
- No public Go API initially — add `pkg/` only if external consumption is needed
