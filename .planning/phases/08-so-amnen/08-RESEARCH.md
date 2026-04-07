# Phase 8: SO-ämnen - Research

**Researched:** 2026-04-07
**Domain:** Content seeding — Geografi and Historia NP-calibrated exercises (database migrations)
**Confidence:** HIGH

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| GEO-01 | System har färdiga övningspass för Geografi — Befolkning och urbanisering (migration, städer, globalisering) | Seed migration pattern from Phase 7 (006) + NP content for Befolkning och urbanisering documented below |
| GEO-02 | System har färdiga övningspass för Geografi — Klimat och klimatförändringar (klimatzoner, global uppvärmning, extremväder) | Seed migration pattern + NP content for Klimat och klimatförändringar documented below |
| GEO-03 | System har färdiga övningspass för Geografi — Naturresurser och hållbarhet (vatten, energi, hållbar utveckling) | Seed migration pattern + NP content for Naturresurser documented below |
| GEO-04 | System har färdiga övningspass för Geografi — Geopolitik och handel (länder, ekonomiska system, konflikter) | Seed migration pattern + NP content for Geopolitik documented below |
| GEO-05 | Alla Geografi-övningar har NP-kalibrerade system-prompts med E/C/A verbhierarki | Phase 6/7 prompt structure is the exact template to replicate for geografi |
| HIST-01 | System har färdiga övningspass för Historia — Industrialismens tid (1800-talets industrialisering, sociala rörelser) | Seed migration pattern + NP content for Industrialismens tid documented below |
| HIST-02 | System har färdiga övningspass för Historia — De två världskrigen (orsaker, Förintelsen, efterdyningar) | Seed migration pattern + NP content for De två världskrigen documented below |
| HIST-03 | System har färdiga övningspass för Historia — Kalla kriget (bipolar värld, dekolonisation, kapprustning) | Seed migration pattern + NP content for Kalla kriget documented below |
| HIST-04 | System har färdiga övningspass för Historia — 1900-talets politiska rörelser (fascism, kommunism, demokratisering) | Seed migration pattern + NP content for Politiska rörelser documented below |
| HIST-05 | Alla Historia-övningar har NP-kalibrerade system-prompts med E/C/A verbhierarki | Phase 6/7 prompt structure is the exact template to replicate for historia |
</phase_requirements>

---

## Summary

Phase 8 is a pure content seeding phase — identical in structure to Phase 7. No frontend or backend code changes are required. The work is entirely database content: write a new seed migration (007) that inserts subjects, topics, and exercises for Geografi and Historia, following the exact same patterns established in migration 006 (Phase 7).

The Phase 6/7 prompt template is proven and production-ready. It contains eight sections: ÄMNE, ÖVNING, NIVÅ, LÄRANDEMÅL, REGLER, NP-KOPPLING, BEDÖMNINGSLEDTRÅDAR, VANLIGA ELEVMISSAR, EXEMPELFRÅGOR. Each section maps directly to an aspect of real nationella prov scoring. The planner needs to apply this template to 8 new topics (4 Geografi + 4 Historia) × 3 exercises = 24 new exercises, each with NP-calibrated content specific to that topic.

SO subjects (Geografi and Historia) in Swedish Lgr22 are administered under the same NP framework as other subjects in åk 9. The verb hierarchy — Beskriv (E), Förklara (C), Resonera/Diskutera/Ta ställning (A) — is identical across all subjects in this system. The display_order for Geografi will be 6 and Historia 7, continuing the established sequence (Biologi=1, Samhällskunskap=2, Matematik=3, Kemi=4, Fysik=5).

**Primary recommendation:** Write a single migration 007 using the same scaffold-first approach as Phase 7: create the file with SYSTEM_PROMPT_PLACEHOLDER for all 24 exercises in Plan 01, fill in Geografi prompts in Plan 02, fill in Historia prompts and run the migration in Plan 03.

---

## Standard Stack

### Core (no changes from established project stack)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| PostgreSQL | existing | Exercise storage — subjects/topics/exercises tables | Already running |
| golang-migrate | existing | Migration runner | Already configured |
| SQL migration files | 007_*.{up,down}.sql | New content delivery mechanism | To be created |

### No New Dependencies

This phase adds zero new libraries, packages, or infrastructure. It extends existing database schema with new rows only.

