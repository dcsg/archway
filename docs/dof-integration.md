# DOF Integration

`/dof:scaffold-go` should delegate to Archway so scaffold output is identical across CLI and MCP flows.

## Wrapper behavior

- `/dof:scaffold-go` collects user choices
- Executes `archway new --language go ...`
- Returns created file summary

## Recommended command shape

```bash
archway new \
  --language go \
  --template go-hexagonal \
  --name <service-name> \
  --module <module-path> \
  --set Transport=http \
  --set DataStore=postgres
```

## MCP mode in DOF

When DOF runs inside Claude Code, prefer MCP calls:

1. `list_templates`
2. `scaffold_project`
3. `analyze_codebase` (optional post-check)
