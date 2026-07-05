---
name: create-yt-summary
description: Summarize YouTube videos by fetching their transcript and running it through an LLM (LM Studio / OpenAI-compatible API). Use this when the user wants to summarize a YouTube video.
---

# create-yt-summary

This skill wraps a CLI tool that reads a YouTube URL from the macOS clipboard, fetches the video transcript via yt-dlp, and generates a summary using an LLM (LM Studio or any OpenAI-compatible API).

## Prerequisites

- macOS (uses `pbpaste` for clipboard access)
- Conda (miniconda/anaconda) installed at `/opt/homebrew/anaconda3/`
- A `config.yaml` file (see below)

## Config

The tool expects a `config.yaml` in the working directory. If one doesn't exist, create it with these keys:

```yaml
output_dir: "~/Documents/yt-summaries"
llm:
  base_url: "http://localhost:1234/v1"
  api_key: "lm-studio"
  model: "local-model"
prompt: "Summarize the following YouTube video transcript. Include the main points and key takeaways."
```

- `output_dir` — where the summary markdown file is saved
- `llm.base_url` — LM Studio or OpenAI-compatible API base URL
- `llm.api_key` — API key (for LM Studio, use any non-empty string)
- `llm.model` — model name to use for summarization
- `prompt` — system prompt sent to the LLM

## Usage

### Step 1: Copy a YouTube URL to clipboard

The user must copy a YouTube URL (watch, shorts, or youtu.be) to the clipboard.

### Step 2: Run the tool

```bash
cd <working-directory-with-config.yaml> && bash repo/run.sh
```

The script:
1. Creates/updates a conda environment named `create-yt-summary`
2. Reads the YouTube URL from the clipboard
3. Fetches the English transcript
4. Sends it to the configured LLM for summarization
5. Writes the summary to `output_dir/<video-title>.md`
6. Prints the output file path on success

### Step 3: Report the result

Tell the user the path of the generated summary file. If the tool fails, report the error message and suggest checking the config or clipboard contents.

## Notes

- The tool only works with English captions (manual or auto-generated)
- Transcripts are truncated to 12,000 characters before sending to the LLM
- The conda environment is created on first run and updated on subsequent runs
