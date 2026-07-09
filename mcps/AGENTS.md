# MCPs

All MCP servers are defined here. Each MCP lives in its own directory. This repo (`agents-skills/mcps/`) is the **canonical home** for every MCP — other repos (e.g. `senor-modal-apps`) symlink to these directories rather than tracking separate submodule pointers.

## Available MCPs

| MCP | Description |
|---|---|
| [`create-image`](create-image/) | Image generation via Modal ComfyUI. Exposes `list_loras`, `list_base_models`, and `generate_image` tools. Canonical repo — `senor-modal-apps/create-image` symlinks here. |
| [`create-video`](create-video/) | Video generation via Modal LTX-2.3. Exposes `generate_video` tool (image → MP4). Canonical repo — `senor-modal-apps/create-video` symlinks here. |
| [`cover-letter-writter`](cover-letter-writter/) | Styled PDF cover letter generation with profile CRUD and history. Exposes 7 tools. |
| [`create-story`](create-story/) | Illustrated PDF + PNG generation from image file paths + markdown text. Exposes `generate_story_pdf` tool. Renders pages via gg, outputs to ~/Desktop/<title>/. Stateless. |
| [`deviantart-mcp`](deviantart-mcp/) | **[IN PROGRESS — not ready]** DeviantArt API integration — browse, search, galleries, collections, messages. |

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
