# Phase 7: NO-ämnen - Research

**Researched:** 2026-04-06
**Domain:** Content seeding — Kemi and Fysik NP-calibrated exercises (database migrations)
**Confidence:** HIGH

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| KEMI-01 | System har färdiga övningspass för Kemi — Materiens uppbyggnad (atomer, molekyler, periodiska systemet) | Seed migration pattern from Phase 3 (003) + NP content for Materiens uppbyggnad documented below |
| KEMI-02 | System har färdiga övningspass för Kemi — Kemiska reaktioner (syror/baser, pH, förbränning) | Seed migration pattern + NP content for Kemiska reaktioner documented below |
| KEMI-03 | System har färdiga övningspass för Kemi — Kemikalier i vardagen (lösningar, kemikaliesäkerhet, blandningar) | Seed migration pattern + NP content for Kemikalier i vardagen documented below |
| KEMI-04 | System har färdiga övningspass för Kemi — Kemi och hållbar utveckling (miljökemi, fossila bränslen, alternativ energi) | Seed migration pattern + NP content for hållbar utveckling documented below |
| KEMI-05 | Alla Kemi-övningar har NP-kalibrerade system-prompts med E/C/A verbhierarki (Beskriv/Förklara/Resonera) | Phase 6 prompt structure is the exact template to replicate for kemi |
| FYS-01 | System har färdiga övningspass för Fysik — Krafter och rörelse (Newtons lagar, hastighet, acceleration) | Seed migration pattern + NP content for Krafter och rörelse documented below |
| FYS-02 | System har färdiga övningspass för Fysik — Elektricitet och magnetism (kretsar, elektromagnetism, induktion) | Seed migration pattern + NP content for Elektricitet documented below |
| FYS-03 | System har färdiga övningspass för Fysik — Energi och energiomvandlingar (mekanisk, termisk, elektrisk energi) | Seed migration pattern + NP content for Energi documented below |
| FYS-04 | System har färdiga övningspass för Fysik — Astronomi och universum (stjärnor, solsystemet, Big Bang) | Seed migration pattern + NP content for Astronomi documented below |
| FYS-05 | Alla Fysik-övningar har NP-kalibrerade system-prompts med E/C/A verbhierarki | Phase 6 prompt structure is the exact template to replicate for fysik |
</phase_requirements>

---

## Summary

Phase 7 is a pure content seeding phase — no frontend or backend code changes required. The work is entirely database content: write a new seed migration (006) that inserts subjects, topics, and exercises for Kemi and Fysik, following the exact same patterns established in migrations 003 (seed structure) and 005 (NP-calibrated prompts).

The Phase 6 prompt template is proven and production-ready. It contains eight sections: ÄMNE, ÖVNING, NIVÅ, LÄRANDEMÅL, REGLER, NP-KOPPLING, BEDÖMNINGSLEDTRÅDAR, VANLIGA ELEVMISSAR, EXEMPELFRÅGOR. Each section maps directly to an aspect of real nationella prov scoring. The planner needs to apply this template to 8 new topics (4 Kemi + 4 Fysik) × 3 exercises = 24 new exercises, each with NP-calibrated content specific to that topic.

The NP for Kemi and Fysik is administered by Umeå University's Department of Applied Educational Science (same group as Biologi). Both subjects use the same two-delprov structure (A=90 min + B=90 min) with a factual layer (Delprov A1-style) and a reasoning layer (Delprov A2-style). The verb hierarchy — Beskriv (E), Förklara (C), Resonera/Diskutera/Ta ställning (A) — is identical across all three NO subjects. The Skolverket/Lgr22 centralt innehåll for Kemi and Fysik maps directly to the 8 topic areas defined in the requirements.

**Primary recommendation:** Write a single migration 006 that inserts Kemi + Fysik subjects, topics, and 24 exercises using the exact same SQL pattern as migration 003 and the same prompt structure as migration 005.

---

## Standard Stack

