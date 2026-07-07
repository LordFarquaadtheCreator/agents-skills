---
name: create-image
description: Generate images via the create-image MCP (Modal ComfyUI workflow). Use when Fahad wants to generate, create, or make an image using LoRAs.
---

# Create Image

Generate images via the `create-image` MCP server. Backed by Modal-hosted ComfyUI with ZImageTurbo + LoRAs.

## Workflow

1. **List LoRAs:** `list_loras` — mandatory before generating. Returns filenames, types, keywords, recommended strengths. Use `keyword` or `type` filters to narrow.
2. **(Optional) List base models:** `list_base_models` — if you need to check available base models.
3. **Generate:** `generate_image` with positive prompt + exactly 3 LoRA slots. All 3 slots required — fill unused with any LoRA at strength 0.0.

## Tools

| Tool | When to use |
|---|---|
| `list_loras` | STEP 1 — always. Discover valid LoRA filenames. |
| `list_base_models` | Optional — check base models by type/architecture |
| `generate_image` | STEP 2 — generate. Requires 3 LoRA slots. |

## generate_image params

| Param | Required | Default | Notes |
|---|---|---|---|
| `positive_prompt` | yes | — | What to generate |
| `lora_filename_1` / `_2` / `_3` | yes | — | From list_loras. Unused slots = any LoRA at 0.0 |
| `lora_strength_1` / `_2` / `_3` | yes | — | 0.0–1.0. Use recommended_strength as starting point. |
| `negative_prompt` | no | `""` | What to avoid |
| `seed` | no | random | For reproducibility |
| `steps` | no | 16 | Sampling steps |
| `width` / `height` | no | 720 / 1024 | Image dimensions |
| `repeat` | no | 1 | Batch with incrementing seeds |
| `output_filename` | no | `mcp_output` | `.png` appended, `_vN` for repeats |

## Timing

60-90s for standard 720x1024. 2+ minutes for larger. Wait accordingly — don't retry mid-generation.

## Output

Saved to `./output/mcp_output/<filename>.png` relative to the MCP server's cwd.

## Tips

- Start with `recommended_strength` from list_loras, adjust from there
- For single LoRA: fill slots 2 and 3 with any LoRA at strength 0.0
- Use `repeat` for batch generation — each gets incrementing seed
