# AI Agentic News Butler Design Spec

- **Date:** 2026-03-19
- **Project:** agentic-news
- **Scope:** MVP design for a single-user, highly personalized, mobile-first daily news H5 product generated from RSS feeds.

## 1. Product Vision

Build a private, daily news-butler product that goes beyond aggregation. The system ingests high-density RSS sources (technology, politics/public affairs, finance, and related professional voices), performs deep AI analysis, and outputs a clean mobile H5 experience similar in editorial feel to NYT/WSJ mobile editions.

The product goal is to improve the user’s:
- Taste (signal discernment)
- Cognition (better mental models and judgement)
- Knowledge breadth (targeted expansion paths)

The system publishes one daily edition before 07:00 local time.

## 2. Confirmed Product Decisions

1. **Daily mode:** Fully automated daily report
2. **RSS scale:** 150–500 sources
3. **Language:** Chinese + English input, Chinese-first output
4. **Analysis style:** Balanced (industry/investment, policy/public affairs, cognition)
5. **Publish deadline:** Morning report, published before 07:00
6. **Daily featured count:** 10–20 primary picks
7. **Personalization:** Strongly personalized for one user
8. **Preference learning:** Dual-track (explicit feedback + implicit behavior)
9. **Confidence policy:** Source-tier-based thresholds
10. **Tech stack:** Go backend + frontend separated static output
11. **Runtime/deploy:** Local scheduled run, then SFTP upload to cloud Nginx
12. **Model strategy:** Quality-first
13. **Auth scope:** No login, single-user private MVP
14. **Architecture choice:** A + C hybrid
   - A: Static daily edition factory
   - C: Structured knowledge artifact output for longitudinal cognition growth

## 3. Architecture Overview (MVP)

### 3.1 System boundary

Two-segment architecture:

- **Local production engine (Go):**
  RSS ingestion, cleansing, AI analysis, scoring/ranking, page rendering, artifact generation, SFTP publishing.

- **Cloud serving layer (Nginx):**
  Static hosting only (HTML/CSS/JS/JSON/images/attachments) via domain access.

### 3.2 Daily artifact layout

Output is grouped by date with complete assets:

- `/YYYY/MM/DD/index.html`
- `/YYYY/MM/DD/articles/{id}.html`
- `/YYYY/MM/DD/assets/...`
- `/YYYY/MM/DD/data/daily.json`
- `/YYYY/MM/DD/data/learning.json`
- `/YYYY/MM/DD/attachments/...` (optional)
- `/YYYY/MM/DD/meta.json`

Also produce/refresh `/latest/` as the stable mobile entry path.

### 3.3 Daily pipeline

1. Read RSS source catalog
2. Ingest last 24h items
3. Deduplicate + source credibility scoring
4. Topic clustering (tech/politics-finance/etc.)
5. Deep AI analysis (summary/insights/opportunities/risks/contrarian)
6. Personalized ranking
7. Select 10–20 featured picks
8. Render H5 + JSON artifacts
9. Generate learning-direction output
10. SFTP upload + latest switch

### 3.4 Resilience and deadline protection

- Source failures: skip and continue, with logging
- AI failures: bounded retries, then degrade to basic-summary mode
- Upload failures: keep complete local package + retryable publish command
- Deadline guard: if still running near 06:50, publish a minimum viable edition rather than miss the day

## 4. Data Model and Scoring

MVP uses file-based structured outputs (JSON), no mandatory database in phase 1.

### 4.1 Core entities

- **Source**
  - `source_id`, `name`, `rss_url`, `domain`, `source_type`
  - `credibility_base`, optional `bias_tags`

- **RawItem**
  - `item_id`, `source_id`, `title`, `url`, `published_at`
  - `raw_content`, `author`, `language`

- **Article**
  - `article_id`, `canonical_url`, `title`, `excerpt`
  - `cover_image`, `content_text`, `keywords`, `entities`
  - `category_primary`, `category_secondary[]`
  - `published_at`, `ingested_at`

- **Insight**
  - `summary_brief`, `summary_deep`
  - `key_points[]`, `viewpoint`
  - `opportunity_risk`, `contrarian_take`
  - `learning_suggestion[]`, `confidence`

- **DailyPick**
  - `rank`, `score_final`, `reason_selected`
  - `reading_time_min`
  - `freshness_score`, `importance_score`
  - `personal_relevance_score`, `source_credibility_score`, `novelty_score`

### 4.2 Ranking formula

Baseline formula:

`FinalScore = 0.30*Importance + 0.25*PersonalRelevance + 0.20*Credibility + 0.15*Novelty + 0.10*Freshness`

Domain-aware dynamic reweighting:
- Tech: raise novelty + industry impact sensitivity
- Public affairs/policy: raise credibility + policy impact sensitivity
- Finance: raise importance + risk signal sensitivity

### 4.3 Source-tier confidence policy

Thresholds by source class:

- **Tier A (high-authority institutions/mainstream high-credibility):** `confidence >= 60`
- **Tier B (professional media/vertical KOL):** `confidence >= 70`
- **Tier C (self-media/opinion-first sources):** `confidence >= 80` + cross-source corroboration

