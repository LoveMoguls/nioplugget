# Phase 1: Foundation - Research

**Researched:** 2026-04-03
**Domain:** Authentication, API key encryption, child onboarding (Go + SvelteKit + PostgreSQL)
**Confidence:** HIGH

## Summary

Phase 1 builds the foundational auth system for Nioplugget: parent registration with Argon2id password hashing, JWT-based sessions, AES-256-GCM encrypted API key storage, and a child invite/PIN-login flow. The stack is Go (Chi) backend, SvelteKit frontend with shadcn-svelte, and PostgreSQL via pgx/sqlc.

All library choices are locked by user decisions (CONTEXT.md). The research focuses on correct usage patterns and integration pitfalls rather than library selection.

**Primary recommendation:** Build horizontally — database schema and crypto services first (they gate everything), then auth endpoints, then frontend flows.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Immediate access after signup — no email verification in v1
- Registration form: email + password only (no parent name)
- Password: minimum 8 characters, no complexity rules
- Argon2id for password hashing
- Live validation: test API call to Claude when key is entered
- Display after saving: masked format (sk-ant-...****)
- Include help text with link + kort guide for getting an API key from console.anthropic.com
- AES-256-GCM encryption at rest, master key in environment variable
- Invite links valid for 72 hours, single-use, atomically invalidated on activation
- Child activates: enters name + chooses 4-digit PIN
- Child login: select name from list + enter PIN (no username to remember)
- First visit after activation: kort valkomstskarm then redirect to subject selection
- Error tone: saklig & tydlig — rak info utan tekniskt sprak, inga "oops" eller emoji
- Wrong PIN: 5 attempts, then 15 min lockout. Show remaining attempts.
- Expired invite link: "Lanken har gatt ut. Be din foralder skapa en ny."
- Invalid/expired API key at chat start: "API-nyckeln fungerar inte. Be din foralder kontrollera den."
- Parent sees no proactive notification about key issues — error surfaces when child tries to chat
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

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| AUTH-01 | Parent registers with email + password | Argon2id hashing, pgx user insert, Chi POST handler |
| AUTH-02 | Parent login with persistent session | JWT generation via go-chi/jwtauth, httpOnly cookie |
| AUTH-03 | Parent logout | Cookie clearing, optional token denylist |
| AUTH-04 | Parent creates child profile with name | DB insert with parent_id FK |
| AUTH-05 | System generates single-use invite link | crypto/rand token, 72h expiry, DB storage |
| AUTH-06 | Child activates account via invite link | Atomic invite invalidation + student record update |
| AUTH-07 | Child login with name + PIN | PIN lookup by parent_id, Argon2id hashed PIN |
| AUTH-08 | PIN rate limiting | Server-side counter with 15-min lockout after 5 fails |
| KEY-01 | Parent enters Claude API key | POST endpoint, live Claude API validation call |
| KEY-02 | API key encrypted with AES-256-GCM | Go crypto/aes + cipher.NewGCM, master key from env |
| KEY-03 | Parent can update/delete API key | PUT/DELETE endpoints with re-encryption |
| KEY-04 | Clear error on invalid/expired key | Error message surfaced at chat start |
| SEC-01 | Backend never logs keys/auth headers | Custom logger middleware that redacts sensitive fields |
| SEC-02 | Invite links single-use + time-bound | Atomic DB update (activated_at + used check) |
| SEC-03 | GDPR consent at registration | Consent checkbox + timestamp in parents table |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-chi/chi/v5 | v5.x | HTTP router | Lightweight, stdlib-compatible, middleware-friendly |
| go-chi/jwtauth/v5 | v5.x | JWT middleware | Built for Chi, wraps lestrrat-go/jwx |
| jackc/pgx/v5 | v5.x | PostgreSQL driver | Pure Go, connection pooling via pgxpool, fastest Go PG driver |
| sqlc | v1.x | Type-safe SQL | Generates Go from SQL queries, zero runtime overhead |
| golang-migrate/migrate/v4 | v4.x | DB migrations | File-based, CLI + library, supports pgx |
| alexedwards/argon2id | latest | Password hashing | Wraps Go stdlib crypto, sensible defaults (OWASP compliant) |
| SvelteKit | 2.x | Frontend framework | SSR + SPA, file-based routing |
| shadcn-svelte | latest | UI components | Accessible, composable, Tailwind-based |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| crypto/aes + crypto/cipher | stdlib | AES-256-GCM encryption | API key encryption at rest |
| crypto/rand | stdlib | Secure random generation | Invite tokens, nonces |
| encoding/base64 | stdlib | Token encoding | URL-safe invite link tokens |
| net/http | stdlib | HTTP client | Claude API key validation call |
| rs/zerolog | v1.x | Structured logging | JSON logs, field-level redaction for SEC-01 |
| golang.org/x/time/rate | latest | Rate limiting | PIN brute-force protection (AUTH-08) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| alexedwards/argon2id | golang.org/x/crypto/argon2 | Raw stdlib — must tune params manually. alexedwards wraps it with OWASP defaults. |
| rs/zerolog | log/slog (stdlib) | slog is simpler but zerolog has better field redaction and structured output |
| go-chi/jwtauth | golang-jwt/jwt | jwtauth integrates with Chi middleware chain natively |

