# Nioplugget

**AI-driven studieapp for nationella proven i arskurs 9.**

Nioplugget hjalper elever att forbereda sig for de nationella proven genom dialog med en AI-larare som staller ledande fragor istallet for att ge fardiga svar. Appen tackar biologi, samhallskunskap och matematik -- mappat mot Skolverkets centrala innehall.

Pedagogisk modell: *"Fragor, inte svar"* kombinerat med spaced repetition (SM-2).

## Features

- **Sokratisk AI-handledning** -- AI:n ger aldrig direkta svar utan guidar eleven med ledande fragor, anpassade efter elevens niva
- **Tre amnen** -- Biologi (ekologi, kroppen, genetik, cellen), Samhallskunskap (demokrati, rattigheter, ekonomi, lag & ratt), Matematik (algebra, geometri, statistik, samband & forandring)
- **Fardiga ovningspass** -- ~50-60 ovningar skraddarsydda med system-prompts, mappade mot kursplanen
- **Spaced repetition (SM-2)** -- Automatisk repetitionsplanering baserat pa resultat
- **Foraldra-barn-modell** -- Foraldern hanterar konto och API-nyckel, barnet studerar via invite-lank + PIN
- **BYOK (Bring Your Own Key)** -- Anvandaren kopplar sin egen Claude API-nyckel, krypterad med AES-256-GCM at rest
- **Progress-tracking** -- Oversikt over styrkor och svagheter per amne for bade elev och foraldrar
- **Mobilanpassat** -- Rent, lugnt UI optimerat for korta ovningspass (10-15 min)
- **Streaming-svar** -- AI-svar streamar i realtid via SSE

## Tech Stack

| Layer    | Technology                          |
|----------|-------------------------------------|
| Backend  | Go (Chi router)                     |
| Frontend | SvelteKit 2 + Svelte 5 (runes)     |
| UI       | shadcn-svelte + Tailwind CSS 4      |
| Database | PostgreSQL                          |
| AI       | Claude API (Anthropic SDK for Go)   |
| Auth     | JWT + Argon2id password hashing     |
| Crypto   | AES-256-GCM for API key encryption  |

## Prerequisites

- **Go** 1.23+
- **Node.js** 20+
- **PostgreSQL** 15+
- **Claude API key** from [Anthropic](https://console.anthropic.com/)
- **golang-migrate** CLI -- `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
- **sqlc** -- `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

## Quick Start

### 1. Clone the repo

```bash
git clone https://github.com/trollstaven/nioplugget.git
cd nioplugget
```

### 2. Set up PostgreSQL

```bash
createdb nioplugget
```

### 3. Set environment variables

```bash
export DATABASE_URL="postgres://user:password@localhost:5432/nioplugget?sslmode=disable"
export ENCRYPTION_KEY="$(openssl rand -hex 16)"  # 32 hex chars = 16 bytes for AES-256-GCM
export JWT_SECRET="$(openssl rand -hex 32)"
```

### 4. Run database migrations

```bash
migrate -path backend/db/migrations -database "$DATABASE_URL" up
```

### 5. Generate sqlc code

```bash
cd backend && sqlc generate
```

### 6. Start the backend

```bash
cd backend && go run cmd/server/main.go
```

The server starts on `http://localhost:8080` by default. Set `PORT` to change it.

### 7. Start the frontend

```bash
cd frontend && npm install && npm run dev
```

The dev server starts on `http://localhost:5173`.

## Environment Variables

| Variable         | Required | Description                                                        |
|------------------|----------|--------------------------------------------------------------------|
| `DATABASE_URL`   | Yes      | PostgreSQL connection string                                       |
| `ENCRYPTION_KEY` | Yes      | Hex-encoded key for AES-256-GCM encryption of stored API keys      |
| `JWT_SECRET`     | Yes      | Secret for signing JWT tokens                                      |
| `PORT`           | No       | Backend server port (default: `8080`)                              |

## Project Structure

```
nioplugget/
├── backend/
│   ├── cmd/server/main.go          # Entrypoint, router setup
│   ├── db/
│   │   ├── migrations/             # PostgreSQL migrations (golang-migrate)
│   │   └── queries/                # SQL queries (sqlc source)
│   ├── internal/
│   │   ├── apikey/                 # API key encryption, storage, validation
│   │   ├── auth/                   # JWT, Argon2id hashing, middleware
│   │   ├── chat/                   # Chat sessions, AI prompts, streaming, scoring
│   │   ├── child/                  # Child profiles, invite links, PIN login
│   │   ├── content/                # Subjects, topics, exercises
│   │   ├── database/               # Connection pool + sqlc generated code
│   │   ├── middleware/             # CORS, request logging
│   │   ├── progress/               # Student progress tracking
│   │   └── srs/                    # SM-2 spaced repetition algorithm
│   └── sqlc.yaml
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api.ts              # Backend API client
│   │   │   ├── components/
│   │   │   │   ├── chat/           # ChatBubble, ChatInput, TypingIndicator
│   │   │   │   └── ui/             # shadcn-svelte components
│   │   │   └── stores/             # Auth and chat state (Svelte stores)
│   │   └── routes/
│   │       ├── chat/[sessionId]/   # AI chat session view
│   │       ├── child/login/        # Child PIN login
│   │       ├── dashboard/          # Parent dashboard + child progress
│   │       ├── invite/[token]/     # Invite link activation
│   │       ├── login/              # Parent login
│   │       ├── progress/           # Student progress overview
│   │       ├── register/           # Parent registration
│   │       ├── setup/              # API key setup
│   │       └── study/              # Subject → topic → exercise navigation
│   └── package.json
└── README.md
```

## How It Works

### Sokratisk AI-handledning

Varje ovningspass har en skraddarsydd system-prompt som instruerar Claude att agera som en pedagogisk handledare. AI:n staller ledande fragor for att guida eleven mot ratt svar istallet for att ge det direkt. Svaren streamar i realtid via Server-Sent Events (SSE).

### BYOK -- Bring Your Own Key

Foraldern laggar in sin egen Claude API-nyckel i appen. Nyckeln krypteras med AES-256-GCM innan den sparas i databasen och dekrypteras enbart vid anrop till Claude API:t. Ingen API-nyckel lagras i klartext.

### Spaced Repetition (SM-2)

Efter varje avslutad ovning bedoms elevens prestation. SM-2-algoritmen beraknar nar ovningen bor repeteras. Pa startsidan visas "Dags att repetera"-kort sa eleven alltid vet vad som bor trainas haernaest.

### Foraldra-barn-modell

Foraldern registrerar sig, laggar in sin API-nyckel och skapar barnprofiler via invite-lankar. Barnet aktiverar sin profil med en PIN-kod och far tillgang till ovningarna. Foraldern kan folja barnets framsteg.

## License

MIT
