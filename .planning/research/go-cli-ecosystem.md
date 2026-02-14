# Go CLI Ecosystem Research for "archway"
## Project Governance & Architecture Enforcement Tool

**Research Date**: February 14, 2026
**Purpose**: Research existing Go CLI tools for building an open-source project governance tool

---

## Executive Summary

This research analyzes 20+ Go CLI tools across five categories: scaffolding, architecture enforcement, brownfield analysis, refactoring/migration, and plugin distribution. Key findings:

1. **Scaffolding Tools**: Most use `text/template` with `embed.FS`; `gonew` offers secure module-based distribution
2. **Architecture Enforcement**: Three distinct approaches exist (YAML config, fluent API, test-based)
3. **Static Analysis**: The `go/analysis` framework is the de facto standard; `inspector` package provides 2.5x speedup
4. **Refactoring**: `dst` (Decorated Syntax Tree) enables comment-preserving transformations
5. **Plugin Distribution**: Module-based approaches preferred over RPC/WASM for complexity and compatibility

---

## 1. Go Scaffolding CLIs

### 1.1 gonew (Official Experimental Tool)

**Repository**: golang.org/x/tools/cmd/gonew
**Status**: Experimental (subject to change)
**Adoption**: Backed by Go team, used by Google Cloud Platform

#### How It Works
```bash
gonew golang.org/x/example/helloserver example.com/myserver
```

- Downloads template module from Go module proxy
- Performs module path substitution (e.g., `golang.org/x/example/helloserver` → `example.com/myserver`)
- Writes new module to local directory
- Leverages checksum database for security

#### Template Distribution
- **Packaging**: Templates are standard Go modules
- **Distribution**: Via Go module proxy (secure, versioned, cached)
- **Discovery**: No official registry; templates shared via documentation/blogs
- **Security**: Checksum verification via sumdb

#### Architecture Decisions Worth Borrowing
✓ Module-based distribution (no custom hosting needed)
✓ Minimal tool design (let users extend via templates)
✓ Leverages existing Go infrastructure (proxy, sumdb)
✗ No template validation or quality checks
✗ Experimental status limits enterprise adoption

**Official Templates**:
- `golang.org/x/example/hello` - CLI tool
- `golang.org/x/example/helloserver` - HTTP server
- Google Cloud Platform: Cloud Functions, Pub/Sub, Cloud Run

---

### 1.2 go-blueprint

**Repository**: github.com/Melkeydev/go-blueprint
**Stars**: 8.6k | **Last Updated**: Jul 10, 2025 (v0.10.11) | **Contributors**: 80+

#### Features
- 6 web frameworks (Chi, Gin, Fiber, HttpRouter, Gorilla/mux, Echo)
- 6 database drivers (MySQL, Postgres, SQLite, MongoDB, Redis, ScyllaDB)
- Advanced features: HTMX/Templ, GitHub Actions, WebSockets, Tailwind CSS, Docker, React

#### Template Approach
- **Method**: Feature flags (`--advanced`, `--feature`) toggle template sections
- **Storage**: Not embedded; likely file-based or git-based templates
- **Language**: Go (92.1%), Shell (7.9%)

#### Distribution
```bash
# Multiple installation methods
go install github.com/melkeydev/go-blueprint@latest
npm install -g @melkeydev/go-blueprint
brew install go-blueprint
```

#### Architecture Insights
- CLI-first with non-interactive mode for CI/CD
- Optional "Blueprint UI" web app for visual preview
- GoReleaser for multi-platform distribution
- No formal plugin API; extensibility via CLI flags

**Lessons**:
✓ Multiple installation methods increase adoption
✓ Web UI lowers barrier for beginners
✓ Feature flags provide flexibility without plugin complexity
✗ Limited to predefined feature combinations

---

### 1.3 Autostrada

**Repository**: autostrada.dev (proprietary web service)
**Adoption**: Moderate (web-based generator)

#### Philosophy
> "Not a third-party framework that you import. The generated code IS your application."

#### Approach
- Web-based project generator (no CLI)
- Generates fully-functioning scaffolds with minimal abstractions
- Customizable: only includes needed features
- Idiomatic Go code with sensible defaults

#### Key Decisions
✓ No framework lock-in (generated code is standalone)
✓ Minimal dependencies → smaller binaries
✓ Includes task-specific helpers (JSON parsing, SQL migrations)
✗ Web-based (no offline usage)
✗ Proprietary (not inspectable/forkable)

---

### 1.4 Template Handling Comparison

| Tool | Template Engine | Storage | Distribution |
|------|----------------|---------|--------------|
| gonew | Module substitution | Go modules | Module proxy |
| go-blueprint | Unknown (likely `text/template`) | Local files | go install / npm / brew |
| Autostrada | Server-side (unknown) | Proprietary | Web UI |
| create-go-app | N/A (deprecated/unmaintained) | - | - |

---

### 1.5 Go Templating Libraries

#### text/template vs. Alternatives

**text/template** (Standard Library)
- **Pros**: Pre-bundled, well-optimized, performant
- **Cons**: Unfamiliar syntax for non-Go developers
- **Performance**: Best overall performance
- **Use case**: Code generation, config files

**pongo2** (Django-like syntax)
- **Repo**: github.com/flosch/pongo2
- **Pros**: Developer-friendly syntax, faster function calls (no reflection)
- **Cons**: Slower if/else logic, external dependency
- **Performance**: ~2.5x slower than `text/template` for conditionals
- **Use case**: HTML templates, developer-facing tools

**Performance Benchmark** (from goTemplateBenchmark):
```
text/template:  Best overall
pongo2:         Fast function calls, slow conditionals
quicktemplate:  Fastest (compile-time), but requires recompilation
liquid:         Slowest
```

**Recommendation for archway**:
- Use `text/template` for code generation (performance, no deps)
- Consider `pongo2` if targeting non-Go developers (friendlier syntax)
- Use `embed.FS` for template storage (single binary distribution)

---

### 1.6 embed Package Best Practices (2026)

**Usage Pattern**:
```go
//go:embed templates/*.tmpl
var templatesFS embed.FS

func main() {
    tmpl, _ := template.ParseFS(templatesFS, "templates/*.tmpl")
}
```

**Advantages**:
✓ Atomic updates (templates + binary version together)
✓ Single-file distribution (no installer needed)
✓ Works with `io/fs` interface (`http.FileServer`, `template.ParseFS`)
✓ No extraction step for users

**Considerations**:
- All files loaded into memory (avoid large files)
- Binary size increases (typically negligible for text templates)
- Use `fs.Sub()` for clean path prefixes
- Consider build tags for development mode (use local files)

**2026 Best Practices**:
- Embed configuration defaults (override at runtime)
- Use glob patterns: `templates/**/*.tmpl`
- Provide `-dev` flag to load from disk (faster iteration)

