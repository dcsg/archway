# PRD-001: TypeScript/Node Language Provider

**Status:** draft
**Date:** 2026-03-08

---

## Problem

Archway currently supports only Go. TypeScript/Node is one of the most popular stacks for backend services, yet developers face the same scaffolding pain: wiring Express/Fastify, TypeORM/Prisma, auth middleware, structured logging, graceful shutdown, Docker, and CI — before writing a single line of business logic. The composition model (architecture + capabilities) that Archway provides for Go should extend to TypeScript/Node to validate the multi-language provider model and serve a much larger developer audience.

## Users

- **Node.js/TypeScript developers** starting new backend services who want production-grade defaults
- **Teams** running polyglot stacks (Go + Node) who want consistent architecture governance across languages
- **Archway contributors** who need a second provider to validate the provider interface is truly language-agnostic

## Goals

- Prove the provider model works for a second language — the Go provider interface (`Scaffold`, `Analyze`, `GetInfo`) should work without changes
- Ship a TypeScript/Node provider with at least 2 architectures and 8+ capabilities
- Reuse the same wizard, suggestion engine, and `archway check` infrastructure

## Non-Goals

- Frontend/React scaffolding — this is backend services only
- Deno or Bun as primary runtimes (Node.js first, others can be added later)
- Full parity with Go provider on day one — start with the most impactful capabilities

## Requirements

### Must Have
- TypeScript/Node provider implementing the `LanguageProvider` interface
- At least 2 architectures: hexagonal (ports & adapters) and flat (simple service)
- Core capabilities: platform (config, logging, lifecycle), bootstrap, http-api, docker, testing, linting
- Templates using TypeScript (not JavaScript) with strict mode
- `archway check` support — dependency rules enforced via import analysis
- Smart suggestions working identically to Go (same suggestion engine, Node-specific rules)
- Per-project `docs/PROJECT.md` generation

### Should Have
- Additional capabilities: mysql/postgres (via Prisma or TypeORM), redis, auth-jwt, rate-limiting
- gRPC capability (via @grpc/grpc-js)
- Kafka consumer capability
- CI GitHub Actions capability
- Pre-commit hooks capability

### Won't Have (v1)
- `archway analyze` for brownfield TypeScript projects (requires TypeScript AST analysis — separate effort)
- Framework-specific templates (NestJS, Next.js API routes) — start framework-agnostic with Express/Fastify choice
- Monorepo support (Nx, Turborepo)
- Runtime alternatives (Deno, Bun)

## User Stories

**As a** TypeScript developer, **I want** to run `archway new my-api --lang typescript` and get a production-ready service **so that** I skip the boilerplate and start with proven patterns.

**As a** tech lead with a polyglot team, **I want** `archway check` to enforce the same architectural rules on both Go and TypeScript services **so that** all services follow consistent structure.

**As a** Node.js developer, **I want** the wizard to suggest capabilities I'm missing (e.g., rate limiting when I pick http-api) **so that** I don't forget production essentials.

## Acceptance Criteria

- [ ] `archway new my-ts-api --lang typescript --arch hexagonal --cap platform,bootstrap,http-api,docker --no-wizard` scaffolds a working TypeScript project
- [ ] Scaffolded project passes `npm run build` (TypeScript compilation) and `npm test`
- [ ] `archway check` validates dependency rules on the scaffolded TypeScript project
- [ ] Smart suggestions fire correctly for TypeScript capability combinations
- [ ] `docs/PROJECT.md` is generated with TypeScript-specific patterns and tools
- [ ] Provider registers correctly and appears in `archway new` language selection
- [ ] Wizard works identically to Go — same flow, TypeScript-specific choices

## Technical Notes

- Provider lives at `providers/typescript/` following the same structure as `providers/golang/`
- Templates use `text/template` + `embed.FS` (same engine as Go)
- Package manager: npm (default), with potential yarn/pnpm support later
- HTTP framework choice: Express or Fastify (wizard question)
- ORM choice: Prisma or TypeORM (wizard question, when database capability selected)
- Testing: Vitest (default) or Jest
- Linting: ESLint + Prettier
- The `archway check` import analysis will need a TypeScript-specific import parser (not Go AST) — likely using regex on import/require statements for v1, with proper TS AST analysis deferred

## Open Questions

- Should we default to Express or Fastify? Fastify has better TypeScript support and performance, but Express has larger ecosystem.
- Should the hexagonal architecture use classes (more traditional OOP) or functional composition (more idiomatic modern TS)?
- What's the minimum Node.js version to target? (Node 20 LTS seems right)
- Should we include a Dockerfile with multi-stage build like Go, or use a simpler single-stage approach?

---

*Written by keel:prd — 2026-03-08*
