# Läxutmaningar — Design Spec

**Datum:** 2026-05-21  
**Status:** Godkänd

## Sammanfattning

Ett gamifierat utmaningsspår där föräldrar och barn laddar upp bilder på läxor (1–6 st), Claude analyserar dem och genererar automatiskt en named kurs med 4–6 Sokratiska övningar. Designat för en 13-årig tjej med ADD — korta bursts, omedelbar visuell belöning, tydlig progress.

Utmaningar är ett **eget spår**, helt separerat från de vanliga ämnena (Biologi, SO, Matte).

---

## Målgrupp & designprinciper

- 13-årig tjej med ADD och koncentrationssvårigheter
- **Korta, självständiga övningar** — varje övning är ett eget kort
- **Omedelbar belöning** — stjärnanimation + XP visas direkt efter varje svar
- **Tydlig progress** — barnet ser alltid hur nära de är att bli klara
- **Ingen väntan** — generation tar ~5–10 sek, sedan är utmaningen redo

---

## Datamodell

### Nya tabeller

```sql
CREATE TABLE challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL REFERENCES parents(id) ON DELETE CASCADE,
    created_by_role TEXT NOT NULL CHECK (created_by_role IN ('parent', 'child')),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    cover_emoji TEXT NOT NULL DEFAULT '📚',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE challenge_exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    system_prompt TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Ändring i befintlig tabell

```sql
ALTER TABLE sessions
    ALTER COLUMN exercise_id DROP NOT NULL,
    ADD COLUMN challenge_exercise_id UUID REFERENCES challenge_exercises(id),
    ADD CONSTRAINT sessions_one_exercise_type CHECK (
        (exercise_id IS NOT NULL) != (challenge_exercise_id IS NOT NULL)
    );
```

En session tillhör antingen en vanlig `exercise` eller en `challenge_exercise`, aldrig båda och aldrig ingen.

---

## Backend

### Nya endpoints

| Method | Path | Auth | Beskrivning |
|--------|------|------|-------------|
| `POST` | `/api/challenges` | parent eller child JWT | Skapa utmaning via bilduppladdning |
| `GET` | `/api/challenges` | parent eller child JWT | Lista utmaningar |
| `GET` | `/api/challenges/{id}` | parent eller child JWT | Hämta utmaning + övningar |
| `DELETE` | `/api/challenges/{id}` | parent JWT only | Ta bort utmaning |

Befintlig `POST /api/sessions` utökas med stöd för `challenge_exercise_id`.

### Uppladdning & generering

1. `POST /api/challenges` tar emot `multipart/form-data` med 1–6 bildfiler (max 5 MB/bild)
2. Bilderna base64-enkodas och skickas till Claude (`claude-sonnet-4-6`) med ett generations-prompt
3. Claude returnerar JSON:
   ```json
   {
     "title": "Fotosyntes-mästaren!",
     "description": "Lär dig hur växter gör sin egen mat",
     "emoji": "🌱",
     "exercises": [
       {
         "title": "Vad behöver en växt?",
         "description": "Utforska de tre ingredienserna",
         "system_prompt": "..."
       }
     ]
   }
   ```
4. Backend sparar challenge + exercises i DB
5. Bilderna sparas **inte** — de skickas direkt till Claude och kastas

### Behörigheter

- **Förälder:** kan skapa, lista och ta bort sina egna utmaningar
- **Barn:** kan skapa (backend slår upp `parent_id` via `students.parent_id`, använder förälderns API-nyckel), lista och starta övningar på utmaningar kopplade till sin förälder
- Barn kan **inte** ta bort utmaningar

### Claude generations-prompt (backend)

```
Du är en pedagog som skapar engagerande studieövningar för en 13-årig elev.
Analysera bilderna och skapa ett JSON-objekt med följande struktur:
- title: En catchy, motiverande titel på svenska (max 40 tecken, gärna med utropstecken)
- description: En kort mening som beskriver vad kursen handlar om
- emoji: En passande emoji
- exercises: Array med 4–6 övningar. Varje övning har:
  - title: Kort övningsrubrik
  - description: En mening som beskriver övningen
  - system_prompt: Fullständig Sokratisk systemPrompt på svenska — ställ frågor, ge aldrig direkta svar

Svara ENBART med JSON, ingen annan text.
```

---

## Frontend

### Nya routes

- `/challenges` — lista alla utmaningar (barn + förälder)
- `/challenges/[id]` — utmaningssida med övningskort

### Uppladdningsmodal

Tillgänglig från:
- Föräldravyn (`/dashboard`) — "Ny utmaning"-knapp
- Barnvyn (`/study`) — "Ny utmaning"-kort i utmaningssektionen

Flöde:
```
Drag-and-drop / filväljare (1–6 bilder)
  → Thumbnails visas + "Skapa utmaning"-knapp
  → Laddningsanimation: "Claude läser din läxa..." (~5–10 sek)
  → Redirect till /challenges/[id] — redo att spela
```

### Utmaningslista (`/challenges`)

- Grid av kort, ett per utmaning
- Varje kort: emoji + titel + "X/Y klara" + intjänade stjärnor
- Nya utmaningar pulserar subtilt
- XP-bar högst upp: totalt intjänat XP över alla utmaningar

### Utmaningssida (`/challenges/[id]`)

- Rad av 4–6 övningskort
- Första kortet är öppet, resten låsta
- Klick på kort → befintlig chat-UI (återanvänder `/chat/[sessionId]`)
- Efter avslutad övning:
  - Stjärnanimation (1–3 ⭐) visas i 2 sekunder
  - "+X XP!" räknas upp
  - Nästa kort låses upp automatiskt

### Stjärnberäkning

| Score (1–5) | Stjärnor | XP |
|-------------|----------|----|
| 1–2 | ⭐ | 10 |
| 3–4 | ⭐⭐ | 20 |
| 5 | ⭐⭐⭐ | 30 |

---

## Felhantering

| Scenario | Beteende |
|----------|----------|
| Otydliga/tomma bilder | Felmeddelande: "Kunde inte läsa bilderna, försök med tydligare foton" |
| Generation timeout (>30 sek) | Felmeddelande, användaren försöker igen |
| API-nyckel saknas | "Din förälder behöver lägga till en API-nyckel först" + länk till inställningar |
| Claude returnerar 0 övningar | Utmaningen tas bort automatiskt, fel visas |
| Fil för stor (>5 MB) | Valideras i frontend innan upload |

---

## Vad som INTE ingår (nu)

- Bilderna sparas inte för senare visning
- Notifikationer när förälder skapar utmaning (barn ser den nästa gång de öppnar appen)
- Nivåsystem baserat på totalt XP (kan läggas till senare)
- Delning av utmaningar mellan konton
