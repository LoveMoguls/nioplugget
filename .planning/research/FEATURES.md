# Feature Research

**Domain:** AI-driven educational study app for Swedish national tests (åk 9)
**Researched:** 2026-04-03
**Confidence:** MEDIUM-HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| User authentication (parent + child) | Any app with accounts needs auth | LOW | Parent logs in with email/password; child logs in with invite link + PIN — no email required for child (Plughorse model) |
| Parent account with child management | Parents are the paying/setup party; they expect oversight | MEDIUM | Parent creates account, generates invite link, sees child profiles |
| Invite-link child onboarding | Kids can't be expected to handle email registration | LOW | Token-based invite link; child picks display name + sets PIN |
| Subject selection (Biologi, Samhälle, Matte) | Users need to navigate to what they want to study | LOW | Top-level navigation; 3 subjects, each with 4 sub-areas |
| Exercise selection per subject area | Granular selection expected once in a subject | LOW | ~50-60 pre-built exercises mapped against Skolverket's centrala innehåll |
| AI dialogue tutor per exercise | Core product experience; users came for this | HIGH | AI asks leading questions, never gives direct answers; system prompt customized per exercise |
| Session persistence (chat history) | Users expect to revisit what they did | MEDIUM | Sessions stored with full message log and score; accessible from history view |
| Score/result after session | Feedback loop on how well session went | MEDIUM | Score computed per session (basis for SM-2 scheduling); shown as summary screen |
| Spaced repetition scheduling (SM-2) | Apps in this space universally do this; users expect "smart" repetition | MEDIUM | SM-2 algorithm schedules next review date per exercise based on session score |
| "Due today" dashboard | Without this, users don't know what to do next | LOW | Cards showing exercises due for repetition today; entry point for daily usage |
| Progress view per subject | Parents and students want to see improvement | MEDIUM | Visual breakdown of strong/weak areas per subject; derived from session history |
| Mobile-responsive UI | Most study sessions happen on phones | LOW | SvelteKit + Tailwind; no native app, but fully usable on mobile |
| Subject-locked AI guardrails | Kids will try to use the AI for off-topic things; parents/teachers expect safety | MEDIUM | System prompt enforces topic scope; AI refuses off-topic requests politely |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Nationella prov-specific curriculum mapping | Generic AI tutors cover everything poorly; depth on åk 9 NP content is rare | HIGH | ~50-60 exercises hand-crafted against Skolverket's centrala innehåll for Bio, SO, Matte — quality beats breadth |
| BYOK (Bring Your Own API Key) | No subscription fee; no vendor lock-in; transparent cost; families pay ~$1-5/month in actual usage | MEDIUM | Parent inputs Claude API key encrypted at rest (AES-256); app uses it for all AI calls |
| Socratic dialogue (questions only, never answers) | Most AI tutors drift toward answer-giving; strict Socratic discipline improves retention | HIGH | System prompt enforces no-answer policy; per-exercise prompts calibrate difficulty and domain vocabulary; this is the hardest part to get right |
| Exercise system prompts tuned per topic | Generic prompts produce generic tutoring; topic-specific prompts activate domain language and correct misconceptions | HIGH | Each of ~50-60 exercises has its own system prompt referencing specific curriculum content (e.g., "Ekologi: Näringsnät och energiflöde") |
| Parent dashboard without nagging | Parents want oversight but don't want to manually prod kids — automated scheduling does this | LOW | Spaced repetition surfaces "due today" exercises without parental intervention; parent sees progress passively |
| Clean, calm UI for studying | Duolingo-style gamification creates anxiety and distraction for exam prep; a calm interface suits focused study | LOW | Minimal decorations; no streaks, no points loss, no shame mechanics |
| Swedish-language app and content | Competitors (Khanmigo, Duolingo) are English-first; Swedish students need Swedish dialogue | MEDIUM | All UI, AI dialogue, and exercise content in Swedish; Claude is capable in Swedish |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Direct AI answers on demand | Students want fast answers; feels like productivity | Defeats the pedagogical purpose; trains passivity; users learn less and realize this eventually | Maintain strict Socratic policy; add hint system that still asks guiding questions |
| Photo upload of homework | Students want to ask about any exercise | Scope creep; hard to OCR/parse reliably; opens cheating path for school assignments | Stick to pre-built exercises; photo upload is out-of-scope per PROJECT.md |
| Leaderboards / competitive scores | Gamification feels motivating | Creates anxiety for weaker students; gamification around exam prep is counterproductive; Duolingo-style shame is specifically avoided here | Show only personal progress, not relative ranking |
| Streaks with loss penalties | Duolingo-style retention hook | Adds anxiety; study consistency should come from spaced repetition scheduling ("due today"), not fear of breaking a streak | "Due today" cards are the retention mechanic — intrinsic motivation, not guilt |
| Real-time chat between students | Social features feel engaging | Changes product category (community platform); moderation burden; distraction from studying | Keep it individual AI dialogue; parents can discuss progress externally |
| In-app subscription / payment | Monetization is natural | Requires payment infrastructure, GDPR/PCI compliance, billing support; adds complexity before validation | BYOK model: user provides API key, no payment infrastructure needed at all |
| Multi-model AI selection (GPT-4, Gemini, etc.) | Power users want choice | API abstraction layer adds complexity; Claude is best-in-class for Socratic dialogue and Swedish; premature optimization | Lock to Claude API initially; add abstraction layer only if validated need arises |
| Push notifications | Retention hook | Requires mobile app or service worker complexity; spaced repetition "due today" on dashboard is sufficient for motivated students | "Due today" dashboard card is the nudge; email digest is acceptable v1.x addition |
| Free tier with limited sessions | Lowers barrier to entry | Without BYOK simplicity, a free tier requires server-side AI costs the project can't absorb; creates two-tier UX complexity | BYOK is the free tier — cost is what the API actually costs, $0-5/month |

