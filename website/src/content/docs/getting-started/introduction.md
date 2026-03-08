---
title: Introduction
description: What Archway is and why it exists
---

## The Problem

Starting a new Go service means making dozens of decisions upfront: project layout, error handling, logging, config, middleware, database setup, graceful shutdown. You either copy-paste from a previous project (carrying its baggage) or spend hours setting up boilerplate before writing a single line of business logic.

## The Solution

Archway is an **architecture-aware service composer**. Instead of monolithic templates, it composes your project from two building blocks:

1. **Architecture** — The structural pattern (hexagonal, flat) that defines your project layout and dependency rules
2. **Capabilities** — Modular features (HTTP API, MySQL, auth, etc.) that plug into your architecture

```
Architecture (hexagonal) + Capabilities (http-api, mysql, docker) = Your Service
```

### What Makes It Different

- **Composable, not monolithic** — Pick only what you need. No unused code, no dead imports.
- **Architecture enforcement** — `archway check` validates dependency rules at any time. Domain code importing adapter code? Caught.
- **Smart suggestions** — Adding an HTTP API? Archway suggests rate limiting and authentication.
- **Production patterns built in** — RFC 7807 errors, structured logging, PII redaction, graceful shutdown, OpenTelemetry — all wired correctly from day one.
- **Project anatomy** — Every scaffolded project gets a `docs/PROJECT.md` documenting exactly what patterns and tools are included.

## Who Is It For

- **Teams** starting new Go microservices who want consistent structure across projects
- **Individual developers** who want production-grade defaults without the setup time
- **Organizations** that want to enforce architectural standards across their Go services

## What It Is Not

- Not a framework — Archway generates plain Go code with no runtime dependency
- Not opinionated about your business logic — it sets up the infrastructure, you write the domain
- Not a code generator that you run repeatedly — scaffold once, own the code
