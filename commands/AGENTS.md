# Commands

Reusable shell scripts for managing this repo.

## `create-skill.sh`

Scaffolds a new skill directory under `skills/`.

```
Usage: ./commands/create-skill.sh <skill-name> [description] [--openai]
```

| Argument | Required | Description |
|---|---|---|
| `skill-name` | Yes | Directory name for the skill. Must be lowercase letters, digits, and hyphens only. |
| `description` | No | One-line description placed in the `SKILL.md` frontmatter under `description:`. |
| `--openai` | No | If present, also creates `agents/openai.yaml` with `display_name`, `short_description`, and `default_prompt` for OpenAI/Agents SDK compatibility. |

**What it creates:**
- `skills/<name>/SKILL.md` — frontmatter with `name` and `description`, plus an H1 heading.
- `skills/<name>/agents/openai.yaml` — (only with `--openai`) interface metadata for OpenAI agents.

## `link-skills.sh`

Symlinks all skills from `skills/` into a target agent's skills directory.

```
Usage: ./commands/link-skills.sh <target-dir>
```

| Argument | Required | Description |
|---|---|---|
| `target-dir` | Yes | Path to the agent's skills directory (e.g., `~/.agents/skills` for Zed, `~/.devin/skills` for Devin, `~/.claude/skills` for Claude Code). |

**What it does:**
- Creates the target directory if it doesn't exist.
- Symlinks each skill directory from `skills/` into the target.
- Skips any existing non-symlink entries (won't overwrite real directories).
- Removes stale symlinks that point to nonexistent skill directories.

> **Zed note:** Zed discovers skills from `~/.agents/skills/` (global, available in all projects) or `<project>/.agents/skills/` (project-local). Run this script with `~/.agents/skills` as the target to make all skills available in every Zed project.

> **Hermes note:** Hermes uses a proprietary skill format (extra metadata fields in `SKILL.md`, category directories with `DESCRIPTION.md`, etc.) that is incompatible with the open skills format in this repo. Symlinking or copying skills from this repo into `~/.hermes/skills/` will not work.
>
> Instead, add `$PWD/skills` to Hermes' `external_dirs` config in `~/.hermes/config.yaml`:
> ```yaml
> skills:
>   external_dirs: ["/absolute/path/to/agents-skills/skills"]
> ```
> Hermes will then load these skills alongside its bundled ones.

---

## Updating these docs

Whenever a new command is added to `commands/`, update both this file (`commands/AGENTS.md`) and the root `AGENTS.md` to reflect the new command. The root `AGENTS.md` only needs a one-line listing; this file should include full usage, arguments, and behavior.
