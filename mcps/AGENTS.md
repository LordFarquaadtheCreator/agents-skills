# MCPs

All MCP servers are defined here. Each MCP lives in its own directory. This repo (`agents-skills/mcps/`) is the **canonical home** for every MCP — other repos (e.g. `senor-modal-apps`) symlink to these directories rather than tracking separate submodule pointers.

## Available MCPs

| MCP | Description |
|---|---|
| [`create-image`](create-image/) | Image generation via Modal ComfyUI. Exposes `list_loras`, `list_base_models`, and `generate_image` tools. Canonical repo — `senor-modal-apps/create-image` symlinks here. |
| [`create-video`](create-video/) | Video generation via Modal LTX-2.3. Exposes `generate_video` tool (image → MP4). Canonical repo — `senor-modal-apps/create-video` symlinks here. |
| [`browserclaw-mcp`](browserclaw-mcp/) | MCP server wrapping browserclaw for AI agent browser automation. Snapshot + ref targeting, no vision model. Exposes 10 tools: `new_page`, `list_pages`, `snapshot`, `click`, `type`, `fill`, `press`, `scroll`, `navigate`, `close_page`. Node + Playwright + Chromium. |
| [`cover-letter-writter`](cover-letter-writter/) | Styled PDF cover letter generation with profile CRUD and history. Exposes 7 tools. |
| [`create-story`](create-story/) | Illustrated PDF + PNG generation from image file paths + markdown text. Exposes `generate_story_pdf` tool. Renders pages via gg, outputs to ~/Desktop/<title>/. Stateless. |
| [`story-tools`](story-tools/) | RAG document retrieval system. Scrapes blogs (Blogger, WordPress, DeviantArt), embeds into local vector store, retrieves documents by semantic query. No story generation — retrieval only. Exposes 7 tools: `search_documents`, `scrape_blogger`, `scrape_wordpress`, `scrape_deviantart`, `process_documents`, `load_vector_db`, `save_vector_db`, `list_vector_dbs`. Requires `config.yaml` (gitignored) with LLM provider config. |
| [`manage-job`](manage-job/) | Job application tracking via Google Sheets backend. Exposes 4 tools: `track_job`, `get_jobs`, `patch_job`, `delete_job`. Proxies to deployed Apps Script web app. Requires `SHEETS_DEPLOYMENT_ID` env var. |
| [`deviantart-mcp`](deviantart-mcp/) | **[IN PROGRESS — not ready]** DeviantArt API integration — browse, search, galleries, collections, messages. |
| [`resume-builder`](resume-builder/) | One-page PDF resume generation with vector-search-based content selection. Exposes 5 tools: `set_embedding_config`, `init_resume`, `get_resume_info`, `search_resume`, `generate_resume`. No LLM dependency — only needs an embedding endpoint. |
| [`photocop`](photocop/) | Copy files dir-to-dir, renaming each to `YYYY-MM-DD@HH.MM.SS.EXT` by mtime. Single binary: `photocop copy` (CLI) + `photocop mcp` (MCP server). Exposes `copy_files` tool with `dry_run` preview. `_N` collision suffix, hidden files skipped, mtime preserved. |
| [`captioner`](captioner/) | Image captioning via Go + opentype. Exposes `caption` tool — composites image with caption text on black bar below. Font fixed 16pt, bar grows to fit. macOS-only (uses `/Library/Fonts/Arial Unicode.ttf`). |
| [`patreon-mcp-server`](patreon-mcp-server/) | Patreon creator API MCP (read-only). Six tools: `fetch_identity`, `fetch_campaigns`, `fetch_campaign`, `fetch_members`, `fetch_posts`, `fetch_post`. Python + FastMCP. Requires `PATREON_ACCESS_TOKEN`. Upstream: KyuRish/patreon-mcp-server. **Exception** — third-party repo, not forked. `AGENTS.md` and `mcp-config.json` live as siblings (`patreon-mcp-server.AGENTS.md`, `patreon-mcp-server.mcp-config.json`) tracked in this parent repo, not inside the submodule. |
| [`pawchive-mcp`](pawchive-mcp/) | Read-only MCP wrapping pawchive.pw's public API. 14 tools covering creators, posts, comments, revisions, flag checks, hash search, app version. In-memory cache with 15min TTL. Go + modelcontextprotocol/go-sdk. No auth, no mutations. |

## Structure

```
mcps/
├── AGENTS.md              # this file
└── <mcp-name>/
    ├── README.md           # comprehensive docs
    ├── AGENTS.md           # agent-facing instructions
    └── mcp-config.json     # copy-pastable MCP config entry
```

## Rules

- Each MCP is a git submodule under `mcps/` — this is the canonical copy.
- Other repos that need an MCP should symlink to `~/agents-skills/mcps/<name>` and gitignore the symlink path. Do not track separate submodule pointers in consumer repos.
- Each MCP should have a `README.md` (human-facing), `AGENTS.md` (agent-facing), and `mcp-config.json`
- When adding a new MCP submodule, update this file and `AGENTS.md` at the repo root
