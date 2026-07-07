---
name: update-create-video-mcp
description: Update the create-video git submodule. Use this when the user wants to pull the latest version of the create-video MCP server, rebuild it, or sync the submodule to its upstream commit.
---

# Update create-video MCP

This skill handles updating the `create-video` git submodule — a video generation MCP server (Modal LTX-2.3) written in Go.

## Submodule locations

```
senor-modal-apps/create-video  →  https://github.com/LordFarquaadtheCreator/create-video.git
agents-skills/mcps/create-video  →  https://github.com/LordFarquaadtheCreator/create-video.git
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
cd "$REPO_ROOT" && git submodule update --remote mcps/create-video
```

This fetches the latest commit from the submodule's default branch and checks it out.

### 2. Verify the update

```bash
cd "$REPO_ROOT/mcps/create-video" && git log --oneline -3
```

### 3. Rebuild (if needed)

**Go binary rebuild:**

```bash
cd "$REPO_ROOT/mcps/create-video" && go build -o create-video .
```

**Docker rebuild (if using Docker):**

```bash
cd "$REPO_ROOT/mcps/create-video" && docker build -t create-video .
```

### 4. Commit the submodule change

```bash
cd "$REPO_ROOT" && git add mcps/create-video && git commit -m "chore: update create-video submodule"
```

## Sync all repos that reference this submodule

The `create-video` submodule is referenced from multiple repos. After pushing changes to the `create-video` source repo, update the submodule pointer in every repo that includes it:

1. **`agents-skills`** — `mcps/create-video` (the canonical skill repo)
2. **`senor-modal-apps`** — `create-video` (the app repo)

For each repo:

```bash
cd <repo_root> && git submodule update --remote <submodule_path>
git add <submodule_path> && git commit -m "update create-video submodule"
git push origin main
```

If you made changes directly inside the submodule working directory (not via `update --remote`), commit and push from inside the submodule first, then update the pointer in each parent repo.

## Troubleshooting

- If the submodule isn't initialized: `cd "$REPO_ROOT" && git submodule init mcps/create-video`
- If the submodule is on a detached HEAD, that's normal after `update --remote`
- If `mcp-config.json` changed, let the user know — they may need to update their local config
- If changes were made inside the submodule, always push the submodule repo first, then update and push parent repos