Below threshold:
- Exclude from primary picks
- Optionally place in “watch/verify” section

### 4.4 Personalization (dual-track)

- **Explicit signals:** like/neutral/disagree, reason correction tags, optional note
- **Implicit signals:** click-through, dwell time, detail expansion, revisit, bookmark

Daily outputs include:
- `why_for_you`
- `taste_growth_hint`
- `knowledge_gap_hint`

## 5. AI Analysis Framework

### 5.1 Layered generation flow

- **Stage A: Understanding layer**
  - Extract facts/entities/time and thematic labels

- **Stage B: Deep analysis layer (core value)**
  - Deep summary, viewpoint, opportunities/risks, contrarian observation

- **Stage C: Personal advisor layer**
  - Why it matters specifically to the user, what to study next

### 5.2 Prompt system

Use versioned Go templates:

- `prompts/extract_facts.tmpl`
- `prompts/deep_analysis.tmpl`
- `prompts/personal_advisor.tmpl`
- `prompts/final_editor.tmpl`

Each prompt enforces:
- Explicit input contract
- Strict JSON output contract
- No fabrication (unknown stays unknown)
- Evidence linkage requirements

### 5.3 Traceability fields

Per analyzed item store:
- `model_name`, `model_version`
- `prompt_version`
- `source_refs[]`
- `evidence_snippets[]`
- `generated_at`

### 5.4 Quality gates

Before item promotion to featured list:
1. Structured schema validation
2. Evidence support validation
3. Confidence + source-tier threshold validation

Failed items degrade or are removed from primary picks.

### 5.5 Learning suggestions (daily)

Generate three recommendation tracks:
- Taste upgrade
- Cognition upgrade (framework-level guidance)
- Knowledge expansion (1–3 next topics)

Also include one actionable “today plan” (short guided reading routine).

## 6. Frontend H5 Information Architecture

### 6.1 Pages

- **Daily index page**
  - Date/issue header + daily keywords
  - Top signals summary
  - Featured list (10–20)
  - Category sections
  - Learning suggestions
  - Footer metadata

- **Article detail page**
  - Category/title/source/time/reading-time header
  - Fact summary
  - Deep AI commentary
  - Opportunity/risk block
  - Personalized relevance block
  - Origin/source links
  - Feedback controls

### 6.2 Card fields

Each featured card shows:
- cover image
- category
- title
- short excerpt (2–3 lines)
- composite score
- mini subscore hints (credibility/relevance)
- one-line AI view
- source + publish time
- deep-read entry

### 6.3 UX principles

- Clean, low-noise, whitespace-rich
- Strong typography hierarchy
- Neutral palette with restrained category accents
- Mobile-first performance
- Preserve scroll context when navigating back
- Basic accessibility semantics and contrast

### 6.4 Personalization feedback UI

In detail page include lightweight controls:
- sentiment: like / neutral / disagree
- correction reason: irrelevant / weak conclusion / unreliable source
- optional note

## 7. Engineering and Deployment Design

### 7.1 Suggested repository layout

- `cmd/daily-builder/`
- `internal/rss/`
- `internal/content/`
- `internal/analyze/`
- `internal/rank/`
- `internal/render/`
- `internal/publish/`
- `internal/profile/`
- `web/templates/`
- `web/static/`
- `config/`
- `output/`
- `state/`

### 7.2 Scheduling

Local scheduler (cron/launchd):
- Trigger around `05:30`
- Hard deadline before `07:00`
- Primary command shape: `daily-builder run --date today --mode morning`

### 7.3 SFTP publish strategy

Two-phase publish to avoid partial exposure:
1. Upload to `/_staging/YYYYMMDD/`
2. Validate completeness, then atomically switch to formal date dir + refresh `/latest/`

### 7.4 Config and secrets

Config files:
- `config/rss_sources.yaml`
- `config/scoring.yaml`
- `config/ai.yaml`

Secrets via environment variables only:
- `SFTP_HOST`, `SFTP_PORT`, `SFTP_USER`, `SFTP_KEY_PATH`, `SFTP_REMOTE_DIR`

### 7.5 Observability

Per-day outputs:
- `meta.json`
- `run.log`

Core metrics:
- source ingest success rate
- AI success/failure ratio
- selected item count
- category distribution
- publish timestamp

### 7.6 Publish gate checks

Automated checks before release:
1. expected files exist
2. internal link integrity
3. JSON schema validation
4. minimum featured count (>=10)
5. remote post-upload smoke check (`latest/index.html`)

## 8. Explicit MVP Exclusions

- No multi-user auth/permissions
- No dynamic online API service requirement
- No admin dashboard in MVP
- No social/community features

## 9. Success Criteria (MVP)

A day is successful when:
1. Daily edition is reachable before 07:00
2. Featured count is within 10–20
3. Each featured item has source + evidence-backed analysis
4. Personalization output is present on each featured item
5. Learning guidance section is generated and readable
6. Publish checks pass with no broken core paths

## 10. Next Step

After this design is approved, proceed to implementation planning (`writing-plans`) and then execute in phased milestones (scaffold → ingestion → analysis → render → publish → verification).
