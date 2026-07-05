# agents-skills

Canonical skill bank for agent skills in the [open skills format](https://inference.sh/blog/skills/agent-skills-overview). Each skill in `skills/` is the source of truth.

## Creating a new skill

```zsh
./commands/create-skill.sh <skill-name> [description] [--openai]
```

This scaffolds a `skills/<name>/` directory with `SKILL.md` and optionally `agents/openai.yaml`. See [`commands/AGENTS.md`](commands/AGENTS.md) for full details.

## Linking skills to an app

Skills are not auto-discovered. Use `link-skills.sh` to symlink all skills into a target agent's skill bank:

```zsh
# For Zed (global — available in all projects)
./commands/link-skills.sh ~/.agents/skills

# For Devin
./commands/link-skills.sh ~/.devin/skills

# For Claude Code
./commands/link-skills.sh ~/.claude/skills
```

**Zed note:** Zed discovers skills from `~/.agents/skills/` (global, available in all projects) or `<project>/.agents/skills/` (project-local). Run `link-skills.sh ~/.agents/skills` to make all skills available in every Zed project.

This is intentional — you control exactly which skills each app can use. Two apps can have different sets of skills without one polluting the other.

Each skill is version-tracked and managed through one source while being available to any number of agents and skill banks.

## Commands

See [`commands/AGENTS.md`](commands/AGENTS.md) for detailed descriptions of all available commands.