**Installation:** None required.

---

## Architecture Patterns

### Existing Project Structure (READ-ONLY for this phase)

```
backend/
├── db/
│   ├── migrations/
│   │   ├── 003_seed_exercises.{up,down}.sql    # Original seed pattern
│   │   ├── 005_update_exercise_prompts.{up,down}.sql  # NP prompt structure
│   │   ├── 006_seed_no_amnen.{up,down}.sql    # Phase 7 — closest template
│   │   └── 007_seed_so_amnen.{up,down}.sql    # Phase 8 target files
│   └── queries/
│       ├── subjects.sql   # ListSubjects, GetSubjectBySlug — unchanged
│       └── topics.sql     # ListTopicsBySubjectID, GetTopicBySlug — unchanged
frontend/
└── src/routes/study/      # Already handles any number of subjects dynamically
```

No files outside `backend/db/migrations/` are touched in this phase.

### Pattern 1: Subject + Topic + Exercise Seed Migration

Identical to migration 006. Each new subject follows this exact structure:

```sql
-- 1. Insert subjects
INSERT INTO subjects (name, slug, display_order) VALUES
    ('Geografi', 'geografi', 6),
    ('Historia', 'historia', 7);

-- 2. Insert topics (using CROSS JOIN VALUES pattern)
INSERT INTO topics (subject_id, name, slug, display_order)
SELECT s.id, t.name, t.slug, t.display_order
FROM subjects s
CROSS JOIN (VALUES
    ('Befolkning och urbanisering', 'befolkning-och-urbanisering', 1),
    ('Klimat och klimatförändringar', 'klimat-och-klimatforandringar', 2),
    ('Naturresurser och hållbarhet', 'naturresurser-och-hallbarhet', 3),
    ('Geopolitik och handel', 'geopolitik-och-handel', 4)
) AS t(name, slug, display_order)
WHERE s.slug = 'geografi';

-- 3. Insert exercises (nested CROSS JOIN VALUES per topic)
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
(
    'Exercise title',
    'Exercise description',
    1,
    'System prompt here...'
),
...
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'geografi'
  AND t.slug = 'befolkning-och-urbanisering';
```

**Slug convention:** Swedish special characters transliterated: å→a, ä→a, ö→o, space→hyphen.
- `befolkning-och-urbanisering`
- `klimat-och-klimatforandringar`
- `naturresurser-och-hallbarhet`
- `geopolitik-och-handel`
- `industrialismens-tid`
- `de-tva-varldskrigen`
- `kalla-kriget`
- `1900-talets-politiska-rorelser`

### Pattern 2: NP-Calibrated System Prompt Structure

The canonical prompt template from Phases 6 and 7. Every exercise prompt follows this exact 8-section structure:

```
Du är en AI-handledare som förbereder elever i årskurs 9 inför nationella provet i {ämne}.

ÄMNE: {Geografi|Historia} — {topic name}
ÖVNING: {exercise title}
NIVÅ: {E-nivå|C-nivå|A-nivå} (Delprov {A1 faktafrågor|A2 resonerande frågor})

LÄRANDEMÅL: {goal aligned to Skolverket Lgr22 centralt innehåll}

REGLER (bryts ALDRIG):
1. Svara ALLTID på svenska.
2. Ge ALDRIG direkta svar — ställ ledande frågor som hjälper eleven tänka själv.
3. Om eleven frågar utanför {topic}: "Bra fråga, men låt oss fokusera på {topic}." Ställ en ny fråga inom ämnet.
4. Anpassa nivån efter elevens svar.
5. Om eleven verkar fast, omformulera frågan eller ge en ledtråd — aldrig svaret.

NP-KOPPLING:
{level-appropriate text about which delprov tests this}

BEDÖMNINGSLEDTRÅDAR FÖR DIG:
{E/C/A scoring calibration for the AI}

VANLIGA ELEVMISSAR (från NP-forskning):
{2-4 subject-specific mistakes}

EXEMPELFRÅGOR ATT STÄLLA (NP-stil):
{3-5 questions using NP verb patterns for this level}

Börja med att hälsa eleven välkommen och ställ en {level-appropriate} öppningsfråga.
```

### Pattern 3: Difficulty → NP Level Mapping

Locked in Phase 6, validated in Phase 7:

