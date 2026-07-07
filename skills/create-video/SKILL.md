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
| `image_base64` | yes | — | Base64-encoded PNG to animate. Generate one first with create-image MCP. |
| `prompt` | yes | — | Motion prompt (e.g. "slow smile, gentle head turn, hair blowing in wind") |
| `negative_prompt` | no | `""` | What to avoid (e.g. "blurry, jittery, distorted") |
| `seed` | no | 42 | Random seed for reproducibility |
| `num_frames` | no | 121 | 121 ≈ 5s at 24fps |
| `guidance_scale` | no | 3.0 | Classifier-free guidance. Higher = more adherence to prompt. 1.0 = no guidance. |
| `num_inference_steps` | no | 30 | Denoising steps. 30 for full model, 8 for distilled. |
| `output_filename` | no | auto | Custom filename (without extension). `.mp4` is forced. |
| `output_mode` | no | `file` | `file` (save to disk), `base64` (return inline as VideoContent), `both` |
| `output_dir` | no | `./output.private/mcp_output` | Directory for saved videos when mode is file/both |
| `poll_interval_ms` | no | 5000 | Polling interval in milliseconds |
| `poll_timeout_sec` | no | 600 | Max time to wait for generation in seconds |

## Timing

Async: submit → poll until ready. H100 GPU. 121 frames at 768x512 takes ~60-120s. Cold start downloads ~46GB on first run — may take minutes. Default poll timeout 600s (10 min) — increase if cold start expected.

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
