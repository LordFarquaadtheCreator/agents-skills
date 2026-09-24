# Agent Instructions — manage-job (skill)

Documentation-only directory: `SKILL.md` for agents using the tracker, this file for agents maintaining the skill. No code, no binary. It is symlinked into `~/.agents/skills/manage-job` and `~/.devin/skills/manage-job`, so edits here reach every agent.

## What this skill wraps

- MCP server: `mcps/manage-job/manage-job-mcp` (Go, stdio). Tools: `track_job`, `get_jobs`, `patch_job`, `delete_job`.
- Backend: a deployed Google Apps Script web app at `https://script.google.com/macros/s/<id>/exec`. All CRUD runs on Google's side. The backend TypeScript source used to live at `skills/manage-job/apps-script/update-beggers-sheet.ts` (added in commit `4778fce`); it is deleted in the current working tree and survives only in git history — recover it from there or from the Apps Script editor if it needs editing.
- Deployment ID: passed via the `SHEETS_DEPLOYMENT_ID` env var, read once at server startup. There is no config file read at runtime. `config/sheets-deployment.yaml` exists only as the local record of the current ID.

## Failure signature

If a tool returns Google Drive HTML ("Page Not Found" / "the file you have requested does not exist" / "unable to open the file at this time"), the calling client's `SHEETS_DEPLOYMENT_ID` is wrong or the deployment was deleted. It is never a sheet-sharing problem. Check every client env block:

- Zed: `~/.config/zed/settings.json` → `context_servers.manage-job.env`
- Hermes: `~/.hermes/config.yaml` → `mcp_servers.manage-job.env`
- Ad-hoc runs: `mcp-call --env SHEETS_DEPLOYMENT_ID=...`

Also confirm the ID itself is live: `curl -s -o /dev/null -w '%{http_code}' https://script.google.com/macros/s/<id>/exec` — expect 302.

## History

The old `manage-job` CLI (`track`, `get`, `patch`, `delete` subcommands) and its `cmd/`, `appscript/`, `apps-script/` files were moved out of this skill dir — the deletions are pending in the working tree. The project is now the MCP submodule only.

## Rebuild / test

```bash
cd /Users/farquaad/agents-skills/mcps/manage-job && go build -o manage-job-mcp . && go test ./...
```
