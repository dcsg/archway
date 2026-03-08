# Plan: Archway v1 — Terraform for Code Architecture

## Overview
**Project:** github.com/dcsg/archway
**Total Phases:** 17 (across 6 parts)
**Source of Truth:** `docs/research/CONTEXT.md`, `docs/architecture/prds/archway-v1.md`
**ADRs:** 001-006 in `docs/architecture/decisions/`

## Progress

| Phase | Title | Status | Updated |
|-------|-------|--------|---------|
| 1.1 | Go Module + Cobra CLI Skeleton | done | 2026-02-14 |
| 1.2 | Provider Interface + Registry | done | 2026-02-15 |
| 1.3 | Config System + archway.yaml Parser | done | 2026-02-16 |
| 1.4 | Language Auto-Detection | done | 2026-02-17 |
| 2.1 | Template Engine + Manifest/Wizard Parsing | done | 2026-02-20 |
| 2.2 | TUI Wizard (Bubbletea + Huh) | done | 2026-02-22 |
| 2.3 | Port Go Templates | done | 2026-02-25 |
| 2.4 | Post-Scaffold Hooks + archway.yaml Gen | done | 2026-02-26 |
| 3.1 | AST Analysis Pipeline | done | 2026-03-01 |
| 3.2 | Architecture Pattern Detection | done | 2026-03-02 |
| 3.3 | Framework + Convention Detection | done | 2026-03-03 |
| 3.4 | Output Formatters | done | 2026-03-04 |
| 3.5 | archway init Wizard | done | 2026-03-05 |
| 4.1 | LLM Provider Abstraction | deferred | 2026-03-08 |
| 4.2 | Configure Command + LLM Analysis | deferred | 2026-03-08 |
| 5.1 | MCP Server + Tools + Resources | deferred | 2026-03-08 |
| 6.1 | GoReleaser + Homebrew + Docs | done | 2026-03-08 |

**IMPORTANT:** Phases 4.1, 4.2, 5.1 deferred from v1 scope. LLM integration and MCP server will be revisited post-v1.

## Model Assignment Strategy

| Phase | Task | Model | Reasoning | Est. Cost |
|-------|------|-------|-----------|-----------|
| 1.1 | Go module + Cobra CLI skeleton | opus | Defines project foundation, package layout from ADR-006 | $0.80 |
| 1.2 | Provider interface + registry | opus | Core contract from ADR-001, must be extraction-ready | $0.80 |
| 1.3 | Config system + archway.yaml parser | opus | Schema design for archway.yaml from ADR-004/005 | $0.80 |
| 1.4 | Language auto-detection | sonnet | Simple pattern matching on manifest files | $0.08 |
| 2.1 | Template engine + manifest/wizard parsing | opus | Engine design from ADR-003, core rendering pipeline | $0.80 |
| 2.2 | TUI wizard (Bubbletea + Huh) | sonnet | UI wiring driven by wizard.yaml schema | $0.08 |
| 2.3 | Port 66 Go templates | sonnet | Bulk file porting, repetitive structure | $0.08 |
| 2.4 | Post-scaffold hooks + archway.yaml generation | sonnet | Straightforward exec calls and YAML writing | $0.08 |
| 3.1 | AST analysis pipeline | opus | Most complex phase, go/packages + inspector design | $0.80 |
| 3.2 | Architecture pattern detection | opus | Heuristic algorithms, confidence scoring | $0.80 |
| 3.3 | Framework + convention detection | opus | AST pattern matching, multiple detection strategies | $0.80 |
| 3.4 | Dependency graph + output formatters | sonnet | Graph construction + JSON/Markdown/terminal output | $0.08 |
| 3.5 | archway init wizard | sonnet | TUI form following Phase 2.2 patterns | $0.08 |
| 4.1 | LLM provider abstraction + OpenAI client | sonnet | **Deferred** — removed from v1 | $0.00 |
| 4.2 | Auto-detection chain + configure command | sonnet | **Deferred** — removed from v1 | $0.00 |
| 5.1 | MCP server + tools + resources | sonnet | **Deferred** — removed from v1 | $0.00 |
| 6.1 | GoReleaser + Homebrew + DOF + docs | sonnet | Configuration files, README, CI setup | $0.08 |

**Estimated Total:** ~$6.40

---

## Part 1: Project Foundation

### Phase 1.1: Go Module + Cobra CLI Skeleton

**Objective:** Initialize the Go module and create the Cobra CLI skeleton with all v1 command stubs following ADR-006 project structure.
**Model:** `opus`
**Max Iterations:** 15
**Completion Promise:** `PHASE 1.1 COMPLETE`
**Dependencies:** None

