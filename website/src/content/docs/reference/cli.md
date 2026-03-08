---
title: CLI Commands
description: Complete CLI reference
---

## `archway new`

Scaffold a new project.

```bash
archway new [name] [flags]
```

### Flags

| Flag | Description | Default |
|------|------------|---------|
| `--name` | Project/service name | positional arg |
| `--arch` | Architecture pattern (`hexagonal`, `flat`) | `hexagonal` |
| `--cap` | Capabilities (comma-separated) | none |
| `--module` | Go module path | `example.com/<name>` |
| `--output-dir` | Output directory | `.` |
| `--no-wizard` | Disable interactive wizard | `false` |
| `--set` | Template variable (key=value), repeatable | none |
| `--language` | Project language | `go` |

### Examples

```bash
# Interactive wizard
archway new my-service

# Non-interactive with all options
archway new my-service \
  --arch hexagonal \
  --cap platform,bootstrap,http-api,mysql,docker \
  --module github.com/myorg/my-service \
  --no-wizard

# With custom variables
archway new my-service --set GoVersion=1.23 --no-wizard
```

## `archway check`

Validate a project against its `archway.yaml` rules.

```bash
archway check [flags]
```

### Flags

| Flag | Description | Default |
|------|------------|---------|
| `--path` | Path to project | `.` |
| `--output` | Output format (`terminal`, `json`) | `terminal` |

### What It Checks

- **Dependency violations** — imports that cross component boundaries
- **Required directories** — components with missing directories
- **Forbidden directories** — directories that shouldn't exist
- **Function complexity** — functions exceeding line/param/return limits
- **Component coverage** — percentage of components with source files

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks pass |
| 1 | One or more violations found |

### Example Output

```
archway check

Architecture: hexagonal
Components: 4 defined

Dependency Violations: 0
Structure Issues: 0
Function Issues: 0

Component Coverage: 100% (4/4)
Compliance: 100%

✓ All checks passed
```

## `archway analyze`

Analyze an existing Go project's architecture.

```bash
archway analyze [flags]
```

### Flags

| Flag | Description | Default |
|------|------------|---------|
| `--path` | Path to project | `.` |
| `--output` | Output format (`terminal`, `json`) | `terminal` |
