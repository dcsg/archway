<!-- keel:start — managed by keel, do not edit manually -->
## Keel

### Project
"Terraform for Code Architecture" — Go CLI for scaffolding, analyzing, and governing code architecture.

### Before Writing Code
1. Read `docs/soul.md` for project context
2. Rules are enforced automatically via `.claude/rules/`
3. If a plan is active, read it in `docs/product/plans/` — check progress table for current state

### Build & Test Commands
```
# Build
go build ./...

# Test
go test ./...

# Lint
golangci-lint run

# Vet
go vet ./...
```

### Keel Commands
When the user asks any of the following, run the corresponding command automatically:

| If the user asks... | Run |
|---------------------|-----|
| "what's our status?", "where are we?", "project status" | `/keel:status` |
| "what's next?", "what should we do next?", "next steps" | `/keel:status` |
| "load context", "remind yourself", "what's this project?" | `/keel:context` |
| "create a plan", "let's plan this", "plan for X" | `/keel:plan` |
| "save this decision", "record this", "capture that" | `/keel:adr` |
| "add an invariant", "that's a hard rule", "never do X" | `/keel:invariant` |
| "write a PRD", "document this feature", "requirements for X" | `/keel:prd` |

### After Compaction
If context was compacted, re-read the active plan file in `docs/product/plans/`. The progress table is the persistent state — it tells you what's done and what's next.
<!-- keel:end -->