**Prompt:**
```
You are building the foundation for Archway, a Go CLI tool. Read these files for full context:
- docs/research/CONTEXT.md (source of truth)
- docs/architecture/prds/archway-v1.md (PRD)
- docs/architecture/decisions/006-project-structure.md (project layout)

Create the Go module and Cobra CLI skeleton:

1. Initialize Go module:
   - go mod init github.com/dcsg/archway
   - Target Go 1.22+

2. Create the directory structure from ADR-006:
   cmd/archway/main.go
   internal/cli/root.go
   internal/cli/new.go
   internal/cli/init.go
   internal/cli/analyze.go
   internal/cli/configure.go
   internal/cli/mcp.go
   internal/cli/version.go
   internal/config/
   internal/provider/
   internal/scaffold/
   internal/analyzer/
   internal/analyzer/detector/
   internal/analyzer/graph/
   internal/llm/
   internal/mcp/
   internal/output/
   providers/golang/

3. Implement cmd/archway/main.go:
   - Minimal main that calls internal/cli.Execute()

4. Implement internal/cli/root.go:
   - Root command: "archway" with description "Terraform for Code Architecture"
   - Add subcommands: new, init, analyze, configure, version, mcp (with serve subcommand)
   - Global flags: --config (config file path), --verbose, --no-color, --output (terminal/json/markdown)

5. Implement command stubs (each in its own file):
   - new.go: "archway new" - Scaffold a new project. Flags: --name, --language, --template, --no-wizard
   - init.go: "archway init" - Initialize archway.yaml for existing project. Flags: --preset
   - analyze.go: "archway analyze" - Analyze existing codebase. Flags: --output, --init (generate archway.yaml)
   - configure.go: "archway configure" with "llm" subcommand. For LLM provider setup
   - mcp.go: "archway mcp serve" - Start MCP server. Flags: --transport (stdio)
   - version.go: "archway version" - Print version info (use ldflags pattern)

6. Each command stub should:
   - Have proper Use, Short, Long, Example fields
   - Print "not yet implemented" when run
   - Accept and validate its flags

7. Add a Makefile with targets:
   - build: go build -o bin/archway ./cmd/archway
   - test: go test ./...
   - lint: golangci-lint run
   - clean: rm -rf bin/

8. Verify: go build ./... succeeds, go vet ./... passes

Dependencies to install: github.com/spf13/cobra, github.com/spf13/viper

When complete, output: PHASE 1.1 COMPLETE
```

---

### Phase 1.2: Provider Interface + Registry

**Objective:** Define the LanguageProvider interface and provider registry that all language providers implement. Must be clean enough for future gRPC extraction per ADR-001.
**Model:** `opus`
**Max Iterations:** 12
**Completion Promise:** `PHASE 1.2 COMPLETE`
**Dependencies:** Phase 1.1

**Prompt:**
```
Read these files for context:
- docs/architecture/decisions/001-embedded-providers-over-plugins.md
- docs/architecture/decisions/005-three-layer-extensibility.md
- docs/research/CONTEXT.md (provider interface section)

Implement the provider interface and registry:

1. Create internal/provider/provider.go:
   - LanguageProvider interface with methods:
     * Scaffold(ctx context.Context, req ScaffoldRequest) (*ScaffoldResponse, error)
     * Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error)
     * Migrate(ctx context.Context, req MigrateRequest) (*MigrateResponse, error)
     * GetInfo(ctx context.Context) (*ProviderInfo, error)
   - Request/Response types:
     * ScaffoldRequest: ProjectName, ModulePath, TemplateName, Options map[string]string, OutputDir
     * ScaffoldResponse: FilesCreated []string, ArchwayYAML []byte
     * AnalyzeRequest: Path string, IncludeLLM bool
     * AnalyzeResponse: Language, Architecture (with confidence), Framework, Conventions, DependencyGraph, Violations
     * MigrateRequest: Path, Strategy string (v2 placeholder)
     * MigrateResponse: Success bool, Changes []string (v2 placeholder)
     * ProviderInfo: Name, Version, Language string, SupportedArchitectures, Templates []TemplateInfo
     * TemplateInfo: Name, Description string, Variables []VariableInfo
     * VariableInfo: Name, Type, Description, Default string, Required bool, Choices []string

2. Create internal/provider/registry.go:
   - Registry struct holding map[string]LanguageProvider
   - Register(language string, provider LanguageProvider)
   - Get(language string) (LanguageProvider, error)
   - List() []string
   - Global default registry instance

3. Create providers/golang/provider.go:
   - GoProvider struct implementing LanguageProvider
   - Scaffold, Analyze, Migrate, GetInfo as stub implementations returning "not yet implemented"
   - Register with global registry in init()

4. Design rules:
   - All types should be serializable to JSON (for MCP and output)
   - AnalyzeResponse should include a Confidence float64 (0-1) for architecture detection
   - DependencyGraph should be its own type with Nodes and Edges
   - Keep Migrate as a placeholder returning ErrNotImplemented

5. Write tests:
   - internal/provider/registry_test.go: register, get, list, get-unknown-returns-error
   - providers/golang/provider_test.go: implements interface, GetInfo returns valid info

When complete, output: PHASE 1.2 COMPLETE
```

---

### Phase 1.3: Config System + archway.yaml Parser

