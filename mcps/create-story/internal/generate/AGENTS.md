# generate

Core rendering and PDF/PNG generation package.

## What it does

- Loads Arial TTF fonts (regular, bold, italic) from system paths
- Decodes images from file paths (PNG/JPEG)
- Extracts dominant color from image, lightens it for page background
- Renders each page as a 2400x1200 PNG via `fogleman/gg`:
  - Scaled image on left half
  - Markdown text on right half with binary-searched font fitting
  - Footer: `<title> #<n>`
- Saves PNGs to `<outDir>/<name>.<n>.png`
- Embeds PNGs into PDF via `go-pdf/fpdf`
- Handles output directory creation with collision detection

## Key types

| Type | Purpose |
|---|---|
| `Input` | Tool input — title, pages, outputDir, fontSize, lightenFactor |
| `Page` | One page — image file path + text |
| `Output` | Result — outputDir, pdfPath, pngPaths, pageCount |
| `fontFamily` | Loaded TTF fonts for regular/bold/italic faces |

## Key functions

| Function | Purpose |
|---|---|
| `Run` | Entry point. Orchestrates font loading, dir resolution, per-page render, PDF write |
| `renderPageImage` | Renders one page to `image.Image` using `gg` |
| `flattenWords` | Splits text into words with line/paragraph break markers |
| `parseMarkdown` | Parses `**bold**` and `*italic*` into segments with face info |
| `getDominantColorFromImage` | Quantizes image colors, picks most common muted tone |
| `sanitizeASCII` | Replaces Unicode punctuation with ASCII, strips non-ASCII |
| `resolveOutputDir` | Finds available dir name, handles collisions (`Title 2`, `Title 3`, ...) |
| `sanitizeFilename` | Strips `/`, `:`, `\`, null from strings for filesystem safety |

## Text processing pipeline

```
raw text
  → sanitizeASCII (Unicode → ASCII)
  → split on \n\n (paragraphs)
  → split on \n (lines)
  → parseMarkdown per line (**bold**, *italic*)
  → strings.Fields (word splitting)
  → []pageItem (word + face + break type)
```

## Binary search font fitting

Starts at `maxFontSize` (default 60px at 2x scale). Binary searches down to find the largest size where all text fits within `availH` (page height minus padding and footer). Min size is 12px. Line height is `1.28 * fontSize`, paragraph padding is `0.4 * lineHeight`.

## Files

| File | Purpose |
|---|---|
| `generate.go` | All rendering, text processing, color extraction, PDF assembly |
| `generate_test.go` | Tests for Run, markdown, ASCII sanitization, collision handling, PNG output |
