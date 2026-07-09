---
name: mcp-bridge
description: Call any MCP server binary from shell without registration. Use when an MCP is not registered but the binary is available on disk. Replaces ad-hoc JSON-RPC scripts.
---

# MCP Bridge

Call any MCP server binary via a CLI wrapper. No registration required. No JSON-RPC knowledge required.

## Prerequisites

Build the wrapper once:

```bash
cd ~/agents-skills/skills/mcp-bridge/scripts/mcp-call/
go build -o mcp-call .
```

Binary lives at `~/agents-skills/skills/mcp-bridge/scripts/mcp-call/mcp-call`.

## Usage

```bash
mcp-call <command...> -- <list|call|describe> [flags]
```

The `--` separator splits the MCP server command from the wrapper subcommand.

### Subcommands

| Subcommand | Purpose | Example |
|---|---|---|
| `list` | Discover available tools | `mcp-call /path/to/binary -- list` |
| `call <tool>` | Call a tool with JSON args | `mcp-call /path/to/binary -- call generate_story_pdf --args '{"title":"Test"}'` |
| `describe <tool>` | Show tool schema + description | `mcp-call /path/to/binary -- describe generate_story_pdf` |

### Flags

| Flag | Purpose | Default |
|---|---|---|
| `--args '<json>'` | JSON arguments for `call` | `{}` |
| `--env KEY=VAL` | Set env var on spawned process. Repeatable. | inherited from parent |
| `--timeout 120s` | Timeout for the operation | `120s` |

## Output

- **Text content** → printed to stdout
- **Image content** → saved to temp file, path printed to stdout
- **Errors** → stderr, exit code 1

## Known MCPs

| MCP | Command | Env vars | Tools |
|---|---|---|---|
| create-story | `~/agents-skills/mcps/create-story/create-story` | none | `generate_story_pdf` |
| create-image | `~/agents-skills/mcps/create-image/create-image` | `COMFYUI_API_URL` | `list_loras`, `list_base_models`, `generate_image` |
| create-video | `~/agents-skills/mcps/create-video/create-video` | (check AGENTS.md) | `generate_video` |
| cover-letter-writter | `~/agents-skills/mcps/cover-letter-writter/cover-letter-writter` | (check AGENTS.md) | 7 tools |

For Docker-based MCPs, use `docker run --rm -i <image>` as the command:

```bash
mcp-call docker run --rm -i create-story -- list
```

## Examples

### List tools

```bash
~/agents-skills/skills/mcp-bridge/scripts/mcp-call/mcp-call \
  ~/agents-skills/mcps/create-story/create-story -- list
```

### Call a tool

```bash
~/agents-skills/skills/mcp-bridge/scripts/mcp-call/mcp-call \
  ~/agents-skills/mcps/create-story/create-story -- \
  call generate_story_pdf --args '{"title":"My Story","pages":[{"image":"/path/to/img.png","text":"Once upon a time."}]}'
```

### Call with env vars

```bash
~/agents-skills/skills/mcp-bridge/scripts/mcp-call/mcp-call \
  ~/agents-skills/mcps/create-image/create-image -- \
  call list_loras \
  --env COMFYUI_API_URL=https://your-endpoint.modal.run
```

### Describe a tool (get schema)

```bash
~/agents-skills/skills/mcp-bridge/scripts/mcp-call/mcp-call \
  ~/agents-skills/mcps/create-story/create-story -- \
  describe generate_story_pdf
```

## Workflow

1. **Discover**: `list` to see available tools
2. **Inspect**: `describe <tool>` if you need arg schema
3. **Call**: `call <tool> --args '<json>'` with the arguments
4. **Parse**: stdout has the result. Image paths on their own lines.

## Tips

- Always use `list` first on an unknown MCP to discover tool names
- Use `describe` to get exact arg schema before constructing `--args`
- For long-running tools (image generation), the default 120s timeout should suffice. Override with `--timeout 300s` if needed
- Multiple `--env` flags can be passed: `--env KEY1=VAL1 --env KEY2=VAL2`
- Args JSON must be valid JSON. Use single quotes around the whole string