**Objective:** Implement the Viper-based config system for user settings and the archway.yaml parser for project architecture rules.
**Model:** `opus`
**Max Iterations:** 12
**Completion Promise:** `PHASE 1.3 COMPLETE`
**Dependencies:** Phase 1.1

**Prompt:**
```
Read these files for context:
- docs/architecture/decisions/004-terraform-mental-model.md
- docs/architecture/decisions/005-three-layer-extensibility.md
- docs/research/CONTEXT.md (archway.yaml example section)

Implement two config layers:

LAYER 1: User config (~/.archway/config.yaml) via Viper

1. Create internal/config/config.go:
   - AppConfig struct:
     * LLM: LLMConfig (Provider, Model, APIKey, BaseURL, Fallbacks)
     * DefaultLanguage string
     * Verbose bool
   - Load() reads from: flags > env vars (ARCHWAY_*) > ~/.archway/config.yaml > defaults
   - Save() writes to ~/.archway/config.yaml with 0600 permissions
   - Uses Viper under the hood

LAYER 2: Project config (archway.yaml) — custom parser

2. Create internal/config/archway_yaml.go:
   - ArchwayConfig struct matching the CONTEXT.md spec:
     * Language string
     * Architecture string
     * Rules: RulesConfig
       - Dependencies []DependencyRule (Layer, Packages []string, MayDependOn []string)
       - Naming []NamingRule (Pattern, MustEndWith, MustStartWith string)
       - Structure StructureConfig (RequiredDirs, ForbiddenDirs []string)
       - Functions FunctionRules (MaxLines, MaxParams, MaxReturnValues int)
     * Extends []string (preset references)
     * Templates TemplateSourceConfig (Source string)
   - LoadArchwayYAML(path string) (*ArchwayConfig, error)
   - SaveArchwayYAML(path string, config *ArchwayConfig) error
   - ValidateArchwayYAML(config *ArchwayConfig) []error
   - FindArchwayYAML(startDir string) (string, error) — walk up directories

3. Write comprehensive tests:
   - config_test.go: load defaults, override with env, save/load round-trip
   - archway_yaml_test.go: parse full example from CONTEXT.md, validate required fields,
     handle missing optional fields, find archway.yaml in parent dirs

4. Create a testdata/archway.yaml with the full example from CONTEXT.md for testing

When complete, output: PHASE 1.3 COMPLETE
```

---

### Phase 1.4: Language Auto-Detection

**Objective:** Detect project language from manifest files.
**Model:** `sonnet`
**Max Iterations:** 5
**Completion Promise:** `PHASE 1.4 COMPLETE`
**Dependencies:** Phase 1.2

**Prompt:**
```
Read docs/architecture/prds/archway-v1.md for context (US-6: auto-detect language).

Implement language auto-detection:

1. Create internal/analyzer/detector/language.go:
   - DetectLanguage(path string) (string, float64, error) — returns language name + confidence
   - Detection rules (check in order):
     * go.mod → "go" (confidence 1.0)
     * composer.json → "php" (confidence 1.0)
     * package.json → "node" (confidence 0.9) — could be frontend-only
     * pyproject.toml or requirements.txt → "python" (confidence 1.0 / 0.8)
     * Cargo.toml → "rust" (confidence 1.0)
     * pom.xml or build.gradle → "java" (confidence 1.0)
   - For v1, only "go" is supported — others return language name but provider will be nil
   - Return ("unknown", 0, nil) if nothing detected

2. Wire into analyze command:
   - Update internal/cli/analyze.go to call DetectLanguage and print result
   - If unknown, suggest user specify --language flag

3. Tests:
   - Create testdata directories with minimal manifest files
   - Test each detection rule
   - Test unknown project
   - Test confidence scores

When complete, output: PHASE 1.4 COMPLETE
```

---

## Part 2: Template Engine + Go Scaffold

### Phase 2.1: Template Engine + Manifest/Wizard Parsing

**Objective:** Build the core template engine that reads manifest.yaml + wizard.yaml + files/ and renders projects using text/template.
**Model:** `opus`
**Max Iterations:** 15
**Completion Promise:** `PHASE 2.1 COMPLETE`
**Dependencies:** Phase 1.1, 1.2