---

## 2. Architecture Enforcement Tools

### 2.1 go-arch-lint

**Repository**: github.com/fe3dback/go-arch-lint
**Stars**: 439 | **Last Updated**: Nov 13, 2025 (v1.14.0) | **Go Version**: 1.25+

#### How It Works

**Step 1: Define Architecture (YAML)**
```yaml
version: 2
workdir: .

components:
  - name: handlers
    package: app/handlers/**
    may-depends-on:
      - services
      - models

  - name: services
    package: app/services/**
    may-depends-on:
      - repositories
      - models

  - name: repositories
    package: app/repositories/**
    may-depends-on:
      - models

  - name: models
    package: app/models/**

common-components:
  - vendor/**
  - pkg/utils/**
```

**Step 2: Validate**
```bash
go-arch-lint check
# Exit code 0: valid
# Exit code 1: violations found
```

#### Validation Mechanism
1. **Package Mapping**: Maps Go packages to defined components (wildcard matching)
2. **Import Analysis**: Analyzes actual import dependencies
3. **Graph Construction**: Builds dependency graph
4. **Comparison**: Validates actual vs. desired dependencies

#### Features
- **Wildcard Patterns**: `*` (single-level), `**` (multi-level)
- **Todo Markers**: Temporary violation legalization during refactoring
- **JSON Output**: `--output-type json` for CI/CD integration
- **Docker Support**: Precompiled binaries available

#### Supported Patterns
- Hexagonal / Onion Architecture
- Domain-Driven Design (DDD)
- MVC / MVVM
- Clean Architecture
- Custom patterns

**Lessons for archway**:
✓ YAML configuration is approachable for most developers
✓ Todo markers enable gradual migration (brownfield-friendly)
✓ JSON output enables CI/CD integration
✓ Wildcard patterns reduce config verbosity
✗ No detection of existing architecture (requires manual definition)

---

### 2.2 archunit (Go port of Java ArchUnit)

**Repository**: github.com/kcmvp/archunit
**Stars**: 22 | **Last Updated**: Jan 11, 2025 | **License**: Apache-2.0

#### API Design (Fluent Interface)

```go
// Define rules
Packages(HaveNameSuffix[Package]("service")).
    ShouldNotRefer(Layers("Repository"))

Types(Implement[Type]("Handler")).
    ShouldBe(Public[Type]())

Functions(HaveNamePrefix[Function]("New")).
    ShouldNotAccept(SimpleTypes[Function]())
```

#### Go Code Modeling

Parses Go code into 6 architectural objects:
- **Layer** (e.g., "Repository", "Service")
- **Package** (Go packages)
- **Type** (structs, interfaces)
- **Function** (functions)
- **Variable** (variables)
- **File** (source files)

Categorized by **Pointcut Interfaces**:
- **Referable**: Layer, Package, Type (enables dependency rules)
- **Exportable**: Type, Function, Variable (enables visibility rules)

#### Key Design Principles

1. **Functional Approach**: Rules are first-class citizens
2. **Type Safety**: Generics prevent invalid rule applications at compile time
3. **Declarative**: Express "what" not "how"
4. **AI-Friendly**: Structured output for AI feedback loops

