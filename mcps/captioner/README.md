# Captioner

MCP server that programmatically captions images. Given image paths and caption text, outputs composited images with the original image on top and caption text on a black bar below.

## How it works

- Image placed untouched in the top area
- Caption text rendered below on a semi-transparent black bar (235 alpha)
- Font size fixed at 16pt — never scales
- Bar height expands to fit all text, no cap
- Text is centered, word-wrapped to image width minus 30px padding
- Output dimensions: `imageWidth x (imageHeight + barHeight)`

## MCP Tool

**Tool name:** `caption`

**Input:**
```json
{
  "captions": [
    {
      "imgPath": "/absolute/path/to/image.png",
      "text": "caption text here"
    }
  ],
  "outputPath": "/optional/output/dir"
}
```

- `outputPath` is optional. If omitted, defaults to `/tmp/captions/<sha256-hash>` based on input content.

**Output:** PNG files saved to the output directory, preserving original filenames. Tool result includes per-image timing (font load, construct, encode, total).

## Build

```bash
go build -o captioner .
```

## Test

```bash
go test -v ./...
```

Tests generate synthetic images and save captioned outputs to `/tmp/caption_test_inspect/` for manual inspection.

## Dependencies

- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) v1.7.0 — MCP server SDK
- [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) — font rendering and image operations
- Font: `/Library/Fonts/Arial Unicode.ttf` (macOS system font)

## Future

- Theme/styles system (gradients, text effects, bar variations)
- Gradient blend between image and caption bar
- Configurable font family/size
