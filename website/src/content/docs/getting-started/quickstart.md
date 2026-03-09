---
title: Quick Start
description: Create your first service in 2 minutes
---

import { Aside, Steps, Tabs, TabItem } from '@astrojs/starlight/components';

<Steps>

1. **Scaffold your service**

   The wizard guides you through every choice:

   ```bash
   archway new my-service
   ```

   You'll pick:
   - Service name and Go module path
   - Architecture — hexagonal or flat
   - Capabilities — multi-select what your service needs
   - Smart suggestions — accept or skip recommended additions

   <Aside type="tip">
   Know exactly what you want? Skip the wizard:

   ```bash
   archway new my-service \
     --arch hexagonal \
     --cap platform,bootstrap,http-api,mysql,docker \
     --module github.com/myorg/my-service \
     --no-wizard
   ```
   </Aside>

2. **Explore what you got**

   ```
   my-service/
   ├── cmd/my-service/main.go          # Thin entry point (15 lines)
   ├── internal/bootstrap/bootstrap.go # All dependency wiring
   ├── domain/                         # Business logic (zero deps)
   │   ├── errors.go                   # Typed domain errors
   │   └── clock.go                    # Testable time abstraction
   ├── port/                           # Interfaces
   │   ├── inbound.go                  # Use case interfaces
   │   └── outbound.go                 # Repository interfaces
   ├── service/                        # Use case implementations
   ├── adapter/                        # External integrations
   │   ├── httphandler/                # REST API (Chi router)
   │   └── mysqlrepo/                  # MySQL repositories
   ├── config/                         # YAML config loading
   ├── platform/                       # Logging, OTel, lifecycle
   ├── docs/PROJECT.md                 # Full project anatomy
   ├── archway.yaml                    # Architecture rules
   └── go.mod
   ```

3. **Run it**

   ```bash
   cd my-service
   go run ./cmd/my-service
   ```

4. **Validate the architecture**

   ```bash
   archway check
   ```

   This catches dependency violations, missing directories, function complexity issues, and anti-patterns — with **11 AST-based detectors**.

</Steps>

## What's Next

<Tabs>
  <TabItem label="Learn">

  - [How It Works](/archway/concepts/how-it-works/) — understand the composition model
  - [Architectures](/archway/concepts/architectures/) — hexagonal vs flat
  - [Capabilities](/archway/concepts/capabilities/) — all 38 capabilities explained

  </TabItem>
  <TabItem label="Build">

  - [Building a REST API](/archway/guides/rest-api/) — step-by-step guide
  - [Adding Capabilities](/archway/guides/adding-capabilities/) — extend an existing project

  </TabItem>
  <TabItem label="Reference">

  - [Capabilities Matrix](/archway/reference/capabilities-matrix/) — everything at a glance
  - [CLI Commands](/archway/reference/cli/) — full flag reference
  - [archway.yaml](/archway/reference/archway-yaml/) — configuration spec

  </TabItem>
</Tabs>
