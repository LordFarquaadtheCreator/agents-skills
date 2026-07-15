---
name: test-mcp
description: Manually test an MCP server via the official MCP Inspector. Use after scaffolding or modifying an MCP to verify handshake, list tools, and call them interactively.
---

# test-mcp

Use the official MCP Inspector to manually verify an MCP server works. Run after scaffolding a new MCP or modifying an existing one. No install needed — `npx` fetches on demand.

## When to use

- After `scaffold-mcp.sh` produces a new server
- After changing tool handlers, input schemas, or transport code
- When a client (Hermes, Claude Code, Zed, etc.) reports the server unreachable or misbehaving
- Before committing MCP changes

## Run

```bash
npx @modelcontextprotocol/inspector <command> [args...]
```

For a stdio binary in this repo:

```bash
npx @modelcontextprotocol/inspector /Users/farquaad/agents-skills/mcps/<name>/<name>
```

For a server that needs args or env, pass them after the command:

```bash
npx @modelcontextprotocol/inspector /Users/farquaad/agents-skills/mcps/flux2-mcp/flux2-mcp mcp
```

## What it does

- Starts a proxy on `localhost:6277` and a web UI on `localhost:6274`
- Prints a URL with a session token — open it in a browser
- Click **Connect** to spawn the server over stdio and run the MCP handshake
- **List Tools** shows every advertised tool with its input schema
- Click a tool to fill in args and call it; response renders inline as JSON
- **Protocol** tab shows raw JSON-RPC request/response log (DevTools-style)

## Verify

1. Handshake succeeds — server shows as connected, no errors in Protocol tab
2. `tools/list` returns the expected tool names
3. Each tool's `inputSchema` matches the Go struct tags
4. Calling a tool with valid args returns the expected output shape
5. Calling with invalid args returns a clear error (not a crash)
6. No stdout pollution — server logs go to stderr, protocol frames stay clean

## Stop

Kill the `npx` process when done. Inspector does not persist state.

## Notes

- Inspector is interactive, not a regression suite. For automated tests use `mcp-server-tester` (`steviec/mcp-server-tester`) with YAML test specs.
- For catching stdout pollution in CI, use `mcp-stdio-guard` (`1Utkarsh1/mcp-stdio-guard`).
