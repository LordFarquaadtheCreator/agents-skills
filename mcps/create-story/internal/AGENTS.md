# create-story — internal packages

Internal Go packages for the create-story MCP server. Not importable outside this module.

## Packages

| Package | Purpose |
|---|---|
| [`generate`](generate/AGENTS.md) | Page rendering, PNG generation, PDF assembly, text layout, color extraction |
| [`mcpserver`](mcpserver/AGENTS.md) | MCP stdio server setup, tool registration, input validation |

## Flow

```
mcpserver.Run()  ──registers──►  generate_story_pdf tool
                   │
                   ▼
                 mcpserver.handleGenerate()  ──validates──►  generate.Run()
                   │
                   ▼
                 generate.Run()  ──renders pages──►  PDF + PNGs on disk
```
