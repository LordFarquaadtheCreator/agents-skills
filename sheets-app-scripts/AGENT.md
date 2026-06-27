# Agent Instructions — sheets-app-scripts

## Compile TS to JS before pushing

Google Apps Script does not run TypeScript. The source of truth is `update-beggers-sheet.ts`.

Before deploying or pushing changes:

```bash
tsgo --project tsconfig.json
```

This compiles `update-beggers-sheet.ts` → `update-beggers-sheet.js`. Push the `.js` file to Apps Script. Never edit the `.js` file directly — always edit `.ts` and recompile.

Both files have this note in their headers.

## API shape

### POST (`doPost`)

Content-Type: `application/json`

**Create:**
```json
{
  "action": "create",
  "companyName": "Acme",
  "link": "https://...",
  "dateApplied": "2024-01-15",
  "industry": "Tech",
  "phoneNumber": "5551234567",
  "email": "a@b.com",
  "status": "Applied Only",
  "notes": ""
}
```

Valid statuses: `Not Started`, `Applied Only`, `Applied + Emailed`, `Applied + Called`, `Applied + Emailed + Called`, `Interview!`, `Got the Job!`, `Didn't Get It`

Valid industries: `Tech`, `Health Care`, `Retail`, `Finance`, `Gig`, `Other`

**Patch:**
```json
{
  "action": "patch",
  "matchBy": { "companyName": "Acme" },
  "update": { "status": "Interview!" }
}
```

**Delete:**
```json
{
  "action": "delete",
  "matchBy": { "companyName": "Acme" }
}
```

`matchBy` and `update` are separate objects. `matchBy` identifies the row (any combination of fields). `update` holds the new values. Both require at least one field.

### GET (`doGet`)

Query params: `page`, `pageSize`, `search`, `industry`, `status`, `order` (`asc` or `desc`, default `desc`).

## Testing

```bash
./test.sh [URL]
```

Runs read, create, patch, delete against the deployed endpoint. Uses unique company name per run. Cleans up on exit.