**Comparison to Java ArchUnit**:
- Similar fluent API design
- Adapted for Go's package system (vs. Java's class hierarchy)
- Smaller community (22 stars vs. Java's 3.2k stars)

**Lessons for archway**:
✓ Fluent API is readable and composable
✓ Type safety prevents config errors
✓ Test-based approach integrates with existing workflows
✗ Small community/adoption
✗ Requires Go knowledge (not accessible to architects unfamiliar with Go)

---

### 2.3 arch-go

**Repository**: github.com/arch-go/arch-go
**Stars**: 245 | **Last Updated**: Feb 3, 2026 (v2.1.2) | **License**: MIT

#### Rule Categories

**1. Dependencies**
```yaml
dependencies_rules:
  - package: "**.impl.**"
    should-not-depend-on:
      - "**.external.**"
```

**2. Package Contents**
```yaml
contents_rules:
  - package: "**.domain.**"
    should-only-contain:
      - interfaces
      - structs
```

**3. Functions**
```yaml
functions_rules:
  - package: "**.handlers.**"
    max-parameters: 3
    max-return-values: 2
    max-lines: 50
    max-public-functions-per-file: 1
```

**4. Naming Conventions**
```yaml
naming_rules:
  - package: "**.repositories.**"
    interfaces-implementing: "Repository"
    should-have-simple-name-ending-with: "Repo"
```

#### Configuration (`arch-go.yml`)
```yaml
version: 1
threshold:
  compliance: 100  # Minimum % of rules a module must satisfy
  coverage: 100    # Minimum % of packages evaluated by at least one rule
```

#### Usage Modes

**CLI**:
```bash
arch-go [flags]
# --verbose: detailed output
# --html: HTML report
# --json: JSON output
```

**Programmatic** (Test Integration):
```go
import "github.com/arch-go/arch-go/api/configuration"

func TestArchitecture(t *testing.T) {
    // Load and validate architecture rules
}
```

#### Distinguishing Features
- Validates **structural patterns** (not just dependencies)
- Enforces **content rules** (what types can exist in packages)
- **Function complexity** limits (parameters, returns, lines)
- **Naming conventions** enforcement

**Comparison to go-arch-lint**:
| Feature | go-arch-lint | arch-go |
|---------|-------------|---------|
| Dependency rules | ✓ | ✓ |
| Content rules | ✗ | ✓ |
| Function limits | ✗ | ✓ |
| Naming conventions | ✗ | ✓ |
| Adoption | 439 stars | 245 stars |

**Lessons for archway**:
✓ Multiple rule types provide comprehensive governance
✓ Thresholds enable gradual adoption (< 100% compliance)
✓ Programmatic API enables test integration
✓ HTML reports aid communication with stakeholders

---

### 2.4 Architecture Enforcement: Three Approaches

| Approach | Example | Pros | Cons |
|----------|---------|------|------|
| **YAML Config** | go-arch-lint, arch-go | Approachable, version-controlled, non-Go users | Less type-safe, limited expressiveness |
| **Fluent API** | archunit | Type-safe, composable, powerful | Requires Go knowledge, steeper learning curve |
| **Test-Based** | archunit (in tests) | Integrates with CI, fails on violation | Requires test runner, less discoverable |

**Recommendation for archway**:
- **Primary**: YAML config (approachable for architects)
- **Advanced**: Optional fluent API for complex rules
- **Integration**: Both should generate test files for CI enforcement

---

## 3. Brownfield Analysis Tools

### 3.1 goda (Go Dependency Analysis)

**Repository**: github.com/loov/goda
**Stars**: 1.6k | **Last Updated**: May 23, 2025 | **Language**: 100% Go

#### Query Syntax

**Package Selection**:
```bash
# All packages in current module
goda tree ./...

# Exclude vendor
goda tree "./...:all - golang.org/x/tools/..."

# Only test dependencies
goda tree "./...:import(test)"
```

**Modifiers**:
- `...` - Wildcard matching
- `:all` - All dependencies (transitive)
- `:import` - Direct imports only
- `:deps` - Dependencies
- `:reach` - Reachability analysis
- `:shared` - Shared dependencies

**Filters**:
```bash
# OS-specific dependencies
goda tree "goos=windows(./...)"

# Purego dependencies
goda tree "purego=1(./...)"
```

#### Capabilities

**Graph Visualization** (requires GraphViz):
```bash
goda graph "./..." | dot -Tpng -o graph.png
```

**Binary Weight Analysis**:
```bash
goda weight ./...
# Shows which dependencies contribute most to binary size
```

**Impact Analysis**:
```bash
# What breaks if we remove package X?
goda tree "reach(./..., github.com/pkg/errors)"
```

#### Performance
- Designed for "fast" complex analysis
- Complements (not replaces) `go list` and `go mod graph`

**Lessons for archway**:
✓ Query language enables ad-hoc exploration
✓ Visual graphs aid understanding
✓ Weight analysis helps prioritize refactoring
✓ Conditional dependency tracking (OS, build tags)
✗ Requires GraphViz for visualization
✗ Query syntax has learning curve

---

### 3.2 Dependency Visualization Tools

| Tool | Stars | Approach | Output |
|------|-------|----------|--------|
| **modgraphviz** | N/A (golang.org/x/exp) | Converts `go mod graph` to DOT | Static GraphViz |
| **modview** | ~100 | Browser-based UI | Interactive web |
| **goda** | 1.6k | Query language + GraphViz | Static/CLI |
| **gomod** | ~400 | Querying + visualization | Static/CLI |
| **godepgraph** | ~1.3k | Package import graph | Static GraphViz |
| **gopkgview** | ~50 | Interactive web UI | Interactive web |

**Key Differences**:
- **Static (GraphViz)**: modgraphviz, goda, godepgraph → Good for documentation
- **Interactive (Web)**: modview, gopkgview → Good for exploration

**Recommendation for archway**:
- Use `go mod graph` + custom parser for module dependencies
- Use `go/packages` for package-level analysis
- Generate both static (SVG) and interactive (HTML) visualizations
- Consider D3.js for interactive graphs (no server needed)

---

### 3.3 go/packages for Codebase Analysis

**Package**: golang.org/x/tools/go/packages

#### Loading Packages
```go
import "golang.org/x/tools/go/packages"

cfg := &packages.Config{
    Mode: packages.NeedName |
          packages.NeedFiles |
          packages.NeedImports |
          packages.NeedDeps |
          packages.NeedTypes,
}

pkgs, err := packages.Load(cfg, "pattern...")
```

#### Load Modes
- `NeedName`: Package name
- `NeedFiles`: Source file list
- `NeedImports`: Direct imports
- `NeedDeps`: Transitive dependencies
- `NeedTypes`: Type information
- `NeedSyntax`: AST (`*ast.File`)
- `NeedTypesInfo`: Type-checking results

#### Use Cases for archway
```go
// Detect existing architecture
func DetectLayers(pkgs []*packages.Package) map[string][]string {
    layers := make(map[string][]string)

    for _, pkg := range pkgs {
        // Analyze package path patterns
        if strings.Contains(pkg.PkgPath, "/handlers/") {
            layers["handlers"] = append(layers["handlers"], pkg.PkgPath)
        }
        // ... detect other patterns
    }

    return layers
}

// Validate dependencies
func ValidateDependencies(pkg *packages.Package, allowed []string) []Violation {
    var violations []Violation

    for importPath := range pkg.Imports {
        if !isAllowed(importPath, allowed) {
            violations = append(violations, Violation{
                Package: pkg.PkgPath,
                Imports: importPath,
                Reason:  "Not in allowed dependencies",
            })
        }
    }

    return violations
}
```

**Lessons for archway**:
✓ Comprehensive package metadata (types, AST, imports)
✓ Respects build constraints (tags, GOOS, GOARCH)
✓ Handles module-aware builds
✗ Slow for large codebases (caching needed)

---

## 4. Migration/Refactoring Tools

### 4.1 dst (Decorated Syntax Tree)

**Repository**: github.com/dave/dst
**Stars**: 1.4k | **Forks**: 62

#### Problem with go/ast

Standard `go/ast` stores comments by byte offset:
```go
// go/ast approach
var a int // comment at byte offset 123
var b int // comment at byte offset 456

// After reordering nodes, comments become misaligned!
```

#### dst Solution

Attaches decorations (comments, whitespace) to nodes:
```go
// dst approach
type Node struct {
    Decs NodeDecs // Attached decorations
}

type NodeDecs struct {
    Start  Decorations
    End    Decorations
    Before Decorations
    After  Decorations
}
```

#### Example Transformation

```go
import (
    "github.com/dave/dst"
    "github.com/dave/dst/decorator"
)

// Parse with decorations
file, err := decorator.Parse(src)

// Modify AST
// ... reverse variable declarations ...

// Restore with decorations intact
restored, err := decorator.Restore(file)

// Result: comments stay with their nodes
var b string // bar
var a int    // foo
```

#### Use Cases
- Code generation (human-readable output)
- Refactoring (preserve original style)
- AST transformations (no full reformat)
- Code analysis (maintain developer intent)

#### Packages
- `github.com/dave/dst` - Decorated AST types
- `github.com/dave/dst/decorator` - Conversion (ast ↔ dst)
- `github.com/dave/dst/dstutil` - Utilities (mirrors `ast/astutil`)

**Lessons for archway**:
✓ Essential for code migrations (preserve comments)
✓ Maintains developer intent (formatting, spacing)
✓ Drop-in replacement for `go/ast` (similar API)
✗ Additional dependency
✗ Slower than standard AST (decoration overhead)

**Recommendation**:
- Use `dst` for code generation/migration features
- Use standard `ast` for read-only analysis (performance)

---

### 4.2 gorename & gomvpkg

**Package**: golang.org/x/tools/refactor/rename

#### gorename (Symbol Renaming)

**Purpose**: Safely rename constants, functions, variables, types

```bash
gorename -from '"pkg/path".OldName' -to NewName
```

**Features**:
- Code-aware (not text search-replace)
- Cross-package renaming
- Updates all references
- Preserves semantics

**Limitations**:
- Requires GOPATH mode (no module support as of 2025)
- Can be slow on large codebases

---

#### gomvpkg (Package Moving)

**Purpose**: Move packages, updating import declarations

```bash
gomvpkg -from old/pkg/path -to new/pkg/path -vcs_mv_cmd "git mv {{.Src}} {{.Dst}}"
```

**Features**:
- Updates all import statements
- Supports VCS integration (git mv, svn mv)
- Cross-module awareness

**Limitations**:
- No module support (GOPATH only)
- Manual VCS command specification

**Status (2025/2026)**:
- Both tools are in maintenance mode
- Community recommends IDE refactoring (GoLand, VSCode) for module-aware projects

**Lessons for archway**:
✓ Demonstrates feasibility of automated refactoring
✗ Outdated architecture (GOPATH-based)
✗ Limited adoption for modern Go (modules)

**Alternative Approach**:
- Use LSP (Language Server Protocol) for refactoring
- Integrate with `gopls` (official Go language server)
- Let IDEs handle complex refactoring; archway focuses on validation

---

### 4.3 golang.org/x/tools/refactor/eg (Example-Based Refactoring)

**Package**: golang.org/x/tools/refactor/eg

#### How It Works

**Step 1: Define Template**
```go
package template

func before(x int) int {
    return x * 2
}

func after(x int) int {
    return x << 1  // Bit shift optimization
}
```

**Step 2: Apply Transformation**
```bash
eg -t template.go -w ./...
```

**Step 3: Result**
```go
// Before
result := value * 2

// After
result := value << 1
```

#### Key Concepts
- **Template**: Package with `before` and `after` functions
- **Pattern Matching**: Matches AST structure (not text)
- **Type-Aware**: Only replaces semantically equivalent code

#### Use Cases
- API migration (old API → new API)
- Performance optimization (e.g., `append` patterns)
- Deprecation replacement (phasing out functions)

#### Limitations
- Requires template authoring (learning curve)
- Limited to expression-level transformations
- Not as flexible as custom AST tools

**Lessons for archway**:
✓ Example-based approach is intuitive
✓ Type-aware matching prevents false positives
✗ Limited to predefined patterns
✗ Requires Go knowledge to write templates

**Recommendation**:
- Consider example-based migrations for common patterns
- Provide template library for typical refactorings (hexagonal → clean architecture)

---

### 4.4 gofmt -r (Rewrite Rules)

**Built-in Tool**: Part of `gofmt`

#### Syntax
```bash
gofmt -r 'pattern -> replacement' -w ./...
```

#### Examples

**Simple Replacement**:
```bash
# Replace fmt.Sprintf with fmt.Sprint
gofmt -r 'fmt.Sprintf("%s", a) -> fmt.Sprint(a)' -w .
```

**Wildcards** (single-letter lowercase):
```bash
# Swap function arguments
gofmt -r 'foo(a, b) -> foo(b, a)' -w .
```

**Complex Patterns**:
```bash
# Simplify slice appending
gofmt -r 'append(x, y[0], y[1:]...) -> append(x, y...)' -w .
```

#### Limitations
- Text-based (not type-aware)
- Limited pattern matching (only single-letter wildcards)
- No semantic validation (can break code)

**When to Use**:
✓ Simple, localized transformations
✓ Quick one-off refactorings
✗ Complex, multi-statement changes
✗ Type-dependent transformations

**Lessons for archway**:
✓ Powerful for simple migrations
✗ Too limited for architecture-level changes
✗ Easy to misuse (breaks code)

**Recommendation**:
- Expose `gofmt -r` as "quick fix" option
- Use `dst` or `eg` for complex migrations

---

### 4.5 gofix (Automated API Migration)

**Built-in Tool**: Part of `go` toolchain (deprecated)

#### Purpose
Automatically update code for Go version upgrades:
```bash
go tool fix -r fixname ./...
```

#### Example Fixes (Historical)
- `context` - Import context from stdlib (not x/net)
- `netipv6zone` - Update IPv6 zone handling
- `printerconfig` - Update printer API

#### Current Status (2026)
- **Deprecated**: Last significant use was Go 1.0 → 1.1
- **Replaced by**: Module versioning + deprecation warnings
- **Learning**: Shows feasibility of automated migrations

#### Directive-Based Approach (Proposal)

Some tools explore `//go:fix` directives:
```go
//go:fix inline
func OldFunc() { ... }

// Automatically replaced with inlined version
```

**Lessons for archway**:
✓ Demonstrates automated migration feasibility
✓ Directive approach enables gradual migration
✗ Deprecated (no active development)
✗ Limited to simple transformations

**Recommendation**:
- Learn from `gofix` design (incremental, reversible)
- Use directives for migration markers (e.g., `//archway:migrate`)

---

## 5. Template/Plugin Distribution

### 5.1 HashiCorp go-plugin (RPC-Based)

**Repository**: github.com/hashicorp/go-plugin
**Stars**: 5.9k | **Adopters**: Terraform, Vault, Nomad, Packer, Waypoint

#### Architecture

**Subprocess Communication**:
```
┌─────────────┐         RPC          ┌─────────────┐
│  Host App   │ ◄─────────────────► │   Plugin    │
│ (CLI Tool)  │  net/rpc or gRPC    │ (subprocess) │
└─────────────┘                      └─────────────┘
```

#### Key Features

**Language Flexibility**:
- Plugins can be written in any language (via gRPC)
- Most use Go (easier interface implementation)

**Security**:
- Plugins run in separate processes (crash isolation)
- Optional checksum verification
- TLS-encrypted communication
- Process sandboxing

**Versioning**:
- Protocol versioning (invalidate incompatible plugins)
- Graceful handling of version mismatches

**Developer Experience**:
```go
// Plugin implementation
type GreeterPlugin struct{}

func (p *GreeterPlugin) Greet() string {
    return "Hello!"
}

// Host usage
client := rpc.NewClient(conn)
var plugin GreeterPlugin
client.Call("Plugin.Greet", &args, &reply)
```

#### Adoption
- **26.9k dependent projects**
- Production-tested across millions of deployments
- Used by all major HashiCorp tools

#### When to Use
✓ Language-agnostic plugins (Python, Ruby, etc.)
✓ Security-critical applications (Vault)
✓ Long-running plugins (daemons)
✗ Performance-critical paths (RPC overhead)
✗ Simple use cases (complexity overkill)

**Lessons for archway**:
✓ Proven at scale (Terraform ecosystem)
✓ Crash isolation prevents cascading failures
✓ gRPC enables language flexibility
✗ Complex setup (gRPC, versioning, handshake)
✗ Overkill for simple validators/generators

---

### 5.2 knqyf263/go-plugin (WASM-Based)

**Repository**: github.com/knqyf263/go-plugin
**Stars**: 720 | **Forks**: 34

#### Architecture

**In-Process Execution**:
```
┌─────────────────────────────────────┐
│         Host Application            │
│  ┌──────────────────────────────┐  │
│  │    Wazero Runtime            │  │
│  │  ┌────────────────────────┐  │  │
│  │  │   Plugin (Wasm)        │  │  │
│  │  └────────────────────────┘  │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

#### How It Works

**Step 1: Define Plugin Interface (Protobuf)**
```protobuf
// go:plugin
service Greeter {
    rpc Greet(GreetRequest) returns (GreetResponse);
}
```

**Step 2: Auto-Generate SDK**
```bash
go-plugin generate
# Generates:
# - plugin.go (plugin implementation)
# - host.go (host loader)
# - proto.go (protobuf definitions)
# - types.go (shared types)
```

**Step 3: Build Plugin**
```bash
GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm
```

**Step 4: Load Plugin**
```go
p, err := host.LoadPlugin("plugin.wasm")
resp, err := p.Greet(ctx, &GreetRequest{Name: "World"})
```

#### Performance Characteristics

**Advantages**:
- Size-efficient (Wasm binary format)
- Memory-safe (sandboxed execution)
- Portable (single compilation target)

**Overhead**:
- In-memory communication (faster than RPC)
- Wasm runtime overhead (slower than native)
- No network/filesystem access by default

#### Limitations vs. RPC
- In-process only (no distributed plugins)
- Go 1.21+ required (wasip1 support)
- Wasm ecosystem still maturing

**Lessons for archway**:
✓ Single binary distribution (no plugin discovery)
✓ Sandboxed execution (safety)
✓ Portable (no multi-arch builds)
✗ Go-only plugins (no language flexibility)
✗ Wasm ecosystem immature (2026)

**When to Use**:
✓ Untrusted plugins (sandboxing critical)
✓ Single-binary distribution
✓ Size-constrained environments
✗ Performance-critical code (native is faster)
✗ Need non-Go plugins

---

### 5.3 golangci-lint Plugin Systems

#### Approach 1: Go Plugin System (Deprecated)

**Method**: Dynamic `.so` libraries
```yaml
# .golangci.yml
linters-settings:
  custom:
    my-linter:
      path: /path/to/plugin.so
      type: goplugin
```

**Build**:
```bash
CGO_ENABLED=1 go build -buildmode=plugin -o linter.so
```

**Limitations**:
- Requires CGO (`CGO_ENABLED=1`)
- Environment must match exactly (OS, Go version, dependencies)
- Dependency version conflicts common
- **Status**: Not recommended (use Module Plugin instead)

---

#### Approach 2: Module Plugin System (Recommended)

**Method**: Compile linters into custom binary

**Configuration** (`.custom-gcl.yml`):
```yaml
version: v1.63.4  # golangci-lint version

plugins:
  # Remote plugin
  - module: github.com/company/custom-linter
    version: v1.2.3

  # Local plugin
  - module: github.com/company/local-linter
    path: ./linters/local-linter
```

**Build**:
```bash
golangci-lint custom
# Generates custom binary with plugins baked in
```

**Advantages**:
✓ No CGO requirement
✓ No version conflicts (single binary)
✓ Reproducible builds
✓ Works with private modules
✗ Requires rebuilding for plugin updates

**Distribution**:
- Commit custom binary to repo
- Or: Build in CI/CD
- Or: Distribute via GitHub Releases

**Lessons for archway**:
✓ Module-based approach avoids plugin complexity
✓ Single binary simplifies distribution
✓ Version pinning ensures reproducibility
✗ Requires rebuild for updates (not truly "pluggable")

---

### 5.4 buf Plugin Architecture

**Repository**: github.com/bufbuild/buf
**Plugin Framework**: bufplugin-go

#### How Buf Plugins Work

**PluginRPC Framework**:
- Protobuf-based RPC (not network RPC)
- CLI arguments + stdin/stdout transport
- Plugins run as subprocesses (like `protoc` plugins)

**Plugin Types**:
1. **Lint Plugins** - Custom protobuf linting rules
2. **Breaking Change Plugins** - API compatibility checks
3. **Code Generation Plugins** - Custom protoc plugins

#### Development Flow

**Step 1: Define Linter**
```go
import "buf.build/go/bufplugin/check"

func main() {
    check.Main(&check.Spec{
        Rules: []*check.RuleSpec{
            {
                ID:      "MY_RULE",
                Default: true,
                Purpose: "Ensures field names are snake_case",
            },
        },
        Check: myCheckFunc,
    })
}
```

**Step 2: Build Plugin**
```bash
go build -o protoc-gen-buf-plugin-my-linter
```

**Step 3: Configure**
```yaml
# buf.yaml
version: v1
lint:
  use:
    - MY_RULE
plugins:
  - plugin: buf.build/company/my-linter
```

#### Integration with Buf CLI

Plugins work like built-in rules:
- Enable/disable via config
- Ignore via comments (`// buf:lint:ignore MY_RULE`)
- Suppress errors per file

**Lessons for archway**:
✓ Subprocess plugins are simpler than RPC
✓ Protobuf API enables tool interoperability
✓ Plugin discovery via registry (buf.build)
✗ Specific to protobuf ecosystem

---

### 5.5 Plugin Distribution: Comparison Matrix

| Approach | Complexity | Flexibility | Distribution | Security | Use Case |
|----------|-----------|-------------|--------------|----------|----------|
| **Module-based** (gonew, golangci-lint) | Low | Low | Go modules | ✓ Checksum verified | Simple, trusted plugins |
| **Subprocess/RPC** (HashiCorp) | High | High | Binaries | ✓ Process isolation | Language-agnostic, untrusted |
| **WASM** (knqyf263) | Medium | Medium | Wasm binaries | ✓✓ Sandboxed | Untrusted, portable |
| **Dynamic .so** (Go plugin) | High | Medium | Shared libraries | ✗ No isolation | Deprecated (avoid) |
| **Embedded** (embed.FS) | Low | Low | Single binary | ✓ Same as binary | Built-in features |

---

### 5.6 Recommendations for archway

**Primary Approach: Module-Based Distribution**
```yaml
# archway.yml
templates:
  - module: github.com/archway/templates/hexagonal
    version: v1.2.3

validators:
  - module: github.com/company/custom-validator
    version: v0.3.0
```

**Rationale**:
✓ Leverages Go module infrastructure (proxy, sumdb)
✓ Simple implementation (no RPC/WASM complexity)
✓ Familiar to Go developers
✓ Works offline (module cache)

**Future: WASM for Untrusted Plugins**
- If archway gains marketplace/community plugins
- WASM provides sandboxing for untrusted code
- Consider when user-contributed validators needed

**Avoid**:
✗ Go plugin system (`.so` files) - too fragile
✗ RPC-based plugins - overkill for simple validators

---

## 6. Static Analysis Foundation

### 6.1 go/analysis Framework

**Package**: golang.org/x/tools/go/analysis

#### Core Types

**Analyzer**:
```go
type Analyzer struct {
    Name             string
    Doc              string
    URL              string
    Flags            flag.FlagSet
    Run              func(*Pass) (interface{}, error)
    RunDespiteErrors bool
    ResultType       reflect.Type
    Requires         []*Analyzer  // Dependencies
    FactTypes        []Fact       // Cross-package facts
}
```

**Pass** (per-package analysis context):
```go
type Pass struct {
    Analyzer   *Analyzer
    Fset       *token.FileSet
    Files      []*ast.File      // Parsed files
    Pkg        *types.Package   // Type info
    TypesInfo  *types.Info
    ResultOf   map[*Analyzer]interface{}  // Dependency results

    Report     func(Diagnostic)
    Reportf    func(token.Pos, string, ...interface{})

    // Fact propagation
    ImportObjectFact  func(types.Object, Fact) bool
    ExportObjectFact  func(types.Object, Fact)
    ImportPackageFact func(*types.Package, Fact) bool
    ExportPackageFact func(Fact)
}
```

#### Example Analyzer

```go
package mycheck

var Analyzer = &analysis.Analyzer{
    Name: "mycheck",
    Doc:  "detects architectural violations",
    Run:  run,
    Requires: []*analysis.Analyzer{
        inspect.Analyzer,  // AST traversal helper
    },
}

func run(pass *analysis.Pass) (interface{}, error) {
    inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

    inspect.Preorder([]ast.Node{(*ast.ImportSpec)(nil)}, func(n ast.Node) {
        imp := n.(*ast.ImportSpec)
        path := strings.Trim(imp.Path.Value, `"`)

        if isForbidden(path) {
            pass.Reportf(imp.Pos(), "forbidden import: %s", path)
        }
    })

    return nil, nil
}
```

#### Analyzer Composition

**Horizontal Dependencies**:
```go
var Analyzer = &analysis.Analyzer{
    Name: "archway",
    Requires: []*analysis.Analyzer{
        inspect.Analyzer,      // AST traversal
        buildssa.Analyzer,     // SSA form
        ctrlflow.Analyzer,     // Control flow graph
    },
    Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
    // Access dependency results
    inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
    ssapkg := pass.ResultOf[buildssa.Analyzer].(*ssa.Package)
    cfg := pass.ResultOf[ctrlflow.Analyzer].(*cfg.CFG)

    // Perform analysis...
}
```

#### Fact Propagation (Cross-Package)

**Export Facts**:
```go
type isPureFunc struct{}

