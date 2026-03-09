---
title: Bootstrap Pattern
description: Testable dependency wiring with a thin main.go
---

import { Aside } from '@astrojs/starlight/components';

The bootstrap pattern separates your application's entry point from its dependency wiring, making the wiring testable and the `main.go` trivially simple.

## The Problem

In a typical Go service, `main.go` grows into a 200+ line function that creates database connections, initializes logging, sets up HTTP servers, registers shutdown hooks, and wires everything together. This code is:

- **Untestable** — you can't unit test `func main()`
- **Hard to read** — mixing infrastructure setup with business wiring
- **Error-prone** — shutdown order mistakes, missing cleanup

## The Solution

Split into two files:

### `cmd/<service>/main.go` — The entry point

```go
package main

import (
    "log/slog"
    "os"
    "github.com/org/my-service/internal/bootstrap"
)

var version = "dev"

func main() {
    if err := bootstrap.Run(version); err != nil {
        slog.Error("application error", "error", err)
        os.Exit(1)
    }
}
```

That's it. 15 lines. Never changes.

### `internal/bootstrap/bootstrap.go` — The wiring

```go
func Run(version string) error {
    ctx, stop := signal.NotifyContext(context.Background(),
        syscall.SIGTERM, syscall.SIGINT)
    defer stop()

    cfg, err := config.Load("config.yaml")
    // ... logger, OTel, database connections ...

    app := lifecycle.New(logger)
    app.Register("http", httpServer)
    app.OnShutdown("mysql", lifecycle.ShutdownFunc(db.Close))
    app.OnShutdown("otel", lifecycle.ShutdownFunc(otelShutdown))

    return app.Run(ctx)
}
```

This function is testable — you can call `Run()` with a test config and verify the wiring works.

## How Capabilities Wire In

Each capability provides **partials** — small code snippets that inject into `bootstrap.go` at specific points:

| Injection Point | What Goes Here |
|----------------|----------------|
| `main_imports` | Import paths for the capability's packages |
| `main_init` | Connection setup (database, cache, etc.) |
| `main_register` | Register servers/consumers with the lifecycle manager |
| `main_shutdown` | Cleanup hooks in reverse dependency order |

For example, when you add `mysql` and `http-api`, the bootstrap file automatically includes MySQL connection initialization, HTTP server registration, and ordered shutdown (HTTP drains first, then MySQL closes).

<Aside type="tip">
The bootstrap pattern is a **capability**, not an architecture concern. It works with any architecture — even flat.
</Aside>

## Why It Matters

| Without Bootstrap | With Bootstrap |
|------------------|----------------|
| Fat `main.go` (100-300 lines) | 15-line `main.go` |
| Untestable wiring | Testable `Run()` function |
| Manual shutdown ordering | Lifecycle manager handles it |
| Capabilities can't auto-wire | Partials inject automatically |