### Core (no changes from established project stack)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| PostgreSQL | existing | Exercise storage — subjects/topics/exercises tables | Already running |
| golang-migrate | existing | Migration runner | Already configured |
| SQL migration files | 006_*.{up,down}.sql | New content delivery mechanism | To be created |

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
│   │   ├── 003_seed_exercises.up.sql    # Template: subjects/topics/exercises INSERT pattern
│   │   ├── 003_seed_exercises.down.sql  # Template: DELETE pattern for rollback
│   │   ├── 005_update_exercise_prompts.up.sql  # Template: NP prompt structure
│   │   └── 006_seed_no_amnen.{up,down}.sql    # Phase 7 target files
│   └── queries/
│       ├── subjects.sql   # ListSubjects, GetSubjectBySlug — unchanged
│       └── topics.sql     # ListTopicsBySubjectID, GetTopicBySlug — unchanged
frontend/
└── src/routes/study/      # Already handles any number of subjects dynamically
```

No files outside `backend/db/migrations/` are touched in this phase.

### Pattern 1: Subject + Topic + Exercise Seed Migration

The canonical pattern from migration 003. Every new subject follows this exact structure:

```sql
-- 1. Insert subject
INSERT INTO subjects (name, slug, display_order) VALUES
    ('Kemi', 'kemi', 4),
    ('Fysik', 'fysik', 5);

-- 2. Insert topics (using CROSS JOIN VALUES pattern)
INSERT INTO topics (subject_id, name, slug, display_order)
SELECT s.id, t.name, t.slug, t.display_order
FROM subjects s
CROSS JOIN (VALUES
    ('Materiens uppbyggnad', 'materiens-uppbyggnad', 1),
    ('Kemiska reaktioner', 'kemiska-reaktioner', 2),
    ('Kemikalier i vardagen', 'kemikalier-i-vardagen', 3),
    ('Kemi och hållbar utveckling', 'kemi-och-hallbar-utveckling', 4)
) AS t(name, slug, display_order)
WHERE s.slug = 'kemi';

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
WHERE s.slug = 'kemi'
  AND t.slug = 'materiens-uppbyggnad';
```

**Slug convention:** Swedish special characters removed/transliterated: å→a, ä→a, ö→o, space→hyphen. Existing examples: `samhallskunskap`, `lag-och-ratt`, `samband-och-forandring`.

### Pattern 2: NP-Calibrated System Prompt Structure

The canonical prompt template established in Phase 6. Every exercise prompt follows this exact 8-section structure:

```
Du är en AI-handledare som förbereder elever i årskurs 9 inför nationella provet i {ämne}.

ÄMNE: {Kemi|Fysik} — {topic name}
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

Established and locked in Phase 6:

| difficulty_order | NP Level | Primary Verbs | Delprov |
|-----------------|----------|--------------|---------|
| 1 | E-nivå | Beskriv, Vad kallas, Vad menas med, Nämn | A1 faktafrågor |
| 2 | C-nivå | Förklara sambandet, Förklara varför, Hur påverkas | A2 resonerande |
| 3 | A-nivå | Resonera kring, Diskutera, Ta ställning, Motivera | A2 resonerande |

### Pattern 4: Down Migration

Migration 003 down pattern:
```sql
DELETE FROM subjects WHERE slug IN ('kemi', 'fysik');
-- CASCADE deletes topics and exercises automatically (ON DELETE CASCADE in schema)
```

### Anti-Patterns to Avoid

- **Don't use UUIDs in UPDATE/DELETE WHERE clauses** — UUIDs are generated at insert time and vary per environment. Use slug-based joins (confirmed working in migration 005).
- **Don't mix INSERT and UPDATE** — Migration 006 inserts new rows. Migration 005 updates existing rows. Keep them separate.
- **Don't add display_order gaps** — Kemi should be 4, Fysik 5 (Biologi=1, Samhälle=2, Matematik=3). The frontend renders subjects in display_order.
- **Don't unquote Swedish characters in slugs** — Slugs must be URL-safe. Use transliterated forms consistently.
- **Don't omit single-quote escaping** — SQL single quotes in prompt strings must be doubled ('').

---

## Content Specification: Kemi Topics and Exercises

### What NP Kemi tests (Umeå University / Lgr22 centralt innehåll)