| difficulty_order | NP Level | Primary Verbs | Delprov |
|-----------------|----------|--------------|---------|
| 1 | E-nivå | Beskriv, Vad kallas, Vad menas med, Nämn | A1 faktafrågor |
| 2 | C-nivå | Förklara sambandet, Förklara varför, Hur påverkas | A2 resonerande |
| 3 | A-nivå | Resonera kring, Diskutera, Ta ställning, Motivera | A2 resonerande |

### Pattern 4: Down Migration

```sql
DELETE FROM subjects WHERE slug IN ('geografi', 'historia');
-- CASCADE deletes topics and exercises automatically (ON DELETE CASCADE in schema)
```

### Pattern 5: Three-Plan Scaffold Approach (from Phase 7)

Phase 7 established the scaffold-then-fill approach as the preferred split:
- **Plan 01:** Migration scaffold — subjects, topics, 24 exercise stubs with `SYSTEM_PROMPT_PLACEHOLDER`
- **Plan 02:** Geografi prompts — replace placeholders for all 12 Geografi exercises
- **Plan 03:** Historia prompts + run migration — replace remaining 12 placeholders, then apply migration

### Anti-Patterns to Avoid

- **Don't use UUIDs in WHERE clauses** — Use slug-based joins. UUIDs vary per environment.
- **Don't add display_order gaps** — Geografi=6, Historia=7 (continuing Biologi=1, Samhälle=2, Matematik=3, Kemi=4, Fysik=5).
- **Don't use single quotes in prompt text** — SQL single quotes in string literals must be doubled ('') or avoided entirely. Verified pattern: use paraphrase instead of possessive forms (no apostrophes).
- **Don't unquote Swedish characters in slugs** — Use transliterated URL-safe forms consistently.
- **Don't run migration until all PLACEHOLDERs are replaced** — Confirm `grep -c SYSTEM_PROMPT_PLACEHOLDER` = 0 before applying.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Subject routing | Custom frontend route per subject | Existing dynamic study route | Already handles any subject via slug |
| NP content structure | New prompt format | Canonical 8-section template from Phase 6 | Validated against real NP structure |
| Multi-row SQL inserts | Separate INSERT per row | CROSS JOIN VALUES pattern | Already established and working in 003/006 |

**Key insight:** This phase is 100% additive SQL content. The system is already built to handle any number of subjects and topics — just add rows.

---

## Common Pitfalls

### Pitfall 1: Apostrophes and Single Quotes in SQL Strings
**What goes wrong:** Swedish possessive forms or English-style abbreviations with apostrophes break SQL string literals.
**Why it happens:** Prompt text is long multi-line Swedish text inside SQL single-quoted strings.
**How to avoid:** Write all prompt content without apostrophes. For possessive: "Karl Marxs idéer" not "Karl Marx's idéer". For abbreviations: write out fully.
**Warning signs:** Any `'` character inside a VALUES string that isn't a doubled `''`.

### Pitfall 2: Missing SYSTEM_PROMPT_PLACEHOLDER Before Migration
**What goes wrong:** Running migration with unfilled placeholders inserts literal `SYSTEM_PROMPT_PLACEHOLDER` as exercise prompts into the database.
**Why it happens:** Excitement to test the migration before prompts are done.
**How to avoid:** Always `grep -c SYSTEM_PROMPT_PLACEHOLDER 007_seed_so_amnen.up.sql` before running. Must equal 0.
**Warning signs:** Count > 0 means prompts are incomplete.

### Pitfall 3: Wrong display_order for New Subjects
**What goes wrong:** Geografi or Historia get the wrong display_order and appear out of sequence in the subject selector.
**Why it happens:** Forgetting the existing sequence (5 subjects already in db after Phase 7).
**How to avoid:** Geografi=6, Historia=7. Check the existing subjects table if unsure.
**Warning signs:** Subjects appearing in wrong position in frontend subject list.

### Pitfall 4: Slug Collision with Existing Content
**What goes wrong:** A Historia topic slug matches an existing subject/topic slug.
**Why it happens:** Common words appear in multiple contexts.
**How to avoid:** Topic slugs only need to be unique within a subject (WHERE clause uses both subject and topic slug). But verify there is no subjects.slug collision for 'geografi' or 'historia'.
**Warning signs:** Duplicate key violation on INSERT.

