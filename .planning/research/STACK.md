# Stack Research

**Domain:** AI-driven educational chat app (Swedish national tests, åk 9)
**Researched:** 2026-04-03
**Confidence:** HIGH (all versions verified against pkg.go.dev and official docs)

---

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.24+ | Backend API server | Required by pgx v5.9.1 and anthropic-sdk-go; Go 1.22+ stdlib mux improvements make chi even leaner |
| go-chi/chi | v5.2.5 | HTTP router + middleware | Zero external dependencies, 100% net/http compatible, 17 163 importers — battle-tested in production. Matches existing trade-analyst experience |
| SvelteKit | latest (Svelte 5) | Frontend framework | Svelte 5 runes model is stable and production-ready; server-side rendering + fine-grained reactivity ideal for chat UI on mobile |
| shadcn-svelte | latest | UI component library | Tailwind v4 + Svelte 5 support landed, copies components into codebase (no runtime dep), matches trade-analyst experience |
| PostgreSQL | 16+ | Primary database | Required by pgx driver; mature, reliable, excellent full-text search for session history |
| anthropic-sdk-go | v1.29.0 | Claude API client | Official Anthropic SDK; native streaming support via SSE, typed message structs, Go 1.23+ required |

### Backend Libraries

| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| go-chi/jwtauth | v5.4.0 | JWT auth middleware for Chi | First-party Chi integration; wraps lestrrat-go/jwx; handles token extraction + validation in one middleware |
| golang.org/x/crypto/argon2 | (x/crypto) | Password hashing | OWASP-recommended Argon2id over bcrypt in 2025; resists GPU/ASIC attacks; stdlib-adjacent |
| crypto/aes + crypto/cipher | stdlib | AES-256-GCM encryption for API keys | Standard library; AES-GCM provides authenticated encryption — integrity + confidentiality without extra deps |
| crypto/rand | stdlib | Cryptographically random nonces/IVs | Must use crypto/rand not math/rand for nonces; prevents reuse attacks |
| jackc/pgx | v5.9.1 | PostgreSQL driver + toolkit | lib/pq is maintenance-only; pgx v5 is the Go-community default; native PostgreSQL features (LISTEN/NOTIFY, COPY) |
| golang-migrate/migrate | v4.19.1 | Database migrations | SQL file pairs (up/down), CLI + embedded library, embeds with Go embed, 20+ DB support |
| sqlc | v1.30.0 | SQL → Go type-safe codegen | Eliminates hand-written scan boilerplate; pgx/v5 target supported; catches SQL errors at build time |

### Frontend Libraries

| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| bits-ui | latest | Headless UI primitives | Powers shadcn-svelte internals; Svelte 5 native |
| @lucide/svelte | latest | Icons | Standard icon set used by shadcn-svelte; tree-shakeable |
| tailwind-variants | latest | Component variant styling | Replaces cva for shadcn-svelte v2 |
| svelte-sonner | latest | Toast notifications | Svelte-native port of Sonner, used by shadcn-svelte |
| formsnap | latest | Form validation integration | Works with sveltekit-superforms or standalone; accessible form components |

### Testing

| Tool | Version | Purpose | Why |
|------|---------|---------|-----|
| stretchr/testify | v1.11.1 | Go test assertions + suites | Standard Go assertion library; require/assert packages; suite setup/teardown for DB tests |
| testcontainers-go | latest | Real Postgres in integration tests | Spins up postgres:16-alpine in Docker per test run; snapshot/restore for fast test isolation; avoids mocks diverging from real DB |
| vitest + @vitest/browser | latest | Frontend unit/component tests | SvelteKit-recommended; Svelte 5 jsdom struggles with runes — browser mode via Playwright is the 2025 solution |
| Playwright | latest | E2E + browser-mode test runner | Powers vitest browser mode; also for full E2E flows |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| sqlc CLI | Generate Go code from SQL | `go generate` integration; configure in `sqlc.yaml` with `sql_package: pgx/v5` |
| golang-migrate CLI | Run migrations from terminal | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| air | Live reload for Go backend | `cosmtrek/air` — standard Go dev server hot-reload |
| pnpm | Frontend package manager | Faster than npm; shadcn-svelte docs use pnpm dlx |