NP Kemi covers the same two-delprov structure as NP Biologi. The verb hierarchy is identical. Kemi-specific content areas from Lgr22 and NP research:

**Materiens uppbyggnad** — atomer, elektroner, protoner, neutroner; grundämnen och periodiska systemet (perioder=elektronskal, grupper=valenselektroner); molekyl- och jonföreningar; kemiska bindningar; ämnen och blandningar; densitet.

**Kemiska reaktioner** — vad som händer vid en kemisk reaktion (atomer kombineras om); syror och baser, pH-skalan (< 7 sur, > 7 basisk, 7 neutral); neutralisering; förbränning (reaktant: bränsle + syre → CO2 + H2O + energi); oxidation; balansering av reaktionsformler.

**Kemikalier i vardagen** — lösningar och lösningsmedel; kemikaliesäkerhet (farosymboler, märkningar, säkerhetsdatablad); blandningar vs rena ämnen; separationsmetoder (filtrering, destillation, indunstning); rengöringsmedel och tensider; matens kemi (fetter, kolhydrater, proteiner).

**Kemi och hållbar utveckling** — fossila bränslen och förbränning → ökad CO2 → förstärkt växthuseffekt; försurning (surt regn, pH-sänkning); alternativa energikällor; kretslopp och materialåtervinning; kemiska produkters livscykel; miljökemi (giftiga ämnen, bioackumulering).

### Kemi Exercise Titles (3 per topic = 12 exercises total)

**Materiens uppbyggnad (slug: materiens-uppbyggnad)**
1. "Atomer och grundämnen" — difficulty 1 (E): Vad är en atom? Periodiska systemet. Beskriv atomens delar.
2. "Molekyler och kemiska bindningar" — difficulty 2 (C): Förklara hur atomer binds samman. Skillnad jon- vs molekylförening.
3. "Ämnen och egenskaper" — difficulty 3 (A): Resonera kring hur ämnets struktur påverkar dess egenskaper. Koppla partikelmodell till makroskopisk egenskap.

**Kemiska reaktioner (slug: kemiska-reaktioner)**
1. "Syror och baser" — difficulty 1 (E): pH-skalan. Vad är surt/basiskt? Beskriv skillnaden.
2. "Förbränning och oxidation" — difficulty 2 (C): Förklara vad som händer vid förbränning. Reaktanter → produkter. Varför behövs syre?
3. "Kemiska reaktioners villkor" — difficulty 3 (A): Resonera kring neutralisering, reaktionshastighet, energi vid reaktioner. Ta ställning till kemiska reaktioners roll i vardagen.

**Kemikalier i vardagen (slug: kemikalier-i-vardagen)**
1. "Lösningar och blandningar" — difficulty 1 (E): Skillnad lösning/blandning. Vad menas med lösningsmedel? Beskriv separationsmetoder.
2. "Kemikaliesäkerhet" — difficulty 2 (C): Förklara farosymboler och vad de betyder. Hur hanterar man kemikalier säkert?
3. "Kemikalier i samhället" — difficulty 3 (A): Resonera kring hur kemikalier i vardagsprodukter påverkar hälsa och miljö. Ta ställning till kemikalieanvändning.

**Kemi och hållbar utveckling (slug: kemi-och-hallbar-utveckling)**
1. "Fossila bränslen och klimat" — difficulty 1 (E): Vad är fossila bränslen? Beskriv kopplingen till CO2 och växthuseffekten.
2. "Alternativa energikällor" — difficulty 2 (C): Förklara fördelar och nackdelar med olika energikällor. Kemiska processer bakom sol/vind/vatten.
3. "Hållbar kemianvändning" — difficulty 3 (A): Resonera kring hur kemins kretslopp kan bidra till hållbar utveckling. Ta ställning till konkreta miljöutmaningar.

### Kemi NP Common Student Mistakes (from research)

