---
name: manage-job
description: Track and retrieve job applications via the manage-job MCP server (Google Sheets backend)
metadata:
  display-name: Manage Job Applications
  enabled: 'true'
---
# Manage Job Applications

Job applications live in one Google Sheet. A deployed Google Apps Script web app does the CRUD; the `manage-job` MCP server is the client. Use the MCP tools — do not touch the sheet or the Apps Script project directly.

## Tools

| Tool | Use |
|---|---|
| `track_job` | Record one or more applications. Use immediately after applying to a job. |
| `get_jobs` | Read rows, with pagination and filters. |
| `patch_job` | Update one or more rows. |
| `delete_job` | Delete one or more rows. |

If a call fails, stop and tell the user what failed.

If a call returns a Google Drive HTML page instead of JSON — "Page Not Found", "Sorry, the file you have requested does not exist", or "Sorry, unable to open the file at this time" — the deployment ID the client was launched with is wrong or the deployment was deleted. It is not a spreadsheet sharing problem. See Configuration.

## track_job

`jobs`: array of 1 or more entries.

| Field | Required | Notes |
|---|---|---|
| `companyName` | yes | Free-form. |
| `link` | yes | URL of the individual job posting, must start with `http://` or `https://`. Get it from the posting's share / copy-link button, not a general careers page. |
| `industry` | yes | Exactly one of `Tech`, `Health Care`, `Retail`, `Finance`, `Gig`, `Other` (case-sensitive). |
| `status` | yes | Exactly one of `Applied Only`, `Applied + Emailed`, `Applied + Called`, `Applied + Emailed + Called`, `Interview!`, `Got the Job!`, `Didn't Get It`, `Not Started` (case-sensitive). |
| `dateApplied` | no | `YYYY-MM-DD`. Defaults to today. |
| `email` | no | Employer contact email. |
| `phoneNumber` | no | Contact phone, 10-15 digits after formatting characters are stripped. |
| `notes` | no | Free-form. Omit if there is nothing worth remarking. |

```json
{
  "jobs": [
    {
      "companyName": "Acme Corp",
      "link": "https://fakejobs.com/quantum-ai-analyst",
      "industry": "Tech",
      "status": "Not Started",
      "email": "email@email.com",
      "phoneNumber": "917-999-1234",
      "notes": "They said to email John at john@company.com"
    }
  ]
}
```

## get_jobs

| Param | Default | Notes |
|---|---|---|
| `page` | 1 | Page number. |
| `pageSize` | 50 | Results per page. |
| `search` | — | Matches companyName, link, email, notes. |
| `industry` | — | Filter. |
| `status` | — | Filter. |
| `order` | desc | Sort by dateApplied: `asc` or `desc`. |

Returns JSON with `rows`, `page`, `pageSize`, `totalPages`, `totalRows`.

```json
{
  "companyName": "Acme Corp",
  "link": "https://fakejobs.com/quantum-ai-analyst",
  "dateApplied": "2026-06-27T04:00:00.000Z",
  "industry": "Tech",
  "phoneNumber": "5551234567",
  "email": "a@b.com",
  "status": "Applied Only",
  "notes": ""
}
```

## patch_job

`patches`: array of 1 or more `{ "matchBy": {...}, "update": {...} }` objects. Any column can appear in either object: `companyName`, `link`, `dateApplied`, `industry`, `phoneNumber`, `email`, `status`, `notes`. `matchBy` needs at least one field; `update` needs at least one field.

```json
{
  "patches": [
    {
      "matchBy": { "companyName": "Acme Corp" },
      "update": { "status": "Interview!" }
    }
  ]
}
```

## delete_job

`deletes`: array of 1 or more `{ "matchBy": {...} }` objects. Use multiple fields when a company name alone could match several rows.

```json
{
  "deletes": [
    { "matchBy": { "companyName": "Acme Corp", "link": "https://example.com" } }
  ]
}
```

## Configuration

The MCP server reads `SHEETS_DEPLOYMENT_ID` from its environment at startup. The current ID is recorded in `config/sheets-deployment.yaml` at the repository root (gitignored):

```yaml
deploymentId: <deployment-id>
```

The web app URL is `https://script.google.com/macros/s/<deployment-id>/exec`.

Client env blocks must match that file:

- Zed: `~/.config/zed/settings.json` → `context_servers.manage-job.env.SHEETS_DEPLOYMENT_ID`
- Hermes: `~/.hermes/config.yaml` → `mcp_servers.manage-job.env.SHEETS_DEPLOYMENT_ID`

When a new Apps Script deployment is created, update the file and every client env block, then restart the client so it re-reads the env var.

## No MCP tools available?

Some agents do not have the manage-job MCP registered. Call the server binary through mcp-bridge instead:

```bash
~/agents-skills/skills/mcp-bridge/scripts/mcp-call/mcp-call \
  /Users/farquaad/agents-skills/mcps/manage-job/manage-job-mcp -- \
  call get_jobs --args '{"pageSize":10}' \
  --env SHEETS_DEPLOYMENT_ID=<deployment-id>
```

The same `list`, `call`, and `describe` subcommands apply as for any other MCP.

## Maintenance

Go source lives in `mcps/manage-job/` (stdio MCP server, no CLI). Build and test:

```bash
cd /Users/farquaad/agents-skills/mcps/manage-job && go build -o manage-job-mcp . && go test ./...
```

The Apps Script backend is deployed and maintained outside this repo; backend changes are made in the Apps Script editor. See `mcps/manage-job/AGENTS.md` for the server layout.
