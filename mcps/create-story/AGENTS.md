# create-story

MCP server for generating illustrated PDF books with per-page PNG images from image files and text.

## What it does

Takes a title and array of pages (each with an image file path + markdown text), renders each page as a PNG using `fogleman/gg`, then embeds those PNGs into a PDF via `go-pdf/fpdf`. Output goes to `~/Desktop/<title>/` — contains `<title>.pdf` and `<title>.<n>.png` per page. Collision handling appends ` 2`, ` 3`, etc. to the directory name.

Stateless — no profiles, no history, no disk state.

## Architecture

```
Agent ──stdio──► MCP Server (internal/mcpserver)
                   │
                   ▼
                 generate.Run
                   │
                   ├── load fonts (Arial TTF from system)
                   ├── resolve output dir (~/Desktop/<title>/ with collision handling)
                   ├── per page:
                   │     ├── decode image from file path
                   │     ├── extract dominant color → lighten for background
                   │     ├── render page via gg (image left, text right, footer)
                   │     ├── save PNG to <title>.<n>.png
                   │     └── embed PNG into PDF
                   └── write PDF to <title>.pdf
```

No external API dependencies. All rendering in-process.

## Directory structure

| Path | Purpose |
|---|---|
| `main.go` | Entry point — calls `mcpserver.Run()` |
| `internal/` | Internal packages — see [`internal/AGENTS.md`](internal/AGENTS.md) |
| `internal/generate/` | Page rendering, PDF/PNG generation — see [`internal/generate/AGENTS.md`](internal/generate/AGENTS.md) |
| `internal/mcpserver/` | MCP stdio server, tool registration, input validation — see [`internal/mcpserver/AGENTS.md`](internal/mcpserver/AGENTS.md) |
| `Dockerfile` | Multi-stage Go build → Alpine runtime |
| `mcp-config.json` | MCP config snippet for agent integration |

## MCP Tool

### `generate_story_pdf`

| Param | Type | Required | Description |
|---|---|---|---|
| `title` | string | yes | Story title. Used as output dir name, PDF filename, and footer label. |
| `pages` | array | yes | Array of pages — each has `image` (absolute file path to PNG/JPEG) and `text` (markdown) |
| `outputDir` | string | no | Base directory. A `<title>/` subdir is created inside. Defaults to `~/Desktop`. |
| `fontSize` | number | no | Max body font size in points. Binary-searched down to fit. Defaults to 30. |
| `lightenFactor` | number | no | Background muting (0.0=original, 1.0=white). Defaults to 0.8. |

Each page:

| Field | Type | Required | Description |
|---|---|---|---|
| `image` | string | yes | Absolute file path to a PNG or JPEG image |
| `text` | string | yes | Story text. Markdown: `**bold**`, `*italic*`, `\n` line breaks, `\n\n` paragraph breaks |

Returns:

```json
{
  "outputDir": "/Users/fahad/Desktop/My Story",
  "pdfPath": "/Users/fahad/Desktop/My Story/My Story.pdf",
  "pngPaths": ["/Users/fahad/Desktop/My Story/My Story.1.png"],
  "pageCount": 1
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

No env vars required. Stdio transport.

## Tests

```bash
go test ./... -count=1 -v
```

| Package | Tests cover |
|---|---|
| `internal/generate` | Empty pages, PDF+PNG output, multiple pages, title subdir, collision handling, invalid image path, markdown, custom font size, ASCII sanitization, nested output dir |
| `internal/mcpserver` | Valid generation, missing title, no pages, missing image, missing text, multiple pages |

## Dependencies

- `github.com/fogleman/gg` — 2D graphics rendering for PNG output
- `github.com/go-pdf/fpdf` — PDF assembly (embeds rendered PNGs)
- `github.com/golang/freetype/truetype` — TTF font parsing for text rendering
- `golang.org/x/image/draw` — high-quality image scaling
- `github.com/modelcontextprotocol/go-sdk/mcp` — MCP SDK for stdio server

macOS only — loads Arial TTF fonts from `/System/Library/Fonts/Supplemental/`.
