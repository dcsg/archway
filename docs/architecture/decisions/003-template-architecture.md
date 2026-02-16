# ADR-003: Template Architecture — text/template + embed.FS + Manifest

**Status:** accepted
**Date:** 2026-02-14
**Deciders:** Daniel Gomes

## Context

Archway scaffolds new projects from templates. The template system must support 66+ Go templates covering hexagonal architecture, CQRS, multiple transports, data stores, auth, email, and observability. Templates must be extensible by users without Go knowledge.

## Constraints

- Templates must be embeddable in a single binary (NF-01)
- Non-Go developers must be able to create templates (Layer 1 extensibility)
- Templates need conditional logic (include/exclude files based on wizard answers)
- Must support a TUI wizard with dynamic questions per template
- Post-MVP: templates loadable from git repos and local directories

## Options Considered

### Option A: text/template + embed.FS + manifest.yaml + wizard.yaml — CHOSEN

Each template is a directory containing:
```
templates/go-hexagonal/
  manifest.yaml          # metadata: name, description, language, variables
  wizard.yaml            # TUI wizard definition: questions, conditions, defaults
  files/                 # template files with {{.Variable}} placeholders
    go.mod.tmpl
    cmd/main.go.tmpl
    internal/domain/...
```

**Pros:**
- Standard library (text/template) — no deps, best performance
- embed.FS for single binary distribution
- manifest.yaml + wizard.yaml are data, not code — anyone can author
- Familiar Go template syntax
- wizard.yaml decouples TUI logic from template content
- Clear separation: manifest (what), wizard (how to ask), files (content)

**Cons:**
- Go template syntax unfamiliar to non-Go developers (`{{if .X}}` vs `{% if x %}`)
- Limited template logic compared to full programming languages
- embed.FS loads all templates into memory (acceptable for text files)

### Option B: gonew-style (module-based)

Templates as Go modules, distributed via Go module proxy.

**Pros:**
- Leverages Go infrastructure (proxy, sumdb, versioning)
- Secure (checksum verified)

**Cons:**
- Templates must be valid Go modules
- No conditional file generation
- No wizard support
- Limited to Go-only templates

### Option C: Cookiecutter-style (Jinja2)

Use Jinja2-compatible templating (e.g., pongo2 library).

**Pros:**
- Familiar syntax for Python/Django developers
- Powerful template logic

**Cons:**
- External dependency (pongo2)
- ~2.5x slower than text/template for conditionals
- Different syntax from Go ecosystem conventions
- No embed.FS integration out of the box

### Option D: Code generation (no templates)

Generate code programmatically via Go AST construction.

**Pros:**
- Type-safe output
- Can validate generated code at generation time

**Cons:**
- Extremely verbose for 66+ templates
- Templates not authorable by non-Go developers
- Much harder to maintain
- No visual correspondence between template and output

## Decision

**Option A: text/template + embed.FS + manifest.yaml + wizard.yaml.** The three-file convention (manifest, wizard, files) makes templates authorable as pure data while keeping the engine simple. The standard library gives us the best performance and zero dependencies.

### Template Format

**manifest.yaml:**
```yaml
name: go-hexagonal
description: Go service with hexagonal architecture
language: go
version: "1.0"
variables:
  - name: ServiceName
    type: string
    required: true
  - name: Transport
    type: choice
    choices: [http, grpc, both]
    default: http
```

**wizard.yaml:**
```yaml
steps:
  - id: basics
    questions:
      - variable: ServiceName
        prompt: "Service name?"
        validate: "^[a-z][a-z0-9-]*$"
      - variable: ModulePath
        prompt: "Go module path?"
  - id: transport
    questions:
      - variable: Transport
        prompt: "Which transport?"
        type: select
        options: [HTTP, gRPC, Both]
```

## Consequences

- Template authors work with YAML + text files, not Go code
- wizard.yaml drives the Bubbletea/Huh TUI dynamically
- embed.FS embeds all official templates at compile time
- Post-MVP: template loader also reads from local dirs and git repos
- Template testing: render each template with default values, verify output compiles
