---
name: create-yt-summary
description: Fetch YouTube video transcripts via yt-dlp. Supports multi-language fallback, text/SRT/VTT/JSON3 output, playlists, search, metadata extraction, and file-based caching. Use this when the user wants to get or read a YouTube video transcript.
---

# create-yt-summary

Go CLI that wraps yt-dlp to fetch YouTube transcripts. Outputs plain text, SRT, VTT, or JSON3. Includes metadata (title, channel, duration, chapters, etc.) and a 7-day file cache.

## Prerequisites

- **yt-dlp** installed and on `$PATH` (`brew install yt-dlp`)
- Go toolchain (for building)

## Build

```bash
cd create-yt-summary && go build -o create-yt-summary .
```

## Usage

```bash
# Basic transcript (English, plain text)
create-yt-summary/create-yt-summary "https://www.youtube.com/watch?v=..."

# Multi-language fallback
create-yt-summary/create-yt-summary -lang de,en "https://www.youtube.com/watch?v=..."

# SRT format
create-yt-summary/create-yt-summary -format srt "https://www.youtube.com/watch?v=..."

# List available subtitle languages
create-yt-summary/create-yt-summary -list-subs "https://www.youtube.com/watch?v=..."

# Playlist (one transcript per video)
create-yt-summary/create-yt-summary "https://www.youtube.com/playlist?list=..."

# Search (top N results)
create-yt-summary/create-yt-summary "ytsearch5:golang tutorial"

# JSON output with metadata + content
create-yt-summary/create-yt-summary -json "https://www.youtube.com/watch?v=..."

# Auth-gated content
create-yt-summary/create-yt-summary -cookies cookies.txt "https://www.youtube.com/watch?v=..."

# Clear cache
create-yt-summary/create-yt-summary -clear-cache
```

## Flags

| Flag | Default | Description |
|----------|----------|-------------|
| `-lang` | `en` | Comma-separated subtitle languages (fallback order) |
| `-format` | `text` | Output format: text, srt, vtt, json3 |
| `-cookies` | | Path to cookies.txt for auth-gated content |
| `-cache-dir` | `$HOME/.cache/yt-transcript` | Cache directory |
| `-no-cache` | `false` | Disable cache |
| `-list-subs` | `false` | List available subtitle languages and exit |
| `-metadata` | `false` | Output metadata as JSON to stderr |
| `-json` | `false` | Output results as JSON array (includes metadata) |
| `-clear-cache` | `false` | Clear cache directory and exit |

## Notes

- Supports any site yt-dlp supports — pass any URL
- Search uses yt-dlp's `ytsearchN:query` syntax
- Cache key: `videoID_lang_format`, TTL: 7 days
- Metadata includes: id, title, channel, duration, upload_date, description, chapters
- Pure stdlib Go — zero external dependencies beyond Go toolchain and yt-dlp
