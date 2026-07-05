---
name: update-yt-summary-tool
description: Update the create-yt-summary git submodule nested under skills/create-yt-summary/. Use this when the user wants to pull the latest version of the YouTube summary CLI tool, rebuild it, or sync the submodule to its upstream commit.
---

# Update create-yt-summary tool

This skill handles updating the `skills/create-yt-summary/create-yt-summary` git submodule — a Go CLI that summarizes YouTube videos using yt-dlp and an LLM.

## Submodule location

```
skills/create-yt-summary/create-yt-summary  →  git@github.com:LordFarquaadtheCreator/create-yt-summary.git
```

Note: this submodule is nested inside the `create-yt-summary` skill directory, not at the repo root.

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
cd "$REPO_ROOT" && git submodule update --remote skills/create-yt-summary/create-yt-summary
```

This fetches the latest commit from the submodule's default branch and checks it out.

### 2. Verify the update

```bash
cd "$REPO_ROOT/skills/create-yt-summary/create-yt-summary" && git log --oneline -3
```

### 3. Rebuild the CLI tool

```bash
cd "$REPO_ROOT/skills/create-yt-summary/create-yt-summary" && go build -o create-yt-summary .
```

### 4. Commit the submodule change

```bash
cd "$REPO_ROOT" && git add skills/create-yt-summary/create-yt-summary && git commit -m "chore: update create-yt-summary submodule"
```

## Troubleshooting

- If the submodule isn't initialized: `cd "$REPO_ROOT" && git submodule init skills/create-yt-summary/create-yt-summary`
- The submodule path includes two levels (`skills/create-yt-summary/create-yt-summary`) — make sure to use the full path in all git commands
- The parent `skills/create-yt-summary/SKILL.md` references the built binary at `create-yt-summary/create-yt-summary` — after rebuilding, the usage instructions in that skill remain valid
