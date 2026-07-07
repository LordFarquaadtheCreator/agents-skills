---
name: cover-letter
description: Generate styled PDF cover letters via the cover-letter-writter MCP. Use when Fahad wants to write, draft, or generate a cover letter for a job application.
---

# Cover Letter

Generate styled PDF cover letters via the `cover-letter-writter` MCP server. Profiles persist on disk — contact info is stored once and reused.

## Workflow

1. **Check for existing profiles:** `list_profiles` — if Fahad's profile exists, skip to step 3
2. **Create profile if needed:** `create_profile` with name, email, address, phone. Save the returned `profileId`.
3. **Draft the body:** Write the cover letter body text. Ask Fahad for the job posting or company details if not provided. Body is plain text — no formatting markers needed.
4. **Generate:** `generate_cover_letter` with `profileId` + `body`. Returns absolute file path.
5. **Record:** Generation is auto-recorded in history. No extra step needed.

## Tools

| Tool | When to use |
|---|---|
| `create_profile` | First time, or new persona/role-target |
| `list_profiles` | Before generating — find existing profileId |
| `get_profile` | Check specific profile details |
| `update_profile` | Contact info changed (new email, phone, address) |
| `delete_profile` | Remove old/unused profiles |
| `generate_cover_letter` | Produce the PDF. Requires profileId + body. |
| `list_history` | Review past cover letters (optionally by profileId) |

## Body drafting guidance

- Ask for the job posting or company name if not given
- Keep to 3-4 paragraphs: intro, relevant experience, why this company, closing
- Plain text only — the PDF template handles layout, salutation ("To Whom it May Concern,"), date, and signature ("Best regards, \<Name\>")
- Smart quotes and em dashes are fine — sanitized to ASCII automatically

## Output

PDF saved to profile's `outputDir` (default `~/Downloads`). Filename defaults to `<Name>NoSpaces>CoverLetter.pdf`. Both overridable per-call via `outputDir` / `filename` params.

## Multiple profiles

One profile per persona is fine — e.g. "tech", "design", "writing". Use `label` to distinguish. History is per-profile, so past letters stay organized.
