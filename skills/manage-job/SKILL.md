---
name: manage-job
description: Track and retrieve job applications via the Google Sheets backend
metadata:
  display-name: Manage Job Applications
  enabled: 'true'
---
# Manage Job Applications

This skill manages job applications via the Google Sheets backend — create, read, update, and delete. The script is a compiled Go binary — no Python or external runtime required.

## Binary

The compiled binary lives at:

```
/Users/farquaad/agents-data/skills/manage-job/manage-job
```

All interaction is via stdio. The binary prints results to stdout and errors to stderr. Exit code 0 means success, exit code 1 means failure. If the command fails, you must stop and tell the user that the command failed.

## Commands

### `track` — Record a new job application

Use this immediately after applying to any job. Creates a new row in the spreadsheet with today's date.

```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job track <companyName> <link> <industry> <status> [email] [phone] [notes]
```

#### Required parameters (in order)

1. **companyName** — Name of the company you applied to. This is a free-form string.
2. **link** — URL of the job posting. Must start with `http://` or `https://`. This must be the link relating to the individual job application, not a general careers page. This link is always attainable by searching through share buttons and copy link buttons on the job posting.
3. **industry** — Must be exactly one of these values (case-sensitive):
   - `Tech`
   - `Health Care`
   - `Retail`
   - `Finance`
   - `Gig`
   - `Other`
4. **status** — Must be exactly one of these values (case-sensitive):
   - `Not Started`
   - `Applied Only`
   - `Applied + Emailed`
   - `Applied + Called`
   - `Applied + Emailed + Called`
   - `Interview!`
   - `Done`

#### Optional parameters (in order, can be omitted)

5. **email** — Employer contact email. Must contain `@` and `.`.
6. **phone** — Contact phone number. Must contain 10-15 digits (formatting characters are stripped).
7. **notes** — Free-form notes about the job. You do not need to fill this out if there is nothing special to remark about the job. All remaining arguments after phone are joined into the notes string.

Optional parameters can be omitted entirely. You cannot skip an optional parameter and provide a later one — if you want to provide notes but not phone, pass an empty string for phone.

#### Examples

*No optional parameters:*
```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job track "Acme Corp" "https://fakejobs.com/quantum-ai-analyst" "Tech" "Not Started"
```

*With email:*
```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job track "Acme Corp" "https://fakejobs.com/quantum-ai-analyst" "Tech" "Not Started" "email@email.com"
```

*With email and phone:*
```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job track "Acme Corp" "https://fakejobs.com/quantum-ai-analyst" "Tech" "Not Started" "email@email.com" "917-999-1234"
```

*All parameters:*
```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job track "Acme Corp" "https://fakejobs.com/quantum-ai-analyst" "Tech" "Not Started" "email@email.com" "917-999-1234" "They said to email \"John\" at \"john@company.com\""
```

#### Output

On success, prints to stdout:
```
Success: {"status":"success"}
```

On failure, prints error to stderr and exits with code 1.

### `get` — Retrieve all tracked job applications

Fetches all job applications from the spreadsheet. Returns JSON to stdout.

```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job get
```

#### Optional query parameters

You can pass key-value pairs as arguments to filter:

```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job get page 1 pageSize 10 search "Acme" industry "Tech" status "Applied Only" order "desc"
```

Supported keys:
- **page** — Page number (default: 1)
- **pageSize** — Results per page (default: 50)
- **search** — Search across companyName, link, email, notes
- **industry** — Filter by industry
- **status** — Filter by status
- **order** — Sort by dateApplied: `asc` or `desc` (default: `desc`)

#### Output

Prints JSON to stdout:
```json
{
  "status": "success",
  "rows": [
    {
      "companyName": "Acme Corp",
      "link": "https://...",
      "dateApplied": "2026-06-27T04:00:00.000Z",
      "industry": "Tech",
      "phoneNumber": "5551234567",
      "email": "a@b.com",
      "status": "Applied Only",
      "notes": ""
    }
  ],
  "page": 1,
  "pageSize": 50,
  "totalPages": 3,
  "totalRows": 121
}
```

### `patch` — Update an existing job application

Updates fields on an existing row. Uses `--matchBy` to find the row and `--update` to specify which fields to change.

```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job patch --matchBy '<json>' --update '<json>'
```

#### Flags

- **`--matchBy`** (required) — JSON object with at least one field to identify the row. Any column can be used: `companyName`, `link`, `dateApplied`, `industry`, `phoneNumber`, `email`, `status`, `notes`.
- **`--update`** (required) — JSON object with at least one field to change. Same columns as above.

#### Examples

*Change status:*
```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job patch --matchBy '{"companyName":"Acme Corp"}' --update '{"status":"Interview!"}'
```

*Update multiple fields:*
```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job patch --matchBy '{"companyName":"Acme Corp","link":"https://example.com"}' --update '{"status":"Done","notes":"Rejected"}'
```

#### Output

On success, prints to stdout:
```
Success: {"status":"success"}
```

On failure, prints error to stderr and exits with code 1.

### `delete` — Delete a job application

Deletes a row from the spreadsheet. Uses `--matchBy` to find the row.

```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job delete --matchBy '<json>'
```

#### Flags

- **`--matchBy`** (required) — JSON object with at least one field to identify the row. Same columns as patch.

#### Examples

*Delete by company name:*
```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job delete --matchBy '{"companyName":"Acme Corp"}'
```

*Delete by multiple fields for precision:*
```bash
/Users/farquaad/agents-data/skills/manage-job/manage-job delete --matchBy '{"companyName":"Acme Corp","link":"https://example.com"}'
```

#### Output

On success, prints to stdout:
```
Success: {"status":"success"}
```

On failure, prints error to stderr and exits with code 1.

## Configuration

The binary reads the Apps Script deployment ID from `config/sheets-deployment.yaml` at the repository root. This file is gitignored and must exist on the local machine. Format:

```yaml
deploymentId: AKfycbwRQ52XCi5htaaHLO1Laizu8-pyYFKI0GEWELSnJHsP1CBDc-9OxNlkWGhlG-8l8tDxIQ
```

The deployment ID is the hash in the Apps Script web app URL: `https://script.google.com/macros/s/<deploymentId>/exec`. When a new deployment is created in Apps Script, update this file with the new ID. If it is missing or malformed, the binary will print an error and exit with code 1.

## Rebuilding

If you modify `main.go`, recompile:

```bash
cd /Users/farquaad/agents-data/skills/manage-job && go build -o manage-job main.go
```
