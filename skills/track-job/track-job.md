---
description: Record a job application after applying
---

Use this skill immediately after applying to any job.

Use the `track-job.py` script to record job applications:

```bash
python /Users/farquaad/agents-data/skills/track-job/track-job.py "<Job Posting URL>" "<email>" "<industry>" "<status>" [phone] [notes]
```

Parameters (in order):
**Link** — URL of the job posting.
**Email** — Employer contact email.
**Industry** — Tech, Health Care, Retail, Finance, Gig, Other.
**Status** — Not Started, Applied Only, Applied + Emailed, Applied + Called, Applied + Emailed + Called, Interview!, Done.
**Phone** — Contact phone number (optional, null if not provided).
**Notes** — Free-form notes on the job (optional, null if not provided).
