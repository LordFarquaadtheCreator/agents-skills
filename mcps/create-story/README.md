# create-story

MCP server that generates illustrated PDF books with per-page PNG images from image files and text. The agent provides absolute file paths to PNG/JPEG images and markdown text per page — the server renders each page as a PNG (image left, text right, muted background extracted from the image), then assembles them into a PDF.

## Architecture

```
Agent (MCP client over stdio)
  │
  ▼
create-story (Go binary)
  └── generate_story_pdf
        ├── load Arial TTF fonts from system
        ├── resolve ~/Desktop/<title>/ (collision handling)
        ├── per page:
        │     ├── decode image from file path
        │     ├── extract dominant color → lighten for background
        │     ├── render via fogleman/gg (image left, text right, footer)
        │     ├── save PNG (<title>.<n>.png)
        │     └── embed PNG into PDF
        └── write PDF (<title>.pdf)
```

Stateless — no profiles, no history, no disk state.

## MCP Tool

### `generate_story_pdf`

| Param | Type | Required | Description |
|---|---|---|---|
| `title` | string | yes | Story title. Used as output dir name, PDF filename, and footer label. |
| `pages` | array | yes | Array of pages — each has `image` (absolute file path) and `text` (markdown) |
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

## MCP Config

Copy `mcp-config.json` into your agent's MCP config.

## Dependencies

- `github.com/fogleman/gg` — 2D graphics rendering
- `github.com/go-pdf/fpdf` — PDF assembly
- `github.com/golang/freetype/truetype` — TTF font parsing
- `golang.org/x/image/draw` — image scaling
- `github.com/modelcontextprotocol/go-sdk/mcp` — MCP SDK

macOS only — loads Arial TTF fonts from `/System/Library/Fonts/Supplemental/`.