func run(pass *analysis.Pass) (interface{}, error) {
    for _, file := range pass.Files {
        for _, decl := range file.Decls {
            if fn, ok := decl.(*ast.FuncDecl); ok {
                if isPure(fn) {
                    obj := pass.TypesInfo.Defs[fn.Name]
                    pass.ExportObjectFact(obj, &isPureFunc{})
                }
            }
        }
    }
}
```

**Import Facts**:
```go
func run(pass *analysis.Pass) (interface{}, error) {
    var fact isPureFunc

    for _, imp := range pass.Pkg.Imports() {
        pass.ImportPackageFact(imp, &fact)
        // Use fact from dependency...
    }
}
```

**Lessons for archway**:
✓ Standard framework (golangci-lint compatible)
✓ Composable analyzers (reuse existing work)
✓ Fact propagation enables cross-package analysis
✓ Driver handles parallelization, caching
✗ Per-package API (need driver for cross-package rules)

---

### 6.2 inspector Package (Fast AST Traversal)

**Package**: golang.org/x/tools/go/ast/inspector

#### Performance

- **2.5x faster** than `ast.Inspect`
- **Amortization**: ~5 traversals to offset construction cost
- **Use case**: Multiple traversals over same AST

#### Construction

```go
import "golang.org/x/tools/go/ast/inspector"