## Feature Dependencies

```
Parent Registration + API Key Storage
    └──requires──> Child Profile Creation (invite link)
                       └──requires──> Child Authentication (PIN login)
                                          └──requires──> Subject/Exercise Selection
                                                             └──requires──> AI Dialogue Session
                                                                                └──requires──> Session Score + Persistence
                                                                                                   └──requires──> SM-2 Scheduling
                                                                                                                      └──requires──> "Due Today" Dashboard

Progress View ──requires──> Session Score + Persistence (needs historical data)

AI Guardrails ──requires──> AI Dialogue Session (enforced via system prompt)

Exercise System Prompts ──requires──> Curriculum Content Mapping (content must exist first)
```

### Dependency Notes

- **Child Authentication requires Parent Registration:** The invite link is generated by the parent account; no parent account = no invite link = no child onboarding.
- **AI Dialogue requires API Key Storage:** Every chat request proxies through the parent's Claude API key; missing key = no chat.
- **SM-2 Scheduling requires Session Score:** SM-2 inputs are quality score (0-5) from each session; without scoring, scheduling is impossible.
- **"Due Today" Dashboard requires SM-2 Scheduling:** The dashboard is the output of the scheduling system; it's meaningless without it.
- **Progress View requires Session Persistence:** Strength/weakness breakdown is derived from aggregated session history; it's a read view, not an input.
- **Exercise System Prompts require Curriculum Mapping:** Cannot write tutor prompts without knowing what each exercise covers; content work precedes prompt work.

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [ ] Parent registration + Claude API key input (encrypted at rest) — gating mechanic; nothing works without this
- [ ] Child profile via invite link + PIN login — the user experience starts here
- [ ] Subject and exercise selection — navigation to the core product
- [ ] AI dialogue session with Socratic system prompt — the product itself
- [ ] Session score + persistence — required for spaced repetition
- [ ] SM-2 spaced repetition scheduling — the learning flywheel
- [ ] "Due today" dashboard — entry point for return visits
- [ ] ~50-60 exercises mapped against Skolverket curriculum (Bio, SO, Matte) — the content without which the product is empty
- [ ] Subject-locked AI guardrails — safety / trust requirement for parents
- [ ] Mobile-responsive UI — most users will be on phone

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] Progress view (strength/weakness per subject) — trigger: users ask "am I getting better?"
- [ ] Parent dashboard with child progress overview — trigger: parents want passive oversight without asking the child
- [ ] Session history browse (past conversations) — trigger: users want to revisit past explanations

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Additional subjects beyond Bio, SO, Matte — trigger: validated demand for expansion
- [ ] Email digest for due exercises — trigger: low return-visit rate despite "due today" cards
- [ ] Multi-language support (English UI) — trigger: non-Swedish user demand
- [ ] Export progress report (PDF for parents) — trigger: school/parent admin request

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Parent registration + API key storage | HIGH | LOW | P1 |
| Child invite link + PIN login | HIGH | LOW | P1 |
| Exercise selection (subject + area) | HIGH | LOW | P1 |
| AI dialogue session (Socratic) | HIGH | HIGH | P1 |
| ~50-60 curriculum-mapped exercises | HIGH | HIGH | P1 |
| Session score + persistence | HIGH | MEDIUM | P1 |
| SM-2 spaced repetition scheduling | HIGH | MEDIUM | P1 |
| "Due today" dashboard | HIGH | LOW | P1 |
| Subject-locked AI guardrails | HIGH | MEDIUM | P1 |
| Mobile-responsive UI | HIGH | LOW | P1 |
| Progress view (strength/weakness) | MEDIUM | MEDIUM | P2 |
| Parent overview dashboard | MEDIUM | LOW | P2 |
| Session history browse | MEDIUM | LOW | P2 |
| Email digest notifications | LOW | MEDIUM | P3 |
| Additional subjects | MEDIUM | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | Khanmigo | Plughorse | Duolingo | Nioplugget approach |
|---------|----------|-----------|----------|---------------------|
| Tutoring model | Socratic (GPT-4); no direct answers | Socratic dialogue; "active learning" | Gamified exercises; AI explanation on demand | Strict Socratic (Claude); per-exercise system prompts |
| Subject coverage | Broad (math, science, coding, humanities) | All åk 7-9 subjects + homework photo upload | Language learning only | Narrow and deep: Bio/SO/Matte åk 9 NP only |
| Content model | Tied to Khan Academy content library | 500+ pre-built modules + user-uploaded homework | Curriculum generated by Duolingo | ~50-60 hand-crafted exercises mapped to Skolverket centrala innehåll |
| Parent/child model | Parent links child account; parent sees chat history | Parent account; invite link for child; no child email needed | Family plan (account sharing, not separate roles) | Parent account + invite link + PIN (no child email); parent stores API key |
| Spaced repetition | Not prominent | Automatic repetition scheduling | Core feature (Half-Life Regression model) | SM-2 algorithm; "due today" cards |
| Progress tracking | Session history; Khan Academy exercise progress | Parent dashboard | XP, streaks, leagues | Session history + score; progress view with strong/weak areas |
| Monetization | $4/month subscription | ~99 SEK/month subscription | Free + subscription tiers | BYOK: user pays API directly; no subscription infrastructure |
| Safety/guardrails | Chat history for parents; content filters | Not publicly documented | Moderated content | Subject-locked prompts; API key controlled by parent |
| Language | English-first | Swedish | 40+ languages | Swedish throughout |
| Gamification | Energy points → avatar customization | Not documented | Heavy (streaks, leagues, XP) | Deliberately minimal; calm UI for exam prep |

