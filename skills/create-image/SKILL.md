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
| `positive_prompt` | yes | — | What to generate. Be descriptive — style, subject, composition, lighting. |
| `lora_filename_1` / `_2` / `_3` | yes | — | From list_loras. Must exist on ComfyUI volume. Unused slots = any LoRA at 0.0 |
| `lora_strength_1` / `_2` / `_3` | yes | — | 0.0–1.0. Use recommended_strength as starting point. |
| `negative_prompt` | no | `""` | What to avoid (e.g. "blurry, low quality, deformed") |
| `seed` | no | random | Fixed seed for reproducibility. For repeat>1, increments by 10 per iteration. |
| `steps` | no | 16 | Sampling steps. Higher = more detail, slower. |
| `width` / `height` | no | 720 / 1024 | Image dimensions in pixels |
| `repeat` | no | 1 | Generate N images with incrementing seeds. Filenames get `_v2`, `_v3` suffixes. |
| `output_filename` | no | auto | Custom filename (without extension). `.png` is forced. |
| `output_mode` | no | `file` | `file` (save to disk), `base64` (return inline as ImageContent), `both` |
| `output_dir` | no | `/tmp/create-image` | Directory for saved images when mode is file/both |

## Timing

60-90s for standard 720x1024. 2+ minutes for larger. Wait accordingly — don't retry mid-generation.

## Output

Saved to `/tmp/create-image/<filename>.png` (configurable via `output_dir`). In `base64` or `both` mode, also returned inline as ImageContent.

## Tips

- Start with `recommended_strength` from list_loras, adjust from there
- For single LoRA: fill slots 2 and 3 with any LoRA at strength 0.0
- Use `repeat` for batch generation — each gets incrementing seed
