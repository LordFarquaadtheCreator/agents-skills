---
description: Generate a styled one-page PDF resume. Use when Fahad wants to build, generate, or tailor a resume for a job application.
---

# Resume Builder Skill

Generate one-page PDF resumes using the `resume-builder` MCP.

## Workflow

1. **One-time setup**: Call `set_embedding_config` with the embedding endpoint (e.g. LM Studio at `http://localhost:1234`, model `text-embedding-embeddinggemma-300m-qat`). This persists on disk.

2. **Initialize resume**: Call `init_resume` with full structured resume data. This stores the resume and builds a vector store embedding every bullet point. Re-calling this overwrites everything.

3. **Ask user**: Auto or manual?
   - **Auto**: Call `generate_resume` with `mode="auto"`, `query="<job description>"`, `template="fahad"`. MCP searches vector store, selects most relevant content, enforces one page, outputs PDF.
   - **Manual**: Call `search_resume` with the job description. Review results with user. Agent tailors content (selects/reorders bullets, drops irrelevant items). Call `generate_resume` with `mode="manual"`, `data=<tailored ResumeData>`, `template="fahad"`.

4. **Output**: PDF saved to `outputDir` (default `/tmp`) as `<Name>Resume.pdf`. Response includes `trimmed` info showing what was dropped for one-page fit.

## Resume Data Format

```json
{
  "name": "Fahad Faruqi",
  "contact": {
    "location": "Queens, NYC",
    "email": "fahadfaruqi1@gmail.com",
    "links": {
      "linkedin": "https://linkedin.com/in/fahadfaruqi42",
      "github": "https://github.com/lordfarquaadthecreator",
      "website": ""
    }
  },
  "education": [
    {
      "institution": "CUNY City College of New York",
      "degree": "B.S. Computer Science - Cum Laude",
      "start": "",
      "end": "Class of 2025",
      "location": "New York, NY",
      "link": ""
    }
  ],
  "skills": [
    { "category": "Languages", "values": "Swift, TypeScript, Go, ..." }
  ],
  "experiences": [
    {
      "company": "Tobi Wealth",
      "role": "Principal Software Engineer",
      "start": "Dec. 2025",
      "end": "Present",
      "location": "Remote",
      "link": "https://tobiwealth.com/",
      "bullets": ["Built a multi-tier caching layer...", "Led migration to..."]
    }
  ],
  "projects": [
    {
      "name": "Google AI Overviews Blocker",
      "tech": "Javascript",
      "date": "Feb. 2026",
      "link": "https://github.com/zbarnz/...",
      "bullets": ["Enhanced detection..."]
    }
  ]
}
```

## Rules

- Empty/null fields are omitted from the PDF render automatically.
- One-page enforcement: trims oldest/lowest-relevance bullets first, then experiences, then projects, then font scaling (floor 11pt).
- Guard rail quotas: max 6 experiences (5 bullets each), 4 projects (2 bullets each), 5 skill groups, 3 education entries.
- Template must be specified. Currently only `"fahad"` available.
- MCP does NOT rewrite resume content. It only selects, orders, and formats.
- No LLM dependency in MCP. Agent handles tailoring in manual mode.
