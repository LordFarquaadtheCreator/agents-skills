# browserclaw-mcp

MCP server wrapping browserclaw. Snapshot + ref targeting for AI agents.

## How it works

stdio MCP server in TypeScript. One BrowserClaw instance per server lifetime, launched on first `new_page` (headless, noSandbox for Docker). Pages tracked by integer id in a Map, mapped to CrawlPage handles (each carries a CDP `targetId`).

```
Agent ──stdio──► MCP Server
                   new_page          ──► BrowserClaw.open(url) → page id
                   snapshot          ──► page.snapshot() → text tree + ref map
                   click/type/fill   ──► ref → Playwright locator → action
                   navigate/scroll   ──► page.goto/back/forward/reload, window.scrollBy
                   close_page        ──► browser.close(targetId)
```

## Architecture

No LLM. No vision. browserclaw provides the snapshot + ref resolution; this server exposes it over MCP. The agent reads the snapshot text, decides which ref to act on, and calls back. Deterministic: same page state → same ref → same element.

`scroll` has no native directional API on CrawlPage, so it uses `page.evaluate('window.scrollBy(dx, dy)')`. `fill` maps directly to browserclaw's `page.fill(FormField[])`.

## Build

```bash
npm install
npm run build
```

## Docker

```bash
docker build -t browserclaw-mcp .
```

## Run

Requires Chromium. Set `PLAYWRIGHT_BROWSERS_PATH` or run `npx playwright install chromium`.

```bash
node dist/index.js
```

## MCP Config

Copy `mcp-config.json` into the agent's MCP config.

## Key files

| File | Purpose |
|---|---|
| `src/index.ts` | MCP server, 10 tool handlers, BrowserClaw lifecycle, graceful shutdown |
| `package.json` | deps + build scripts |
| `Dockerfile` | Node build → slim runtime with Chromium |
| `mcp-config.json` | MCP config snippet |
