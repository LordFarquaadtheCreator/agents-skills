# Agents

Guidelines for AI agents working on this codebase.

## Project

Go MCP server that captions images. Exposes a single `caption` tool over stdio. Image on top, caption text on black bar below. Font fixed at 16pt, bar height grows to fit text.

## Structure

- `main.go` — MCP server, tool handler, image compositing, word wrap
- `main_test.go` — tests with synthetic images, outputs to `/tmp/caption_test_inspect/`
- `go.mod` — module name `captioner`, Go 1.25.6

## Build & Test

```bash
go build -o captioner .
go test -v ./...
```

## Rules

- Font size is fixed. Never scale it. Bar height adjusts to fit text.
- No bar height cap. All text must be visible.
- Keep the MCP tool input schema as-is: `captions` array + optional `outputPath`.
- `outputPath` defaults to `/tmp/captions/<hash>` when omitted.
- Font path is `/Library/Fonts/Arial Unicode.ttf` (macOS).
- Tests save inspectable output to `/tmp/caption_test_inspect/`.
- No comments unless explaining non-obvious behavior.
- Minimal edits. Follow existing style.
