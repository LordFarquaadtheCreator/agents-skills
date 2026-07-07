---
name: update-cover-letter-writter-mcp
description: Update the cover-letter-writter git submodule under mcps/. Use this when the user wants to pull the latest version of the cover-letter-writter MCP server, rebuild it, or sync the submodule to its upstream commit.
---

# Update cover-letter-writter MCP

This skill handles updating the `mcps/cover-letter-writter` git submodule — a cover letter PDF generation MCP server (profile CRUD + generate + history) written in Go.

## Submodule location

```
mcps/cover-letter-writter  →  https://github.com/LordFarquaadtheCreator/cover-letter-writter.git
```

## Before you begin: resolve the repo root

This skill's directory is `<directory>`. Resolve the absolute path to the `agents-skills` repo root:

```bash
REPO_ROOT="$(cd "$(dirname "$(dirname "$(dirname "<directory>")")")" && pwd)"
```

Alternatively, if the skill is symlinked into an agent's skills directory, use git to find the parent repo:

```bash
cd "<directory>/../.." && REPO_ROOT="$(git rev-parse --show-toplevel)"
```

Verify the root looks correct:

```bash
echo "$REPO_ROOT"
ls "$REPO_ROOT"/.gitmodules
```

If `.gitmodules` is not found, the path resolution is wrong — manually confirm where the `agents-skills` repo lives.

## Update steps

Run all commands from `$REPO_ROOT`.

### 1. Pull the latest submodule commit

```bash
cd "$REPO_ROOT" && git submodule update --remote mcps/cover-letter-writter
```

This fetches the latest commit from the submodule's default branch and checks it out.

### 2. Verify the update

```bash
cd "$REPO_ROOT/mcps/cover-letter-writter" && git log --oneline -3
```

### 3. Rebuild (if needed)

**Go binary rebuild:**

```bash
cd "$REPO_ROOT/mcps/cover-letter-writter" && go build -o cover-letter-writter .
```

**Docker rebuild (if using Docker):**

```bash
cd "$REPO_ROOT/mcps/cover-letter-writter" && docker build -t cover-letter-writter .
```

### 4. Run tests (recommended)

```bash
cd "$REPO_ROOT/mcps/cover-letter-writter" && go test ./... -count=1
```

### 5. Commit the submodule change

```bash
cd "$REPO_ROOT" && git add mcps/cover-letter-writter && git commit -m "chore: update cover-letter-writter submodule"
```

## Sync all repos that reference this submodule

The `cover-letter-writter` submodule may be referenced from multiple repos. After pushing changes to the `cover-letter-writter` source repo, update the submodule pointer in every repo that includes it:

1. **`agents-skills`** — `mcps/cover-letter-writter` (the canonical skill repo)

For each repo:

```bash
cd <repo_root> && git submodule update --remote <submodule_path>
git add <submodule_path> && git commit -m "update cover-letter-writter submodule"
git push origin main
```

If you made changes directly inside the submodule working directory (not via `update --remote`), commit and push from inside the submodule first, then update the pointer in each parent repo.

## Troubleshooting

- If the submodule isn't initialized: `cd "$REPO_ROOT" && git submodule init mcps/cover-letter-writter`
- If the submodule is on a detached HEAD, that's normal after `update --remote`
- If `mcp-config.json` changed, let the user know — they may need to update their local MCP config
- If changes were made inside the submodule, always push the submodule repo first, then update and push parent repos
- `profiles.json` and `history.json` are runtime data files (gitignored) — they will not appear in the submodule
