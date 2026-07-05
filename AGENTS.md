# agents-skills

This repo is a container/repository for all things related to agentic development.
Here you will find skills in the [open skills format](https://inference.sh/blog/skills/agent-skills-overview) and mcp servers in their respective directories.
This is used to track and link skills amongst different apps such as claude code, hermes, devin, etc.

## Commands

Reusable shell scripts live in [`commands/`](commands/). See [`commands/AGENTS.md`](commands/AGENTS.md) for detailed descriptions of each command.

- [`create-skill.sh`](commands/create-skill.sh) — scaffold a new skill
- [`link-skills.sh`](commands/link-skills.sh) — symlink skills into a target agent's skills directory
