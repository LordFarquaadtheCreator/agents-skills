---
name: scaffold-mcp
description: Scaffold a new MCP server in this repo. Triggers on phrases like "create a new mcp", "scaffold an mcp", "add an mcp server", "new mcp", or when the user wants to build a new Model Context Protocol server.
---

# Scaffold MCP

Scaffold a new MCP server under `mcps/` following the patterns used by existing MCPs in this repo.

## When to Use

- User says "create a new MCP" or "scaffold an MCP"
- User wants to build a new Model Context Protocol server
- User asks to add a new MCP to the repo

## Step 1: Gather Info

Ask the user for:

1. **MCP name** — lowercase, hyphens only (e.g. `my-cool-mcp`). Used as directory name, Go binary name, and Docker image name.
2. **Description** — one-line description of what the MCP does.
3. **Tool names + descriptions** — what tools the MCP should expose. Each tool needs a name, description, and input params.
4. **Env vars** — any required environment variables (e.g. API keys, endpoints).
5. **GitHub repo URL** — the remote repo for the submodule (e.g. `git@github.com:LordFarquaadtheCreator/my-cool-mcp.git`). If the user hasn't created one yet, the script will init a local git repo and print instructions.

## Step 2: Run the Scaffold Script

```bash
~/agents-skills/commands/scaffold-mcp.sh <mcp-name> [description]
```

This creates:

```
mcps/<mcp-name>/
├── .gitignore
├── AGENTS.md
├── Dockerfile
├── README.md
├── go.mod
├── main.go
├── mcp-config.json
└── internal/
    └── mcpserver/
        └── server.go
```

The script also initializes a local git repo inside the directory.

## Step 3: Implement the Tools

Open `internal/mcpserver/server.go`. The scaffold provides a single example tool (`example_tool`). Replace it with the real tools.

### Pattern to follow

Every MCP in this repo uses the same structure:

1. **Input structs** with `jsonschema` tags for tool args:
   ```go
   type MyToolInput struct {
       Field string `json:"field" jsonschema:"required,Description of the field"`
   }
   ```

2. **Output structs** for structured results:
   ```go
   type MyToolResult struct {
       Result string `json:"result"`
   }
   ```

3. **Tool registration** in `Run()`:
   ```go
   mcp.AddTool(server, &mcp.Tool{
       Name:        "my_tool",
       Description: "What it does. Be specific — the agent reads this to decide when to call it.",
   }, func(ctx context.Context, req *mcp.CallToolRequest, args MyToolInput) (*mcp.CallToolResult, MyToolResult, error) {
       return handleMyTool(ctx, req, args)
   })
   ```

4. **Handler functions** that validate input, do work, return JSON:
   ```go
   func handleMyTool(ctx context.Context, req *mcp.CallToolRequest, args MyToolInput) (*mcp.CallToolResult, MyToolResult, error) {
       if args.Field == "" {
           return nil, MyToolResult{}, fmt.Errorf("field is required")
       }
       // do work
       return jsonResult(MyToolResult{Result: "done"})
   }
   ```

5. **`jsonResult` helper** (already in the scaffold):
   ```go
   func jsonResult[T any](out T) (*mcp.CallToolResult, T, error) {
       b, err := json.MarshalIndent(out, "", "  ")
       if err != nil {
           return nil, out, fmt.Errorf("marshal result: %w", err)
       }
       return &mcp.CallToolResult{
           Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
       }, out, nil
   }
   ```

### If the MCP needs env vars

Add env var checks in `main.go`:
```go
func main() {
    apiURL := os.Getenv("MY_API_URL")
    if apiURL == "" {
        log.Fatal("MY_API_URL env var is required")
    }
    if err := mcpserver.Run(apiURL); err != nil {
        fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
        os.Exit(1)
    }
}
```

Pass env vars through `Run()` as params — don't read `os.Getenv` inside handlers.

### If the MCP needs internal packages

Create sub-packages under `internal/`:
```
internal/
├── mcpserver/    # MCP server setup, tool handlers, validation
└── mypkg/        # business logic
```

Add an `internal/AGENTS.md` listing the packages. See `mcps/create-story/internal/AGENTS.md` for the format.

