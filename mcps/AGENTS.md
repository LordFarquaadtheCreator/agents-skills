# MCPs

All MCP servers are defined here as git submodules. Each MCP lives in its own directory.

## Available MCPs

| MCP | Description |
|---|---|
| [`create-image`](create-image/) | Image generation via Modal ComfyUI. Exposes `list_loras` and `generate_image` tools. |
| [`deviantart-mcp`](deviantart-mcp/) | DeviantArt integration. |

## Structure

```
mcps/
├── AGENTS.md              # this file
└── <mcp-name>/
    ├── README.md           # comprehensive docs
    ├── AGENTS.md           # agent-facing instructions
    ├── Dockerfile          # builds and runs the MCP server
    └── mcp-config.json     # copy-pastable MCP config entry
```

## Rules

- Each MCP is a git submodule under `mcps/`
- Each MCP should have a `README.md` (human-facing), `AGENTS.md` (agent-facing), `Dockerfile`, and `mcp-config.json`
- When adding a new MCP submodule, update this file and `AGENTS.md` at the repo root
