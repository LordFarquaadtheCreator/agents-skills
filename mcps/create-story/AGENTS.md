# create-story

MCP server for generating illustrated PDF books from base64 images and text.

## How it works

Stdio-based MCP server written in Go. Exposes one tool:

- `generate_story_pdf` — takes a title and array of pages (each with base64 image + text), builds a PDF with image-left/text-right layout and muted background colors extracted from images, writes to disk.

Stateless — no profiles, no history, no disk state. Agent provides everything inline.

## Architecture

No external API dependencies. PDF generation is in-process via `go-pdf/fpdf`. Base64 images are decoded to temp files, rendered into the PDF, then cleaned up.

```
Agent ──stdio──► MCP Server
                   generate_story_pdf  ── decode base64 → temp files → build PDF → write to disk
```

## Tests

```bash
go test ./... -count=1
```

Tests across 2 packages:

| Package | Tests cover |
|---|---|
| `internal/generate` | Empty pages, PDF written, multiple pages, default filename, nested output dir, invalid base64, markdown formatting, custom font size |
| `internal/mcpserver` | Valid generation, missing title, no pages, missing image, missing text, multiple pages |

## Build

```bash
go build -o create-story .
```

## Docker

```bash
docker build -t create-story .
```

## Run

No env vars required.

```bash
./create-story
```

## MCP Config

Copy `mcp-config.json` into the agent's MCP config.

## Key files

| File | Purpose |
|---|---|
| `main.go` | Entry point |
| `internal/mcpserver/server.go` | MCP server setup, tool handler, input validation |
| `internal/generate/generate.go` | PDF generation: base64 decode, color extraction, text layout, page rendering |
| `Dockerfile` | Multi-stage Go build → Alpine runtime |
| `mcp-config.json` | MCP config snippet |
| `internal/generate/generate_test.go` | PDF generation tests |
| `internal/mcpserver/server_test.go` | Handler integration tests |
