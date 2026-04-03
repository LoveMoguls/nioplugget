# Phase 6: NP-Based Exercise Prompts - Context

**Gathered:** 2026-04-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Rewrite all ~36 exercise system prompts to match real nationella prov (NP) patterns in level, tonality, and question types. Use research from .planning/research/NP-QUESTIONS.md as the foundation. This is content work, not code architecture.

</domain>

<decisions>
## Implementation Decisions

### Prompt structure per difficulty level
- Difficulty 1 (E-level): AI uses "Beskriv" verb — asks factual recall questions
- Difficulty 2 (C-level): AI uses "Förklara" verb — asks for explanations and connections
- Difficulty 3 (A-level): AI uses "Resonera/Diskutera/Ta ställning" — asks for reasoning with multiple perspectives

### Tonality
- Match NP formulation style: formal but accessible Swedish
- Use exact NP phrasing patterns: "Resonera kring...", "Förklara sambandet mellan...", "Ge exempel som visar..."
- Questions should feel like real NP tasks, not textbook exercises

### Per-subject patterns (from NP research)
- **Biologi:** Ecological reasoning chains, body system cause-effect, genetics probability + ethics, cell process comparisons
- **Samhällskunskap:** Source analysis, democratic process reasoning, rights vs responsibilities, economic cause-effect chains
- **Matematik:** Multi-step problem solving with explanation requirement, "visa hur du tänker" emphasis, real-world context problems

### Scoring structure awareness
- System prompts should guide students toward the NP scoring pattern: position → grund → exempel → (A-level) motargument
- Math prompts should emphasize showing working ("poäng för förtjänster, inte avdrag för fel")

### Implementation approach
- Update the existing seed migration (or add a new migration) with rewritten system prompts
- Each prompt includes: subject context, difficulty-appropriate verbs, example NP-style questions to ask, common student mistakes to address
- Keep the Socratic "never give direct answers" constraint from Phase 2

### Claude's Discretion
- Exact prompt wording details beyond the patterns above
- How many example questions to include per prompt
- Balance between curriculum coverage and NP-specific focus

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `.planning/research/NP-QUESTIONS.md` — comprehensive NP question pattern research
- `backend/db/migrations/002_seed_exercises.up.sql` — current exercise seed data with existing prompts
- `backend/db/queries/exercises.sql` — sqlc queries for exercises

### Established Patterns
- Exercises stored in PostgreSQL with system_prompt text column
- Seeded via golang-migrate migration files
- 12 topic areas across 3 subjects, 3-5 exercises each

### Integration Points
- New migration file (e.g., 005_update_exercise_prompts.up.sql) to UPDATE existing exercise system_prompt values
- No frontend or backend code changes needed — just database content

</code_context>

<specifics>
## Specific Ideas

- Use real NP question formulations as templates in the system prompts
- Each prompt should include 3-5 example questions the AI should ask, modeled on actual NP tasks
- Math prompts must emphasize "visa hur du tänker" and reward partial understanding
- Samhälle prompts must include source-critical thinking patterns
- Bio prompts must include ecological reasoning chains

</specifics>

<deferred>
## Deferred Ideas

None

</deferred>

---

*Phase: 06-np-based-exercise-prompts*
*Context gathered: 2026-04-03*
