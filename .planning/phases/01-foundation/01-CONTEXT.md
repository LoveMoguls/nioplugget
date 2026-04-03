# Phase 1: Foundation - Context

**Gathered:** 2026-04-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Parents can register, store an encrypted Claude API key, and onboard their child via invite link. Child can log in with name + PIN. Covers AUTH-01..08, KEY-01..04, SEC-01..03.

</domain>

<decisions>
## Implementation Decisions

### Registration flow
- Immediate access after signup — no email verification in v1
- Registration form: email + password only (no parent name)
- Password: minimum 8 characters, no complexity rules
- Argon2id for password hashing

### Post-registration experience
- Claude's Discretion: Dashboard with setup guide vs onboarding wizard — pick best UX approach
- Must guide parent through: 1) Add API key 2) Create child profile

### API key setup
- Part of first-time setup, immediately after registration
- Live validation: test API call to Claude when key is entered
- Display after saving: masked format (sk-ant-...****)
- Include help text with link + kort guide for getting an API key from console.anthropic.com
- AES-256-GCM encryption at rest, master key in environment variable

### Child onboarding
- Invite links valid for 72 hours, single-use, atomically invalidated on activation
- Child activates: enters name + chooses 4-digit PIN
- Child login: select name from list + enter PIN (no username to remember)
- First visit after activation: kort välkomstskärm ("Välkommen [namn]! Välj ett ämne för att börja.") then redirect to subject selection

### Error responses
- Tone: saklig & tydlig — rak info utan tekniskt språk, inga "oops" eller emoji
- Wrong PIN: 5 attempts, then 15 min lockout. Show remaining attempts.
- Expired invite link: "Länken har gått ut. Be din förälder skapa en ny."
- Invalid/expired API key at chat start: "API-nyckeln fungerar inte. Be din förälder kontrollera den."
- Parent sees no proactive notification about key issues — error surfaces when child tries to chat

### Security
- Backend never logs API keys or Authorization headers
- Invite links are single-use and time-bound (72h)
- GDPR consent collected explicitly at parent registration
- PIN rate limiting enforced server-side

### Claude's Discretion
- Post-registration UX flow (wizard vs dashboard with guide)
- Exact form styling and layout
- JWT token expiry duration
- Session cookie configuration
- GDPR consent copy text

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- None — greenfield project

### Established Patterns
- Go (Chi) + SvelteKit + PostgreSQL stack decided
- shadcn-svelte for UI components
- Argon2id for password hashing (from stack research)
- chi-jwtauth for JWT middleware (from stack research)
- pgx for PostgreSQL driver (from stack research)
- golang-migrate for migrations (from stack research)
- sqlc for type-safe queries (from stack research)

### Integration Points
- Database schema is the foundation all other phases build on
- Auth middleware gates all student and parent routes
- EncryptService (AES-256-GCM) is required before Phase 2 can proxy Claude API calls

</code_context>

<specifics>
## Specific Ideas

- Child login should feel effortless — select name from a list, type 4 digits, done
- Error messages should be direct and instructive, never cute or technical
- API key guide should be simple enough for a non-technical parent

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-foundation*
*Context gathered: 2026-04-03*
