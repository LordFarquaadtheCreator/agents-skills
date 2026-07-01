# create-image

MCP server that exposes image generation tools using the Modal ComfyUI generation workflow.

## Files

| File | Purpose |
|---|---|
| `main.go` | Entry point: reads `COMFYUI_API_URL` env var, starts MCP server over stdio. |
| `mcp.go` | MCP server setup, tool handlers (`list_loras`, `generate_image`), HTTP client, file saving. |
| `types.go` | Typed structs: `Request`, `ModelCard`, `LoraEntry`, `BaseModel`, `Defaults`, MCP input/output. |
| `model_card.yaml` | Source of truth for LoRAs and base models on the Modal volume. |
| `Dockerfile` | Builds and runs the MCP server in a container. |
| `mcp-config.json` | Copy-pastable entry for `mcpServers` in `~/.codeium/windsurf/mcp_config.json`. |

## MCP Tools

- `list_loras` — lists all LoRAs from `model_card.yaml` for the Modal ComfyUI generation workflow, with full metadata.
- `generate_image` — generates an image via the Modal ComfyUI generation workflow. Required: `positive_prompt`, `lora_filename_1/2/3`, `lora_strength_1/2/3`. Optional: `negative_prompt`, `seed`, `steps`, `width`, `height`, `repeat`, `output_filename`.

## Build

```bash
go build -o create-image .
```

## Docker

```bash
docker build -t create-image .
```

## Run

The binary is an MCP server over stdio. It requires `COMFYUI_API_URL` env var.

```bash
COMFYUI_API_URL=https://your-modal-api-url ./create-image
```

## Config

Copy `mcp-config.json` into `~/.codeium/windsurf/mcp_config.json` under `mcpServers`. Set `COMFYUI_API_URL` to your Modal ComfyUI generation workflow endpoint.
