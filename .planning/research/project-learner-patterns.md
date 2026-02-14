# Project Learner Patterns Research

**Source:** Background research agent (completed Feb 14, 2026)

## Architecture Pattern Detection

### Hexagonal Architecture
- Detect via `ports/`, `adapters/`, `domain/` directory patterns
- Import graph analysis: domain doesn't import adapters (unidirectional)
- Look for port interfaces + adapter implementations

### Clean/Layered
- Unidirectional dependency flow: infrastructure → application → domain
- Detect `usecase/`, `entity/`, `repository/` patterns

### MVC
- `models/`, `views/`, `controllers/` directories

### Detection Techniques
- Dependency Structure Matrix (DSM) for coupling analysis
- Community detection algorithms (label propagation) for tightly coupled components
- DFS for cycle detection in dependency graphs

## Convention Extraction

### Error Handling Patterns
- Sentinel errors: `var ErrXxx = errors.New(...)`
- Typed errors: `type XxxError struct{}`
- Wrapping: `%w` with `errors.Is/As`
- Detection: AST scan for `var` declarations matching `Err*` pattern

### Logging
- Identify library from go.mod: zap, zerolog, logrus, slog
- Detect structured vs unstructured from initialization patterns
- Scan for `log.With()` / `slog.With()` (structured) vs `log.Printf()` (unstructured)

### Config Management
- Detect: Viper, godotenv, koanf, env vars, YAML/JSON
- Look for config struct tags, init patterns

### Testing Patterns
- Table-driven: loop over test cases with `t.Run()`
- BDD: Ginkgo, Godog imports
- Unit/integration ratio: count files matching `*_test.go` vs `*_integration_test.go`

### Naming Conventions
- Parse AST identifiers, classify: camelCase, snake_case, PascalCase, kebab-case
- Check consistency across codebase

## Framework Identification

### Go HTTP Frameworks
| Framework | Detection Signal |
|-----------|-----------------|
| Chi | `chi.NewRouter()` + `github.com/go-chi/chi` in go.mod |
| Gin | `gin.Default()` + `github.com/gin-gonic/gin` in go.mod |
| Echo | `echo.New()` + `github.com/labstack/echo` in go.mod |
| Fiber | `fiber.New()` + `github.com/gofiber/fiber` in go.mod |
| stdlib | `http.ListenAndServe()` + no framework in go.mod |

### Database Libraries
| Library | Detection Signal |
|---------|-----------------|
| GORM | `gorm.io/gorm` in go.mod |
| sqlx | `github.com/jmoiron/sqlx` in go.mod |
| pgx | `github.com/jackc/pgx` in go.mod |
| database/sql | `database/sql` imports, no ORM |

### Other Language Frameworks
- **PHP**: Laravel (`artisan` file), Symfony (`symfony.lock`)
- **Node**: Express, Fastify, NestJS from `package.json` dependencies
- **Python**: Django (`manage.py`), Flask, FastAPI from `requirements.txt`

## ADR Extraction Sources

### From Dependencies
- go.mod/package.json → infer framework choice, database, logging decisions
- Each major dependency is an implicit architectural decision

### From Config Files
- `.golangci.yaml` → code quality priorities
- `Dockerfile` → infrastructure decisions
- CI/CD workflows → deployment strategy
- `.editorconfig` → code style decisions

### From Explicit ADRs
- Parse `doc/adr/`, `docs/adr/`, `.claude/decisions/` directories
- Extract status, context, decision, consequences from Markdown

## Invariant Detection

### Validation Tags
- Extract business rules from struct tags: `validate:"required,email,min=8"`
- These encode domain constraints

### Function Validation
- Analyze if-conditions for constraints: `age >= 18`, `price > 0`
- Guard clauses at function entry points

### Test Assertions
- Extract invariants from `assert.Greater`, `assert.Positive`, `assert.NotNil`
- Test names encode expected behavior

### Linter Configs
- `.golangci.yaml` → code quality invariants (max complexity, function length)
- These are enforced rules

## AST Analysis Pipeline (Go)

### Recommended Multi-Stage Pipeline

1. **Static file analysis** — directory structure, config files, manifest files
2. **AST-based analysis** — parse source, build dependency graph, extract patterns
3. **Semantic analysis** — type checking, call graphs, interface implementations
4. **Temporal analysis** — git history, hotspots, change frequency
5. **Synthesis** — combine findings, generate reports with confidence scores

### Key Go Packages
- `go/ast`, `go/types`, `go/parser` — core AST
- `golang.org/x/tools/go/packages` — load packages with types
- `golang.org/x/tools/go/analysis` — analyzer framework
- `golang.org/x/tools/go/ast/inspector` — 2.5x faster traversal
- `github.com/dave/dst` — comment-preserving AST transformations

### Load Modes for go/packages
- `NeedName` — package name
- `NeedFiles` — source file list
- `NeedSyntax` — AST (crucial for analysis)
- `NeedTypes` — type information
- `NeedTypesInfo` — type-checking results
- `NeedDeps` — transitive dependencies

## Existing Tools Reference

| Tool | Focus | Key Capability |
|------|-------|---------------|
| CodeScene | Temporal | Behavioral analysis, hotspot detection |
| SonarQube | Static | Code quality, security vulnerabilities |
| Lattix | Architecture | DSM, 10+ partitioning algorithms |
| Understand | Visualization | TreeMap, Sunburst, dependency views |
| CodeCharta | Visualization | 3D code city metaphor |

## Output Format Recommendation

Analysis results should be JSON with:
- Confidence scores per finding
- Evidence (file paths, line numbers)
- Architecture pattern classification
- Convention summary
- Framework list with versions
- Inferred ADRs
- Detected business rules/invariants