**Installation:**
```bash
# Backend
go get github.com/go-chi/chi/v5
go get github.com/go-chi/jwtauth/v5
go get github.com/jackc/pgx/v5
go get github.com/alexedwards/argon2id
go get github.com/rs/zerolog
go get golang.org/x/time/rate

# Frontend
npm create svelte@latest frontend
npx shadcn-svelte@latest init
```

## Architecture Patterns

### Recommended Project Structure
```
nioplugget/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # Entry point
│   ├── internal/
│   │   ├── auth/
│   │   │   ├── handler.go        # HTTP handlers (register, login, logout)
│   │   │   ├── jwt.go            # Token generation/validation
│   │   │   └── middleware.go     # Auth middleware (parent, child)
│   │   ├── child/
│   │   │   ├── handler.go        # Invite, activate, PIN login
│   │   │   └── ratelimit.go      # PIN rate limiting
│   │   ├── apikey/
│   │   │   ├── handler.go        # CRUD endpoints
│   │   │   ├── encrypt.go        # AES-256-GCM encrypt/decrypt
│   │   │   └── validate.go       # Claude API test call
│   │   ├── database/
│   │   │   ├── pool.go           # pgxpool setup
│   │   │   └── queries/          # sqlc generated code
│   │   └── middleware/
│   │       └── logger.go         # Redacting logger
│   ├── db/
│   │   ├── migrations/           # golang-migrate files
│   │   └── queries/              # sqlc SQL files
│   └── sqlc.yaml
├── frontend/
│   ├── src/
│   │   ├── routes/
│   │   │   ├── +page.svelte          # Landing/login
│   │   │   ├── (parent)/
│   │   │   │   ├── dashboard/
│   │   │   │   │   └── +page.svelte  # Parent dashboard
│   │   │   │   └── setup/
│   │   │   │       └── +page.svelte  # API key setup
│   │   │   ├── (child)/
│   │   │   │   └── login/
│   │   │   │       └── +page.svelte  # Child PIN login
│   │   │   └── invite/
│   │   │       └── [token]/
│   │   │           └── +page.svelte  # Invite activation
│   │   └── lib/
│   │       ├── api.ts                # Backend API client
│   │       └── stores/
│   │           └── auth.ts           # Auth state store
│   └── svelte.config.js
└── db/
    └── migrations/
```

### Pattern 1: Chi Middleware Chain
**What:** Layered middleware for auth, logging, and rate limiting
**When to use:** Every route group that needs protection

```go
r := chi.NewRouter()
r.Use(middleware.Logger)        // Custom redacting logger
r.Use(middleware.Recoverer)

// Public routes
r.Group(func(r chi.Router) {
    r.Post("/api/auth/register", auth.Register)
    r.Post("/api/auth/login", auth.Login)
    r.Post("/api/child/login", child.PINLogin)
    r.Post("/api/invite/{token}/activate", child.Activate)
})

// Parent-protected routes
r.Group(func(r chi.Router) {
    r.Use(jwtauth.Verifier(tokenAuth))
    r.Use(jwtauth.Authenticator(tokenAuth))
    r.Use(auth.ParentOnly)
    r.Post("/api/auth/logout", auth.Logout)
    r.Route("/api/apikey", func(r chi.Router) {
        r.Post("/", apikey.Store)
        r.Put("/", apikey.Update)
        r.Delete("/", apikey.Delete)
    })
    r.Route("/api/children", func(r chi.Router) {
        r.Post("/", child.Create)
        r.Get("/", child.List)
        r.Post("/{id}/invite", child.GenerateInvite)
    })
})
```

### Pattern 2: AES-256-GCM Envelope Encryption
**What:** Encrypt API keys at rest with a master key from environment
**When to use:** Storing the Claude API key

