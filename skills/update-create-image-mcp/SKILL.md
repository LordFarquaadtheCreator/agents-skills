---
name: update-create-image-mcp
description: Update the create-image git submodule under mcps/. Use this when the user wants to pull the latest version of the create-image MCP server, rebuild it, or sync the submodule to its upstream commit.
---

# Update create-image MCP

This skill handles updating the `mcps/create-image` git submodule — an image generation MCP server (Modal ComfyUI) written in Go.

## Submodule location

```
mcps/create-image  →  git@github.com:LordFarquaadtheCreator/create-image.git
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
cd "$REPO_ROOT" && git submodule update --remote mcps/create-image
```

This fetches the latest commit from the submodule's default branch and checks it out.

### 2. Verify the update

```bash
cd "$REPO_ROOT/mcps/create-image" && git log --oneline -3
```

### 3. Rebuild (if needed)

**Go binary rebuild:**

```bash
cd "$REPO_ROOT/mcps/create-image" && go build -o create-image .
```

**Docker rebuild (if using Docker):**

```bash
cd "$REPO_ROOT/mcps/create-image" && docker build -t create-image .
```

### 4. Commit the submodule change

```bash
cd "$REPO_ROOT" && git add mcps/create-image && git commit -m "chore: update create-image submodule"
```

## Sync all repos that reference this submodule

The `create-image` submodule is referenced from multiple repos. After pushing changes to the `create-image` source repo, update the submodule pointer in every repo that includes it:

1. **`agents-skills`** — `mcps/create-image` (the canonical skill repo)
2. **`senor-modal-apps`** — `create-image` (the app repo)

For each repo:

```bash
cd <repo_root> && git submodule update --remote <submodule_path>
git add <submodule_path> && git commit -m "update create-image submodule"
git push origin main
```

If you made changes directly inside the submodule working directory (not via `update --remote`), commit and push from inside the submodule first, then update the pointer in each parent repo.

## Troubleshooting

- If the submodule isn't initialized: `cd "$REPO_ROOT" && git submodule init mcps/create-image`
- If the submodule is on a detached HEAD, that's normal after `update --remote`
- If `model_card.yaml` or `mcp-config.json` changed, let the user know — they may need to update their local config
- If changes were made inside the submodule, always push the submodule repo first, then update and push parent repos