**Prompt:**
```
Read these files for context:
- docs/architecture/decisions/003-template-architecture.md
- docs/architecture/decisions/005-three-layer-extensibility.md

Build the template engine:

1. Create internal/scaffold/manifest.go:
   - Manifest struct:
     * Name, Description, Language, Version string
     * Variables []VariableDefinition
   - VariableDefinition: Name, Type (string/bool/choice), Description, Default string, Required bool, Choices []string
   - ParseManifest(data []byte) (*Manifest, error)

2. Create internal/scaffold/wizard.go (data model only, not TUI):
   - WizardConfig struct:
     * Steps []WizardStep
   - WizardStep: ID string, Questions []WizardQuestion
   - WizardQuestion: Variable, Prompt, Type (input/select/confirm/multiselect), Validate string, Options []string, When string (conditional expression)
   - ParseWizard(data []byte) (*WizardConfig, error)

3. Create internal/scaffold/renderer.go:
   - Renderer struct:
     * Takes an fs.FS (works with embed.FS or os.DirFS)
   - RenderTemplate(templateDir string, outputDir string, vars map[string]interface{}) (*RenderResult, error):
     * Read manifest.yaml from templateDir
     * Walk files/ directory
     * For each .tmpl file: parse with text/template, execute with vars, write to outputDir (strip .tmpl suffix)
     * For non-.tmpl files: copy as-is
     * Support directory names with {{.Variables}} (e.g., "cmd/{{.ServiceName}}/")
     * Return RenderResult with list of created files
   - Template functions: camelCase, snakeCase, pascalCase, kebabCase, lower, upper, title, contains, hasPrefix, hasSuffix, join, split

4. Create a minimal test template:
   - providers/golang/templates/go-minimal/ with manifest.yaml, wizard.yaml, and a few .tmpl files
   - Use this for testing the renderer

5. Tests:
   - renderer_test.go: render minimal template, verify output files, verify variable substitution,
     verify template functions, verify .tmpl suffix stripping, verify directory name substitution
   - manifest_test.go: parse valid manifest, handle missing fields
   - wizard_test.go: parse valid wizard config

When complete, output: PHASE 2.1 COMPLETE
```

---

### Phase 2.2: TUI Wizard (Bubbletea + Huh)

**Objective:** Build the interactive TUI wizard driven by wizard.yaml that collects user input for scaffolding.
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `PHASE 2.2 COMPLETE`
**Dependencies:** Phase 2.1

**Prompt:**
```
Read docs/architecture/decisions/003-template-architecture.md for wizard.yaml format context.

Build the TUI wizard using charmbracelet/huh:

1. Install dependencies: github.com/charmbracelet/huh, github.com/charmbracelet/bubbletea

2. Create internal/scaffold/tui.go:
   - RunWizard(wizardConfig *WizardConfig, manifest *Manifest) (map[string]interface{}, error)
   - Dynamically builds huh.Form from WizardConfig steps:
     * input type → huh.NewInput()
     * select type → huh.NewSelect()
     * confirm type → huh.NewConfirm()
     * multiselect type → huh.NewMultiSelect()
   - Apply validation rules from WizardQuestion.Validate (regex patterns)
   - Apply defaults from Manifest.Variables
   - Handle conditional questions (When field) — skip questions when condition not met
   - Return collected values as map

3. Wire into archway new command:
   - Update internal/cli/new.go:
     * List available templates (from provider.GetInfo())
     * If --no-wizard flag: collect all required vars from flags
     * Else: load wizard.yaml for selected template, run TUI wizard
     * Pass collected vars to renderer
     * Print summary of created files

4. Support non-interactive mode:
   - --name, --language, --template flags bypass wizard
   - Additional template variables passed as --set key=value

5. Tests:
   - tui_test.go: test form building from wizard config (unit test the config-to-form mapping,
     not the actual TUI interaction)

When complete, output: PHASE 2.2 COMPLETE
```

---

### Phase 2.3: Port 66 Go Templates

**Objective:** Port the existing DOF scaffold-go templates into the manifest.yaml + wizard.yaml + files/ format.
**Model:** `sonnet`
**Max Iterations:** 20
**Completion Promise:** `PHASE 2.3 COMPLETE`
**Dependencies:** Phase 2.1

**Prompt:**
```
Read docs/research/CONTEXT.md for Go provider template details:
- Hexagonal architecture, CQRS, Chi, slog, OTel, koanf, franz-go
- Full wizard: service type, transports, data stores, auth, email, etc.

The existing 66 templates live at ~/dotfiles/claude/templates/go-service/.
Read all files in that directory to understand the current template structure.

Port all templates to providers/golang/templates/go-hexagonal/:

1. Create providers/golang/templates/go-hexagonal/manifest.yaml:
   - All variables the wizard needs (ServiceName, ModulePath, Transport, DataStore, Auth, Email, etc.)
   - Each variable with type, description, default, choices

2. Create providers/golang/templates/go-hexagonal/wizard.yaml:
   - Multi-step wizard matching current DOF scaffold-go interview flow
   - Steps: basics, architecture, transport, datastore, auth, email, observability, extras

3. Create providers/golang/templates/go-hexagonal/files/:
   - Port each template file, converting to text/template syntax if needed
   - Maintain hexagonal directory structure: cmd/, internal/domain/, internal/port/, internal/adapter/
   - Support conditional file inclusion (e.g., kafka files only when Transport includes kafka)

4. Embed templates in the Go provider:
   - providers/golang/templates.go with //go:embed templates/* directive
   - Wire into GoProvider.Scaffold()

5. Create a simple "go-minimal" template as well:
   - Minimal Go service (just main.go + go.mod) for quick starts

6. Verify: render go-hexagonal with default values, ensure output compiles with go build

When complete, output: PHASE 2.3 COMPLETE
```

