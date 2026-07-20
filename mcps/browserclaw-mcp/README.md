# browserclaw-mcp

MCP server wrapping [browserclaw](https://github.com/idan-rubin/browserclaw) for AI agent browser automation. Snapshot + ref targeting, no vision model in the loop.

## How it works

stdio-based MCP server in TypeScript. Spawns a headless Chromium via Playwright + browserclaw on first `new_page` call. Exposes 10 tools for page management, snapshot, and ref-based interaction.

```
Agent ──stdio──► MCP Server
                   new_page          ──► BrowserClaw.open(url) → page id
                   snapshot          ──► page.snapshot() → text tree + ref map
                   click/type/fill   ──► ref → Playwright locator → action
                   navigate/scroll   ──► page navigation / window.scrollBy
                   close_page        ──► browser.close(targetId)
```

## Tools

| Tool | Purpose |
|---|---|
| `new_page` | Open a tab at url, returns integer page id |
| `list_pages` | List open tabs (browser.tabs + our id map) |
| `snapshot` | Accessibility tree + ref map for a page |
| `click` | Click element by ref |
| `type` | Type text into element by ref, optional `submit`/`slowly` |
| `fill` | Fill multiple form fields: `[{ref, type, value}]` |
| `press` | Press key combo, optional ref to focus first |
| `scroll` | Scroll viewport: up/down/left/right by N pixels |
| `navigate` | url / back / forward / reload |
| `close_page` | Close a page by id |

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

Requires Chromium. Playwright installs its own via `npx playwright install chromium`, or set `PLAYWRIGHT_BROWSERS_PATH` to a system Chromium.

```bash
node dist/index.js
```

Browser launches headless with `noSandbox: true` (Docker/CI friendly). Override in `src/index.ts` if needed.

## MCP Config

Copy `mcp-config.json` into the agent's MCP config.

## Key files

| File | Purpose |
|---|---|
| `src/index.ts` | MCP server, 10 tool handlers, BrowserClaw lifecycle |
| `package.json` | deps: browserclaw, playwright, mcp sdk |
| `Dockerfile` | Node build → slim runtime with Chromium |
| `mcp-config.json` | MCP config snippet |