```go
func Encrypt(plaintext []byte, masterKey []byte) ([]byte, error) {
    block, err := aes.NewCipher(masterKey)
    if err != nil { return nil, err }
    aesGCM, err := cipher.NewGCM(block)
    if err != nil { return nil, err }
    nonce := make([]byte, aesGCM.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil { return nil, err }
    return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext []byte, masterKey []byte) ([]byte, error) {
    block, err := aes.NewCipher(masterKey)
    if err != nil { return nil, err }
    aesGCM, err := cipher.NewGCM(block)
    if err != nil { return nil, err }
    nonceSize := aesGCM.NonceSize()
    nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
    return aesGCM.Open(nil, nonce, ct, nil)
}
```

### Pattern 3: Atomic Invite Invalidation
**What:** Single SQL UPDATE that checks validity AND activates in one operation
**When to use:** When child activates an invite link

```sql
-- Atomically activate invite: returns the row only if valid + unused + not expired
UPDATE students
SET pin_hash = $2, activated_at = NOW(), invite_code = NULL
WHERE invite_code = $1
  AND activated_at IS NULL
  AND invite_expires_at > NOW()
RETURNING id, name, parent_id;
```

### Anti-Patterns to Avoid
- **SELECT then UPDATE for invites:** Race condition — two browser tabs could both validate. Use atomic UPDATE with WHERE clause.
- **Logging request bodies containing API keys:** Use a redacting logger that strips known sensitive fields.
- **Storing PIN as plaintext:** Hash PINs with Argon2id same as passwords, even though they're short.
- **JWT in localStorage:** Use httpOnly cookies to prevent XSS token theft.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Password hashing | Custom bcrypt/scrypt wrapper | alexedwards/argon2id | Handles salt generation, timing-safe comparison, OWASP-compliant params |
| JWT generation | Manual HMAC signing | go-chi/jwtauth | Handles token creation, verification, Chi middleware integration |
| Rate limiting | Custom counter map with mutex | golang.org/x/time/rate | Token bucket algorithm, goroutine-safe, well-tested |
| Connection pooling | Manual pgx.Connect per request | pgxpool.New | Health checks, idle timeout, max connections |
| SQL generation | String concatenation | sqlc | Type-safe, injection-proof, compile-time checked |
| Random tokens | math/rand | crypto/rand | Cryptographically secure for invite tokens |

**Key insight:** Auth and crypto are the two domains where hand-rolling most reliably produces vulnerabilities. Use established libraries for everything.

## Common Pitfalls

### Pitfall 1: Argon2id Parameter Tuning
**What goes wrong:** Using default parameters that are too weak for production or too strong for development
**Why it happens:** alexedwards/argon2id ships reasonable defaults but they may need adjustment
**How to avoid:** Use `argon2id.DefaultParams` which provides OWASP-recommended settings (memory=64MB, iterations=1, parallelism=2). Do NOT reduce for dev — keep consistent.
**Warning signs:** Login taking >1 second (params too high) or test suite timing out

### Pitfall 2: JWT Secret in Source Code
**What goes wrong:** Hardcoding the JWT signing secret or using a weak key
**Why it happens:** Quick setup during development
**How to avoid:** Generate a 256-bit random key, store in environment variable (`JWT_SECRET`). Use `crypto/rand` to generate. Never commit.
**Warning signs:** Same token works across environments

### Pitfall 3: AES Master Key Length
**What goes wrong:** Using a key that's not exactly 32 bytes for AES-256
**Why it happens:** Environment variable encoding confusion (hex vs raw bytes)
**How to avoid:** Store master key as 64-character hex string in env, decode with `hex.DecodeString`. Validate length == 32 at startup.
**Warning signs:** `crypto/aes: invalid key size` error at runtime

### Pitfall 4: Missing CORS Configuration
**What goes wrong:** Frontend SvelteKit dev server can't reach Go backend
**Why it happens:** Different ports in development (SvelteKit on 5173, Go on 8080)
**How to avoid:** Add CORS middleware to Chi with explicit origin allowlist. In production, serve everything from same origin via reverse proxy.
**Warning signs:** Preflight OPTIONS requests fail with 403

### Pitfall 5: Invite Token Entropy
**What goes wrong:** Guessable invite tokens
**Why it happens:** Using short random strings or sequential IDs
**How to avoid:** Generate 32 bytes from crypto/rand, encode as URL-safe base64 (43 characters). This gives 256 bits of entropy.
**Warning signs:** Token length < 20 characters

