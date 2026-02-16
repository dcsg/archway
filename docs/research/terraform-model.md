# Terraform Provider Architecture Research for Archway CLI

## Executive Summary

This research analyzes Terraform's provider architecture and similar plugin systems to provide practical implementation guidance for "archway," a code scaffolding and project governance CLI tool. The goal is to architect a language provider system that supports templates, brownfield analysis, and migration strategies across multiple languages.

## Search Summary

```json
{
  "search_summary": {
    "platforms_searched": ["github", "hashicorp docs", "developer blogs", "package registries"],
    "repositories_analyzed": 15,
    "docs_reviewed": 8
  }
}
```

---

## 1. Terraform's Provider Model

### 1.1 Core Architecture: Plugin Protocol

Terraform achieves complete decoupling between core and providers through a versioned gRPC-based plugin protocol.

**Key Architecture Decisions:**

- **Process Isolation**: Providers are separate executables launched as subprocesses
- **Communication**: gRPC over loopback interface (127.0.0.1)
- **Protocol**: Protocol Buffers for interface definitions
- **Handshake**: stdout-based discovery and port negotiation

**How It Works:**

1. Terraform Core launches provider binary as subprocess
2. Provider prints handshake information to stdout: `CORE-PROTOCOL-VERSION | APP-PROTOCOL-VERSION | NETWORK-TYPE | NETWORK-ADDR | PROTOCOL`
3. Terraform Core connects as gRPC client to indicated port
4. RPC calls flow over gRPC for all provider operations

**Protocol Versions:**

- Protocol v6 (current): Compatible with Terraform CLI 1.0+
- Supports multiple protocol versions simultaneously
- Negotiation during handshake selects mutually-supported version

### 1.2 Provider Discovery & Distribution

**Filesystem Discovery:**

Terraform searches for providers in this order:

1. **Local plugin directory**: `.terraform/plugins/`
2. **Current working directory**: `terraform.d/plugins/`
3. **User plugin directory**: Platform-specific (e.g., `~/.terraform.d/plugins/`)
4. **Filesystem mirrors**: Configured in CLI config file
5. **Remote registry**: As fallback for missing providers

**Binary Naming Convention:**

```
terraform-provider-<NAME>_v<VERSION>
terraform-provider-aws_v5.1.0
```

**Multi-platform Support:**

Providers organized by OS/arch:
```
registry.terraform.io/
  hashicorp/
    aws/
      5.1.0/
        darwin_amd64/
        darwin_arm64/
        linux_amd64/
        windows_amd64/
```

### 1.3 Registry Protocol

**Provider Requirements Format (HCL):**

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
```

**Registry Discovery Flow:**

1. Parse `source` into hostname/namespace/type (e.g., `registry.terraform.io/hashicorp/aws`)
2. Query registry API: `GET /v1/providers/{namespace}/{type}/versions`
3. Receive available versions list
4. Resolve version constraints using semantic versioning
5. Download platform-specific binary: `GET /v1/providers/{namespace}/{type}/{version}/download/{os}/{arch}`
6. Verify GPG signature
7. Cache locally in `.terraform/providers/`

**Registry Response Format (JSON):**

```json
{
  "versions": [
    {
      "version": "5.1.0",
      "protocols": ["5.0", "6.0"],
      "platforms": [
        {"os": "darwin", "arch": "amd64"},
        {"os": "linux", "arch": "amd64"}
      ]
    }
  ]
}
```

### 1.4 Versioning & Lock Files

**Lock File** (`.terraform.lock.hcl`):

```hcl
provider "registry.terraform.io/hashicorp/aws" {
  version     = "5.1.0"
  constraints = "~> 5.0"
  hashes = [
    "h1:abc123...",
    "zh:def456...",
  ]
}
```

Ensures deterministic provider versions across team members and CI/CD.

---

## 2. HashiCorp go-plugin Library

The foundation of Terraform's plugin system is the `github.com/hashicorp/go-plugin` library.

### 2.1 Core Components

**Handshake Configuration:**

```go
var handshakeConfig = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "ARCHWAY_PLUGIN",
    MagicCookieValue: "e3b0c44298fc1c14", // Random value for security
}
```

**Purpose:**
- **Magic Cookie**: Prevents accidental execution of wrong binaries
- **Protocol Version**: Enables version negotiation

**Plugin Interface Definition:**

```go
// Define the plugin interface
type LanguageProvider interface {
    Scaffold(project string, options map[string]interface{}) error
    Analyze(path string) (*AnalysisResult, error)
    Migrate(path string, strategy string) error
}

// Plugin map for registration
var pluginMap = map[string]plugin.Plugin{
    "provider": &LanguageProviderPlugin{},
}
```

### 2.2 gRPC Implementation Pattern

**Server Side (Plugin):**

```go
package main

import (
    "github.com/hashicorp/go-plugin"
    "archway/proto"
)

type GoProvider struct{}