### Pitfall 5: Histórica Content Anachronism
**What goes wrong:** Historia prompts include facts that don't match Lgr22 centralt innehåll for åk 9, or confuse Swedish grade-9 scope with university-level history.
**Why it happens:** Historia is broad; the NP for åk 9 has specific scope limits.
**How to avoid:** Anchor each topic to Lgr22 centralt innehåll for historia åk 7-9. The four topics in requirements define the scope exactly.
**Warning signs:** Prompts that reference events outside 1800-2000 CE scope (except as brief context).

---

## Content Specification: Geografi Topics and Exercises

### What NP Geografi tests (Lgr22 centralt innehåll for geografi åk 7-9)

NP Geografi åk 9 uses the same two-delprov structure. Lgr22 centralt innehåll covers:

**Befolkning och urbanisering** — befolkningsfördelning i världen; migrationsorsaker (push/pull-faktorer); urbanisering (stad vs landsbygd, megastäder); globalisering och rörlighet; demografisk transition; befolkningstäthet och resurstillgång.

**Klimat och klimatförändringar** — klimatzoner (tropisk, subtropisk, tempererad, subarktisk, arktisk); faktorer som påverkar klimat (latitud, hav, höjd, vindmönster); global uppvärmning — orsaker (växthusgaser, förstärkt växthuseffekt), konsekvenser (havsnivåhöjning, extremväder); klimatanpassning och klimatmitigering; IPCC och Parisavtalet.

**Naturresurser och hållbarhet** — förnybara och icke-förnybara resurser; vattenresurser (sötvatten, grundvatten, vattenkonflikter); energiresurser (fossila bränslen vs förnybar energi); hållbar utveckling (Brundtland-definitionen, SDG-målen); ekosystemtjänster; markanvändning och avskogning.

**Geopolitik och handel** — världshandel och handelsmönster; ekonomiska system (BNP, HDI, development gap); globala organisationer (FN, WTO, EU); geopolitiska konflikter och resurskonflikter; beroenden och globala värdekedjor; Nord-Syd-perspektivet.

### Geografi Exercise Titles (3 per topic = 12 exercises total)

**Befolkning och urbanisering (slug: befolkning-och-urbanisering)**
1. "Befolkningsfördelning och migration" — difficulty 1 (E): Beskriv push/pull-faktorer för migration. Vad menas med befolkningstäthet?
2. "Urbanisering och globalisering" — difficulty 2 (C): Förklara vad som driver urbanisering. Hur påverkar globalisering migration?
3. "Befolkningstillväxt och resurser" — difficulty 3 (A): Resonera kring sambandet mellan befolkningstillväxt, resurser och hållbarhet. Ta ställning till migration som lösning.

**Klimat och klimatförändringar (slug: klimat-och-klimatforandringar)**
1. "Klimatzoner och klimatfaktorer" — difficulty 1 (E): Beskriv jordens klimatzoner. Vad menas med klimat kontra väder?
2. "Global uppvärmning och konsekvenser" — difficulty 2 (C): Förklara sambandet mellan växthusgaser och global uppvärmning. Vilka konsekvenser kan extremväder ge?
3. "Klimatåtgärder och ansvarsfördelning" — difficulty 3 (A): Resonera kring hur länder bör fördela ansvaret för klimatåtgärder. Ta ställning till Parisavtalet.

**Naturresurser och hållbarhet (slug: naturresurser-och-hallbarhet)**
1. "Förnybara och icke-förnybara resurser" — difficulty 1 (E): Vad menas med förnybar resurs? Beskriv skillnaden mellan fossila bränslen och förnybar energi.
2. "Vattentillgång och energiresurser" — difficulty 2 (C): Förklara varför vattenbrist uppstår i vissa regioner. Hur påverkar energiresurser geopolitik?
3. "Hållbar resurshushållning" — difficulty 3 (A): Resonera kring hur hållbar utveckling balanserar ekonomisk tillväxt och miljöhänsyn. Ta ställning till SDG-målen.