---

### Phase 2.4: Post-Scaffold Hooks + archway.yaml Generation

**Objective:** Run post-scaffold hooks and generate archway.yaml alongside scaffolded projects.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `PHASE 2.4 COMPLETE`
**Dependencies:** Phase 2.1, 1.3

**Prompt:**
```
Read docs/research/CONTEXT.md: "Generates archway.yaml as part of scaffold (desired state is set from day one)"

Implement post-scaffold hooks and archway.yaml generation:

1. Create internal/scaffold/hooks.go:
   - RunPostScaffoldHooks(outputDir string, hooks []string) error
   - Default hooks for Go projects:
     * go mod init {ModulePath}
     * go mod tidy
     * git init (if not already in a git repo)
   - Each hook runs in outputDir with proper error handling
   - Hooks are defined in manifest.yaml under a "hooks" key
   - Print each hook as it runs: "Running: go mod init..."

2. Generate archway.yaml as part of scaffold:
   - After rendering template files, also generate archway.yaml in outputDir
   - archway.yaml contents based on template's architecture:
     * For go-hexagonal: dependency rules matching hexagonal layers
     * Language: go
     * Architecture: hexagonal
     * Rules matching the generated structure
   - Use internal/config/archway_yaml.go to save

3. Wire into the full archway new flow:
   - Render templates → generate archway.yaml → run hooks → print summary

4. Tests:
   - hooks_test.go: test hook execution, test failure handling
   - Integration test: full scaffold flow produces valid project with archway.yaml

When complete, output: PHASE 2.4 COMPLETE
```

---

## Part 3: Deterministic Go Analyzer

### Phase 3.1: AST Analysis Pipeline

**Objective:** Build the core analysis pipeline using go/packages and ast/inspector to load and traverse Go source code.
**Model:** `opus`
**Max Iterations:** 15
**Completion Promise:** `PHASE 3.1 COMPLETE`
**Dependencies:** Phase 1.2

**Prompt:**
```
Read these files for context:
- docs/research/go-cli-ecosystem.md (sections on go/packages, inspector, go/analysis)
- docs/research/project-learner-patterns.md (AST analysis pipeline)
- docs/architecture/decisions/006-project-structure.md

Build the AST analysis pipeline:

1. Install dependency: golang.org/x/tools (for go/packages and ast/inspector)

2. Create internal/analyzer/analyzer.go:
   - Analyzer struct:
     * Holds loaded packages, inspector, analysis results
   - LoadPackages(path string) error:
     * Use packages.Load with NeedName | NeedFiles | NeedImports | NeedDeps | NeedSyntax | NeedTypes
     * Load ./... from the given path
     * Handle errors gracefully (partial results OK)
   - Analyze(ctx context.Context) (*AnalysisResult, error):
     * Orchestrates all detection passes
     * Returns unified AnalysisResult
   - AnalysisResult struct:
     * Language string
     * Architecture ArchitectureResult (Pattern string, Confidence float64, Evidence []string)
     * Framework FrameworkResult (Name, Version string, Confidence float64)
     * Conventions ConventionResults (ErrorHandling, Logging, Config, Testing patterns)
     * Dependencies DependencyGraph
     * PackageCount, FileCount, FunctionCount int
     * Metadata map[string]string

3. Create internal/analyzer/graph/graph.go:
   - DependencyGraph struct:
     * Nodes []PackageNode (Path, Name, IsInternal bool, Layer string)
     * Edges []DependencyEdge (From, To string, ImportType string)
   - BuildGraph(pkgs []*packages.Package) *DependencyGraph
   - FindCycles() [][]string
   - LayerViolations(rules []DependencyRule) []Violation

4. Wire GoProvider.Analyze() to create an Analyzer, load packages, and return results.

5. Tests:
   - Use testdata/ directories with small Go projects for testing
   - analyzer_test.go: load valid project, handle invalid path, handle empty project
   - graph_test.go: build graph, detect cycles, find violations

When complete, output: PHASE 3.1 COMPLETE
```

---

### Phase 3.2: Architecture Pattern Detection

**Objective:** Detect architecture patterns (hexagonal, clean, DDD, layered, flat) from package structure and import graph with confidence scoring.
**Model:** `opus`
**Max Iterations:** 12
**Completion Promise:** `PHASE 3.2 COMPLETE`
**Dependencies:** Phase 3.1