### Pitfall 6: PIN Hash vs Password Hash Confusion
**What goes wrong:** Treating PINs differently from passwords in the hash scheme
**Why it happens:** 4-digit PINs feel simpler
**How to avoid:** Hash PINs with the same Argon2id function as passwords. The rate limiting (5 attempts / 15 min lockout) provides the brute-force protection that the low entropy of 4-digit PINs requires.
**Warning signs:** PINs stored as plaintext or simple SHA-256 without salt

## Code Examples

### JWT Token Generation with go-chi/jwtauth
```go
import "github.com/go-chi/jwtauth/v5"

var tokenAuth *jwtauth.JWTAuth

func init() {
    secret := os.Getenv("JWT_SECRET")
    tokenAuth = jwtauth.New("HS256", []byte(secret), nil)
}

func GenerateToken(userID string, role string) (string, error) {
    claims := map[string]interface{}{
        "sub":  userID,
        "role": role, // "parent" or "child"
    }
    jwtauth.SetExpiry(claims, time.Now().Add(24*time.Hour))
    _, tokenString, err := tokenAuth.Encode(claims)
    return tokenString, err
}
```

### Rate Limiter per Student
```go
import "golang.org/x/time/rate"

type PINRateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*studentLimiter
}

type studentLimiter struct {
    attempts  int
    lockedAt  time.Time
    lastReset time.Time
}

func (rl *PINRateLimiter) Allow(studentID string) (bool, int) {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    sl, exists := rl.limiters[studentID]
    if !exists {
        sl = &studentLimiter{lastReset: time.Now()}
        rl.limiters[studentID] = sl
    }

    // Check if lockout has expired (15 min)
    if !sl.lockedAt.IsZero() && time.Since(sl.lockedAt) < 15*time.Minute {
        return false, 0
    } else if !sl.lockedAt.IsZero() {
        sl.lockedAt = time.Time{}
        sl.attempts = 0
    }

    sl.attempts++
    remaining := 5 - sl.attempts
    if sl.attempts >= 5 {
        sl.lockedAt = time.Now()
        return false, 0
    }
    return true, remaining
}
```

### Claude API Key Validation
```go
func ValidateAPIKey(apiKey string) error {
    req, _ := http.NewRequest("GET", "https://api.anthropic.com/v1/models", nil)
    req.Header.Set("x-api-key", apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return fmt.Errorf("connection failed") }
    defer resp.Body.Close()
    if resp.StatusCode == 401 { return fmt.Errorf("invalid API key") }
    if resp.StatusCode != 200 { return fmt.Errorf("unexpected status: %d", resp.StatusCode) }
    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| bcrypt | Argon2id | OWASP 2023 recommendation | Memory-hard, resists GPU attacks |
| JWT in localStorage | httpOnly cookie | Ongoing best practice | Prevents XSS token theft |
| pgx v4 | pgx v5 | 2023 | New API, better generics support |
| go-chi/jwtauth v1 | go-chi/jwtauth v5 | 2023 | Wraps lestrrat-go/jwx/v2 |

**Deprecated/outdated:**
- bcrypt: Still works but Argon2id is the recommended choice for new projects
- database/sql with lib/pq: pgx is the modern Go PostgreSQL driver

## Open Questions

1. **JWT refresh token strategy**
   - What we know: Single JWT with expiry is simplest. Refresh tokens add complexity.
   - What's unclear: Whether to implement refresh tokens in Phase 1 or add later.
   - Recommendation: Start with a single JWT (24h expiry) in httpOnly cookie. No refresh tokens in Phase 1 — add if needed. This is in Claude's Discretion per CONTEXT.md.

2. **Child session token format**
   - What we know: Children log in with name + PIN. They need a session.
   - What's unclear: Same JWT format as parents or simpler session?
   - Recommendation: Same JWT with `role: "child"` claim. Simplifies middleware — just check role in the same token.

## Sources

### Primary (HIGH confidence)
- Go stdlib crypto/aes, crypto/cipher — AES-256-GCM pattern is well-documented
- Go stdlib crypto/rand — secure random generation
- alexedwards/argon2id README — default params and usage
- go-chi/jwtauth README — middleware integration pattern

### Secondary (MEDIUM confidence)
- OWASP password storage cheat sheet — Argon2id recommended parameters
- Anthropic API docs — models endpoint for key validation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries locked by user decision, well-established Go ecosystem
- Architecture: HIGH — standard Go project layout with Chi routing patterns
- Pitfalls: HIGH — crypto and auth pitfalls are extensively documented

**Research date:** 2026-04-03
**Valid until:** 2026-05-03 (stable domain, established libraries)
