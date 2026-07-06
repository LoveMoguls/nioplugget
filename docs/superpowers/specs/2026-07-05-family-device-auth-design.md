# Familjeenhet + profilväljare — design

Datum: 2026-07-05
Status: Godkänd av användaren

## Syfte

Nioplugget används av en enda familj men nås på en publik URL. Dagens auth
(förälder: mejl+lösenord; barn: namn + PIN; JWT som går ut efter 24 h) ger
daglig inloggningsfriktion. Målet: **noll friktion i vardagen** — öppna
sajten, tryck på din avatar, inne — utan att släppa in främlingar som kan
chatta på förälderns Claude API-nyckel.

## Koncept

Netflix-modellen bakom en engångs-upplåsning per enhet:

1. **Ny enhet**: sida "Ange familjekod" → koden skrivs in en gång →
   enheten blir betrodd i 365 dagar (HttpOnly-cookie).
2. **Betrodd enhet**: profilväljare med alla familjens profiler
   (föräldrar + barn) → tryck på avatar → inloggad (30 dagars JWT).
3. **Byte av familjekod** kastar ut alla betrodda enheter (epoch-räknare).

## Backend

### Datamodell (migration 013)

```sql
CREATE TABLE family_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    code_hash TEXT NOT NULL,
    device_epoch INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

En rad. `code_hash` = Argon2id (samma `auth.HashPassword`/`VerifyPassword`
som lösenord/PIN). Ingen familjekod satt ⇒ tabellen tom ⇒ upplåsning
omöjlig tills föräldern satt koden (bootstrap via lösenordslogin).

### Nytt paket `internal/device`

- `POST /api/device/unlock {code}` (publik, rate-limitad 5 försök/15 min
  per IP via samma mekanism som `child.PINRateLimiter`):
  verifierar koden mot `code_hash`; vid träff sätts HttpOnly-cookie
  `device` = JWT `{role:"device", epoch:N}`, giltighet 365 dagar.
  Fel kod ⇒ 401 `{"error":"Fel familjekod"}`.
- `GET /api/profiles` (kräver giltig device-cookie med `epoch` ==
  aktuell epoch): alla föräldrar + alla studenter:
  `[{id, name, role: "parent"|"child"}]`. Avatarer genereras i frontend
  (initial + färg-hash) — ingen schemaändring.
- `POST /api/profile/login {id, role}` (kräver device-cookie):
  slår upp användaren, sätter ordinarie `jwt`-cookie med rätt roll.
- `POST /api/device/set-code {newCode, currentCode?}` (förälder-JWT krävs):
  sätter/byter familjekoden. Om en kod redan finns krävs `currentCode`
  (eller att anropet verifieras mot den). Byte bumpar `device_epoch`
  (alla enheter måste låsas upp igen). Minst 6 tecken.

### Ändringar i befintlig kod

- `auth.GenerateToken`: giltighet 24 h → **30 dagar** (gäller alla roller;
  lösenords- och PIN-login får samma förlängning automatiskt).
- Device-JWT signeras med samma `JWT_SECRET`/jwtauth-instans men läses av
  egen middleware i device-paketet (rollen "device" ger INTE tillgång till
  vanliga API:er — `ParentOnly`/`ChildOnly` släpper inte igenom den).
- **Känsliga åtgärder kräver familjekoden i request-bodyn** när en kod
  finns satt: `POST/PUT/DELETE /api/apikey` får fältet `familyCode` och
  verifierar det (utan satt kod: oförändrat beteende). `set-code` enligt
  ovan.

## Frontend

- **`/las-upp`**: enda fält "Familjekod" + knapp. Vid succé → `/profiler`.
- **`/profiler`**: rutnät med avatarer (initial + deterministisk färg ur
  namnet), namn under. Tryck → `POST /api/profile/login` → förälder till
  `/dashboard`, barn till `/study`.
- **Rotflödet `/`**: har giltig jwt → redirect till app (roll avgör).
  Annars har device-cookie → `/profiler`. Annars → `/las-upp`.
  (Avgörs med `GET /api/auth/me` resp. `GET /api/profiles` — 401 ⇒ nästa steg.)
- **"Byt profil"** i appens meny: `POST /api/auth/logout`-motsvarighet för
  båda roller (rensar jwt) → `/profiler`.
- **Föräldrapanelen**: nytt kort "Familjekod" — sätt/byt kod (två fält:
  ny kod ×2; nuvarande kod krävs vid byte). Info: "Byte loggar ut alla
  enheter."
- **Bort ur UI:t**: `/child/login` (PIN-sidan), invite-flödet
  (generera-invite-knappen i föräldrapanelen), länkar till `/login` från
  landningssidan ersätts med flödet ovan. `/login` (mejl+lösenord) behålls
  som dold fallback-URL för bootstrap/nödfall. Backend-endpoints för
  PIN/invite lämnas orörda.

## Säkerhet

- Familjekoden rate-limitas (5/15 min/IP) och lagras som Argon2id-hash.
- Device-cookien är HttpOnly, Secure (via befintlig cookie-hjälpare),
  365 d; jwt-cookien 30 d. Epoch-bump = omedelbar fjärrutkastning.
- Rollen "device" ger bara `GET /api/profiles` + `POST /api/profile/login`.
- API-nyckelhantering kräver familjekoden — barn på betrodd enhet kan
  plugga men inte röra nyckeln.
- Kvarvarande risk (accepterad): den som kan familjekoden och URL:en
  kommer in som vilken profil som helst. Familjeintern tjänst; koden är
  bytbar och utkastning global.
- Profilvalda sessioner (via profilväljaren på en betrodd enhet)
  epoch-stämplas vid inloggning och kastas ut omedelbart vid kodbyte
  (epoch-bump). Lösenordsinloggade sessioner (bootstrap via `/login`)
  saknar epoch-claim och påverkas medvetet inte av kodbyten.
- `CF-Connecting-IP` litas på för rate limit-nyckeln (unlock-endpointen).
  Vid direktåtkomst på det lokala nätverket (utan Cloudflare-tunneln
  emellan) kan headern förfalskas av klienten. Accepterad risk för en
  familjeintern tjänst på LAN.

## Testning

- device-paketet: unlock (rätt/fel kod/ingen kod satt/rate limit),
  epoch-mismatch avvisas, profiles/login kräver device-cookie,
  set-code kräver förälder + nuvarande kod + bumpar epoch.
- apikey: familyCode-krav när kod finns, oförändrat annars.
- Befintliga auth/child-tester ska fortsätta passera (24h→30d justeras).

## Utanför scope

Passkeys/WebAuthn, Telegram-godkännande av enheter, per-profil-PIN,
avatarval/emoji, borttagning av PIN/invite-backend, flera familjer.