---

## Installation

```bash
# Backend Go modules
go get github.com/go-chi/chi/v5@v5.2.5
go get github.com/go-chi/jwtauth/v5@v5.4.0
go get github.com/anthropics/anthropic-sdk-go@v1.29.0
go get github.com/jackc/pgx/v5@v5.9.1
go get github.com/golang-migrate/migrate/v4@v4.19.1
go get golang.org/x/crypto

# Backend testing
go get github.com/stretchr/testify@v1.11.1
go get github.com/testcontainers/testcontainers-go/modules/postgres

# sqlc (code generator, install as CLI tool)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# golang-migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Frontend (in frontend/ directory)
pnpm dlx sv create . --add tailwindcss
pnpm dlx shadcn-svelte@latest init

# Frontend testing
pnpm add -D vitest @vitest/browser vitest-browser-svelte playwright
```

---

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Auth | go-chi/jwtauth + argon2 | session cookies + bcrypt | JWT stateless for parent/child dual-role model; argon2 strictly better than bcrypt vs modern attacks |
| DB driver | pgx v5 | database/sql + lib/pq | lib/pq explicitly maintenance-only; pgx v5 is community default; sqlc targets pgx natively |
| DB migrations | golang-migrate | Atlas, goose | golang-migrate is simplest SQL-file-per-migration; Atlas requires schema HCL; goose is fine but less documentation |
| SQL | sqlc | GORM, sqlx | GORM hides SQL and adds magic; sqlx has no compile-time safety; sqlc catches errors at build time and produces idiomatic Go |
| Streaming | SSE (net/http) | WebSockets | Claude API streams SSE natively; SSE is one-way server-push which is exactly what's needed; WebSockets add bidirectional complexity for no gain |
| Frontend testing | vitest browser mode | @testing-library/svelte + jsdom | jsdom cannot handle Svelte 5 runes reactivity reliably; browser mode is the official 2026 recommendation |
| Password hashing | argon2id (x/crypto) | bcrypt | OWASP 2025 recommends argon2id as gold standard; bcrypt is legacy-acceptable but weaker against GPU attacks |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| GORM | Hides SQL, reflection magic, hard to debug N+1 queries | sqlc for type-safe generated queries |
| lib/pq | Explicitly in maintenance mode since 2022; new features won't land | pgx v5 |
| math/rand for crypto | Not cryptographically secure; nonce reuse catastrophic for AES-GCM | crypto/rand (stdlib) |
| AES-CBC without HMAC | No authentication; malleable ciphertext; CBC padding oracle attacks | AES-256-GCM (authenticated encryption, stdlib) |
| WebSockets for chat | Bidirectional not needed; Claude API uses SSE; adds conn management complexity | SSE via net/http (flush + `text/event-stream`) |
| @testing-library/svelte | jsdom + Svelte 5 runes = flaky tests; community migrating away in 2025 | vitest browser mode + playwright |
| golang-jwt/jwt directly | Manual integration with Chi; jwtauth already wraps it with Chi middleware | go-chi/jwtauth v5 |
| Storing API keys as plaintext | Security vulnerability; Claude API keys grant billing access | AES-256-GCM encrypt at rest, key from env |

---

## Critical Implementation Notes

### AES-256-GCM for Claude API Key Storage

The parent's Claude API key must be encrypted before storing in Postgres:

```go
// Use crypto/aes + crypto/cipher (GCM mode) — stdlib only, no extra deps
// Key: 32-byte secret from environment (never hardcoded)
// Nonce: 12-byte random from crypto/rand per encryption
// Output: nonce || ciphertext || tag stored as bytea in Postgres
```

Use `ENCRYPTION_KEY` env var (32 bytes, base64-encoded). Rotate via re-encryption job if key changes.

### SSE Streaming Pattern for Chat

Claude API returns SSE events. The Go backend proxies these to the browser:

```
Browser <--SSE (text/event-stream)--> Go backend <--SSE--> Claude API
```