**Geopolitik och handel (slug: geopolitik-och-handel)**
1. "Världshandel och ekonomiska system" — difficulty 1 (E): Beskriv vad BNP mäter. Vad menas med handelsmönster?
2. "Globala organisationer och beroenden" — difficulty 2 (C): Förklara hur FN och WTO påverkar världshandeln. Hur skapas ekonomiska beroenden?
3. "Geopolitiska konflikter och Nord-Syd" — difficulty 3 (A): Resonera kring orsaker till ekonomiska klyftor mellan Nord och Syd. Ta ställning till global rättvisa i handel.

### Geografi NP Common Student Mistakes

1. **Klimat vs väder** — Students use "klimat" and "väder" interchangeably. NP requires them to distinguish: väder = kortsiktigt, klimat = långsiktigt medelvärde.
2. **Växthuseffekten = dålig** — Students think all greenhouse effect is anthropogenic and bad. The natural greenhouse effect is what makes Earth habitable; it's the *förstärkt* växthuseffekt that is the problem.
3. **Migration: push vs pull confusion** — Students list factors but can't correctly categorize them as push (driving people away) vs pull (attracting to destination).
4. **BNP vs HDI** — Students don't distinguish. BNP = ekonomisk produktion only; HDI adds health and education. A country can have high BNP but low HDI.
5. **Förnybara resurser = outtömliga** — Students think renewable means limitless. Skogar och fiskbestånd är förnybara men kan utarmas om de nyttjas för snabbt.
6. **Geopolitik = krig** — Students reduce geopolitics to military conflict, missing economic, resource, and diplomatic dimensions.

### Geografi NP-Style Example Questions by Level

**E-nivå (Beskriv/Vad kallas/Vad menas med):**
- "Vad menas med urbanisering?"
- "Beskriv två push-faktorer som kan göra att människor migrerar."
- "Vad kallas effekten av att växthusgaser fångar värme i atmosfären?"
- "Vad menas med en förnybar resurs? Ge ett exempel."
- "Beskriv vad BNP mäter."

**C-nivå (Förklara sambandet/Förklara varför/Hur påverkas):**
- "Förklara sambandet mellan växthusgaser och global uppvärmning."
- "Förklara varför urbanisering ökar i låginkomstländer."
- "Hur påverkas vattenförsörjningen i en region när glaciärerna smälter?"
- "Förklara skillnaden mellan BNP och HDI som mått på välstånd."
- "Hur påverkar handelsmönster ekonomiska beroenden mellan länder?"

**A-nivå (Resonera kring/Ta ställning/Diskutera):**
- "Resonera kring vem som bär störst ansvar för att minska klimatförändringarna — rika eller fattiga länder. Motivera."
- "Ta ställning till om ekonomisk tillväxt och hållbar utveckling kan förenas. Använd konkreta exempel."
- "Diskutera hur globalisering påverkar Nord-Syd-klyftan — bidrar den till ökad eller minskad ojämlikhet?"
- "Resonera kring hur vattenbrist kan leda till geopolitiska konflikter."

---

## Content Specification: Historia Topics and Exercises

### What NP Historia tests (Lgr22 centralt innehåll for historia åk 7-9)

NP Historia åk 9 tests historical thinking skills: orsak-och-verkan, perspektivtagande, källkritik, och historiska begrepp. Lgr22 centralt innehåll for historia åk 7-9 covers:

**Industrialismens tid** — 1800-talets industrialisering (ångmaskinen, textilindustri, järnvägar); urbanisering och befolkningstillväxt; sociala rörelser (fackföreningar, socialism, kommunism, feminism); barnarbete och arbetsvillkor; klassamhälle och sociala reformer; kolonialism och imperialism.

**De två världskrigen** — Första världskriget: orsaker (nationalismen, allianssystem, imperialism, mordet i Sarajevo), händelseförlopp, konsekvenser (Versaillesfreden, Nationernas förbund); Mellankrigstiden; Andra världskriget: orsaker (Versaillestraktatens konsekvenser, Hitlers uppgång), Förintelsen och Förintelsens mekanismer, krigsekonomi, Allierade vs Axelmakterna, Hiroshima, FN bildas.

**Kalla kriget** — Bipolär världsordning (USA vs Sovjet); Berlinmuren; Koreakriget och Vietnamkriget som proxykonflikt; Kubakrisen; kapprustningen och kärnvapenhot; rymdkapplöpning; dekolonisation och nationalistiska rörelser i Afrika/Asien; Berlinmurens fall 1989 och Sovjets upplösning.