// One-time construction
in := inspector.New(files)  // []*ast.File

// Subsequent traversals are fast
```

#### API Styles

**1. Classic Preorder Traversal**:
```go
// Filter by node type (nil = all types)
in.Preorder([]ast.Node{
    (*ast.FuncDecl)(nil),
    (*ast.CallExpr)(nil),
}, func(n ast.Node) {
    switch n := n.(type) {
    case *ast.FuncDecl:
        // Handle function
    case *ast.CallExpr:
        // Handle call
    }
})
```

**2. Modern Cursor API** (v0.34.0+):
```go
// Cursor provides navigation (parent, siblings, children)
for c := range in.Root().Preorder((*ast.CallExpr)(nil)) {
    call := c.Node().(*ast.CallExpr)

    // Navigate to parent
    if parent, ok := c.Parent(); ok {
        // ...
    }

    // Navigate to siblings
    if next, ok := c.NextSibling(); ok {
        // ...
    }
}
```

**3. Generic Iterator** (v0.26.0+):
```go
import "golang.org/x/tools/go/ast/inspector"

for call := range inspector.All[*ast.CallExpr](in) {
    // Type-safe iteration
}
```

#### Advanced Features

**Find by Position**:
```go
root := in.Root()
cursor, found := root.FindByPos(start, end)
if found {
    node := cursor.Node()
}
```

**Enclosing Stack**:
```go
for c := range cursor.Enclosing() {
    // c represents node and all ancestors
    fmt.Println(c.Node())
}
```

**Edge Information**:
```go
kind := cursor.ParentEdgeKind()  // Field name in parent struct
index := cursor.ParentEdgeIndex() // Index in slice field
```

#### When to Use

| Scenario | Tool |
|----------|------|
| One-off traversal | `ast.Inspect` |
| Multiple traversals | `inspector.Inspector` |
| Need parent/sibling access | `Cursor` API |
| Type filtering | `Preorder(types)` |
| Position-based lookup | `FindByPos()` |

**Lessons for archway**:
✓ Use `inspector` for multi-pass analysis
✓ Cursor API simplifies navigation logic
✓ Type filtering reduces boilerplate
✗ Overkill for single-pass analysis

---

## 7. Cross-Cutting Insights

### 7.1 Common Patterns Across Tools

#### Pattern 1: YAML + Code Dual Interface
- **Examples**: arch-go, go-arch-lint
- **Rationale**: YAML for approachability, programmatic API for power users
- **Recommendation**: Offer both in archway

#### Pattern 2: Analyzer Composition
- **Examples**: go/analysis framework, golangci-lint
- **Rationale**: Reusable components, parallel execution
- **Recommendation**: Build archway validators as `analysis.Analyzer`s

#### Pattern 3: Gradual Adoption Mechanisms
- **Examples**: arch-go (thresholds), go-arch-lint (todo markers)
- **Rationale**: Brownfield codebases can't achieve 100% compliance immediately
- **Recommendation**: Support `archway:ignore` comments, compliance thresholds

#### Pattern 4: Module-Based Distribution
- **Examples**: gonew, golangci-lint module plugins
- **Rationale**: Leverages Go ecosystem (proxy, sumdb, versioning)
- **Recommendation**: Primary distribution mechanism for archway

#### Pattern 5: Multi-Format Output
- **Examples**: arch-go (JSON/HTML), go-arch-lint (JSON)
- **Rationale**: JSON for CI/CD, HTML for stakeholders
- **Recommendation**: Support JSON, SARIF, Markdown, HTML

---

### 7.2 Best Practices from Ecosystem

#### Code Quality
- Use `golangci-lint` for linting (archway itself)
- Provide comprehensive tests (analysis tools are complex)
- Document analyzer behavior with examples

#### Performance
- Use `inspector` for multi-pass AST traversal
- Cache `go/packages` results (expensive to load)
- Parallelize package analysis (go/analysis does this)

#### User Experience
- Provide actionable error messages (not just "violation detected")
- Suggest fixes (`analysis.SuggestedFix`)
- Generate migration guides (not just validation)

#### Distribution
- Multi-platform binaries (GoReleaser)
- Multiple install methods (`go install`, `brew`, `docker`)
- Version compatibility matrix (which Go versions supported)

#### Extensibility
- Module-based templates/validators (primary)
- Well-documented plugin API
- Example plugins in separate repo

---

### 7.3 Pitfalls to Avoid

#### Pitfall 1: Go Plugin System (`.so`)
- **Why**: Fragile (version conflicts, CGO requirement)
- **Alternative**: Module-based approach

#### Pitfall 2: Custom Template Language
- **Why**: Learning curve, limited ecosystem
- **Alternative**: Use `text/template` or `pongo2`

#### Pitfall 3: Monolithic Design
- **Why**: Hard to extend, test, maintain
- **Alternative**: Composable analyzers (go/analysis)

#### Pitfall 4: GOPATH Assumptions
- **Why**: Go modules are standard (since Go 1.11)
- **Alternative**: Use `go/packages`, module-aware tools

#### Pitfall 5: Ignoring Brownfield Migration
- **Why**: Most codebases are brownfield
- **Alternative**: Gradual adoption (thresholds, ignores, migration plans)

---

## 8. Recommended Architecture for archway

### 8.1 Core Components

```
archway/
├── cmd/
│   └── archway/         # CLI entry point
├── pkg/
│   ├── analyzer/        # go/analysis.Analyzer implementations
│   ├── detector/        # Detect existing architecture
│   ├── validator/       # Validate against rules
│   ├── migrator/        # Generate migration plans
│   ├── scaffolder/      # Generate new projects
│   └── template/        # Template engine (embed.FS)
├── templates/           # Built-in templates (embed.FS)
│   ├── hexagonal/
│   ├── clean/
│   └── ddd/
└── analyzers/           # Built-in validators
    ├── dependencies/
    ├── naming/
    └── structure/