Chi handles this with standard `http.Flusher` interface — no library needed. Set `Cache-Control: no-cache`, `Content-Type: text/event-stream`, flush after each write.

### JWT for Dual-Role Auth (Parent + Child)

The parent/child model requires role claims in JWT:

```
parent: { role: "parent", user_id: "..." }
child:  { role: "child",  child_id: "...", parent_id: "..." }
```

jwtauth middleware verifies token; route-level middleware checks role claim. Children log in via invite link + PIN (not email/password), so no argon2 needed for child auth — PIN validated against hashed value in DB.

### sqlc Configuration

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "./db/queries"
    schema: "./db/migrations"
    gen:
      go:
        package: "db"
        out: "./internal/db"
        sql_package: "pgx/v5"
```

---

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| pgx v5.9.1 | Go 1.25+, PostgreSQL 14+ | Verified on pkg.go.dev, published Mar 2026 |
| anthropic-sdk-go v1.29.0 | Go 1.23+ | Requires Go 1.23 minimum; published Apr 2026 |
| go-chi/chi v5.2.5 | Go stdlib net/http | Published Feb 2026; 17k+ importers |
| go-chi/jwtauth v5.4.0 | chi v5 + lestrrat-go/jwx | Published Feb 2026 |
| golang-migrate v4.19.1 | Go 1.24, 1.25 | Published Nov 2025 |
| testify v1.11.1 | Go 1.19+ | Published Aug 2025 |
| shadcn-svelte | Svelte 5, Tailwind v4 | Tailwind v4 + Svelte 5 support confirmed in changelog |

---

## Sources

- [pkg.go.dev/github.com/anthropics/anthropic-sdk-go](https://pkg.go.dev/github.com/anthropics/anthropic-sdk-go) — v1.29.0, published Apr 1 2026, Go 1.23+ requirement (HIGH)
- [pkg.go.dev/github.com/go-chi/chi/v5](https://pkg.go.dev/github.com/go-chi/chi/v5) — v5.2.5, published Feb 5 2026 (HIGH)
- [pkg.go.dev/github.com/go-chi/jwtauth/v5](https://pkg.go.dev/github.com/go-chi/jwtauth/v5) — v5.4.0, published Feb 26 2026 (HIGH)
- [pkg.go.dev/github.com/jackc/pgx/v5](https://pkg.go.dev/github.com/jackc/pgx/v5) — v5.9.1, published Mar 22 2026 (HIGH)
- [pkg.go.dev/github.com/golang-migrate/migrate/v4](https://pkg.go.dev/github.com/golang-migrate/migrate/v4) — v4.19.1, published Nov 29 2025 (HIGH)
- [pkg.go.dev/github.com/stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify) — v1.11.1, published Aug 27 2025 (HIGH)
- [pkg.go.dev/crypto/aes](https://pkg.go.dev/crypto/aes) — stdlib, AES-256-GCM pattern (HIGH)
- [pkg.go.dev/golang.org/x/crypto/argon2](https://pkg.go.dev/golang.org/x/crypto/argon2) — Argon2id OWASP recommendation (HIGH)
- [shadcn-svelte.com/docs/installation/sveltekit](https://www.shadcn-svelte.com/docs/installation/sveltekit) — Svelte 5 + Tailwind v4 confirmed (HIGH)
- [platform.claude.com/docs/en/build-with-claude/streaming](https://platform.claude.com/docs/en/build-with-claude/streaming) — SSE streaming events (HIGH)
- [golang.testcontainers.org/modules/postgres/](https://golang.testcontainers.org/modules/postgres/) — Postgres module, snapshot support (MEDIUM)
- [docs.sqlc.dev/en/stable/guides/using-go-and-pgx.html](https://docs.sqlc.dev/en/stable/guides/using-go-and-pgx.html) — pgx/v5 target (HIGH)
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html) — Argon2id parameters (HIGH)

---

*Stack research for: Nioplugget — AI educational chat app (Go + SvelteKit + PostgreSQL + Claude API)*
*Researched: 2026-04-03*
