---
title: archway.yaml
description: Project configuration reference
---

Every Archway project has an `archway.yaml` at its root. This file defines the architecture rules that `archway check` enforces.

## Full Example

```yaml
language: go
architecture: hexagonal

capabilities:
  - platform
  - bootstrap
  - http-api
  - mysql

components:
  - name: domain
    in: ["domain/**"]
    may_depend_on: []

  - name: ports
    in: ["port/**"]
    may_depend_on: [domain]

  - name: service
    in: ["service/**"]
    may_depend_on: [domain, ports]

  - name: adapters
    in: ["adapter/**"]
    may_depend_on: [ports, domain]

rules:
  max_function_lines: 80
  max_function_params: 5
  max_function_returns: 3
  required_dirs: []
  forbidden_dirs: []
```

## Fields

### `language`

The project's programming language. Currently only `go` is supported.

### `architecture`

The architecture pattern used to scaffold the project (`hexagonal`, `flat`).

### `capabilities`

List of capabilities that were composed into the project. Informational — used for documentation and future `archway update` support.

### `components`

The core of architecture enforcement. Each component defines:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Component identifier |
| `in` | string[] | Glob patterns matching this component's packages |
| `may_depend_on` | string[] | Other components this one is allowed to import from |

**Dependency rules are enforced transitively.** If `service` may depend on `domain` and `ports`, but not `adapters`, then any import from a `service/` package into an `adapter/` package is a violation.

### `rules`

Optional rules for function complexity and project structure:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_function_lines` | int | 80 | Maximum lines per function |
| `max_function_params` | int | 5 | Maximum parameters per function |
| `max_function_returns` | int | 3 | Maximum return values per function |
| `required_dirs` | string[] | [] | Directories that must exist |
| `forbidden_dirs` | string[] | [] | Directories that must not exist |

## Validation

The configuration is validated when running `archway check`:

- Component names must be unique
- Components cannot depend on themselves
- Dependencies must reference existing components
- `in` patterns must be valid globs
