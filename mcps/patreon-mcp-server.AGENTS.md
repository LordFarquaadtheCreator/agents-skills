# patreon-mcp-server

MCP server for the Patreon creator API. Read-only access to your campaigns, members, and posts from any MCP-compatible client. Upstream: https://github.com/KyuRish/patreon-mcp-server

## How it works

stdio-based MCP server written in Python (FastMCP). Requires a Patreon **Creator Access Token** (your own campaign data only). Six tools, all read-only:

- `fetch_identity` — authenticated user profile
- `fetch_campaigns` — all your campaigns
- `fetch_campaign` — single campaign with tier breakdown
- `fetch_members` — paginated patron list (100/page, use `next_cursor`)
- `fetch_posts` — paginated post list (20/page, use `next_cursor`)
- `fetch_post` — single post by ID

No patron emails, no creator notes, no writes, no caching. Patron data is sent to whatever AI provider the agent runs on — Fahad is responsible for compliance with Patreon's Creator Privacy Promise.

## Required env vars

- `PATREON_ACCESS_TOKEN` — Creator Access Token from https://www.patreon.com/portal/registration/register-clients

Optional:
- `TRANSPORT` — `stdio` (default) or other FastMCP transports

## Build

```bash
uv sync
```

## Run

```bash
PATREON_ACCESS_TOKEN=your_token uv run src/patreon_mcp_server/server.py
```

## MCP Config

Copy `mcp-config.json` into the agent's MCP config. Set `PATREON_ACCESS_TOKEN`.

## Key files

| File | Purpose |
|---|---|
| `src/patreon_mcp_server/server.py` | Entry point, loads `.env`, runs FastMCP |
| `src/patreon_mcp_server/mcp_server.py` | FastMCP instance + PatreonClient init |
| `src/patreon_mcp_server/tools.py` | `@mcp.tool()` definitions |
| `src/patreon_mcp_server/models.py` | Pydantic models + JSON:API parsers |
| `src/patreon_mcp_server/utils/client.py` | PatreonClient HTTP layer |
| `mcp-config.json` | MCP config snippet |
