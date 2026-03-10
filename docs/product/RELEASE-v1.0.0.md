# Archway v1.0.0

**Terraform for Code Architecture** -- Archway composes production-ready services from architecture patterns and capability modules, then enforces the rules you set.

## Four Pillars

### Guide (`archway guide`)
AI-native architecture instructions. Generates context-aware guidance for AI coding agents based on your project's `archway.yaml`, so LLMs respect your architecture decisions.

### Compose (`archway new`)
Scaffold production-ready services from architecture patterns. Interactive wizard with 38 capabilities across 8 categories (transport, data, resilience, patterns, security, observability, infrastructure, quality). Smart suggestions flag missing capabilities and warn about problematic combinations.

### Analyze (`archway analyze`)
Detect architecture patterns in existing codebases. AST-based analysis identifies hexagonal and flat architectures, maps dependency graphs, and generates an `archway.yaml` from what it finds.

### Enforce (`archway check`)
Validate projects against `archway.yaml` rules. 11 built-in anti-pattern detectors catch architecture violations before they reach code review.

## Supported Architectures

- **Hexagonal** -- Ports and adapters with clear domain boundaries
- **Flat** -- Simple, pragmatic structure for smaller services

## Supported Languages

- **Go** -- Full template support with hexagonal and flat architectures

## Installation

### Homebrew

```bash
brew install dcsg/tap/archway
```

### Binary

Download from the [GitHub Releases](https://github.com/dcsg/archway/releases/tag/v1.0.0) page.

## Quick Start

```bash
# Scaffold a new service
archway new my-service

# Analyze an existing codebase
archway analyze ./my-project

# Check architecture rules
archway check ./my-project

# Generate AI agent instructions
archway guide ./my-project
```
