<!-- keel:status:start — updated by /keel:status, do not edit manually -->
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 KEEL STATUS — Archway
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

 ACTIVE PLAN
 ───────────
 Archway v1 — Terraform for Code Architecture
 Progress: 12/17 phases (70%)

 | Phase | Title                                    | Status      |
 |-------|------------------------------------------|-------------|
 | 1.1   | Go Module + Cobra CLI Skeleton            | done        |
 | 1.2   | Provider Interface + Registry             | done        |
 | 1.3   | Config System + archway.yaml Parser       | done        |
 | 1.4   | Language Auto-Detection                   | done        |
 | 2.1   | Template Engine + Manifest/Wizard Parsing | done        |
 | 2.2   | TUI Wizard (Bubbletea + Huh)              | done        |
 | 2.3   | Port Go Templates                         | done        |
 | 2.4   | Post-Scaffold Hooks + archway.yaml Gen    | done        |
 | 3.1   | AST Analysis Pipeline                     | done        |
 | 3.2   | Architecture Pattern Detection            | done        |
 | 3.3   | Framework + Convention Detection          | done        |
 | 3.4   | Output Formatters                         | done        |
 | 3.5   | archway init Wizard                       | done        |
 | 4.1   | LLM Provider Abstraction + OpenAI Client  | -           |
 | 4.2   | Configure Command + LLM-Enhanced Analysis | -           |
 | 5.1   | MCP Server + Tools + Resources            | -           |
 | 6.1   | GoReleaser + Homebrew + Docs              | in-progress |

 WHAT'S NEXT
 ───────────
 Phase 4.1 — LLM Provider Abstraction + OpenAI Client
   - Implement LLMProvider interface and OpenAI-compatible client
   - Build auto-detection chain (Ollama, OpenAI, config file)
   - Create NoopProvider for graceful degradation when no LLM available
   Run: /keel:plan to start or /keel:context to load context first

 RULES
 ─────
 7 packs installed:
   architecture.md    code-quality.md    error-handling.md
   go.md              linter-golangci.md security.md
   testing.md

 GOVERNANCE
 ──────────
 Soul:        exists
 Decisions:   7 records
 Invariants:  1 constraint
 Product:     spec exists + 2 PRDs
 Tickets:     Not configured

 TEAM
 ────
 Shared (committed to git):
   Rules:    7 packs in .claude/rules/
   Agents:   9 agents in .claude/agents/
   MCP:      not configured

 Members need:
   No env vars required

 Run /keel:team setup to validate your environment.
 Run /keel:team to see full onboarding instructions.

 WARNINGS
 ────────
 All clear — governance is healthy.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
<!-- keel:status:end -->
