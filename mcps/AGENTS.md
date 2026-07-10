# MCPs

All MCP servers are defined here. Each MCP lives in its own directory. This repo (`agents-skills/mcps/`) is the **canonical home** for every MCP — other repos (e.g. `senor-modal-apps`) symlink to these directories rather than tracking separate submodule pointers.

## Available MCPs

| MCP | Description |
|---|---|
| [`create-image`](create-image/) | Image generation via Modal ComfyUI. Exposes `list_loras`, `list_base_models`, and `generate_image` tools. Canonical repo — `senor-modal-apps/create-image` symlinks here. |
| [`create-video`](create-video/) | Video generation via Modal LTX-2.3. Exposes `generate_video` tool (image → MP4). Canonical repo — `senor-modal-apps/create-video` symlinks here. |
| [`cover-letter-writter`](cover-letter-writter/) | Styled PDF cover letter generation with profile CRUD and history. Exposes 7 tools. |
| [`create-story`](create-story/) | Illustrated PDF + PNG generation from image file paths + markdown text. Exposes `generate_story_pdf` tool. Renders pages via gg, outputs to ~/Desktop/<title>/. Stateless. |
| [`story-tools`](story-tools/) | RAG system for story generation. Scrapes blogs (Blogger, WordPress, DeviantArt), embeds into local vector store, generates stories via LLM. Exposes 9 tools: `generate_story`, `search_documents`, `scrape_blogger`, `scrape_wordpress`, `scrape_deviantart`, `process_documents`, `load_vector_db`, `save_vector_db`, `list_vector_dbs`. Requires `config.yaml` (gitignored) with LLM provider config. |
| [`manage-job`](manage-job/) | Job application tracking via Google Sheets backend. Exposes 4 tools: `track_job`, `get_jobs`, `patch_job`, `delete_job`. Proxies to deployed Apps Script web app. Requires `SHEETS_DEPLOYMENT_ID` env var. |
| [`deviantart-mcp`](deviantart-mcp/) | **[IN PROGRESS — not ready]** DeviantArt API integration — browse, search, galleries, collections, messages. |
| [`resume-builder`](resume-builder/) | One-page PDF resume generation with vector-search-based content selection. Exposes 5 tools: `set_embedding_config`, `init_resume`, `get_resume_info`, `search_resume`, `generate_resume`. No LLM dependency — only needs an embedding endpoint. |

## Structure

```
mcps/
├── AGENTS.md              # this file
└── <mcp-name>/
    ├── README.md           # comprehensive docs
    ├── AGENTS.md           # agent-facing instructions
    ├── Dockerfile          # builds and runs the MCP server
    └── mcp-config.json     # copy-pastable MCP config entry
```

## Rules

- Each MCP is a git submodule under `mcps/` — this is the canonical copy.
- Other repos that need an MCP should symlink to `~/agents-skills/mcps/<name>` and gitignore the symlink path. Do not track separate submodule pointers in consumer repos.
- Each MCP should have a `README.md` (human-facing), `AGENTS.md` (agent-facing), `Dockerfile`, and `mcp-config.json`
- When adding a new MCP submodule, update this file and `AGENTS.md` at the repo root
