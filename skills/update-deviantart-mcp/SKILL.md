---
name: update-deviantart-mcp
description: Update the deviantart-mcp git submodule under mcps/. Use this when the user wants to pull the latest version of the deviantart-mcp MCP server, rebuild it, or sync the submodule to its upstream commit.
---

# Update deviantart-mcp MCP

This skill handles updating the `mcps/deviantart-mcp` git submodule — a DeviantArt MCP server written in Go.

## Submodule location

```
mcps/deviantart-mcp  →  git@github.com:LordFarquaadtheCreator/deviantart-mcp.git
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
cd "$REPO_ROOT" && git submodule update --remote mcps/deviantart-mcp
```

This fetches the latest commit from the submodule's default branch and checks it out.

### 2. Verify the update

```bash
cd "$REPO_ROOT/mcps/deviantart-mcp" && git log --oneline -3
```

### 3. Rebuild the MCP server (if needed)

If the user needs a fresh binary:

```bash
cd "$REPO_ROOT/mcps/deviantart-mcp" && go build -o deviantart-mcp cmd/deviantart-mcp/main.go
```

### 4. Commit the submodule change

The updated submodule pointer should be committed in the parent repo:

```bash
cd "$REPO_ROOT" && git add mcps/deviantart-mcp && git commit -m "chore: update deviantart-mcp submodule"
```

## Troubleshooting

- If the submodule isn't initialized: `cd "$REPO_ROOT" && git submodule init mcps/deviantart-mcp`
- If the submodule is on a detached HEAD (normal after `update --remote`), that's expected — the parent repo tracks the commit hash, not a branch
- If the user needs a specific tag or branch, use: `cd "$REPO_ROOT/mcps/deviantart-mcp" && git checkout <tag-or-branch>`
