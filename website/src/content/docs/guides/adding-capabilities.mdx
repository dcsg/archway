---
title: Adding Capabilities
description: How to add new capabilities to an existing project
---

import { Aside, Steps } from '@astrojs/starlight/components';

Archway capabilities are designed to be composable at scaffold time. This guide explains how to understand and manually add capabilities to an existing project.

## At Scaffold Time

The easiest way to include capabilities is when creating your project:

```bash
archway new my-service \
  --arch hexagonal \
  --cap platform,bootstrap,http-api,mysql,redis,docker
```

Use the interactive wizard to explore all 38 capabilities with descriptions and smart suggestions:

```bash
archway new my-service
```

## Understanding Capability Structure

Each capability provides:

1. **Template files** — Go source files that get rendered into your project
2. **Partials** — Code snippets that wire into `bootstrap.go`
3. **Config fields** — Struct fields added to `config/config.go`
4. **Manifest** — Metadata about requirements and suggestions

### Example: What `mysql` Provides

```
capabilities/mysql/
├── capability.yaml              # name, requires, suggests
├── files/
│   └── adapter/mysqlrepo/
│       └── connection.go.tmpl   # Connection pooling setup
└── _partials/
    ├── main_imports.go.tmpl     # import "github.com/org/svc/adapter/mysqlrepo"
    ├── main_init.go.tmpl        # db, err := mysqlrepo.NewConnection(...)
    └── main_shutdown.go.tmpl    # app.OnShutdown("mysql", ...)
```

## Manually Adding a Capability

If you want to add a capability to an existing project:

<Steps>

1. **Add the source files**

   Look at the capability's `files/` directory for what to add. For example, adding Redis means creating `adapter/redisrepo/connection.go`.

2. **Update config**

   Add the capability's config struct and fields to `config/config.go`:

   ```go
   type RedisConfig struct {
       Addr     string `yaml:"addr"`
       Password string `yaml:"password"`
       DB       int    `yaml:"db"`
   }
   ```

3. **Wire into bootstrap**

   Add the initialization and shutdown code to `internal/bootstrap/bootstrap.go`:

   ```go
   // In imports
   import "github.com/myorg/my-service/adapter/redisrepo"

   // In Run(), after config loading
   redisClient, err := redisrepo.NewConnection(redisrepo.ConnectionConfig{
       Addr:     cfg.Redis.Addr,
       Password: cfg.Redis.Password,
       DB:       cfg.Redis.DB,
   }, logger)
   if err != nil {
       return fmt.Errorf("connect redis: %w", err)
   }

   // Register shutdown
   app.OnShutdown("redis", lifecycle.ShutdownFunc(func(ctx context.Context) error {
       return redisClient.Close()
   }))
   ```

4. **Update archway.yaml**

   Add the capability to your project's capability list:

   ```yaml
   capabilities:
     - platform
     - bootstrap
     - http-api
     - mysql
     - redis  # Added
   ```

5. **Update config.yaml.example**

   Add the new configuration section:

   ```yaml
   redis:
     addr: localhost:6379
     password: ""
     db: 0
   ```

</Steps>

## Checking Compatibility

Before adding a capability, check its `capability.yaml` for:

- **`requires`** — Other capabilities that must be present
- **`suggests`** — Recommended companions
- **`conflicts`** — Capabilities that can't coexist

```yaml
# Example: auth-jwt requires http-api
name: auth-jwt
requires:
  - http-api
suggests: []
conflicts: []
```

<Aside type="tip">
Use `archway check` after adding a capability to make sure your project still passes all architecture rules.
</Aside>