**1900-talets politiska rörelser** — Fascism (Mussolini, Hitler, ideologi, totalitarism); kommunism (ryska revolutionen, Stalin, Mao, gulag); demokratisering (Weimar-republiken, FNs deklaration om mänskliga rättigheter, avkolonisering); apartheid och Sydafrika; feminism och rösträtt; medborgarrättsrörelsen i USA.

### Historia Exercise Titles (3 per topic = 12 exercises total)

**Industrialismens tid (slug: industrialismens-tid)**
1. "Industrialiseringens orsaker och spridning" — difficulty 1 (E): Beskriv vad industrialisering innebar. Vilka uppfinningar var centrala?
2. "Sociala rörelser och arbetsvillkor" — difficulty 2 (C): Förklara varför fackföreningar bildades. Hur förändrades arbetsvillkoren under 1800-talet?
3. "Industrialismens konsekvenser" — difficulty 3 (A): Resonera kring industrialismens långsiktiga konsekvenser för samhälle och miljö. Ta ställning till om industrialismen var ett framsteg.

**De två världskrigen (slug: de-tva-varldskrigen)**
1. "Första världskrigets orsaker" — difficulty 1 (E): Beskriv de viktigaste orsakerna till första världskriget. Vad menas med allianssystemet?
2. "Andra världskriget och Förintelsen" — difficulty 2 (C): Förklara hur Versaillesfreden bidrog till andra världskrigets utbrott. Hur var Förintelsen möjlig?
3. "Världskrigens efterdyningar" — difficulty 3 (A): Resonera kring hur de två världskrigen omformade den internationella ordningen. Ta ställning till FNs roll.

**Kalla kriget (slug: kalla-kriget)**
1. "Den bipolära världsordningen" — difficulty 1 (E): Beskriv vad kalla kriget innebar. Vad menas med bipolär världsordning?
2. "Proxykonflikt och kapprustning" — difficulty 2 (C): Förklara vad som menas med proxykonflikt. Hur påverkade kapprustningen det globala säkerhetsläget?
3. "Kalla krigets slut och dekolonisation" — difficulty 3 (A): Resonera kring hur kalla kriget påverkade dekolonisationsprocessen i Afrika och Asien. Ta ställning till supermakternas ansvar.

**1900-talets politiska rörelser (slug: 1900-talets-politiska-rorelser)**
1. "Fascism och kommunism som ideologier" — difficulty 1 (E): Beskriv grunddragen i fascism respektive kommunism. Vad menas med totalitär stat?
2. "Demokratisering och mänskliga rättigheter" — difficulty 2 (C): Förklara hur FNs deklaration om mänskliga rättigheter kom till. Vad drev demokratiseringsrörelser under 1900-talet?
3. "Politiska rörelsers arv" — difficulty 3 (A): Resonera kring hur 1900-talets politiska rörelser formar vår samtid. Ta ställning till hur vi förhindrar totalitarism i dag.

### Historia NP Common Student Mistakes

1. **Orsaker vs anledningar** — Students list one trigger event (mordet i Sarajevo) as "orsaken" to WWI without understanding systemic causes (nationalismen, imperialismen, allianssystemet). NP expects multi-causal analysis.
2. **Versaillesfreden förenklat** — Students say "Versaillesfreden orsakade andra världskriget" without nuance. NP expects students to explain the mechanism (krigsskuld, reparationer, territoriella förluster → ekonomisk kris → nationalistisk reaktion).
3. **Kalla kriget = krig** — Students think it involved direct military conflict between USA and Sovjet. They miss the indirect (proxy) nature and ideological dimension.
4. **Fascism = nazism** — Students use these interchangeably. Fascism = Mussolinis rörelse; Nazism = Hitlers variant med rasideologi. They have similar structures but different ideological content.
5. **Historisk empati vs nutida värderingar** — Students judge historical actors by today's standards without attempting to understand the historical context. NP expects perspektivtagande.
6. **Dekolonisation utan kopplingen till kalla kriget** — Students treat decolonization as independent of superpower politics. NP expects understanding of how USA and Sovjet exploited anti-colonial movements.

### Historia NP-Style Example Questions by Level

