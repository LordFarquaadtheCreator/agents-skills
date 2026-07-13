#!/bin/bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MCPS_DIR="$REPO_ROOT/mcps"

usage() {
  echo "Usage: $0 <mcp-name> [description]"
  echo ""
  echo "Positional arguments:"
  echo "  mcp-name    Required. Name of the MCP (used as directory, binary, and Docker image name)."
  echo "  description Optional. One-line description for AGENTS.md and README.md."
  echo ""
  echo "Creates a new MCP server scaffold under mcps/<mcp-name>/ with:"
  echo "  main.go, internal/mcpserver/server.go, go.mod, Dockerfile,"
  echo "  mcp-config.json, AGENTS.md, README.md, .gitignore"
  echo "  Also initializes a local git repo."
  exit 1
}

MCP_NAME="${1:-}"
DESCRIPTION="${2:-}"

if [ -z "$MCP_NAME" ]; then
  echo "Error: mcp-name is required."
  echo ""
  usage
fi

# Validate name: only lowercase letters, digits, and hyphens
if ! [[ "$MCP_NAME" =~ ^[a-z0-9-]+$ ]]; then
  echo "Error: mcp-name must contain only lowercase letters, digits, and hyphens."
  exit 1
fi

MCP_DIR="$MCPS_DIR/$MCP_NAME"

if [ -d "$MCP_DIR" ]; then
  echo "Error: MCP '$MCP_NAME' already exists at $MCP_DIR"
  exit 1
fi

# --- Create directory structure ---
mkdir -p "$MCP_DIR/internal/mcpserver"

# --- .gitignore ---
cat > "$MCP_DIR/.gitignore" <<EOF
$MCP_NAME
.DS_Store
EOF

# --- go.mod ---
cat > "$MCP_DIR/go.mod" <<EOF
module github.com/LordFarquaadtheCreator/$MCP_NAME

go 1.25.0

require github.com/modelcontextprotocol/go-sdk v1.6.1

require (
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
)
EOF

# --- main.go ---
cat > "$MCP_DIR/main.go" <<EOF
package main

import (
	"fmt"
	"os"

	"github.com/LordFarquaadtheCreator/$MCP_NAME/internal/mcpserver"
)

func main() {
	if err := mcpserver.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}
EOF

# --- internal/mcpserver/server.go ---
cat > "$MCP_DIR/internal/mcpserver/server.go" <<'EOF'
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- Tool inputs ---

type ExampleToolInput struct {
	Message string `json:"message" jsonschema:"required,Example input field"`
}

// --- Tool outputs ---

type ExampleToolResult struct {
	Result string `json:"result"`
}

// Run starts the stdio MCP server.
func Run() error {
	server := mcp.NewServer(&mcp.Implementation{Name: "MCP_NAME", Version: "1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "example_tool",
		Description: "Replace this with your tool description. Be specific — the agent reads this to decide when to call it.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ExampleToolInput) (*mcp.CallToolResult, ExampleToolResult, error) {
		return handleExampleTool(ctx, req, args)
	})

	return server.Run(context.Background(), &mcp.StdioTransport{})
}

// --- Handlers ---

func handleExampleTool(ctx context.Context, req *mcp.CallToolRequest, args ExampleToolInput) (*mcp.CallToolResult, ExampleToolResult, error) {
	if args.Message == "" {
		return nil, ExampleToolResult{}, fmt.Errorf("message is required")
	}

	return jsonResult(ExampleToolResult{Result: "echo: " + args.Message})
}

// jsonResult marshals the structured output as pretty JSON in the text content.
func jsonResult[T any](out T) (*mcp.CallToolResult, T, error) {
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, out, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}, out, nil
}
EOF

# Fix the MCP_NAME placeholder in server.go
sed -i '' "s/MCP_NAME/$MCP_NAME/g" "$MCP_DIR/internal/mcpserver/server.go"

# --- Dockerfile ---
cat > "$MCP_DIR/Dockerfile" <<EOF
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o $MCP_NAME .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/$MCP_NAME .
ENTRYPOINT ["./$MCP_NAME"]
EOF

# --- mcp-config.json ---
cat > "$MCP_DIR/mcp-config.json" <<EOF
{
  "$MCP_NAME": {
    "command": "docker",
    "args": ["run", "--rm", "-i", "$MCP_NAME"]
  },
  "$MCP_NAME-stdio": {
    "command": "/Users/farquaad/agents-skills/mcps/$MCP_NAME/$MCP_NAME",
    "args": []
  }
}
EOF

# --- AGENTS.md ---
if [ -n "$DESCRIPTION" ]; then
  cat > "$MCP_DIR/AGENTS.md" <<EOF
# $MCP_NAME

$DESCRIPTION

## How it works

TODO: Describe what this MCP does and how it works.