```

### 8.2 Configuration Format

```yaml
version: 1

# Project metadata
project:
  name: my-service
  architecture: hexagonal  # or: clean, ddd, custom

# Architecture definition
layers:
  - name: domain
    packages: ["internal/domain/**"]
    may-depend-on: []

  - name: ports
    packages: ["internal/ports/**"]
    may-depend-on: [domain]

  - name: adapters
    packages: ["internal/adapters/**"]
    may-depend-on: [ports, domain]

  - name: application
    packages: ["cmd/**"]
    may-depend-on: [adapters, ports, domain]

# Custom validators (module-based)
validators:
  - module: github.com/company/custom-validator
    version: v1.0.0

# Gradual adoption
thresholds:
  compliance: 80  # Allow 20% violations temporarily
  coverage: 90    # 90% of packages must be covered by rules

# Ignore patterns
ignore:
  - "vendor/**"
  - "**/*_test.go"
```

### 8.3 CLI Commands

```bash
# Scaffold new project
archway new myservice --template hexagonal

# Detect existing architecture
archway detect --output archway.yml

# Validate architecture
archway check
archway check --json  # CI/CD-friendly output
archway check --html  # Stakeholder-friendly report

# Generate migration plan
archway migrate --from current --to hexagonal --output migration.md

# Visualize architecture
archway graph --output arch.svg