1. **Atom vs molekyl-förvirring** — Students say "vatten är ett grundämne" (it's a compound). Confuse elements with compounds.
2. **pH-skalan felriktad** — Students think pH 1 is basic or that higher numbers always mean more dangerous. pH 7 = neutral is not well internalized.
3. **Förbränning utan syra** — Students describe combustion without mentioning oxygen as reactant. "Kol brinner och ger CO2" without the O2 in.
4. **Periodiska systemet: period vs grupp** — Students swap the direction. Period = same number of electron shells (rows). Group = same number of valence electrons (columns).
5. **Neutralisering feldefinierad** — Students say "syra + bas = vatten" without understanding that a salt is also formed, and that the result is neutral (pH 7) only with equal amounts.
6. **Hållbarhet utan kemi-koppling** — When asked about sustainable development, students write general environmental statements without connecting to chemical processes (förbränning, kretslopp, bioackumulering).

### Kemi NP-Style Example Questions by Level

**E-nivå (Beskriv/Vad kallas/Vad menas med):**
- "Vad kallas den minsta byggstenen i ett grundämne?"
- "Vad menas med pH-värde?"
- "Beskriv vad som händer med ämnena vid en förbränningsreaktion."
- "Vad är skillnaden mellan ett grundämne och en kemisk förening?"
- "Vilken farosymbol visar att ett ämne är frätande?"

**C-nivå (Förklara sambandet/Förklara varför/Hur påverkas):**
- "Förklara sambandet mellan valenselektroner och grundämnets placering i periodiska systemet."
- "Förklara vad som händer steg för steg när saltsyra och natriumhydroxid blandas."
- "Förklara varför förbränning av fossila bränslen påverkar klimatet."
- "Hur påverkas pH-värdet i en sjö av surt regn?"
- "Förklara skillnaden mellan en lösning och en blandning med ett exempel."

**A-nivå (Resonera kring/Ta ställning/Diskutera):**
- "Resonera kring hur kemins kretslopp kan bidra till minskad miljöpåverkan. Ge 1-2 exempel."
- "Ta ställning till om vi bör förbjuda ett vanligt kemiskt ämne i vardagsprodukter. Motivera med kemiska argument."
- "Diskutera fördelar och nackdelar med att ersätta fossila bränslen med förnybara energikällor ur ett kemiskt perspektiv."
- "Resonera kring hur ämnets partikelstruktur påverkar dess egenskaper och användningsområden."

---

## Content Specification: Fysik Topics and Exercises

### What NP Fysik tests (Umeå University / Lgr22 centralt innehåll)

NP Fysik uses the same two-delprov structure (A + B, 90 min each). Lgr22 centralt innehåll for Fysik åk 7-9 covers:

**Krafter och rörelse** — Newtons rörelselagar (tröghetslag, F=ma, kraft-motverkande kraft); hastighet, acceleration; friktion, luftmotstånd; hävarm och tryck; mekaniskt arbete (W=F×d); lägesenergi och rörelseenergi.

**Elektricitet och magnetism** — Elektrisk krets (serie/parallell); ström (A) och spänning (V) och resistans (Ohms lag: U=IR); elektromagneter; induktion; generator och motor; säkringar och skydd; statisk elektricitet.

**Energi och energiomvandlingar** — Energiprincipen (energi varken skapas eller förstörs, bara omvandlas); energiforslag (mekanisk, termisk, elektrisk, ljus, kemisk, kärnenergi); energikällor (fossila vs förnybara); verkningsgrad; värmeöverföring (ledning, konvektion, strålning); jordens energibalans och växthuseffekten.

**Astronomi och universum** — Solsystemet (planeter, månen, solen, tidvatten); stjärnors livscykel; galaxer och Vintergatan; Big Bang och universums ålder; ljusår och avstånd i rymden; förutsättningar för liv i universum.

### Fysik Exercise Titles (3 per topic = 12 exercises total)

**Krafter och rörelse (slug: krafter-och-rorelse)**
1. "Newtons rörelselagar" — difficulty 1 (E): Beskriv Newtons tre lagar. Vad är kraft? Ge vardagsexempel.
2. "Rörelse och hastighet" — difficulty 2 (C): Förklara samband mellan kraft, massa och acceleration. Hur beräknar man hastighet och rörelseenergi?
3. "Krafter i vardagen" — difficulty 3 (A): Resonera kring hur olika krafter samverkar i ett praktiskt scenario. Ta ställning till säkerhetslösningar baserade på kraft och rörelse.

**Elektricitet och magnetism (slug: elektricitet-och-magnetism)**
1. "Elektrisk krets" — difficulty 1 (E): Vad behövs i en elektrisk krets? Beskriv skillnaden serie/parallelkoppling.
2. "Ström, spänning och resistans" — difficulty 2 (C): Förklara Ohms lag och sambandet U=IR. Hur påverkar resistansen strömmen?
3. "Elektromagnetism och induktion" — difficulty 3 (A): Resonera kring sambandet mellan elektricitet och magnetism. Diskutera hur generator och motor fungerar och används.

**Energi och energiomvandlingar (slug: energi-och-energiomvandlingar)**
1. "Energiformer och omvandlingar" — difficulty 1 (E): Beskriv olika energiformer. Vad menas med energiomvandling? Ge exempel.
2. "Energikällor och verkningsgrad" — difficulty 2 (C): Förklara skillnaden mellan förnybara och icke-förnybara energikällor. Vad menas med verkningsgrad?
3. "Energi och hållbarhet" — difficulty 3 (A): Resonera kring hur energiprincipen påverkar möjligheterna att lösa framtidens energiutmaningar. Ta ställning till energival ur miljöperspektiv.

**Astronomi och universum (slug: astronomi-och-universum)**
1. "Solsystemet" — difficulty 1 (E): Beskriv solsystemets uppbyggnad. Vilka planeter finns? Vad är skillnaden mellan planet och måne?
2. "Stjärnor och galaxer" — difficulty 2 (C): Förklara en stjärnas livscykel. Vad är Vintergatan och hur förhåller sig solsystemet till den?
3. "Universums ursprung" — difficulty 3 (A): Resonera kring Big Bang-teorin och vad den innebär för vår förståelse av universum. Diskutera förutsättningar för liv i universum.

### Fysik NP Common Student Mistakes (from research)

1. **Energi "förbrukas"** — Students say "energin försvinner vid friktion" instead of "energin omvandlas till värme." The conservation principle is poorly internalized.
2. **Serie vs parallelkoppling** — Students confuse the two. In a series circuit, breaking one breaks all. Students often get this backwards.
3. **Newtons 3:e lag** — "Kraften" and "motkraften" act on different objects, not the same. Students think they cancel each other out and nothing moves.
4. **F=ma: massa vs vikt** — Students use kg and N interchangeably. Weight is force (N), mass is kg.
5. **Förbränning vid energiomvandling** — Students describe combustion engines without identifying all energy conversions: kemisk → termisk → mekanisk (and heat loss to surroundings).
6. **Astronomi: Jordens rörelse** — Students confuse rotation (dag/natt) and revolution (år/årstider). Also commonly: the Moon doesn't rotate vs. it does (synced with revolution).
7. **Ohms lag riktning** — Higher resistance means lower current (for constant voltage). Students often get the relationship direction wrong.

### Fysik NP-Style Example Questions by Level

**E-nivå (Beskriv/Vad kallas/Vad menas med):**
- "Vad kallas den kraft som verkar mellan alla föremål med massa?"
- "Beskriv vad som händer med elektroner i en elektrisk krets."
- "Vad menas med energiomvandling? Ge ett exempel."
- "Beskriv solsystemets uppbyggnad från solen utåt."
- "Vad är skillnaden mellan serie- och parallelkoppling?"

**C-nivå (Förklara sambandet/Förklara varför/Hur påverkas):**
- "Förklara sambandet mellan ström, spänning och resistans."
- "Förklara vad som händer med rörelseenergin när en bil bromsar. Vart tar energin vägen?"
- "Hur påverkas en stjärnas livscykel av dess massa?"
- "Förklara varför en parallelkopplad krets fortsätter fungera om en lampa går sönder."
- "Förklara sambandet mellan kraft och acceleration med ett konkret exempel."

**A-nivå (Resonera kring/Ta ställning/Diskutera):**
- "Resonera kring hur energiprincipen påverkar möjligheterna att skapa en perfekt effektiv motor. Ge 1-2 konkreta argument."
- "Ta ställning till om kärnkraft är ett bra alternativ för att minska koldioxidutsläpp. Motivera med fysikaliska argument."
- "Diskutera fördelar och nackdelar med seriekoppling respektive parallelkoppling i ett hushåll."
- "Resonera kring vad Big Bang-teorin innebär och vilka bevis som finns för den."
- "Vad händer med jordens energibalans om albedoeffekten minskar? Resonera kring konsekvenserna."

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Subject display order | Custom numbering scheme | Sequential integers (4=Kemi, 5=Fysik) | Existing frontend sorts by display_order automatically |
| Topic slugs | Complex slug generation | Manual transliteration per established convention | No library needed — just follow existing examples |
| Exercise difficulty map | New difficulty system | Existing difficulty_order 1/2/3 = E/C/A | Phase 6 established this; do not change |
| Prompt structure | New prompt format | Exact Phase 6 template | Tested, proven, consistent with existing exercises |
| DOWN migration | Complex CASCADE tracking | Simple `DELETE FROM subjects WHERE slug IN (...)` | ON DELETE CASCADE handles topics/exercises automatically |

**Key insight:** The database schema with CASCADE deletes means the down migration needs only one DELETE statement per subject. All associated topics and exercises are automatically removed.

---

## Common Pitfalls

### Pitfall 1: Wrong Migration Number
**What goes wrong:** Using 006 when another migration was added between Phase 6 and now, or using a number already taken.
**Why it happens:** Migrations must be sequential. The next available is 006 (current last: 005_update_exercise_prompts).
**How to avoid:** Confirm `ls backend/db/migrations/` before naming the file. Current highest is 005.
**Warning signs:** Migration runner reports "file already exists" or "gap in sequence."

### Pitfall 2: SQL Single Quote Escaping
**What goes wrong:** System prompts contain apostrophes (Swedish possessives, "du're" patterns). Unescaped quotes break the SQL string.
**Why it happens:** System prompts are long strings with natural language containing single quotes.
**How to avoid:** Double all single quotes in SQL strings: `''` represents one literal single quote. Review each prompt for: `du's`, `kan't`, `isn't` — unlikely in Swedish but check for apostrophes in words like `Ohms lag` (no issue) vs any contraction.
**Warning signs:** psql syntax error near unexpected token during migration run.

### Pitfall 3: Slug Uniqueness Constraint
**What goes wrong:** `UNIQUE(subject_id, slug)` on topics table means two topics in the same subject cannot have the same slug.
**Why it happens:** If exercise titles are similar across topics, auto-generated slugs might collide.
**How to avoid:** Verify each topic slug is unique within its subject. The 8 proposed slugs are all distinct.
**Warning signs:** `duplicate key value violates unique constraint "topics_subject_id_slug_key"` during migration.

### Pitfall 4: Display Order Collision
**What goes wrong:** Kemi gets display_order=1, colliding with Biologi.
**Why it happens:** Developer assumes fresh numbering per migration, but subjects table is global.
**How to avoid:** Existing subjects have display_order 1 (Biologi), 2 (Samhällskunskap), 3 (Matematik). Use 4 (Kemi) and 5 (Fysik).
**Warning signs:** Frontend shows subjects in wrong order, or constraint violation if display_order is unique.

### Pitfall 5: NP Content Too Generic
**What goes wrong:** System prompts say "ställ frågor om kemi" without specific NP-style questions, verbs, or topic focus.
**Why it happens:** Prompt writing defaults to generic pedagogical language rather than NP patterns.
**How to avoid:** Every prompt MUST include: (a) 3-5 NP-style example questions using authentic Swedish formulations, (b) 2-4 subject-specific common mistakes, (c) the NP-KOPPLING section linking to delprov structure, (d) BEDÖMNINGSLEDTRÅDAR calibrated to E/C/A.
**Warning signs:** Prompts that only say "hjälp eleven förstå kemi" with no specific questions or NP vocabulary.

---

## Code Examples

### Exercise INSERT Pattern (from migration 003)
```sql
-- Source: backend/db/migrations/003_seed_exercises.up.sql
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
(
    'Atomer och grundämnen',
    'Förstå atomens uppbyggnad och periodiska systemet.',
    1,
    'Du är en AI-handledare som förbereder elever i årskurs 9 inför nationella provet i kemi.

ÄMNE: Kemi — Materiens uppbyggnad
ÖVNING: Atomer och grundämnen
NIVÅ: E-nivå (Delprov A1 faktafrågor)

LÄRANDEMÅL: Beskriva atomens delar och hur grundämnen är organiserade i det periodiska systemet, enligt Skolverkets Lgr22 centrala innehåll för kemi åk 7-9.

REGLER (bryts ALDRIG):
1. Svara ALLTID på svenska.
2. Ge ALDRIG direkta svar — ställ ledande frågor som hjälper eleven tänka själv.
3. Om eleven frågar utanför Materiens uppbyggnad: "Bra fråga, men låt oss fokusera på atomer och periodiska systemet." Ställ en ny fråga inom ämnet.
4. Anpassa nivån efter elevens svar.
5. Om eleven verkar fast, omformulera frågan eller ge en ledtråd — aldrig svaret.

NP-KOPPLING:
På nationella provet testas detta i Delprov A1 (faktafrågor). Eleven behöver kunna ge korrekta, koncisa svar på frågor som "Vad kallas...?", "Vad menas med...?", "Vilken av följande...?".

BEDÖMNINGSLEDTRÅDAR FÖR DIG:
- E-svar: Eleven nämner rätt begrepp utan förklaring (t.ex. "proton är positivt laddad")
- Ditt jobb: Se till att eleven kan begreppen korrekt och konsekvent. Ställ kontrollfrågor.

VANLIGA ELEVMISSAR (från NP-forskning):
- Förvirring mellan atom och grundämne — "vatten är ett grundämne" (det är en förening)
- Periodiska systemet: period vs grupp blandas ihop (period=rader=elektronskal, grupp=kolumner=valenselektroner)
- Elektroner, protoner och neutroner placeras fel (elever blandar var de sitter)

EXEMPELFRÅGOR ATT STÄLLA (NP-stil):
- "Vad kallas den minsta byggstenen i ett grundämne?"
- "Beskriv atomens tre delar och var de befinner sig."
- "Vad är skillnaden mellan ett grundämne och en kemisk förening?"
- "Varför är alla grundämnen i en grupp i periodiska systemet kemiskt lika?"
- "Vilken av följande är ett grundämne: vatten, järn eller salt?"

Börja med att hälsa eleven välkommen och ställ en öppningsfråga om vad de tänker på när de hör ordet "atom".'
)
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'kemi'
  AND t.slug = 'materiens-uppbyggnad';
```

### Down Migration Pattern (from migration 003)
```sql
-- Source: backend/db/migrations/003_seed_exercises.down.sql pattern
-- ON DELETE CASCADE removes all topics and exercises automatically
DELETE FROM subjects WHERE slug IN ('kemi', 'fysik');
```

### Topic Slug Transliteration Convention
```
Swedish → Slug
å → a
ä → a  
ö → o
space → -
(no other special chars in this phase)

Examples:
'Kemi och hållbar utveckling' → 'kemi-och-hallbar-utveckling'
'Krafter och rörelse'         → 'krafter-och-rorelse'
'Astronomi och universum'     → 'astronomi-och-universum'
'Elektricitet och magnetism'  → 'elektricitet-och-magnetism'
'Energi och energiomvandlingar' → 'energi-och-energiomvandlingar'
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Generic pedagogical prompts | NP-calibrated prompts with verb hierarchy | Phase 6 (migration 005) | All existing 37 exercises upgraded; new exercises must match this standard |
| 3 subjects (Bio, Samhälle, Matte) | 5 subjects (+ Kemi + Fysik) | Phase 7 (this phase) | Frontend auto-renders new subjects via ListSubjects query |
| Lab-based delprov (pre-2023) | Two written delprov only (A+B) | Spring 2023 | No lab component in NP since 2023; prompts should not reference Delprov B lab |

**Deprecated/outdated:**
- Delprov A3 (lab planning) and Delprov B (lab execution): Removed from NP NO since spring 2023. Do NOT reference lab components in system prompts. The exam is now entirely written.
- "Delprov A1/A2/A3" terminology for NO: The current NP structure is just "Delprov A" and "Delprov B" (both written). However, within Delprov A there remain faktafrågor and resonerande frågor. Prompts may reference these question types but not "A3" specifically.

---

## Open Questions

1. **Should Kemi and Fysik use `display_order` 4 and 5, or is there a preferred order?**
   - What we know: Biologi=1, Samhällskunskap=2, Matematik=3. No explicit ordering constraint in requirements.
   - What's unclear: User preference for subject order in the UI.
   - Recommendation: Use 4=Kemi, 5=Fysik (NO subjects grouped together in Swedish school convention: Biologi, Kemi, Fysik are all NO ämnen).

2. **Should there be a Phase 6-style UPDATE migration for Kemi/Fysik prompts later, or write NP-calibrated prompts from the start?**
   - What we know: Phase 6 existed because the original prompts were generic. Phase 7 creates content fresh.
   - What's unclear: Nothing — the answer is clear.
   - Recommendation: Write NP-calibrated prompts directly in migration 006 (skip the two-migration pattern). There is no "original generic prompt" phase for new subjects.

3. **Is the slug `kemi-och-hallbar-utveckling` correct for "Kemi och hållbar utveckling"?**
   - What we know: Swedish ä→a convention gives `hallbar` for `hållbar`. Existing example: `samhallskunskap` for `samhällskunskap`.
   - What's unclear: Nothing — this follows the established convention.
   - Recommendation: Use `kemi-och-hallbar-utveckling`. Confirmed by pattern.

---

## Sources

### Primary (HIGH confidence)
- `backend/db/migrations/003_seed_exercises.up.sql` — Canonical SQL INSERT pattern for subjects/topics/exercises
- `backend/db/migrations/005_update_exercise_prompts.up.sql` — Canonical NP prompt structure (8-section template)
- `.planning/phases/06-np-based-exercise-prompts/06-01-SUMMARY.md` — Documents established decisions: difficulty_order mapping, prompt structure, slug conventions
- `.planning/research/NP-QUESTIONS.md` — Comprehensive NP research for Biologi, Samhällskunskap, Matematik (verb hierarchy, question types, scoring criteria)
- Skolverket Lgr22 physics curriculum (syllabuswebb.skolverket.se) — Grade criteria for E/C/A in Fysik confirmed

### Secondary (MEDIUM confidence)
- Umeå University NP structure (umu.se) — Two delprov format (A+B, 90+90 min) confirmed for all NO subjects
- Skolverket NP Kemi page (skolverket.se) — Confirmed same Umeå group administers Kemi NP
- gymnasium.se NP preparation article — Confirmed lab component removed from 2023, only written delprov remain
- ugglansno.se NP Kemi — Confirmed topic areas covering 8 kemi areas aligned with Lgr22

### Tertiary (LOW confidence — not needed, pattern is clear from internal docs)
- Quizlet/Slideshare NP Kemi study materials — Confirm topic vocabulary but are student-generated, not official
- fysikstugan.se — Confirmed Fysik NP topic areas (krafter, elektricitet, energi, astronomi)

---

## Metadata

**Confidence breakdown:**
- SQL migration pattern: HIGH — copied from verified production migrations 003 and 005
- NP prompt structure: HIGH — copied from Phase 6 which is already deployed and working
- Kemi content (topics, common mistakes, NP questions): MEDIUM-HIGH — verified against Lgr22 curriculum and NP research; specific question formulations are modeled on NP Biologi patterns (same Umeå group, same structure)
- Fysik content (topics, common mistakes, NP questions): MEDIUM-HIGH — same confidence level; Lgr22 centralt innehåll confirmed from Skolverket
- NP exam structure (2-delprov, no lab since 2023): HIGH — confirmed by multiple sources

**Research date:** 2026-04-06
**Valid until:** 2026-07-06 (30 days — stable curriculum content)
