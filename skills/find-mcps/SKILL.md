---
name: find-mcps
description: use this skill when the user wants to discover, install, or configure MCP servers. triggers on phrases like "find an mcp", "add an mcp", "what mcps are available", "install mcp", or when setting up a new MCP server.
---

# Find MCPs

This skill helps you discover and install MCP servers from the `mcps/` directory in this repo.

## Step 1: List available MCPs

Read `~/agents-skills/mcps/` to see what MCP servers exist. Each subdirectory is an MCP server.

For each MCP, read its `AGENTS.md` to understand what it does and its `mcp-config.json` for the config entry.

```bash
ls ~/agents-skills/mcps/
```

## Step 2: Present options

Show the user what's available. For each MCP:
- Name (directory name)
- What it does (from `AGENTS.md`)
- Required env vars or config

## Step 3: Install an MCP

When the user picks an MCP:

1. Read its `mcp-config.json` — this is the copy-pastable entry for `mcpServers`
2. Read `~/.codeium/windsurf/mcp_config.json` OR wherever the agent's platform's mcp config is located — this is where MCP servers are registered
3. Merge the new MCP entry into the `mcpServers` object
4. Write the updated config back

Do not overwrite existing entries. Only add or update the selected MCP.

Example flow:
```bash
# read the mcp's config
cat ~/agents-skills/mcps/<mcp-name>/mcp-config.json

# read the current mcp config
cat ~/.codeium/windsurf/mcp_config.json

# merge and write back
```

## Step 4: Build and verify

If the MCP has a `Dockerfile` and the user wishes to use it via docker, build it:
```bash
cd ~/agents-skills/mcps/<mcp-name> && docker build -t <mcp-name> .
```

If the MCP is a Go binary, build it:
```bash
cd ~/agents-skills/mcps/<mcp-name> && go build -o <mcp-name> .
```

Verify the MCP appears in the config:
```bash
cat ~/.codeium/windsurf/mcp_config.json | grep <mcp-name>
```

## Structure

```
mcps/
├── AGENTS.md              # global rules for MCPs
└── <mcp-name>/
    ├── Dockerfile          # builds and runs the MCP server
    ├── AGENTS.md           # what this MCP does, how to build/run
    └── mcp-config.json     # copy-pastable entry for mcpServers
```
