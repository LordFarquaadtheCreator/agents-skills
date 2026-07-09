# mcpserver

MCP stdio server setup, tool registration, and input validation.

## What it does

- Creates an MCP server with stdio transport
- Registers the `generate_story_pdf` tool with schema from `generate.Input`
- Validates input: title required, at least one page, each page needs image + text
- Calls `generate.Run` and returns JSON-encoded `generate.Output`

## Key functions

| Function | Purpose |
|---|---|
| `Run` | Creates server, registers tool, starts stdio transport |
| `handleGenerate` | Validates input, calls `generate.Run`, marshals output to JSON |

## Tool description

The `generate_story_pdf` tool description tells the agent:
- Provide absolute file paths to PNG/JPEG images (not base64)
- Text supports markdown: `**bold**`, `*italic*`, `\n`, `\n\n`
- Output goes to `~/Desktop/<title>/` with PDF + per-page PNGs
- Returns output directory, PDF path, and PNG paths

## Files

| File | Purpose |
|---|---|
| `server.go` | Server setup, tool registration, validation, JSON output |
| `server_test.go` | Tests for handleGenerate — valid generation, missing fields, multiple pages |
