# Agents

This repo contains scripts for managing GitHub tokens and tracking job applications.

## Scripts

### set-gh-cli-token.py

Swaps GitHub CLI token using `gh auth login --with-token`.

**Usage:**
```bash
python set-gh-cli-token.py <work_mode|personal_mode>
```

**Requirements:**
- `gh` CLI installed and authenticated
- `~/agents-data/config/gh-pats.json` exists with `work_PAT` and `personal_PAT` keys

**Behavior:**
- Reads PAT from config file
- Runs `gh auth login --with-token` to set CLI token
- Prints success message

### set-gh-mcp-token.py

Swaps GitHub PAT in Windsurf MCP config for GitHub server.

**Usage:**
```bash
python set-gh-mcp-token.py <work_mode|personal_mode>
```

**Requirements:**
- `~/.codeium/windsurf/mcp_config.json` exists with github server config
- `~/agents-data/config/gh-pats.json` exists with `work_PAT` and `personal_PAT` keys

**Behavior:**
- Reads PAT from config file
- Updates `mcp_config.json` github server Authorization header
- Prints success message

### track-job.py

Records a job application to Google Sheets.

**Usage:**
```bash
python skills/track-job/track-job.py "<Job Posting URL>" "<email>" "<industry>" "<status>" [phone] [notes]
```

**Parameters:**
- Link - Job posting URL
- Email - Employer contact email
- Industry - Tech, Health Care, Retail, Finance, Gig, Other
- Status - Application status
- Phone - Contact phone number (optional, null if not provided)
- Notes - Free-form notes (optional, null if not provided)

**Behavior:**
- Validates all inputs
- Auto-sets date to today
- Optional fields (phone, notes) default to null if not provided
- Posts to Google Apps Script
- Returns exit code

### get-jobs.py

Retrieves all job applications from Google Sheets.

**Usage:**
```bash
python skills/get-jobs/get-jobs.py
```

**Behavior:**
- GET request to Google Apps Script
- Returns JSON array of all job applications
- Handles errors gracefully

## Config

### config/gh-pats.json

Contains GitHub PATs for different contexts:
```json
{
  "work_PAT": "ghp_xxx...",
  "personal_PAT": "ghp_yyy..."
}
```

**Note:** This file is gitignored for security.