## Sources

- [Khanmigo Learners Page](https://www.khanmigo.ai/learners) — MEDIUM confidence (official marketing page)
- [Khanmigo Review 2026 (kidsaitools.com)](https://www.kidsaitools.com/en/articles/khanmigo-review-parents-complete-2026) — MEDIUM confidence (review site)
- [Plughorse.com](https://www.plughorse.com) — MEDIUM confidence (official product page, fetched directly)
- [Duolingo AI Innovations](https://foralink.io/blogs/duolingos-ai-innovations-transforming-language-learning-and-beyond) — LOW confidence (third-party blog)
- [Duolingo Spaced Repetition Research Paper](https://research.duolingo.com/papers/settles.acl16.pdf) — HIGH confidence (official Duolingo research)
- [SocraticAI: Transforming LLMs into Guided CS Tutors](https://arxiv.org/abs/2512.03501) — HIGH confidence (peer-reviewed arxiv)
- [AI tutoring RCT in UK classrooms](https://arxiv.org/html/2512.23633v1) — HIGH confidence (research paper)
- [SM-2 Algorithm Explained](https://tegaru.app/en/blog/sm2-algorithm-explained) — MEDIUM confidence (multiple corroborating sources)
- [Khanmigo Parents Safety Features](https://www.khanmigo.ai/parents) — MEDIUM confidence (official)
- [BYOK model overview](https://www.byok.tech/) — MEDIUM confidence (corroborated by JetBrains, Vercel, and others)
- [Matematik 9 app (App Store)](https://apps.apple.com/se/app/matematik-9/id825478026) — MEDIUM confidence (evidence of Swedish NP study app market)

---
*Feature research for: AI-driven educational study app for Swedish nationella prov åk 9*
*Researched: 2026-04-03*
