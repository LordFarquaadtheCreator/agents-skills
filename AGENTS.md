# agents-skills

This repo is a container/repository for all things related to agentic development.
Here you will find skills in the [open skills format](https://inference.sh/blog/skills/agent-skills-overview) and mcp servers in their respective directories.
This is used to track and link skills amongst different apps such as claude code, hermes, devin, zed, etc.

## Commands

Reusable shell scripts live in [`commands/`](commands/). See [`commands/AGENTS.md`](commands/AGENTS.md) for detailed descriptions of each command.

- [`create-skill.sh`](commands/create-skill.sh) — scaffold a new skill
- [`scaffold-mcp.sh`](commands/scaffold-mcp.sh) — scaffold a new MCP server
- [`link-skills.sh`](commands/link-skills.sh) — symlink skills into a target agent's skills directory (e.g. `~/.agents/skills/` for Zed, `~/.devin/skills/` for Devin)

## Skills

Skills live in [`skills/`](skills/). Each skill has a `SKILL.md` with frontmatter + instructions. Notable skills:

- [`mcp-bridge`](skills/mcp-bridge/) — CLI wrapper (`mcp-call`) that lets agents call any MCP server binary without registration. Replaces ad-hoc JSON-RPC scripts.

## MCPs

MCP servers live in [`mcps/`](mcps/) as git submodules. This repo is the **canonical home** for all MCPs — other repos (e.g. `senor-modal-apps`) symlink to `mcps/<name>` rather than tracking their own submodule pointers. See [`mcps/AGENTS.md`](mcps/AGENTS.md) for the full list and per-MCP docs.