func (p *GoProvider) Scaffold(req *proto.ScaffoldRequest) (*proto.ScaffoldResponse, error) {
    // Template rendering logic
    return &proto.ScaffoldResponse{Success: true}, nil
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: handshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "provider": &ProviderGRPCPlugin{Impl: &GoProvider{}},
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

**Client Side (Host):**

```go
package main

import (
    "os/exec"
    "github.com/hashicorp/go-plugin"
)

func main() {
    client := plugin.NewClient(&plugin.ClientConfig{
        HandshakeConfig: handshakeConfig,
        Plugins:         pluginMap,
        Cmd:             exec.Command("./plugins/archway-provider-go"),
        AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
    })
    defer client.Kill()

    rpcClient, err := client.Client()
    if err != nil {
        panic(err)
    }

    raw, err := rpcClient.Dispense("provider")
    if err != nil {
        panic(err)
    }

    provider := raw.(LanguageProvider)
    provider.Scaffold("my-service", options)
}
```

### 2.3 Key Benefits

1. **Language Agnostic**: Plugins can be written in any language with gRPC support
2. **Process Isolation**: Plugin crashes don't crash the host
3. **Security**: Sandboxed execution
4. **Versioning**: Multiple plugin versions can coexist
5. **Cross-platform**: Works on Windows, Linux, macOS

### 2.4 Performance Considerations

**From Eli Bendersky's analysis:**

- **RPC Overhead**: 10-100x slower than direct function calls
- **Serialization Cost**: Protocol Buffer encoding/decoding
- **Network**: Even loopback has latency vs in-process

**When RPC is Worth It:**
- Plugin needs isolation (security, stability)
- Multi-language support required
- Plugin updates without recompiling host
- Long-running operations (overhead amortized)

---

## 3. Alternative Plugin Architectures

### 3.1 Yeoman Generators (npm-based)

**Architecture:**

```json
{
  "repositories": [
    {
      "citation": "[1] Yeoman. 'Yeoman Generator Documentation.' npm, 2025. https://yeoman.io/authoring/",
      "platform": "npm",
      "stats": {
        "package_count": "5000+ generators",
        "ecosystem": "JavaScript/Node.js"
      },
      "key_features": [
        "npm package distribution",
        "Convention-based directory structure",
        "Composability (generators can call generators)",
        "Yeoman Environment for plugin discovery",
        "Prompts and user interaction built-in"
      ],
      "architecture": "Dynamic require() loading of npm packages at runtime",
      "code_quality": {
        "testing": "good",
        "documentation": "excellent",
        "maintenance": "active"
      },
      "usage_example": "yo name - runs app/ generator from generator-name package",
      "limitations": [
        "Node.js only (single language)",
        "No type safety",
        "npm dependency hell",
        "Security concerns with arbitrary code execution"
      ]
    }
  ]
}
```

**Directory Structure:**

```
generator-name/
  app/               # Default generator (yo name)
    index.js
    templates/
  router/            # Sub-generator (yo name:router)
    index.js
    templates/
  package.json
```

**Generator Class:**

```javascript
const Generator = require('yeoman-generator');

module.exports = class extends Generator {
  async prompting() {
    this.answers = await this.prompt([{
      type: 'input',
      name: 'name',
      message: 'Project name?'
    }]);
  }

  writing() {
    this.fs.copyTpl(
      this.templatePath('index.js'),
      this.destinationPath(`${this.answers.name}/index.js`),
      { title: this.answers.name }
    );
  }
};
```

**Discovery Mechanism:**

1. Search npm registry for packages matching `generator-*`
2. Load via `require('generator-name/app')`
3. Instantiate and run generator methods

**Pros:**
- Simple distribution (npm publish)
- Large ecosystem
- Easy to write (just JavaScript)
- Built-in CLI (`yo`)

**Cons:**
- Limited to Node.js
- No sandboxing
- Version conflicts
- No compile-time type checking

### 3.2 Cookiecutter (Python/Jinja)

**Architecture:**

```json
{
  "repositories": [
    {
      "citation": "[2] Cookiecutter. 'Template Extensions Documentation.' Python, 2025. https://cookiecutter.readthedocs.io/",
      "platform": "github/pypi",
      "stats": {
        "stars": "22000+",
        "templates": "6000+"
      },
      "key_features": [
        "Git repository as template source",
        "Jinja2 templating engine",
        "Template inheritance (v2.2+)",
        "Custom Jinja extensions",
        "Hooks for pre/post generation",
        "JSON configuration (cookiecutter.json)"
      ],
      "architecture": "Git clone + Jinja2 rendering",
      "code_quality": {
        "testing": "comprehensive",
        "documentation": "excellent",
        "maintenance": "active"
      },
      "usage_example": "cookiecutter gh:user/template or cookiecutter /path/to/template",
      "limitations": [
        "No plugin system (templates are the unit)",
        "Limited conditional logic in templates",
        "Git-centric (hard to version template code separately)",
        "No brownfield analysis support"
      ]
    }
  ]
}
```

**Configuration Format:**

```json
{
  "project_name": "My Project",
  "project_slug": "{{ cookiecutter.project_name.lower().replace(' ', '_') }}",
  "language": ["Go", "Python", "Node"],
  "_extensions": ["jinja2_time.TimeExtension"],
  "_copy_without_render": ["*.binary"]
}
```

**Template Structure:**

```
cookiecutter-project/
  cookiecutter.json
  {{ cookiecutter.project_slug }}/
    README.md
    src/
      main.{{ cookiecutter.language }}.jinja
  hooks/
    pre_gen_project.py
    post_gen_project.py
```

**Custom Extensions:**

```python
# extensions/custom_filters.py
def snake_case(text):
    return text.lower().replace(' ', '_')

# Referenced in cookiecutter.json
{
  "_extensions": ["extensions.custom_filters"]
}
```

**Pros:**
- Simple mental model (git + templates)
- Wide language support (any text file)
- Mature ecosystem
- No runtime dependencies (just Python)

**Cons:**
- Templates only (no code analysis)
- Git-heavy workflow
- Limited extensibility
- No plugin versioning

### 3.3 Nx Plugins (Monorepo Tooling)

**Architecture:**

```json
{
  "repositories": [
    {
      "citation": "[3] Nx. 'Plugin Architecture Documentation.' Monorepo Platform, 2025. https://nx.dev/docs/plugin-registry",
      "platform": "npm",
      "stats": {
        "stars": "23000+",
        "official_plugins": "20+",
        "community_plugins": "100+"
      },
      "key_features": [
        "NPM packages with metadata",
        "Generators for code scaffolding",
        "Executors for task automation",
        "Migrations for codebase updates",
        "Workspace generators (internal conventions)",
        "Plugin composition"
      ],
      "architecture": "TypeScript-based plugin system with schema validation",
      "code_quality": {
        "testing": "excellent",
        "documentation": "excellent",
        "maintenance": "very active"
      },
      "usage_example": "nx generate @nx/react:component MyComponent",
      "limitations": [
        "Tied to Nx workspace concept",
        "JavaScript/TypeScript focused",
        "Complex for simple use cases"
      ]
    }
  ]
}
```

**Plugin Structure:**

```
nx-plugin-myframework/
  package.json
  generators/           # Code scaffolding
    application/
      schema.json       # Generator options (validated)
      generator.ts      # Implementation
      files/            # Templates
  executors/            # Build/test/deploy tasks
    build/
      schema.json
      executor.ts
  migrations/           # Update scripts
    update-1-0-0/
      migration.ts
```

**Generator Implementation:**

```typescript
import { Tree, formatFiles, generateFiles } from '@nx/devkit';
import * as path from 'path';

interface Schema {
  name: string;
  directory?: string;
}

export default async function (tree: Tree, schema: Schema) {
  const normalizedOptions = normalizeOptions(tree, schema);

  generateFiles(
    tree,
    path.join(__dirname, 'files'),
    normalizedOptions.projectRoot,
    normalizedOptions
  );

  await formatFiles(tree);
}
```

**Schema Validation:**

```json
{
  "$schema": "http://json-schema.org/schema",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "Component name",
      "$default": { "$source": "argv", "index": 0 }
    },
    "directory": {
      "type": "string",
      "description": "Directory to generate into"
    }
  },
  "required": ["name"]
}
```

**Pros:**
- Strong typing (TypeScript)
- Schema validation
- Excellent composition
- Migration support
- Rich ecosystem

**Cons:**
- Node.js ecosystem only
- Heavy dependency tree
- Monorepo-centric design

### 3.4 Buf Plugins (Protobuf Ecosystem)

**Architecture:**

```json
{
  "repositories": [
    {
      "citation": "[4] Buf. 'Remote Plugins Documentation.' Buf Schema Registry, 2025. https://buf.build/docs/bsr/remote-plugins/",
      "platform": "buf.build",
      "stats": {
        "plugins": "50+",
        "focus": "Protobuf code generation"
      },
      "key_features": [
        "Remote plugin execution (BSR-hosted)",
        "Local plugin support",
        "Plugin versioning with @ syntax",
        "Docker-based isolation",
        "Language-agnostic (protoc plugins)",
        "Centralized plugin registry"
      ],
      "architecture": "Plugin registry + Docker + protoc plugin protocol",
      "code_quality": {
        "testing": "good",
        "documentation": "excellent",
        "maintenance": "very active"
      },
      "usage_example": "buf generate --template buf.gen.yaml",
      "limitations": [
        "Specific to Protobuf ecosystem",
        "Requires Docker for remote plugins",
        "Registry dependency"
      ]
    }
  ]
}
```

**Configuration (buf.gen.yaml):**

```yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go:v1.31.0
    out: gen/go
    opt: paths=source_relative
  - local: protoc-gen-custom
    out: gen/custom
```

**Plugin Discovery:**

1. Parse `buf.gen.yaml`
2. For remote plugins: query `registry.go-semantic-release.xyz` or `buf.build/plugins`
3. Download plugin binary or Docker image
4. Execute with protoc protocol

**Pros:**
- Centralized registry
- Version pinning
- Remote execution (no local install)
- Language-agnostic

**Cons:**
- Domain-specific (Protobuf)
- Network dependency
- Docker overhead

---

## 4. Go Plugin Systems: Detailed Comparison

### 4.1 Native Go Plugins (.so)

**Implementation:**

```go
// plugin/myplugin.go
package main

import "fmt"

func Scaffold(name string) error {
    fmt.Printf("Scaffolding %s\n", name)
    return nil
}

// Required: no main() in plugin mode
```

**Build:**

```bash
go build -buildmode=plugin -o myplugin.so plugin/myplugin.go
```

**Load:**

```go
package main

import (
    "plugin"
    "fmt"
)

func main() {
    p, err := plugin.Open("myplugin.so")
    if err != nil {
        panic(err)
    }

    scaffold, err := p.Lookup("Scaffold")
    if err != nil {
        panic(err)
    }

    scaffoldFunc := scaffold.(func(string) error)
    scaffoldFunc("my-project")
}
```

**Technical Reality:**

```json
{
  "technical_insights": {
    "common_patterns": [
      "Used sparingly in production due to limitations",
      "Best for controlled environments (same Go version)",
      "Requires exact package version matching"
    ],
    "best_practices": [
      "Define clear plugin interfaces",
      "Version binaries with Go version metadata",
      "Test on target platforms"
    ],
    "pitfalls": [
      "ABI instability across Go versions",
      "Platform limitations (no Windows)",
      "Dependency version conflicts",
      "Symbol resolution errors",
      "Debugging difficulties"
    ]
  }
}
```

**Verdict:**
- **Performance**: Excellent (in-process calls)
- **Portability**: Poor (Linux/macOS only, version-sensitive)
- **Ease of Use**: Poor (fragile)
- **Production Readiness**: Low

**Recommendation**: Avoid for archway unless you control entire deployment.

### 4.2 gRPC Subprocess (hashicorp/go-plugin)

**Full Implementation Example:**

**Protocol Definition (proto/provider.proto):**

```protobuf
syntax = "proto3";

package provider;
option go_package = "archway/proto";

service LanguageProvider {
  rpc Scaffold(ScaffoldRequest) returns (ScaffoldResponse);
  rpc Analyze(AnalyzeRequest) returns (AnalyzeResponse);
  rpc Migrate(MigrateRequest) returns (MigrateResponse);
}

message ScaffoldRequest {
  string project_name = 1;
  string template = 2;
  map<string, string> options = 3;
}

message ScaffoldResponse {
  bool success = 1;
  repeated string files_created = 2;
  string error = 3;
}

message AnalyzeRequest {
  string path = 1;
}

message AnalyzeResponse {
  string language = 1;
  string framework = 2;
  repeated string conventions_violated = 3;
  map<string, string> metrics = 4;
}

message MigrateRequest {
  string path = 1;
  string strategy = 2;
}

message MigrateResponse {
  bool success = 1;
  repeated string changes = 2;
}
```

**Shared Interface (shared/interface.go):**

```go
package shared

import (
    "context"
    "github.com/hashicorp/go-plugin"
    "archway/proto"
    "google.golang.org/grpc"
)

var Handshake = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "ARCHWAY_PLUGIN",
    MagicCookieValue: "7d8a912e4c6b3f1a",
}

// LanguageProvider is the interface implemented by plugins
type LanguageProvider interface {
    Scaffold(ctx context.Context, req *proto.ScaffoldRequest) (*proto.ScaffoldResponse, error)
    Analyze(ctx context.Context, req *proto.AnalyzeRequest) (*proto.AnalyzeResponse, error)
    Migrate(ctx context.Context, req *proto.MigrateRequest) (*proto.MigrateResponse, error)
}

// Plugin is the implementation of plugin.GRPCPlugin
type ProviderPlugin struct {
    plugin.Plugin
    Impl LanguageProvider
}

func (p *ProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
    proto.RegisterLanguageProviderServer(s, &GRPCServer{Impl: p.Impl})
    return nil
}

func (p *ProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
    return &GRPCClient{client: proto.NewLanguageProviderClient(c)}, nil
}

// GRPCClient is the client side
type GRPCClient struct {
    client proto.LanguageProviderClient
}

func (m *GRPCClient) Scaffold(ctx context.Context, req *proto.ScaffoldRequest) (*proto.ScaffoldResponse, error) {
    return m.client.Scaffold(ctx, req)
}

func (m *GRPCClient) Analyze(ctx context.Context, req *proto.AnalyzeRequest) (*proto.AnalyzeResponse, error) {
    return m.client.Analyze(ctx, req)
}

func (m *GRPCClient) Migrate(ctx context.Context, req *proto.MigrateRequest) (*proto.MigrateResponse, error) {
    return m.client.Migrate(ctx, req)
}

// GRPCServer is the server side
type GRPCServer struct {
    proto.UnimplementedLanguageProviderServer
    Impl LanguageProvider
}

func (m *GRPCServer) Scaffold(ctx context.Context, req *proto.ScaffoldRequest) (*proto.ScaffoldResponse, error) {
    return m.Impl.Scaffold(ctx, req)
}

func (m *GRPCServer) Analyze(ctx context.Context, req *proto.AnalyzeRequest) (*proto.AnalyzeResponse, error) {
    return m.Impl.Analyze(ctx, req)
}

func (m *GRPCServer) Migrate(ctx context.Context, req *proto.MigrateRequest) (*proto.MigrateResponse, error) {
    return m.Impl.Migrate(ctx, req)
}
```

**Plugin Implementation (plugins/go-provider/main.go):**

```go
package main

import (
    "context"
    "github.com/hashicorp/go-plugin"
    "archway/proto"
    "archway/shared"
)

type GoProvider struct{}

func (p *GoProvider) Scaffold(ctx context.Context, req *proto.ScaffoldRequest) (*proto.ScaffoldResponse, error) {
    // Template rendering logic
    // embed.FS for templates, text/template for rendering
    files := []string{
        req.ProjectName + "/go.mod",
        req.ProjectName + "/main.go",
        req.ProjectName + "/README.md",
    }

    return &proto.ScaffoldResponse{
        Success:      true,
        FilesCreated: files,
    }, nil
}

func (p *GoProvider) Analyze(ctx context.Context, req *proto.AnalyzeRequest) (*proto.AnalyzeResponse, error) {
    // AST parsing with go/ast, go/parser
    // Check for conventions:
    // - Project structure (cmd/, internal/, pkg/)
    // - go.mod presence
    // - Linting with golangci-lint rules

    return &proto.AnalyzeResponse{
        Language:  "Go",
        Framework: "standard library",
        Metrics: map[string]string{
            "go_version": "1.21",
            "modules":    "enabled",
        },
    }, nil
}

func (p *GoProvider) Migrate(ctx context.Context, req *proto.MigrateRequest) (*proto.MigrateResponse, error) {
    // Code transformation with go/ast printer
    // Refactor patterns

    return &proto.MigrateResponse{
        Success: true,
        Changes: []string{"Added internal/ structure", "Updated go.mod"},
    }, nil
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: shared.Handshake,
        Plugins: map[string]plugin.Plugin{
            "provider": &shared.ProviderPlugin{Impl: &GoProvider{}},
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

**Host CLI (cmd/archway/main.go):**

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "github.com/hashicorp/go-plugin"
    "archway/proto"
    "archway/shared"
)

func main() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: archway <language> <command>")
        os.Exit(1)
    }

    language := os.Args[1]
    command := os.Args[2]

    // Discover plugin binary
    pluginPath := findPlugin(language)
    if pluginPath == "" {
        fmt.Printf("Plugin for %s not found\n", language)
        os.Exit(1)
    }

    // Launch plugin
    client := plugin.NewClient(&plugin.ClientConfig{
        HandshakeConfig:  shared.Handshake,
        Plugins:          map[string]plugin.Plugin{"provider": &shared.ProviderPlugin{}},
        Cmd:              exec.Command(pluginPath),
        AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
    })
    defer client.Kill()

    rpcClient, err := client.Client()
    if err != nil {
        panic(err)
    }

    raw, err := rpcClient.Dispense("provider")
    if err != nil {
        panic(err)
    }

    provider := raw.(shared.LanguageProvider)

    // Execute command
    switch command {
    case "scaffold":
        resp, err := provider.Scaffold(context.Background(), &proto.ScaffoldRequest{
            ProjectName: "my-service",
            Template:    "microservice",
            Options:     map[string]string{"framework": "chi"},
        })
        if err != nil {
            panic(err)
        }
        fmt.Printf("Created: %v\n", resp.FilesCreated)

    case "analyze":
        resp, err := provider.Analyze(context.Background(), &proto.AnalyzeRequest{
            Path: ".",
        })
        if err != nil {
            panic(err)
        }
        fmt.Printf("Language: %s, Framework: %s\n", resp.Language, resp.Framework)
    }
}

func findPlugin(language string) string {
    // Search order:
    // 1. .archway/plugins/
    // 2. ~/.archway/plugins/
    // 3. /usr/local/lib/archway/plugins/

    pluginName := fmt.Sprintf("archway-provider-%s", language)

    searchPaths := []string{
        filepath.Join(".archway", "plugins", pluginName),
        filepath.Join(os.Getenv("HOME"), ".archway", "plugins", pluginName),
        filepath.Join("/usr/local/lib/archway/plugins", pluginName),
    }

    for _, path := range searchPaths {
        if _, err := os.Stat(path); err == nil {
            return path
        }
    }

    return ""
}
```

**Verdict:**
- **Performance**: Good (RPC overhead acceptable for code generation)
- **Portability**: Excellent (works everywhere)
- **Ease of Use**: Good (well-documented)
- **Production Readiness**: High (proven at scale)

**Recommendation**: Strong choice for archway.

### 4.3 WebAssembly (WASM)

**Implementation (go-plugin WASM variant):**

```go
// plugin/wasm/main.go
package main

import (
    "github.com/knqyf263/go-plugin/types/known/emptypb"
    "archway/proto"
)

//export scaffold
func scaffold(ptr, size uint32) uint64 {
    // Decode request from memory
    req := decodeScaffoldRequest(ptr, size)

    // Business logic
    resp := &proto.ScaffoldResponse{
        Success:      true,
        FilesCreated: []string{"main.go"},
    }

    // Encode response to memory
    return encodeScaffoldResponse(resp)
}

func main() {}
```

**Host:**

```go
package main

import (
    "context"
    "github.com/tetratelabs/wazero"
    "archway/proto"
)

func main() {
    ctx := context.Background()

    // Create runtime
    r := wazero.NewRuntime(ctx)
    defer r.Close(ctx)

    // Load WASM module
    wasm, _ := os.ReadFile("plugin.wasm")
    mod, _ := r.Instantiate(ctx, wasm)

    // Call exported function
    scaffold := mod.ExportedFunction("scaffold")

    // Marshal request to memory
    reqBytes := marshal(req)
    ptr, _ := mod.Memory().Write(0, reqBytes)

    // Call
    results, _ := scaffold.Call(ctx, ptr, uint64(len(reqBytes)))

    // Read response from memory
    respPtr, respSize := results[0], results[1]
    respBytes := mod.Memory().Read(respPtr, respSize)
    resp := unmarshal(respBytes)
}
```

**Verdict:**
- **Performance**: Excellent (near-native, sandboxed)
- **Portability**: Excellent (WASM runs anywhere)
- **Ease of Use**: Poor (immature ecosystem, complex FFI)
- **Production Readiness**: Medium (emerging)

**Recommendation**: Interesting for future, but too bleeding-edge for MVP.

---

## 5. Mapping to Archway Requirements

### 5.1 Language Provider Interface

**Required Capabilities:**

```go
type LanguageProvider interface {
    // Greenfield: Create new projects from templates
    Scaffold(ctx context.Context, req *ScaffoldRequest) (*ScaffoldResponse, error)

    // Brownfield: Understand existing code
    Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error)

    // Brownfield: Transform existing code
    Migrate(ctx context.Context, req *MigrateRequest) (*MigrateResponse, error)

    // Metadata
    GetInfo(ctx context.Context) (*ProviderInfo, error)
}

type ProviderInfo struct {
    Name         string
    Version      string
    Language     string
    Frameworks   []string
    Templates    []TemplateInfo
    Analyzers    []AnalyzerInfo
    Migrations   []MigrationInfo
}

type TemplateInfo struct {
    Name        string
    Description string
    Variables   []VariableInfo
}

type VariableInfo struct {
    Name        string
    Type        string // string, bool, choice
    Description string
    Default     string
    Required    bool
    Choices     []string // for type=choice
}
```

### 5.2 Provider Implementations by Language

**Go Provider:**

```go
type GoProvider struct {
    templates embed.FS // Embedded templates
}

func (p *GoProvider) Scaffold(ctx context.Context, req *ScaffoldRequest) (*ScaffoldResponse, error) {
    // Templates:
    // - microservice (standard-lib HTTP)
    // - grpc-service (gRPC + Buf)
    // - cli (Cobra)
    // - library (pkg structure)

    tmpl := req.Template // e.g., "microservice"
    vars := req.Options  // e.g., {"framework": "chi", "db": "postgres"}

    // Use text/template or html/template
    // Create file tree based on template

    return &ScaffoldResponse{
        Success:      true,
        FilesCreated: files,
    }, nil
}

func (p *GoProvider) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
    // Use go/ast, go/parser, go/types
    // Check for:
    // - Project layout (standard vs domain-driven)
    // - go.mod structure
    // - Test coverage
    // - Linting issues (golangci-lint)
    // - Dependency management

    violations := []string{}

    // Example checks:
    if !hasGoMod(req.Path) {
        violations = append(violations, "Missing go.mod")
    }

    if !hasCorrectLayout(req.Path) {
        violations = append(violations, "Non-standard project layout")
    }

    return &AnalyzeResponse{
        Language:             "Go",
        Framework:            detectFramework(req.Path),
        ConventionsViolated: violations,
        Metrics:             collectMetrics(req.Path),
    }, nil
}

func (p *GoProvider) Migrate(ctx context.Context, req *MigrateRequest) (*MigrateResponse, error) {
    // Strategies:
    // - adopt-standard-layout: Reorganize into cmd/, internal/, pkg/
    // - add-observability: Inject OpenTelemetry
    // - modernize-deps: Update go.mod, fix breaking changes

    strategy := req.Strategy

    switch strategy {
    case "adopt-standard-layout":
        return adoptStandardLayout(req.Path)
    case "add-observability":
        return addObservability(req.Path)
    }

    return &MigrateResponse{Success: false}, nil
}
```

**PHP Provider:**

```go
type PHPProvider struct {
    templates embed.FS
}

func (p *PHPProvider) Scaffold(ctx context.Context, req *ScaffoldRequest) (*ScaffoldResponse, error) {
    // Templates:
    // - symfony-api
    // - laravel-app
    // - slim-microservice

    framework := req.Options["framework"] // symfony, laravel, slim

    // Generate composer.json
    // Generate directory structure
    // Generate framework-specific files

    return &ScaffoldResponse{Success: true}, nil
}

func (p *PHPProvider) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
    // Use nikic/php-parser (call via exec or embed interpreter)
    // Or: exec PHP script that uses php-parser library

    // Detect framework from:
    // - composer.json dependencies
    // - Directory structure
    // - Config files (config/app.php for Laravel)

    framework := "unknown"
    if hasFile(req.Path, "artisan") {
        framework = "laravel"
    } else if hasFile(req.Path, "symfony.lock") {
        framework = "symfony"
    }

    return &AnalyzeResponse{
        Language:  "PHP",
        Framework: framework,
    }, nil
}
```

### 5.3 Template Management

**Option 1: Embedded Templates (Simple)**

```go
import "embed"

//go:embed templates/*
var templates embed.FS

func (p *GoProvider) Scaffold(ctx context.Context, req *ScaffoldRequest) (*ScaffoldResponse, error) {
    tmplPath := fmt.Sprintf("templates/%s", req.Template)

    // Walk embedded FS
    fs.WalkDir(templates, tmplPath, func(path string, d fs.DirEntry, err error) error {
        if d.IsDir() {
            return nil
        }

        // Read template file
        content, _ := templates.ReadFile(path)

        // Render with text/template
        tmpl, _ := template.New("file").Parse(string(content))

        var buf bytes.Buffer
        tmpl.Execute(&buf, req.Options)

        // Write to destination
        destPath := strings.Replace(path, tmplPath, req.ProjectName, 1)
        os.WriteFile(destPath, buf.Bytes(), 0644)

        return nil
    })

    return &ScaffoldResponse{Success: true}, nil
}
```

**Option 2: Git-Based Templates (Flexible)**

```go
func (p *GoProvider) Scaffold(ctx context.Context, req *ScaffoldRequest) (*ScaffoldResponse, error) {
    // Clone template repository
    repoURL := fmt.Sprintf("https://github.com/archway-templates/go-%s", req.Template)

    tmpDir, _ := os.MkdirTemp("", "archway-template-")
    defer os.RemoveAll(tmpDir)

    cmd := exec.Command("git", "clone", "--depth=1", repoURL, tmpDir)
    cmd.Run()

    // Process cookiecutter.json or similar
    config := readTemplateConfig(tmpDir)

    // Render templates
    renderTemplates(tmpDir, req.ProjectName, req.Options)

    return &ScaffoldResponse{Success: true}, nil
}
```

### 5.4 Brownfield Analysis Architecture

**AST-Based Analysis (Go example):**

```go
import (
    "go/ast"
    "go/parser"
    "go/token"
    "golang.org/x/tools/go/packages"
)

func analyzeGoProject(path string) (*AnalyzeResponse, error) {
    // Load packages
    cfg := &packages.Config{
        Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes,
        Dir:  path,
    }

    pkgs, err := packages.Load(cfg, "./...")
    if err != nil {
        return nil, err
    }

    violations := []string{}
    metrics := map[string]string{}

    // Check project structure
    if !hasDirectory(path, "cmd") {
        violations = append(violations, "Missing cmd/ directory")
    }

    if !hasDirectory(path, "internal") {
        violations = append(violations, "Consider using internal/ for private packages")
    }

    // Analyze code
    totalFuncs := 0
    totalTests := 0

    for _, pkg := range pkgs {
        for _, file := range pkg.Syntax {
            ast.Inspect(file, func(n ast.Node) bool {
                switch x := n.(type) {
                case *ast.FuncDecl:
                    totalFuncs++
                    if strings.HasPrefix(x.Name.Name, "Test") {
                        totalTests++
                    }
                }
                return true
            })
        }
    }

    metrics["total_functions"] = strconv.Itoa(totalFuncs)
    metrics["total_tests"] = strconv.Itoa(totalTests)
    metrics["test_ratio"] = fmt.Sprintf("%.2f%%", float64(totalTests)/float64(totalFuncs)*100)

    return &AnalyzeResponse{
        Language:             "Go",
        ConventionsViolated: violations,
        Metrics:             metrics,
    }, nil
}
```

**Pattern Detection (Framework identification):**

```go
func detectGoFramework(path string) string {
    goModPath := filepath.Join(path, "go.mod")
    content, err := os.ReadFile(goModPath)
    if err != nil {
        return "unknown"
    }

    modContent := string(content)

    frameworks := map[string]string{
        "github.com/gin-gonic/gin":    "gin",
        "github.com/gofiber/fiber":    "fiber",
        "github.com/labstack/echo":    "echo",
        "github.com/go-chi/chi":       "chi",
        "google.golang.org/grpc":      "grpc",
    }

    for dep, framework := range frameworks {
        if strings.Contains(modContent, dep) {
            return framework
        }
    }

    return "standard-library"
}
```

---

## 6. Practical Recommendation for Archway

### 6.1 MVP Architecture

```json
{
  "implementation_recommendations": [
    {
      "scenario": "MVP (first 3 months)",
      "recommended_solution": "Embedded Go providers with hashicorp/go-plugin architecture",
      "rationale": "Start simple with Go-only providers compiled into binary, add plugin support later"
    },
    {
      "scenario": "Post-MVP (6-12 months)",
      "recommended_solution": "gRPC-based external plugins with GitHub release distribution",
      "rationale": "Enable community contributions and multi-language providers"
    },
    {
      "scenario": "Long-term (12+ months)",
      "recommended_solution": "Custom registry + remote plugins (Buf-style)",
      "rationale": "Centralized plugin management, version control, telemetry"
    }
  ]
}
```

### 6.2 Recommended Tech Stack

**MVP Stack:**

```go
// Dependencies
require (
    github.com/hashicorp/go-plugin v1.6.0
    google.golang.org/grpc v1.60.0
    google.golang.org/protobuf v1.32.0
    github.com/spf13/cobra v1.8.0
    github.com/Masterminds/semver/v3 v3.2.1

    // For Go provider
    golang.org/x/tools v0.16.0 // go/ast, go/packages

    // For template rendering
    // text/template (stdlib) is sufficient
)
```

**Directory Structure:**

```
archway/
  cmd/
    archway/
      main.go                 # CLI entry point

  internal/
    plugin/
      manager.go              # Plugin discovery & lifecycle
      registry.go             # Version resolution

  proto/
    provider/
      v1/
        provider.proto        # Plugin interface
        provider.pb.go
        provider_grpc.pb.go

  shared/
    interface.go              # Go interface + gRPC glue

  providers/
    go/
      main.go                 # Go provider implementation
      analyzer.go
      scaffolder.go
      migrator.go
      templates/              # Embedded templates
        microservice/
        cli/
        library/

    php/                      # Future
    node/                     # Future

  pkg/
    archway/
      client.go               # Public API for embedding

  Makefile
  buf.yaml                    # Protobuf management
  buf.gen.yaml
```

**Build Strategy (MVP):**

```makefile
# Makefile

.PHONY: build
build: proto
	# Build main CLI (embeds providers for MVP)
	go build -o bin/archway ./cmd/archway

.PHONY: build-plugins
build-plugins: proto
	# Build external plugins (post-MVP)
	go build -o bin/plugins/archway-provider-go ./providers/go
	go build -o bin/plugins/archway-provider-php ./providers/php

.PHONY: proto
proto:
	buf generate

.PHONY: install
install:
	cp bin/archway /usr/local/bin/
	mkdir -p ~/.archway/plugins
	cp bin/plugins/* ~/.archway/plugins/
```

### 6.3 MVP Implementation: Embedded Providers

**Rationale:**
- Faster development (no plugin discovery complexity)
- Single binary distribution
- Easier debugging
- Lower barrier for users

**Code:**

```go
// cmd/archway/main.go
package main

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "archway/providers/go"
    "archway/proto/provider/v1"
)

var providers = map[string]v1.LanguageProvider{
    "go":     &goprovider.Provider{},
    "php":    &phpprovider.Provider{},
    "node":   &nodeprovider.Provider{},
    "python": &pythonprovider.Provider{},
}

func main() {
    rootCmd := &cobra.Command{
        Use:   "archway",
        Short: "Code scaffolding and governance CLI",
    }

    scaffoldCmd := &cobra.Command{
        Use:   "scaffold <language> <template>",
        Short: "Create new project from template",
        Args:  cobra.ExactArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
            language := args[0]
            template := args[1]

            provider, ok := providers[language]
            if !ok {
                fmt.Printf("Unknown language: %s\n", language)
                return
            }

            name, _ := cmd.Flags().GetString("name")

            resp, err := provider.Scaffold(context.Background(), &v1.ScaffoldRequest{
                ProjectName: name,
                Template:    template,
                Options:     map[string]string{},
            })

            if err != nil {
                fmt.Printf("Error: %v\n", err)
                return
            }

            fmt.Printf("Created %d files\n", len(resp.FilesCreated))
        },
    }
    scaffoldCmd.Flags().String("name", "", "Project name")

    analyzeCmd := &cobra.Command{
        Use:   "analyze",
        Short: "Analyze existing project",
        Run: func(cmd *cobra.Command, args []string) {
            // Auto-detect language
            language := detectLanguage(".")

            provider := providers[language]
            resp, _ := provider.Analyze(context.Background(), &v1.AnalyzeRequest{
                Path: ".",
            })

            fmt.Printf("Language: %s\n", resp.Language)
            fmt.Printf("Framework: %s\n", resp.Framework)
            fmt.Printf("Violations: %v\n", resp.ConventionsViolated)
        },
    }

    rootCmd.AddCommand(scaffoldCmd, analyzeCmd)
    rootCmd.Execute()
}

func detectLanguage(path string) string {
    if fileExists(path, "go.mod") {
        return "go"
    }
    if fileExists(path, "package.json") {
        return "node"
    }
    if fileExists(path, "composer.json") {
        return "php"
    }
    if fileExists(path, "requirements.txt") || fileExists(path, "pyproject.toml") {
        return "python"
    }
    return "unknown"
}
```

### 6.4 Post-MVP: External Plugins

**Plugin Discovery:**

```go
// internal/plugin/manager.go
package plugin

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "github.com/hashicorp/go-plugin"
    "archway/shared"
)

type Manager struct {
    pluginDir string
}

func NewManager() *Manager {
    return &Manager{
        pluginDir: getPluginDir(),
    }
}

func getPluginDir() string {
    // Search order:
    // 1. ARCHWAY_PLUGIN_DIR env var
    // 2. ./.archway/plugins
    // 3. ~/.archway/plugins
    // 4. /usr/local/lib/archway/plugins

    if dir := os.Getenv("ARCHWAY_PLUGIN_DIR"); dir != "" {
        return dir
    }

    if _, err := os.Stat(".archway/plugins"); err == nil {
        return ".archway/plugins"
    }

    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".archway", "plugins")
}

func (m *Manager) LoadProvider(language string, version string) (shared.LanguageProvider, error) {
    // Find plugin binary
    pluginName := fmt.Sprintf("archway-provider-%s", language)
    if version != "" {
        pluginName = fmt.Sprintf("%s_%s", pluginName, version)
    }

    pluginPath := filepath.Join(m.pluginDir, pluginName)

    if _, err := os.Stat(pluginPath); err != nil {
        return nil, fmt.Errorf("plugin not found: %s", pluginPath)
    }

    // Launch plugin
    client := plugin.NewClient(&plugin.ClientConfig{
        HandshakeConfig:  shared.Handshake,
        Plugins:          map[string]plugin.Plugin{"provider": &shared.ProviderPlugin{}},
        Cmd:              exec.Command(pluginPath),
        AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
    })

    rpcClient, err := client.Client()
    if err != nil {
        return nil, err
    }

    raw, err := rpcClient.Dispense("provider")
    if err != nil {
        return nil, err
    }

    return raw.(shared.LanguageProvider), nil
}
```

**Version Resolution:**

```go
// internal/plugin/registry.go
package plugin

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/Masterminds/semver/v3"
)

type Registry struct {
    baseURL string
}

func NewRegistry(url string) *Registry {
    if url == "" {
        url = "https://registry.archway.dev"
    }
    return &Registry{baseURL: url}
}

func (r *Registry) ResolveVersion(provider string, constraint string) (string, error) {
    // Query registry API
    url := fmt.Sprintf("%s/v1/providers/%s/versions", r.baseURL, provider)

    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var data struct {
        Versions []string `json:"versions"`
    }

    json.NewDecoder(resp.Body).Decode(&data)

    // Parse constraint
    c, err := semver.NewConstraint(constraint)
    if err != nil {
        return "", err
    }

    // Find matching version
    var best *semver.Version
    for _, v := range data.Versions {
        ver, err := semver.NewVersion(v)
        if err != nil {
            continue
        }

        if c.Check(ver) {
            if best == nil || ver.GreaterThan(best) {
                best = ver
            }
        }
    }

    if best == nil {
        return "", fmt.Errorf("no version matches constraint: %s", constraint)
    }

    return best.String(), nil
}

func (r *Registry) DownloadProvider(provider string, version string, os string, arch string) (string, error) {
    // Download from registry or GitHub releases
    url := fmt.Sprintf("%s/v1/providers/%s/%s/download/%s/%s",
        r.baseURL, provider, version, os, arch)

    // Download to temp file
    tmpFile := fmt.Sprintf("/tmp/archway-provider-%s_%s", provider, version)

    // HTTP GET + save to file
    // TODO: verify checksum/signature

    return tmpFile, nil
}
```

### 6.5 Distribution Strategy

**GitHub Releases (MVP):**

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build
        run: |
          # Build for multiple platforms
          GOOS=linux GOARCH=amd64 go build -o archway-linux-amd64 ./cmd/archway
          GOOS=linux GOARCH=arm64 go build -o archway-linux-arm64 ./cmd/archway
          GOOS=darwin GOARCH=amd64 go build -o archway-darwin-amd64 ./cmd/archway
          GOOS=darwin GOARCH=arm64 go build -o archway-darwin-arm64 ./cmd/archway
          GOOS=windows GOARCH=amd64 go build -o archway-windows-amd64.exe ./cmd/archway

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            archway-*
          generate_release_notes: true
```

**Homebrew (macOS):**

```ruby
# Formula/archway.rb
class Archway < Formula
  desc "Code scaffolding and governance CLI"
  homepage "https://github.com/yourorg/archway"
  url "https://github.com/yourorg/archway/archive/v0.1.0.tar.gz"
  sha256 "abc123..."
  license "MIT"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "bin/archway"
  end

  test do
    assert_match "archway version", shell_output("#{bin}/archway version")
  end
end
```

**Install Script (Linux/macOS):**

```bash
#!/bin/bash
# install.sh

set -e

VERSION="${ARCHWAY_VERSION:-latest}"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

BINARY="archway-${OS}-${ARCH}"
URL="https://github.com/yourorg/archway/releases/download/${VERSION}/${BINARY}"

echo "Downloading archway ${VERSION} for ${OS}/${ARCH}..."
curl -L "$URL" -o /tmp/archway
chmod +x /tmp/archway
sudo mv /tmp/archway /usr/local/bin/archway

echo "✓ Installed to /usr/local/bin/archway"
archway version
```

---

## 7. Technical Insights & Best Practices

```json
{
  "technical_insights": {
    "common_patterns": [
      "Process isolation via gRPC subprocess (Terraform, Vault)",
      "Embedded templates with text/template or embed.FS (Go ecosystem)",
      "Git-based template repositories (Cookiecutter, go-scaffold)",
      "Registry-based plugin discovery (Terraform, Buf)",
      "Semantic versioning for all plugins (universal)",
      "Lock files for reproducibility (Terraform, npm, Cargo)"
    ],
    "best_practices": [
      "Start with embedded providers, extract to plugins later",
      "Use Protocol Buffers for stable plugin interfaces",
      "Version plugin protocol separately from plugins",
      "Provide both local and remote plugin sources",
      "Implement plugin timeouts and health checks",
      "Sign plugin binaries for security",
      "Cache downloaded plugins locally",
      "Support air-gapped environments (filesystem mirrors)",
      "Generate plugin SDKs for easy development",
      "Provide plugin scaffolding tools (meta-plugins)"
    ],
    "pitfalls": [
      "Avoid native Go plugins (.so) - too fragile",
      "Don't require Docker for basic operations",
      "Don't build registry before validating demand",
      "Don't over-engineer MVP - YAGNI",
      "Avoid tight coupling between host and plugin versions",
      "Don't ignore plugin security (arbitrary code execution)",
      "Don't forget Windows support (common oversight)",
      "Avoid complex template logic - keep Jinja simple"
    ],
    "emerging_trends": [
      "WASM for portable, sandboxed plugins (early stage)",
      "Remote plugin execution (Buf model)",
      "AI-assisted code generation (GitHub Copilot)",
      "Policy-as-code for governance (OPA/Rego integration)",
      "Language servers for IDE integration (LSP)"
    ]
  }
}
```

---

## 8. Complexity Comparison

| Approach | MVP Time | Flexibility | Ecosystem | Maintenance |
|----------|----------|-------------|-----------|-------------|
| **Embedded providers** | 2 weeks | Low | Single binary | Easy |
| **gRPC plugins (local)** | 1 month | High | Multi-binary | Medium |
| **gRPC plugins (registry)** | 3 months | Very High | Registry + CLI | High |
| **WASM plugins** | 4 months | Very High | Emerging | High |
| **Native .so plugins** | 1 month | Low | Fragile | Very High |

---

## 9. Final Recommendation

### Phase 1: MVP (Months 1-3)

**Architecture:**
- Embedded Go providers in single binary
- Direct function calls (no plugins yet)
- Templates in `embed.FS`
- Simple CLI with Cobra

**Implementation:**

```go
// Single binary, providers as packages
import (
    goprovider "archway/providers/go"
    phpprovider "archway/providers/php"
)

var providers = map[string]Provider{
    "go":  &goprovider.Provider{},
    "php": &phpprovider.Provider{},
}
```

**Deliverables:**
- `archway scaffold go microservice`
- `archway analyze` (basic AST analysis for Go)
- 3-5 Go templates
- GitHub releases for distribution
- Install script

**Why:**
- Fast to market
- Easy to debug
- Proven UX before investing in plugins
- Single binary = happy users

### Phase 2: Plugin Architecture (Months 4-6)

**Architecture:**
- Extract providers to `hashicorp/go-plugin` gRPC plugins
- Keep embedded providers as fallback
- Local plugin discovery from `~/.archway/plugins`
- GitHub releases for plugins

**Migration:**

```go
// Detect if plugin exists, else use embedded
provider, err := pluginManager.LoadProvider(lang, version)
if err != nil {
    // Fallback to embedded
    provider = embeddedProviders[lang]
}
```

**Deliverables:**
- Plugin SDK (protobuf + Go helpers)
- `archway plugin install <name>`
- Community contribution guide
- 2-3 community plugins (PHP, Node)

**Why:**
- Community can contribute
- Faster iteration on individual languages
- Plugins can be language-specific (PHP plugin in PHP)

### Phase 3: Registry (Months 7-12)

**Architecture:**
- Custom registry (simple HTTP API)
- Versioned plugins with semver constraints
- Lock file (`.archway.lock`)
- Remote plugin execution (optional)

**Registry API:**

```
GET /v1/providers
GET /v1/providers/:name/versions
GET /v1/providers/:name/:version/download/:os/:arch
```

**Deliverables:**
- Registry service (Go HTTP server)
- `archway.toml` for provider requirements
- `.archway.lock` generation
- Plugin marketplace website

**Why:**
- Centralized discoverability
- Version management
- Telemetry (usage stats)
- Potential monetization

---

## 10. Code Structure Recommendation

**Recommended Repository Layout:**

```
archway/
  cmd/
    archway/              # Main CLI
    registry/             # Registry server (Phase 3)

  internal/
    plugin/
      manager.go          # Plugin lifecycle
      loader.go           # Discovery & loading
      registry.go         # Version resolution

    template/
      renderer.go         # Template engine
      validator.go

    analyzer/
      detector.go         # Language detection
      runner.go           # Analysis orchestration

  proto/
    provider/v1/
      provider.proto
      provider.pb.go
      provider_grpc.pb.go

  shared/
    interface.go          # LanguageProvider interface
    grpc.go               # gRPC client/server glue

  providers/
    go/
      main.go
      scaffold.go
      analyze.go
      migrate.go
      templates/
        microservice/
          go.mod.tmpl
          main.go.tmpl
          README.md.tmpl

    php/
      # Similar structure

    node/
      # Similar structure

  pkg/
    archway/              # Public Go API
      client.go

  scripts/
    install.sh
    build.sh

  .github/
    workflows/
      release.yml
      test.yml

  docs/
    plugin-development.md
    template-syntax.md
    architecture.md

  Makefile
  go.mod
  buf.yaml
  buf.gen.yaml
  README.md
```

---

## Sources

[1] HashiCorp. "Terraform Plugin Protocol." Terraform Docs, 2025. https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol

[2] HashiCorp. "go-plugin: Golang plugin system over RPC." GitHub, 2025. https://github.com/hashicorp/go-plugin

[3] Bendersky, Eli. "RPC-based plugins in Go." Eli Bendersky's Website, 2023. https://eli.thegreenplace.net/2023/rpc-based-plugins-in-go/

[4] HashiCorp. "Provider Registry Protocol." Terraform Docs, 2025. https://developer.hashicorp.com/terraform/internals/provider-registry-protocol

[5] Yeoman. "Writing Your Own Yeoman Generator." Yeoman Docs, 2025. https://yeoman.io/authoring/

[6] Cookiecutter. "Template Extensions Documentation." Cookiecutter Docs, 2025. https://cookiecutter.readthedocs.io/en/stable/advanced/template_extensions.html

[7] Nx. "Plugin Architecture Documentation." Nx Docs, 2025. https://nx.dev/docs/plugin-registry

[8] Buf. "Remote Plugins Documentation." Buf Docs, 2025. https://buf.build/docs/bsr/remote-plugins/usage/

[9] knqyf263. "go-plugin: Go Plugin System over WebAssembly." GitHub, 2025. https://github.com/knqyf263/go-plugin

[10] Bendersky, Eli. "Plugins in Go." Eli Bendersky's Website, 2021. https://eli.thegreenplace.net/2021/plugins-in-go/

[11] HashiCorp. "terraform-provider-scaffolding-framework." GitHub, 2025. https://github.com/hashicorp/terraform-provider-scaffolding-framework

[12] Go Semantic Release. "Plugin Registry." GitHub, 2025. https://github.com/go-semantic-release/plugin-registry

[13] Various. "Writing multi-package analysis tools for Go." Eli Bendersky's Website, 2020. https://eli.thegreenplace.net/2020/writing-multi-package-analysis-tools-for-go/

[14] Medium. "Hashicorp Plugin System Design and Implementation." Medium, 2024. https://zerofruit-web3.medium.com/hashicorp-plugin-system-design-and-implementation-5f939f09e3b3

[15] Medium. "Plugins, Extensions, and WASM: Making Your Go App Extensible." Medium, 2024. https://medium.com/@hamlet_dev/plugins-extensions-and-wasm-making-your-go-app-extensible-without-regret-85d55c957b9c
