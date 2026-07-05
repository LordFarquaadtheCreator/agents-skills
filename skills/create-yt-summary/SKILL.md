---
name: create-yt-summary
description: Summarize YouTube videos by fetching their transcript via yt-dlp and running it through an LLM (LM Studio / OpenAI-compatible API). Use this when the user wants to summarize or understand a YouTube video.
---

# create-yt-summary

This skill provides a Go CLI that takes a YouTube URL, fetches the video transcript via yt-dlp, and prints an LLM-generated summary to stdout.

## Prerequisites

- `yt-dlp` installed and on `$PATH` (`brew install yt-dlp`)

## Build

```bash
cd repo && go build -o create-yt-summary .
```

## Config (environment variables)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `LLM_MODEL` | yes | — | Model name (e.g. `gpt-4o`, `llama-3`) |
| `LLM_BASE_URL` | no | `http://localhost:1234/v1` | API base URL |
| `LLM_API_KEY` | no | `lm-studio` | API key |
| `SUMMARY_PROMPT` | no | `"Summarize the following YouTube video transcript..."` | System prompt |

## Usage

### Primary: pass URL as argument

```bash
repo/create-yt-summary "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```

The summary is printed to stdout as markdown (with the video title as an H1).

### Fallback: URL from clipboard (macOS)

If no argument is given, it reads the clipboard via `pbpaste`:

```bash
repo/create-yt-summary
```

### As an agent

Just build the binary once, then invoke it with the URL. Read the stdout for the summary and present it to the user. If it fails, read stderr for the error.

## Notes

- Only English captions are supported (manual or auto-generated)
- Transcripts are truncated to 12,000 characters before the LLM call
- Pure stdlib Go — zero dependencies beyond the Go toolchain and `yt-dlp`
