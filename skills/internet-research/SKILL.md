---
name: internet-research
description: Thorough internet research using BrowserOS MCP and built-in tools. Use this when the user wants you to research a topic, find facts, compare opinions, investigate claims, or gather evidence from the web — especially when they ask for "research", "look into", "find out", "what's the consensus on", or "tell me about X".
---

# Internet Research Skill

Use this skill when you need to research a topic on the internet. It provides a structured methodology for gathering evidence, evaluating sources, and forming conclusions.

## Before You Start: Clarify the Research Scope

From the user's request and context, determine which of the following apply. If ambiguous, ask the user briefly:

1. **Reddit / Internet forums** — Should I check Reddit, Hacker News, or similar forums for community sentiment and lived experiences?
2. **YouTube videos** — Should I look at YouTube content for explanations, reviews, tutorials, or walkthroughs?
3. **Official sources** — Should I check official documentation, company pages, government sites, academic papers, or authoritative sources?

**Default assumption:** If the user doesn't specify, assume **all three** are relevant and proceed with the full methodology below.

## Research Methodology

### Source Hierarchy & Weighting

Sources are ranked by reliability. Weight your conclusions accordingly:

| Tier | Source Type | Examples | Weight |
|------|-------------|----------|--------|
| **Primary** | Official / authoritative | Documentation, specs, whitepapers, academic papers, government data, company announcements, official APIs, standards bodies, first-party testimonials (lived experiences from forums can also count here) | High |
| **Secondary** | Aggregated / curated | YouTube summaries, news articles, blog posts, review roundups, forum discussions | Medium |

### Evidence Requirements

Before forming a conclusion, gather at minimum:

- **5 pieces of primary evidence** — official documentation, authoritative sources, direct testimonials
- **10 pieces of secondary evidence** — forum posts, blog articles, YouTube content, news coverage

### Step 1: Use BrowserOS MCP to Navigate the Web

The primary tool for this skill is **BrowserOS MCP** (or any available browser/browsing tool). Use it to:

- Search Google, DuckDuckGo, or other search engines
- Navigate to specific pages
- Extract content from pages
- Take screenshots of pages when visual information matters

### Step 2: Reddit & Internet Forum Research

When researching Reddit or similar forums:

1. **Cast a wide net** — Search across multiple subreddits and threads. Don't settle for the first few results. Look for:
   - The top-voted comments (but see caveat below)
   - Controversial comments (sorted by "controversial" — these often surface dissenting opinions)
   - Recent threads (within the last year, preferably months)
   - Multiple threads asking the same question (to gauge consensus)

2. **Take everything with a grain of salt** — Internet users can be dramatic, hyperbolic, or overly negative. Consider:
   - Selection bias: people with strong opinions (especially negative ones) are more likely to post
   - Vocal minority: a loud complaint doesn't represent the silent majority
   - Astroturfing: some posts/comments may be astroturfed by interested parties
   - Hivemind effects: early downvotes can bury valid counterpoints

3. **Lived experiences can be primary sources** — Direct, detailed testimonials ("I've been using X for 3 years and here's my experience") can be treated as **primary evidence** when they describe specific, verifiable facts about someone's experience. Vague complaints ("X sucks") are secondary at best.

4. **Cross-reference forum claims** — If a Reddit post claims "X broke my setup", look for:
   - Corroboration from other users
   - Official responses from maintainers
   - Whether the issue was resolved in later updates

### Step 3: YouTube Research

When researching YouTube content:

1. **Use the YouTube summary tool** to extract key information from videos without watching the full video. The tool lives at:
   - `agents-skills/skills/create-yt-summary/SKILL.md` — instructions for building and using it. You can assume it is already built. 

2. Look for:
   - Reviews and comparisons (e.g., "X vs Y" videos)
   - Tutorials and walkthroughs
   - Conference talks and official presentations (these are primary sources)
   - Deep-dive analyses

3. **YouTube videos are secondary evidence** unless they are:
   - Official presentations from the source/company (primary)
   - Conference talks by authoritative figures (primary)
   - First-hand recorded experiences/testimonials (primary)

### Step 4: Official Sources Research

Official sources are the most reliable. Prioritize:

1. **Official documentation** — API docs, user manuals, spec sheets, READMEs on official repos
2. **Company/government websites** — `.gov`, `.edu`, official company domains
3. **Academic papers** — Peer-reviewed research (arXiv, IEEE, ACM, Google Scholar)
4. **Standards bodies** — IETF RFCs, W3C specs, ISO standards
5. **Official announcements** — Blog posts, release notes, changelogs, press releases
6. **Source code** — The actual code in official repositories is the ultimate truth for technical questions

### Step 5: Synthesize & Conclude

After gathering evidence:

1. **List your evidence** — Present a clear summary of what you found, labeled by tier (primary/secondary)
2. **Weight toward primary sources** — When drawing conclusions, primary evidence gets significantly more weight. If official docs say X but Reddit says Y, go with X.
3. **Acknowledge uncertainty** — If evidence is contradictory or insufficient, say so. Don't overstate confidence.
4. **Note consensus vs. outliers** — For secondary evidence, distinguish between widespread consensus and minority opinions.
5. **Flag source quality issues** — Note if a source is dated, biased, speculative, or otherwise questionable.

## Summary Checklist

Checklist that must be satisfied before concluding research:

- [ ] Research scope clarified (or assumed all three)
- [ ] At least **5 primary evidence** sources gathered
- [ ] At least **10 secondary evidence** sources gathered
- [ ] Reddit/forums: wide net cast across multiple threads/subreddits
- [ ] YouTube: summary tool used where applicable
- [ ] Official sources: checked and prioritized
- [ ] Conclusion drawn with greater weight to primary evidence
- [ ] Uncertainty and source quality acknowledged