**Prompt:**
```
Read docs/research/project-learner-patterns.md for detection heuristics.

Implement architecture pattern detection:

1. Create internal/analyzer/detector/architecture.go:
   - DetectArchitecture(graph *DependencyGraph, pkgs []*packages.Package) *ArchitectureResult
   - Pattern detectors (each returns pattern name + confidence 0-1 + evidence):

   a. Hexagonal:
     - Directory signals: port/, ports/, adapter/, adapters/, domain/
     - Import signals: domain doesn't import adapters (unidirectional)
     - Port interfaces with adapter implementations
     - Confidence: 0.9 if dirs + imports match, 0.6 if only dirs

   b. Clean Architecture:
     - Directory signals: usecase/, entity/, repository/, infrastructure/
     - Import signals: unidirectional flow infrastructure → application → domain
     - Confidence: 0.9 if dirs + imports, 0.6 if only dirs

   c. DDD (Domain-Driven Design):
     - Directory signals: domain/, application/, infrastructure/
     - Code signals: aggregate, valueobject, repository patterns in naming
     - Confidence: 0.8 if dirs + naming patterns

   d. Layered/MVC:
     - Directory signals: models/, views/, controllers/, handlers/, services/
     - Confidence: 0.7 if 2+ of these exist

   e. Flat:
     - No subdirectories or only cmd/ + single package
     - Confidence: 0.9 if all Go files in 1-2 packages

2. Scoring logic:
   - Run all detectors
   - Return highest confidence match
   - If multiple patterns > 0.5, return primary with "also resembles X" in evidence
   - If all < 0.3, return "unrecognized" with suggestions

3. Tests with testdata/:
   - Create minimal project structures for each pattern
   - Verify correct detection and reasonable confidence scores
   - Test edge cases: hybrid patterns, empty projects

When complete, output: PHASE 3.2 COMPLETE
```

---

### Phase 3.3: Framework + Convention Detection

**Objective:** Detect HTTP framework, database libraries, and coding conventions (error handling, logging, config, testing patterns).
**Model:** `opus`
**Max Iterations:** 12
**Completion Promise:** `PHASE 3.3 COMPLETE`
**Dependencies:** Phase 3.1

**Prompt:**
```
Read docs/research/project-learner-patterns.md for detection signals.

Implement framework and convention detection:

1. Create internal/analyzer/detector/framework.go:
   - DetectFramework(goModContent string, pkgs []*packages.Package) *FrameworkResult
   - Detection from go.mod imports:
     * github.com/go-chi/chi → Chi
     * github.com/gin-gonic/gin → Gin
     * github.com/labstack/echo → Echo
     * github.com/gofiber/fiber → Fiber
     * google.golang.org/grpc → gRPC
     * net/http without framework → stdlib
   - Also detect database libraries:
     * gorm.io/gorm → GORM
     * github.com/jmoiron/sqlx → sqlx
     * github.com/jackc/pgx → pgx
     * database/sql without ORM → database/sql
   - Return list of detected frameworks/libraries with versions from go.mod

2. Create internal/analyzer/detector/convention.go:
   - DetectConventions(pkgs []*packages.Package) *ConventionResults
   - ConventionResults struct with sub-results for each category:

   a. Error handling:
     - Scan for var Err* = errors.New(...) → sentinel errors
     - Scan for type *Error struct → typed errors
     - Scan for fmt.Errorf with %w → wrapping
     - Report dominant pattern

   b. Logging:
     - Detect from go.mod: slog, zap, zerolog, logrus
     - Detect structured vs unstructured from usage patterns
     - slog.With() / zap.With() → structured
     - log.Printf() → unstructured

   c. Config:
     - Detect from go.mod: viper, koanf, godotenv, envconfig
     - Detect struct tags: mapstructure, env, yaml

   d. Testing:
     - Detect table-driven: loop with t.Run()
     - Detect BDD: ginkgo, godog imports
     - Count *_test.go files, report test ratio

3. Tests:
   - Create testdata Go files exercising each convention pattern
   - Verify correct detection for each category

When complete, output: PHASE 3.3 COMPLETE
```

---

### Phase 3.4: Output Formatters

**Objective:** Implement terminal, JSON, and Markdown output formatters for analysis results.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `PHASE 3.4 COMPLETE`
**Dependencies:** Phase 3.1

**Prompt:**
```
Implement output formatters for analysis results:

1. Create internal/output/formatter.go:
   - Formatter interface: Format(result *AnalysisResult) (string, error)
   - NewFormatter(format string) (Formatter, error) — factory for "terminal", "json", "markdown"

2. Create internal/output/terminal.go:
   - Pretty-printed terminal output with sections:
     * Project Summary (language, files, packages)
     * Architecture (pattern, confidence, evidence)
     * Framework (name, version)
     * Conventions (error handling, logging, config, testing)
     * Dependency Issues (violations, cycles)
   - Use color/bold for headers (respect --no-color flag)
   - Use unicode box-drawing for structure

3. Create internal/output/json.go:
   - JSON output of full AnalysisResult (json.MarshalIndent)
   - Suitable for CI/CD parsing and MCP tool responses

4. Create internal/output/markdown.go:
   - Markdown report with headers, tables, code blocks
   - Suitable for saving to docs/ or pasting into PRs

5. Wire into archway analyze command:
   - Update internal/cli/analyze.go to run full analysis pipeline
   - Use --output flag to select formatter (default: terminal)
   - Print formatted result

6. Tests:
   - Format a sample AnalysisResult with each formatter
   - Verify JSON is valid, Markdown has expected headers

When complete, output: PHASE 3.4 COMPLETE
```