# List available templates
archway templates list

# Add custom validator
archway validators add github.com/company/validator@v1.0.0
```

### 8.4 Technology Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **CLI Framework** | `cobra` | Industry standard (kubectl, gh, hugo) |
| **Config Parsing** | `viper` | Supports YAML, ENV, flags |
| **Static Analysis** | `go/analysis` | Standard framework, composable |
| **AST Traversal** | `inspector` | 2.5x faster than `ast.Inspect` |
| **AST Transformation** | `dst` | Preserves comments/formatting |
| **Template Engine** | `text/template` + `embed.FS` | Stdlib, performant, single binary |
| **Dependency Analysis** | `go/packages` | Module-aware, comprehensive |
| **Graph Visualization** | DOT + D3.js | Static (DOT) + Interactive (D3) |
| **Plugin Distribution** | Go modules | Familiar, secure, versioned |
| **Output Formats** | JSON, SARIF, Markdown, HTML | CI/CD + human-friendly |

---

## 9. Implementation Roadmap

### Phase 1: MVP (Validation Only)
- [x] Research ecosystem (this document)
- [ ] Implement YAML config parser
- [ ] Build dependency validator (go/analysis)
- [ ] CLI: `archway check`
- [ ] Output: JSON, terminal
- [ ] Documentation + examples

### Phase 2: Detection (Brownfield Support)
- [ ] Implement architecture detector
- [ ] CLI: `archway detect`
- [ ] Support for common patterns (hexagonal, clean, DDD)
- [ ] Confidence scoring

### Phase 3: Scaffolding (Greenfield Support)
- [ ] Template engine (embed.FS + text/template)
- [ ] Built-in templates (hexagonal, clean, DDD)
- [ ] CLI: `archway new`
- [ ] Module-based template distribution

### Phase 4: Migration (Brownfield Transformation)
- [ ] Migration plan generator
- [ ] CLI: `archway migrate`
- [ ] AST transformations (dst)
- [ ] Step-by-step guides

### Phase 5: Extensibility
- [ ] Module-based validator loading
- [ ] Validator API documentation
- [ ] Example validators
- [ ] Template marketplace (docs)

### Phase 6: Polish
- [ ] Interactive graph visualization (D3.js)
- [ ] HTML reports
- [ ] IDE integrations (LSP?)
- [ ] CI/CD integrations (GitHub Actions, GitLab CI)

---

## 10. Key Repositories Reference

### Scaffolding
- [gonew](https://go.dev/blog/gonew) - Official Go project templating (experimental)
- [go-blueprint](https://github.com/Melkeydev/go-blueprint) - 8.6k stars, feature-rich scaffolder
- [autostrada](https://autostrada.dev/) - Web-based generator, no framework lock-in

### Architecture Enforcement
- [go-arch-lint](https://github.com/fe3dback/go-arch-lint) - 439 stars, YAML-based validation
- [archunit](https://github.com/kcmvp/archunit) - 22 stars, fluent API, Go port of Java ArchUnit
- [arch-go](https://github.com/arch-go/arch-go) - 245 stars, comprehensive rules (deps, content, naming)

### Static Analysis
- [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis) - Standard analyzer framework
- [inspector](https://pkg.go.dev/golang.org/x/tools/go/ast/inspector) - Fast AST traversal (2.5x speedup)
- [go/packages](https://pkg.go.dev/golang.org/x/tools/go/packages) - Load package metadata

### Refactoring
- [dst](https://github.com/dave/dst) - 1.4k stars, decorated syntax tree (preserves comments)
- [eg](https://pkg.go.dev/golang.org/x/tools/refactor/eg) - Example-based refactoring
- [gorename](https://pkg.go.dev/golang.org/x/tools/refactor/rename) - Safe symbol renaming (GOPATH only)

### Dependency Analysis
- [goda](https://github.com/loov/goda) - 1.6k stars, advanced query language
- [modgraphviz](https://pkg.go.dev/golang.org/x/exp/cmd/modgraphviz) - Visualize `go mod graph`
- [modview](https://github.com/bayraktugrul/modview) - Interactive browser visualization

### Plugin Systems
- [go-plugin (HashiCorp)](https://github.com/hashicorp/go-plugin) - 5.9k stars, RPC-based, used by Terraform
- [go-plugin (knqyf263)](https://github.com/knqyf263/go-plugin) - 720 stars, WASM-based
- [golangci-lint modules](https://golangci-lint.run/docs/plugins/module-plugins/) - Recommended plugin approach

---

## 11. Sources

### Scaffolding & Templates
- [Go-Blueprint Docs](https://docs.go-blueprint.dev/)
- [Experimenting with project templates - Go Blog](https://go.dev/blog/gonew)
- [GitHub - Melkeydev/go-blueprint](https://github.com/Melkeydev/go-blueprint)
- [Autostrada: Code generator for Go projects](https://autostrada.dev/)
- [Go template libraries: A performance comparison](https://blog.logrocket.com/golang-template-libraries-performance-comparison/)
- [Using Pongo2 Templates in Go](https://zetcode.com/golang/pongo2/)
- [How to Bundle Static Resources Inside Go Binaries](https://oneuptime.com/blog/post/2026-01-23-go-embed-static-resources/view)

### Architecture Enforcement
- [GitHub - fe3dback/go-arch-lint](https://github.com/fe3dback/go-arch-lint)
- [GitHub - kcmvp/archunit](https://github.com/kcmvp/archunit)
- [GitHub - arch-go/arch-go](https://github.com/arch-go/arch-go)
- [GitHub - roblaszczak/go-cleanarch](https://github.com/roblaszczak/go-cleanarch)

### Static Analysis
- [analysis package - golang.org/x/tools/go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- [inspector package - golang.org/x/tools/go/ast/inspector](https://pkg.go.dev/golang.org/x/tools/go/ast/inspector)
- [How to create your own Go static analyzer?](https://pvs-studio.com/en/blog/posts/go/1329/)
- [Static Analysis with Go: The First Steps](https://kat.bio/blog/go-static-analysis)
- [Writing multi-package analysis tools for Go](https://eli.thegreenplace.net/2020/writing-multi-package-analysis-tools-for-go/)

### Refactoring & AST
- [GitHub - dave/dst](https://github.com/dave/dst)
- [dst package - github.com/dave/dst](https://pkg.go.dev/github.com/dave/dst)
- [Rewriting Go with AST transformation](https://xdg.me/rewriting-go-with-ast-transformation/)
- [Rewriting Go source code with AST tooling](https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/)
- [eg package - golang.org/x/tools/refactor/eg](https://pkg.go.dev/golang.org/x/tools/refactor/eg)
- [gorename: easy refactoring tool for Golang](https://texlution.com/post/gorename/)

### Dependency Analysis
- [GitHub - loov/goda](https://github.com/loov/goda)
- [GitHub - bayraktugrul/modview](https://github.com/bayraktugrul/modview)
- [modgraphviz command - golang.org/x/exp/cmd/modgraphviz](https://pkg.go.dev/golang.org/x/exp/cmd/modgraphviz)
- [GitHub - Helcaraxan/gomod](https://github.com/Helcaraxan/gomod)

### Plugin Systems
- [GitHub - hashicorp/go-plugin](https://github.com/hashicorp/go-plugin)
- [plugin package - github.com/hashicorp/go-plugin](https://pkg.go.dev/github.com/hashicorp/go-plugin)
- [GitHub - knqyf263/go-plugin](https://github.com/knqyf263/go-plugin)
- [RPC-based plugins in Go](https://eli.thegreenplace.net/2023/rpc-based-plugins-in-go/)
- [Go Plugin System – Golangci-lint](https://golangci-lint.run/docs/plugins/go-plugins/)
- [Module Plugin System – Golangci-lint](https://golangci-lint.run/docs/plugins/module-plugins/)
- [Introducing custom lint and breaking change plugins for Buf](https://buf.build/blog/buf-custom-lint-breaking-change-plugins)

### Migration Tools
- [go:fix A Revolutionary Tool for Automated Code Migration](https://huizhou92.com/p/gofix-a-revolutionary-tool-for-automated-code-migration/)
- [Codebase Refactoring (with help from Go)](https://go.dev/talks/2016/refactor.article)
- [GitHub - golang-migrate/migrate](https://github.com/golang-migrate/migrate)

---

## 12. Conclusion

The Go ecosystem provides robust building blocks for **archway**:

1. **Scaffolding**: `gonew` demonstrates module-based distribution; `go-blueprint` shows feature-rich generation
2. **Validation**: Three approaches (YAML, fluent API, test-based) - recommend YAML for accessibility
3. **Analysis**: `go/analysis` + `inspector` provide performant, composable foundation
4. **Transformation**: `dst` enables comment-preserving code generation
5. **Distribution**: Module-based approach preferred over RPC/WASM for simplicity

**Key Differentiators for archway**:
- **Brownfield-first**: Detect existing architecture (not just validate)
- **Migration support**: Generate transformation plans (not just check compliance)
- **Multi-paradigm**: Support hexagonal, clean, DDD, custom patterns
- **Gradual adoption**: Thresholds, ignores, step-by-step guides

**Next Steps**:
1. Build MVP validator (Phase 1)
2. Test on real codebases (brownfield validation)
3. Gather community feedback
4. Iterate on detection and migration features

---

**Document Version**: 1.0
**Last Updated**: February 14, 2026
**Researcher**: Claude Code (Technical Researcher Agent)
