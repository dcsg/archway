# Claude Code MCP Configuration

Add Archway as an MCP server in your Claude Code MCP settings:

```json
{
  "mcpServers": {
    "archway": {
      "command": "archway",
      "args": ["mcp", "serve", "--transport", "stdio"]
    }
  }
}
```

Available tools:

- `analyze_codebase`
- `detect_architecture`
- `list_templates`
- `scaffold_project`

Available resources:

- `archway://config`
- `archway://analysis`