## Tools

- \`example_tool\` — TODO: Replace with your tool description.

## Build

\`\`\`bash
go build -o $MCP_NAME .
\`\`\`

## Docker

\`\`\`bash
docker build -t $MCP_NAME .
\`\`\`

## Run

\`\`\`bash
./$MCP_NAME
\`\`\`

## Key files

| File | Purpose |
|---|---|
| \`main.go\` | Entry point |
| \`internal/mcpserver/server.go\` | MCP server setup, tool handlers, validation |
| \`Dockerfile\` | Multi-stage Go build → Alpine runtime |
| \`mcp-config.json\` | MCP config snippet |
EOF
else
  cat > "$MCP_DIR/AGENTS.md" <<EOF
# $MCP_NAME

TODO: Description of what this MCP does.

## How it works

TODO: Describe what this MCP does and how it works.

## Tools

- \`example_tool\` — TODO: Replace with your tool description.

## Build

\`\`\`bash
go build -o $MCP_NAME .
\`\`\`

## Docker

\`\`\`bash
docker build -t $MCP_NAME .
\`\`\`

## Run

\`\`\`bash
./$MCP_NAME
\`\`\`

## Key files

| File | Purpose |
|---|---|
| \`main.go\` | Entry point |
| \`internal/mcpserver/server.go\` | MCP server setup, tool handlers, validation |
| \`Dockerfile\` | Multi-stage Go build → Alpine runtime |
| \`mcp-config.json\` | MCP config snippet |
EOF
fi

# --- README.md ---
if [ -n "$DESCRIPTION" ]; then
  cat > "$MCP_DIR/README.md" <<EOF
# $MCP_NAME

$DESCRIPTION

## Architecture

\`\`\`
Agent (MCP client over stdio)
  │
  ▼
$MCP_NAME (Go binary)
  └── example_tool
\`\`\`

## MCP Tools

### \`example_tool\`

TODO: Document the tool.

| Param | Type | Required | Description |
|---|---|---|---|
| \`message\` | string | yes | Example input field |

## Build

\`\`\`bash
go build -o $MCP_NAME .
\`\`\`

## Docker

\`\`\`bash
docker build -t $MCP_NAME .
\`\`\`

## Run

\`\`\`bash
./$MCP_NAME
\`\`\`

## MCP Config

Copy \`mcp-config.json\` into your agent's MCP config.

## Dependencies

- \`github.com/modelcontextprotocol/go-sdk/mcp\` — MCP SDK for stdio server
EOF
else
  cat > "$MCP_DIR/README.md" <<EOF
# $MCP_NAME

TODO: Description of what this MCP does.

## Architecture

\`\`\`
Agent (MCP client over stdio)
  │
  ▼
$MCP_NAME (Go binary)
  └── example_tool
\`\`\`

## MCP Tools

### \`example_tool\`

TODO: Document the tool.

| Param | Type | Required | Description |
|---|---|---|---|
| \`message\` | string | yes | Example input field |

## Build

\`\`\`bash
go build -o $MCP_NAME .
\`\`\`

## Docker

\`\`\`bash
docker build -t $MCP_NAME .
\`\`\`

## Run

\`\`\`bash
./$MCP_NAME
\`\`\`

## MCP Config

Copy \`mcp-config.json\` into your agent's MCP config.

## Dependencies

- \`github.com/modelcontextprotocol/go-sdk/mcp\` — MCP SDK for stdio server
EOF
fi

# --- Initialize git repo ---
cd "$MCP_DIR"
git init
git add -A
git commit -m "scaffold $MCP_NAME MCP server"

# --- Summary ---
echo ""
echo "Created MCP '$MCP_NAME' at $MCP_DIR"
echo "  .gitignore"
echo "  AGENTS.md"
echo "  Dockerfile"
echo "  README.md"
echo "  go.mod"
echo "  main.go"
echo "  mcp-config.json"
echo "  internal/mcpserver/server.go"
echo ""
echo "Git repo initialized with initial commit."
echo ""
echo "Next steps:"
echo "  1. cd $MCP_DIR && go mod tidy"
echo "  2. Implement your tools in internal/mcpserver/server.go"
echo "  3. Update AGENTS.md and README.md"
echo "  4. Create a GitHub repo and push:"
echo "     git remote add origin git@github.com:LordFarquaadtheCreator/$MCP_NAME.git"
echo "     git push -u origin main"
echo "  5. Add as submodule to agents-skills:"
echo "     cd $REPO_ROOT"
echo "     rm -rf mcps/$MCP_NAME"
echo "     git submodule add git@github.com:LordFarquaadtheCreator/$MCP_NAME.git mcps/$MCP_NAME"
echo "  6. Update mcps/AGENTS.md with a row for $MCP_NAME"