---

### Phase 3.5: archway init Wizard

**Objective:** Implement `archway init` to generate archway.yaml for existing projects via interactive wizard.
**Model:** `sonnet`
**Max Iterations:** 8
**Completion Promise:** `PHASE 3.5 COMPLETE`
**Dependencies:** Phase 3.1, 3.2, 2.2, 1.3

**Prompt:**
```
Read docs/architecture/prds/archway-v1.md (US-2: Initialize desired architecture).

Implement archway init:

1. Update internal/cli/init.go:
   - Check if archway.yaml already exists (warn + confirm overwrite)
   - Optionally run archway analyze first to pre-populate suggestions
   - Launch TUI wizard using huh:
     Step 1: Language (auto-detected, confirm)
     Step 2: Architecture pattern (hexagonal/clean/ddd/layered/custom)
     Step 3: Dependency rules (pre-populated from analysis or preset)
     Step 4: Structure rules (required/forbidden dirs)
     Step 5: Naming conventions (optional)
     Step 6: Preset extends (optional, e.g., archway/go-hexagonal-strict)
   - Generate archway.yaml using internal/config/archway_yaml.go
   - Print summary of generated config

2. If --preset flag provided, load preset and skip most wizard questions

3. Tests:
   - Test init generates valid archway.yaml
   - Test with pre-existing archway.yaml (overwrite confirmation)

When complete, output: PHASE 3.5 COMPLETE
```

---

## Part 4: LLM Integration

### Phase 4.1: LLM Provider Abstraction + OpenAI Client

**Objective:** Implement the LLM provider interface and OpenAI-compatible client per ADR-002.
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `PHASE 4.1 COMPLETE`
**Dependencies:** Phase 1.3

**Prompt:**
```
Read docs/architecture/decisions/002-openai-compatible-api-for-llm.md for full context.

Implement LLM integration:

1. Install dependency: github.com/sashabaranov/go-openai

2. Create internal/llm/provider.go:
   - LLMProvider interface:
     * Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
     * Available() bool
   - CompletionRequest: SystemPrompt, UserPrompt string, MaxTokens int
   - CompletionResponse: Content string, TokensUsed int

3. Create internal/llm/openai.go:
   - OpenAIProvider struct (wraps sashabaranov/go-openai client)
   - NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider
   - Implements LLMProvider
   - Supports any OpenAI-compatible API (OpenAI, Groq, Together, Ollama)

4. Create internal/llm/detect.go:
   - DetectProvider(cfg *config.AppConfig) (LLMProvider, error)
   - Auto-detection chain:
     1. Check Ollama at localhost:11434 (HTTP GET /api/tags)
     2. Check OPENAI_API_KEY or ARCHWAY_LLM_API_KEY env var
     3. Check ~/.archway/config.yaml llm section
     4. Return nil (no LLM available — graceful degradation)
   - Log which provider was detected

5. Create internal/llm/noop.go:
   - NoopProvider that returns ErrNoLLM for all calls
   - Used when no LLM is detected (allows code to call LLM without nil checks)

6. Tests:
   - Mock HTTP server for OpenAI API tests
   - Test detection chain with various env configurations
   - Test NoopProvider returns appropriate error

When complete, output: PHASE 4.1 COMPLETE
```

---

### Phase 4.2: Configure Command + LLM-Enhanced Analysis

**Objective:** Implement `archway configure llm` and wire LLM into analysis for ADR generation and invariant extraction.
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `PHASE 4.2 COMPLETE`
**Dependencies:** Phase 4.1, 3.1

**Prompt:**
```
Implement LLM configuration and enhanced analysis:

1. Update internal/cli/configure.go:
   - archway configure llm subcommand:
     * TUI wizard with huh:
       - Provider: Ollama (local) / OpenAI / Groq / Together / Custom
       - If Ollama: auto-detect available models, let user pick
       - If cloud: prompt for API key, model name
       - If custom: prompt for base URL, API key, model
     * Test connection (send simple prompt, verify response)
     * Save to ~/.archway/config.yaml
     * Print success with provider name and model

2. Create LLM-enhanced analysis functions:
   - internal/llm/analysis.go:
     * GenerateADRs(ctx, analysisResult, llmProvider) ([]ADR, error)
       - Send analysis summary to LLM with prompt to extract architectural decisions
       - Parse response into structured ADR format
     * ExtractInvariants(ctx, analysisResult, llmProvider) ([]Invariant, error)
       - Send validation patterns and test assertions to LLM
       - Parse response into structured invariant format
     * SemanticAssessment(ctx, analysisResult, llmProvider) (string, error)
       - Send full analysis to LLM for narrative assessment
       - Return human-readable architecture assessment

3. Wire into archway analyze:
   - If LLM available and --no-llm not set:
     * Run LLM-enhanced analysis after deterministic analysis
     * Append LLM results to AnalysisResult
   - If no LLM: skip silently (graceful degradation)
   - Warn before sending code to cloud provider (respect NF-09)

4. Tests:
   - Test configure flow with mock responses
   - Test LLM analysis functions with mock LLM provider
   - Test graceful degradation when no LLM available

When complete, output: PHASE 4.2 COMPLETE
```

