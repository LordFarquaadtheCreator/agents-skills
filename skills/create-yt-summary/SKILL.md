---
name: create-yt-summary
description: Summarize YouTube videos by fetching their transcript via yt-dlp and running it through an LLM (LM Studio / OpenAI-compatible API). Use this when the user wants to summarize or understand a YouTube video.
---

# create-yt-summary

This skill provides a Go CLI that takes a YouTube URL, fetches the video transcript via yt-dlp, and prints an LLM-generated summary to stdout.

## Prerequisites

- **yt-dlp** installed and on `$PATH` (`brew install yt-dlp`)
- Go toolchain (for building)

## Build

```bash
cd create-yt-summary && go build -o create-yt-summary .
```

## Config (environment variables)

| Variable | Required | Description |
|----------|----------|-------------|
| `LLM_MODEL` | yes | Model name (e.g. `gpt-4o`, `llama-3`) |
| `LLM_BASE_URL` | yes | API base URL (e.g. `http://localhost:1234`) — code appends `/v1/chat/completions` |
| `LLM_API_KEY` | no | API key; omitted from request if not set |
| `SUMMARY_PROMPT` | no | System prompt for summarization |

## Usage

```bash
create-yt-summary/create-yt-summary "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```

The summary is printed to stdout as markdown (video title as H1). Errors go to stderr.

## Notes

- **English only** — the code hardcodes `"en"` subtitles (manual first, auto-generated as fallback)
- **OpenAI API format only** — calls `/chat/completions`; Anthropic, Gemini, etc. need a compatible proxy
- Requires `yt-dlp` on `$PATH` — all transcript fetching delegates to it
- Pure stdlib Go — zero external dependencies beyond the Go toolchain and yt-dlp
