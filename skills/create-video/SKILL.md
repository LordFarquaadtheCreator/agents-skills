---
name: create-video
description: Generate video from an image via the create-video MCP (Modal LTX-2.3 workflow). Use when Fahad wants to animate, generate video from, or bring motion to an image.
---

# Create Video

Generate video from a source image via the `create-video` MCP server. Backed by Modal-hosted LTX-2.3 Diffusers. Image-to-video only — requires a source PNG (typically from create-image).

## Workflow

1. **Have a source image:** Generate one via create-image MCP, or Fahad provides one. Image is center-cropped to 768x512 landscape.
2. **Generate:** `generate_video` with image (base64 or file path) + motion prompt.

## Tool

Only one tool: `generate_video`

## generate_video params

| Param | Required | Default | Notes |
|---|---|---|---|
| `image_base64` | yes | — | Base64-encoded PNG source image |
| `prompt` | yes | — | Motion description (e.g. "slow pan right, waves crashing") |
| `negative_prompt` | no | `worst quality, inconsistent motion...` | What to avoid |
| `seed` | no | 42 | Reproducibility |
| `num_frames` | no | 121 | 121 ≈ 5s at 24fps |
| `num_inference_steps` | no | 30 | 30 full model, 8 distilled |
| `output_mode` | no | file | `file`, `base64`, `both` |

## Timing

Async: submit → poll until ready. H100 GPU. 121 frames at 768x512 takes ~60-120s. Cold start downloads ~46GB on first run — may take minutes.

## Motion prompt tips

- Describe natural, slow movements: "gentle smile", "hair blowing in wind", "slow camera pan"
- Avoid complex multi-action prompts — LTX-2.3 handles single motions best
- Source image style/subject is preserved — match motion to image content

## Output

MP4 saved to disk (file mode) or returned inline (base64 mode).

## Pipeline

```
create-image (LoRA-styled PNG) → create-video (animate PNG → MP4)
```

Use create-image first to get a styled image, then create-video to animate it.