---

## Part 5: MCP Server

### Phase 5.1: MCP Server + Tools + Resources

**Objective:** Expose Archway's analysis capabilities as an MCP server for AI client integration.
**Model:** `sonnet`
**Max Iterations:** 12
**Completion Promise:** `PHASE 5.1 COMPLETE`
**Dependencies:** Phase 3.1, 3.2, 3.3, 2.1

**Prompt:**
```
Read docs/research/CONTEXT.md (MCP server section) and docs/architecture/prds/archway-v1.md (US-4).

Install dependency: github.com/modelcontextprotocol/go-sdk

Implement MCP server:

1. Create internal/mcp/server.go:
   - NewServer() creates MCP server instance
   - RegisterTools() registers all Archway tools
   - RegisterResources() registers all resources
   - Serve(transport string) error — starts server (stdio transport)

2. Create internal/mcp/tools.go — MCP tool definitions:

   a. analyze_codebase:
     - Input: path (string, required), output_format (string, optional: json/markdown)
     - Runs full Analyzer pipeline on given path
     - Returns AnalysisResult as JSON

   b. detect_architecture:
     - Input: path (string, required)
     - Returns architecture pattern + confidence + evidence

   c. list_templates:
     - Input: language (string, optional)
     - Returns available templates with metadata from provider.GetInfo()

   d. scaffold_project:
     - Input: template (string), name (string), module_path (string), options (object), output_dir (string)
     - Runs scaffold pipeline
     - Returns list of created files

3. Create internal/mcp/resources.go — MCP resource definitions:

   a. archway://config — current archway.yaml content
   b. archway://analysis — latest analysis results (cached)

4. Update internal/cli/mcp.go:
   - archway mcp serve starts the MCP server
   - --transport flag (default: stdio)

5. Create example Claude Code configuration:
   - docs/claude-code-config.md showing how to add Archway as MCP server

6. Tests:
   - Test tool handlers with mock inputs
   - Test resource handlers
   - Integration test: start server, call tool, verify response

When complete, output: PHASE 5.1 COMPLETE
```

---

## Part 6: Polish + Distribution

### Phase 6.1: GoReleaser + Homebrew + DOF + Documentation

**Objective:** Set up distribution, DOF integration, and user documentation.
**Model:** `sonnet`
**Max Iterations:** 10
**Completion Promise:** `PHASE 6.1 COMPLETE`
**Dependencies:** All previous phases

**Prompt:**
```
Final polish and distribution setup:

1. Create .goreleaser.yaml:
   - Build for: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
   - ldflags: -s -w -X main.version={{.Version}} -X main.commit={{.Commit}}
   - Archive format: tar.gz (linux/mac), zip (windows)
   - Homebrew tap configuration
   - Changelog generation from git commits

2. Create Homebrew formula template in Formula/archway.rb

3. Update Makefile:
   - release: goreleaser release
   - snapshot: goreleaser release --snapshot --clean
   - install-local: go install ./cmd/archway

4. Create .github/workflows/:
   - ci.yml: go test, go vet, golangci-lint on push/PR
   - release.yml: goreleaser on tag push

5. Create README.md:
   - Project description ("Terraform for Code Architecture")
   - Quick start (install, archway new, archway analyze)
   - Commands reference
   - archway.yaml example
   - MCP server setup for Claude Code
   - Contributing guide
   - License (decide: MIT or Apache 2.0)

6. DOF integration:
   - Document how /dof:scaffold-go will call archway new go
   - Create docs/dof-integration.md with MCP configuration

7. Create .golangci.yml with reasonable linter configuration

8. Final checks:
   - go build ./... succeeds
   - go test ./... passes
   - golangci-lint run passes
   - go vet ./... passes

When complete, output: PHASE 6.1 COMPLETE
```

---

## Execution Order

```
Phase 1.1 ──┬── Phase 1.2 ──── Phase 3.1 ──┬── Phase 3.2
             │                               ├── Phase 3.3
             ├── Phase 1.3 ──── Phase 4.1    ├── Phase 3.4
             │                  └── Phase 4.2 └── Phase 3.5
             ├── Phase 1.4
             │
             └── Phase 2.1 ──┬── Phase 2.2
                             ├── Phase 2.3
                             └── Phase 2.4
                                              Phase 5.1 (after Part 2 + Part 3)
                                              Phase 6.1 (after all)
```

**Parallel opportunities:**
- Phase 1.2 and 1.3 can run in parallel (both depend on 1.1 only)
- Phase 2.1-2.4 and Phase 3.1-3.5 can run in parallel (independent tracks)
- Phase 4.1 can start as soon as 1.3 is done (parallel with Part 2 and Part 3)
