# create-story

MCP server that generates illustrated PDF books from base64 images and text. The agent provides images (base64-encoded PNG or JPEG) and story text per page — the server builds a PDF with image-left, text-right layout and muted background colors extracted from each image.

## Architecture

```
Agent (MCP client over stdio)
  │
  ▼
create-story (Go binary)
  └── generate_story_pdf  ── decodes base64 images, builds PDF, writes to disk
```

Stateless — no profiles, no history, no disk state. The agent provides everything inline.

## MCP Tool

### `generate_story_pdf`

Generate a PDF book from base64 images and text. Each page: image left, story text right, muted background extracted from the image.

| Param | Type | Required | Description |
|---|---|---|---|
| `title` | string | yes | Story title shown in the page footer |
| `pages` | array | yes | Array of pages — each has `image` (base64 PNG/JPEG, no data URI prefix) and `text` |
| `outputDir` | string | no | Directory to save the PDF. Defaults to `~/Downloads`. |
| `filename` | string | no | PDF filename. Defaults to `<title>.pdf`. |
| `fontSize` | number | no | Max body font size in points. Binary-searched down to fit text on page. Defaults to 30. |
| `lightenFactor` | number | no | How muted the background (0.0=original, 1.0=white). Defaults to 0.8. |

Each page object:

| Field | Type | Required | Description |
|---|---|---|---|
| `image` | string | yes | Base64-encoded PNG or JPEG (no `data:image/png;base64,` prefix) |
| `text` | string | yes | Story text. Markdown supported: `**bold**`, `*italic*`, `\n` for line breaks, `\n\n` for paragraph breaks. |

Returns:

```json
{
  "outputPath": "/Users/fahad/Downloads/My Story.pdf",
  "pageCount": 10,
  "filename": "My Story.pdf"
}
```

## Build

```bash
go build -o create-story .
```

## Docker

```bash
docker build -t create-story .
```

## Run

```bash
./create-story
```

No env vars required. The server starts immediately over stdio.

## Tests

```bash
go test ./... -count=1
```

Tests across 2 packages:

| Package | Tests cover |
|---|---|
| `internal/generate` | Empty pages, PDF written, multiple pages, default filename, nested output dir, invalid base64, markdown formatting, custom font size |
| `internal/mcpserver` | Valid generation, missing title, no pages, missing image, missing text, multiple pages |

## MCP Config

Copy `mcp-config.json` into your agent's MCP config.

## Files

| File | Purpose |
|---|---|
| `main.go` | Entry point |
| `internal/mcpserver/server.go` | MCP server setup, tool handler, validation |
| `internal/generate/generate.go` | PDF generation: base64 decode, color extraction, text layout, page rendering |
| `Dockerfile` | Multi-stage Go build → Alpine runtime |
| `mcp-config.json` | MCP config snippet |
