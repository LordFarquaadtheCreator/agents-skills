---
name: photocop
description: Copy files dir-to-dir via the photocop MCP, renaming each to YYYY-MM-DD@HH.MM.SS.EXT by mtime. Use when Fahad wants to copy photos/files and rename by timestamp.
---

# PhotoCopy

Copy files from one directory to another, renaming each to `YYYY-MM-DD@HH.MM.SS.EXT` based on its mtime. Backed by the `photocop` MCP server (`photocop mcp` subcommand).

## When to use

- Fahad wants to copy files/photos from one dir to another
- Fahad wants files renamed by their timestamp (mtime)
- Fahad mentions the Photography folder, "to edit", or similar copy+rename workflows

## Workflow

1. **(Optional) Preview:** `copy_files` with `dry_run=true` — see what would happen without copying.
2. **Copy:** `copy_files` with `src` + `dst` — performs the copy.

## Tools

| Tool | When to use |
|---|---|
| `copy_files` | Copy + rename. Set `dry_run=true` first to preview. |

## copy_files params

| Param | Required | Default | Notes |
|---|---|---|---|
| `src` | yes | — | Source directory. `~` expanded. |
| `dst` | yes | — | Destination directory. Created if missing. |
| `dry_run` | no | `false` | Preview without copying. Returns same JSON with `status: "preview"`. |

## Behavior

- Renames each file to `YYYY-MM-DD@HH.MM.SS.EXT` using file mtime, local 24h
- Collisions (same name incl extension) get `_2`, `_3`, ... before extension
- Hidden files (dot-prefixed) skipped
- Files sorted by name for deterministic `_N` assignment
- mtime preserved on copied files (idempotent re-runs)
- Subdirectories inside `src` are skipped — only files copied

## Output

JSON with `copied`, `skipped`, `total`, `dry_run`, and `files[]` (each with `original`, `new_name`, `status`, optional `error`).

## Tips

- Always run `dry_run=true` first if unsure what will happen
- Destination is created if missing — no need to pre-create
- Re-running on same src/dst is safe (mtime preserved, collisions handled)
