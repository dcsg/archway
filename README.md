# Archway

Terraform for Code Architecture.

Archway is a Go CLI for scaffolding production-ready services and analyzing existing codebases with a Terraform-inspired workflow: declare desired architecture (`archway.yaml`), inspect current state (`archway analyze`), and evolve safely.

## Features (v1)

- `archway new`: scaffold Go services from embedded templates (`go-hexagonal`, `go-minimal`)
- `archway init`: generate `archway.yaml` for an existing project
- `archway analyze`: detect architecture, frameworks, conventions, and dependency relationships
- `archway configure llm`: configure OpenAI-compatible LLM providers (Ollama/OpenAI/Groq/Together/custom)
- `archway mcp serve`: expose tools/resources over stdio for AI clients

## Install

### From source

```bash
go install ./cmd/archway
```

### Homebrew (tap)

```bash
brew tap dcsg/tap
brew install archway
```

## Quick Start

```bash
# Scaffold a service
archway new --name orders --module github.com/acme/orders --template go-hexagonal --no-wizard

# Analyze current codebase
archway analyze --path .

# Initialize desired architecture config for existing project
archway init --language go --architecture hexagonal --no-wizard
```

## Command Reference

- `archway new [flags]`
- `archway init [flags]`
- `archway analyze [flags]`
- `archway configure llm [flags]`
- `archway mcp serve --transport stdio`
- `archway version`

## `archway.yaml` Example

```yaml
language: go
architecture: hexagonal
rules:
  dependencies:
    - layer: domain
      packages: ["domain/**"]
      may_depend_on: []
    - layer: ports
      packages: ["port/**"]
      may_depend_on: [domain]
    - layer: adapters
      packages: ["adapter/**"]
      may_depend_on: [ports, domain]
  structure:
    required_dirs: [cmd/, internal/domain/, internal/port/, internal/adapter/]
    forbidden_dirs: [utils/, helpers/]
  functions:
    max_lines: 80
    max_params: 4
extends:
  - archway/go-hexagonal-strict
templates:
  source: archway/go-hexagonal
```

## MCP Integration

See [docs/claude-code-config.md](docs/claude-code-config.md) and [docs/dof-integration.md](docs/dof-integration.md).

## Development

```bash
make test
make build
```

## Contributing

1. Fork the repo
2. Create a branch (`codex/<feature-name>`)
3. Add tests for behavior changes
4. Open a PR

## License

MIT