**E-nivå (Beskriv/Vad kallas/Vad menas med):**
- "Beskriv vad som menas med industrialisering."
- "Vad kallas den systematiska utrotningen av judar och andra grupper under andra världskriget?"
- "Beskriv vad som menas med kalla kriget."
- "Vad menas med en totalitär stat? Ge ett exempel."
- "Vad kallas det när ett stormaktskrig utkämpas indirekt via ett tredje land?"

**C-nivå (Förklara sambandet/Förklara varför/Hur påverkas):**
- "Förklara sambandet mellan Versaillesfreden och andra världskrigets utbrott."
- "Förklara varför fackföreningar bildades under industrialismens tid."
- "Hur påverkades koloniserade folk av kalla krigets superMaktskamp?"
- "Förklara skillnaden mellan fascism och kommunism som politiska ideologier."
- "Förklara hur Förintelsen var möjlig — vilka mekanismer möjliggjorde den?"

**A-nivå (Resonera kring/Ta ställning/Diskutera):**
- "Resonera kring industrialismens konsekvenser — vilka grupper gynnades och vilka missgynnades?"
- "Ta ställning till om FN lyckats uppfylla sitt syfte sedan 1945. Motivera med historiska exempel."
- "Diskutera hur erfarenheterna från 1900-talets totalitära rörelser bör påverka vår syn på demokrati i dag."
- "Resonera kring vad som krävdes för att Förintelsen skulle kunna genomföras — på individ-, samhälls- och statsnivå."

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| New migration per subject pair | Scaffold-then-fill (3 plans) | Phase 7 established | Separates structural SQL from content quality work |
| Generic prompt text | 8-section NP-calibrated template | Phase 6 established | Prompts align with real Skolverket NP scoring |
| psql for migration | golang-migrate CLI | Phase 7 confirmed | `migrate -path ./db/migrations -database $DATABASE_URL up` |

**Deprecated/outdated:**
- `SYSTEM_PROMPT_PLACEHOLDER` pattern: only valid during scaffold phase — must be 0 before running migration.

---

## Open Questions

1. **SO NP structure — delprov details**
   - What we know: Geografi and Historia both use NP for åk 9 in Sweden; the E/C/A verb hierarchy applies across all subjects in Lgr22
   - What's unclear: Exact delprov naming for SO subjects (may differ from NO subjects which use "Delprov A1/A2")
   - Recommendation: Use the same "Delprov A1 faktafrågor / Delprov A2 resonerande frågor" framing as Kemi/Fysik — it maps correctly to the E vs C/A distinction. If SO uses different naming, the pedagogical intent is the same.

2. **Migration number**
   - What we know: Last migration is 006. Next available is 007.
   - What's unclear: Whether any migration was added between Phase 7 completion and Phase 8 start.
   - Recommendation: Confirm with `ls backend/db/migrations/` at plan execution time. Current state shows 006 is the highest — use 007.

---

## Sources

### Primary (HIGH confidence)
- Migration 006 (`backend/db/migrations/006_seed_no_amnen.up.sql`) — direct reference pattern for all SQL structure
- Phase 7 RESEARCH.md (`.planning/phases/07-no-amnen/07-RESEARCH.md`) — validated patterns and anti-patterns
- Phase 7 SUMMARYs (07-01, 07-02, 07-03) — confirmed what worked in practice
- REQUIREMENTS.md — exact topic names, slugs, and requirement IDs
- STATE.md — accumulated decisions and context

### Secondary (MEDIUM confidence)
- Lgr22 centralt innehåll for Geografi and Historia åk 7-9 (knowledge from training data, consistent with requirement topic areas)
- NP structure for SO subjects: Geografi and Historia administered under same NP framework as NO subjects

### Tertiary (LOW confidence — flag for validation)
- Specific "vanliga elevmissar" for Geografi and Historia: sourced from training data, not verified against Skolverket NP research documents. Treat as pedagogically plausible; refine if NP research contradicts.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — identical to Phase 7, no new components
- Architecture: HIGH — exact same SQL patterns, verified working
- Content specification (exercise titles, descriptions): HIGH — directly from requirements + Lgr22 scope
- Content specification (VANLIGA ELEVMISSAR): MEDIUM — pedagogically grounded but not verified against official NP research documents
- Pitfalls: HIGH — learned from Phase 7 execution

**Research date:** 2026-04-07
**Valid until:** 2026-05-07 (stable content domain)