## Step 4: Update mcp-config.json

The scaffold creates a basic `mcp-config.json` with both Docker and stdio entries. Update it:

- Add env vars to the `env` block if needed
- Adjust Docker volume mounts if the MCP needs filesystem access
- The stdio entry uses `~/agents-skills/mcps/<name>/<name>` as the binary path — correct if the binary name differs

Example (from `create-image`):
```json
{
  "create-image": {
    "command": "docker",
    "args": ["run", "--rm", "-i", "-e", "COMFYUI_API_URL", "create-image"],
    "env": {
      "COMFYUI_API_URL": "YOUR_MODAL_COMFYUI_API_URL"
    }
  },
  "create-image-stdio": {
    "command": "/Users/farquaad/senor-modal-apps/create-image/create-image",
    "args": [],
    "env": {
      "COMFYUI_API_URL": "YOUR_MODAL_COMFYUI_API_URL"
    }
  }
}
```

## Step 5: Update AGENTS.md and README.md

The scaffold creates template `AGENTS.md` and `README.md`. Fill them in:

- **AGENTS.md** — agent-facing: what the MCP does, tool list, how to build/run, key files table
- **README.md** — human-facing: architecture diagram, tool param tables, build/run/docker instructions, dependencies

See `mcps/manage-job/AGENTS.md` and `mcps/manage-job/README.md` for the simplest examples.

## Step 6: Build and Test

```bash
cd ~/agents-skills/mcps/<mcp-name>
go build -o <mcp-name> .
go test ./... -count=1
```

Write tests following the pattern in `mcps/create-story/internal/mcpserver/server_test.go` — test handler validation, missing fields, and success paths.

## Step 7: Add as Git Submodule

If the GitHub repo doesn't exist yet:

1. Create the repo on GitHub
2. Push the scaffolded code:
   ```bash
   cd ~/agents-skills/mcps/<mcp-name>
   git remote add origin git@github.com:LordFarquaadtheCreator/<mcp-name>.git
   git push -u origin main
   ```
3. Remove the directory from the parent repo and re-add as submodule:
   ```bash
   cd ~/agents-skills
   rm -rf mcps/<mcp-name>
   git submodule add git@github.com:LordFarquaadtheCreator/<mcp-name>.git mcps/<mcp-name>
   ```

If the repo already exists, just add the submodule:
```bash
cd ~/agents-skills
git submodule add git@github.com:LordFarquaadtheCreator/<mcp-name>.git mcps/<mcp-name>
```

## Step 8: Update mcps/AGENTS.md

Add a row to the table in `mcps/AGENTS.md`:

```markdown
| [`<mcp-name>`](<mcp-name>/) | <description> |
```

## Step 9: Register with Agents

Copy the `mcp-config.json` entry into the agent's MCP config:

- **Windsurf**: `~/.codeium/windsurf/mcp_config.json`
- **Claude Code**: `~/.claude/claude_desktop_config.json`
- **Zed**: project `.zed/settings.json` under `mcp_servers`
- **Devin**: session config

Merge into the `mcpServers` object — don't overwrite existing entries.

## Patterns Summary

| File | Purpose |
|---|---|
| `main.go` | Entry point, env var checks, calls `mcpserver.Run()` |
| `internal/mcpserver/server.go` | MCP server setup, tool registration, handlers, validation, `jsonResult` helper |
| `internal/mcpserver/server_test.go` | Handler tests |
| `Dockerfile` | Multi-stage Go build → Alpine runtime |
| `mcp-config.json` | Copy-pastable MCP config (Docker + stdio entries) |
| `AGENTS.md` | Agent-facing docs |
| `README.md` | Human-facing docs |
| `.gitignore` | Ignore the compiled binary |
| `go.mod` | Go module — uses `github.com/modelcontextprotocol/go-sdk/mcp` |

All MCPs use:
- Go + `github.com/modelcontextprotocol/go-sdk/mcp` v1.6.1
- stdio transport (`mcp.StdioTransport{}`)
- Multi-stage Dockerfile (golang:1.25-alpine builder → alpine:latest runtime)
- `jsonResult[T]` generic helper for JSON output
- `jsonschema` struct tags for tool input schemas
