# comfy

Generic MCP server for the **comfy-headless** Modal app. Every workflow in the shared store becomes its own MCP tool — nothing is hardcoded per workflow: the tool list, each tool's parameters, and the request wiring are all generated from the store files at startup.

```
Agent (MCP client over stdio)
  │
  ▼
comfy (Go binary)
  ├── z_image_turbo_txt2img      ── runs z-image-turbo-txt2img.workflow.json
  ├── flux2_klein_9b_i2i         ── runs flux2-klein-9b-i2i.workflow.json
  ├── flux2_klein_9b_inpaint     ── runs flux2-klein-9b-inpaint.workflow.json
  ├── flux2_klein_9b_pulid       ── runs flux2-klein-9b-pulid.workflow.json
  ├── wan22_5b_ti2v              ── runs wan22-5b-ti2v.workflow.json
  ├── z_image_chord_texture      ── runs z-image-chord-texture.workflow.json
  └── flux2_klein_9b_dataset     ── runs flux2-klein-9b-dataset.workflow.json
        (…all discovered from ~/senor-modal-apps/comfy-headless/workflows)
```

## How it works

1. **Load** — at startup the server reads every `*.workflow.json` in the workflow store (`COMFY_WORKFLOWS_DIR`, default `~/senor-modal-apps/comfy-headless/workflows`). A mis-shaped store (unparsable file, tool-name collision, workflow with no tokens) fails startup loudly.
2. **Generate tools** — each file becomes one MCP tool named from its filename (`z-image-turbo-txt2img.workflow.json` → `z_image_turbo_txt2img`). The input schema is generated from the file:
   - one **parameter per `{{token}}`** — `number` for the standard's numeric tokens (`seed`, `steps`, `cfg`, `width`, `height`, `lora_strength_N`, …), `string` otherwise; tokens where an empty string is legal (`negative_prompt`, `clothes`) are optional and default to `""`;
   - one **path parameter per `LoadImage` input** (`mask`, `ref`, …) — the file is read and uploaded as base64 with the job.
3. **Run** — the tool call fills the template (refusing any unfilled `{{token}}`), POSTs to comfy-headless `/infer-async`, polls `/status`, and returns the outputs: every file is saved under `COMFY_MCP_OUTPUT_DIR` (default `/tmp/comfy-mcp/<tool>-<timestamp>/`), PNG/JPEG/WebP/GIF outputs are also returned inline, and the summary text lists every path.

Adding a workflow to the store needs **zero code changes** — restart the server (or `/reload-mcp` in Hermes) and the new tool appears.

## Build

```bash
go build -o comfy .
```

## Run

```bash
./comfy
```

Env (all optional; defaults target the deployed app):

| Variable | Default | Description |
|---|---|---|
| `COMFY_HEADLESS_URL` | `https://lordfarquaadthecreator--comfy-headless` | comfy-headless base URL |
| `MODAL_REGION` | `us-east` | region the endpoint slugs resolve in |
| `COMFY_WORKFLOWS_DIR` | `~/senor-modal-apps/comfy-headless/workflows` | the workflow store |
| `COMFY_MCP_OUTPUT_DIR` | `/tmp/comfy-mcp` | where run outputs are saved |
| `COMFY_HEADLESS_SUBMIT_URL` + `COMFY_HEADLESS_STATUS_URL` | — | full endpoint overrides (both together), for tests |

## Tests

```bash
go test ./... -count=1
```

Hermetic across 3 packages:

| Package | Tests cover |
|---|---|
| `internal/store` | workflow discovery, tool-name slugging + collisions, token extraction (dedup, embedded tokens), numeric typing (name sets + `lora_strength_N` patterns), allow-empty tokens, LoadImage discovery + parameter dedup/guards, error paths |
| `internal/server` | schema generation (types, required, image params), tool listing, full tool call against a mock comfy-headless (values fill, no raw tokens shipped, base64 image upload, output files + inline images), missing/non-numeric values, missing/unreadable images, headless error propagation |
| `internal/headless` | the comfy-headless client (token fill, submit/poll protocol, base64 inputs, status mapping) |

## MCP Config

```json
{
  "comfy": {
    "command": "/Users/farquaad/agents-skills/mcps/comfy/comfy",
    "args": [],
    "env": {}
  }
}
```

Registered in Hermes as `comfy` (`hermes mcp add comfy --command …`). **Set a per-server `timeout` of 1800s** — renders take minutes and the default 120s per-tool-call timeout cancels them.

## Key files

| File | Purpose |
|---|---|
| `main.go` | entry point: config → store load → stdio MCP server |
| `internal/store/store.go` | store loader: tools, token specs, image inputs |
| `internal/server/server.go` | dynamic tool schemas, argument collection, run + output handling |
| `internal/headless/headless.go` | comfy-headless client (shared lineage with create-image/generate-texture MCPs) |
| `internal/config/config.go` | env resolution |
