# agentic-news
### Your AI news butler for cognition, not just aggregation

Every morning, agentic-news turns high-density RSS streams into a clean mobile H5 daily briefing.
It goes beyond “what happened” and focuses on “why it matters, how to reason, and what to learn next.”

---

## Core Value

- **High-signal curation**: 10–20 featured items worth reading daily
- **Deep analysis**: summary, viewpoint, risk/opportunity, contrarian insights
- **Strong personalization**: dual-track learning from feedback and behavior
- **Learning guidance**: actionable paths for cognition and knowledge growth
- **Mobile-first reading**: clean, concise, editorial-style experience

---

## Workflow

1. Ingest RSS (tech / public affairs / finance)
2. Clean and deduplicate content
3. Run staged AI analysis (facts → insights → personal advisor)
4. Rank with weighted scoring
5. Render daily H5 and archive by `YYYY/MM/DD`
6. Publish via SFTP to Nginx and refresh `/latest`

---

## Status

🚧 MVP in active development (Go + static frontend + local scheduler + SFTP publishing)

---

## Repository

- GitHub: https://github.com/nikkofu/agentic-news
- 中文页面: [index.md](./index.md)
