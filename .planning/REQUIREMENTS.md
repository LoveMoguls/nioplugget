# Requirements: Nioplugget

**Defined:** 2026-04-03
**Core Value:** Eleven lär sig genom dialog — AI:n ger aldrig direkta svar utan guidar eleven med ledande frågor

## v1 Requirements

### Authentication

- [ ] **AUTH-01**: Förälder kan registrera sig med e-post och lösenord
- [ ] **AUTH-02**: Förälder kan logga in och sessionen bevaras över sidladdning
- [ ] **AUTH-03**: Förälder kan logga ut
- [ ] **AUTH-04**: Förälder kan skapa barnprofil med namn
- [ ] **AUTH-05**: System genererar en engångs-invite-länk (tidsbegränsad) för barnprofilen
- [ ] **AUTH-06**: Barn kan aktivera konto via invite-länk och välja PIN-kod
- [ ] **AUTH-07**: Barn kan logga in med sitt namn + PIN
- [ ] **AUTH-08**: PIN-inloggning har rate limiting mot brute force

### API Key Management

- [ ] **KEY-01**: Förälder kan lägga in sin Claude API-nyckel
- [ ] **KEY-02**: API-nyckeln krypteras med AES-256-GCM innan den sparas i databasen
- [ ] **KEY-03**: Förälder kan uppdatera eller ta bort sin API-nyckel
- [ ] **KEY-04**: System visar tydligt felmeddelande om API-nyckeln är ogiltig eller utgången

### Content

- [ ] **CONT-01**: System har färdiga övningspass för Biologi (Ekologi, Kroppen, Genetik, Cellen)
- [ ] **CONT-02**: System har färdiga övningspass för Samhällskunskap (Demokrati, Rättigheter, Ekonomi, Lag & rätt)
- [ ] **CONT-03**: System har färdiga övningspass för Matematik (Algebra, Geometri, Statistik, Samband & förändring)
- [ ] **CONT-04**: Varje ämnesområde har 3-5 övningar med stigande svårighetsgrad
- [ ] **CONT-05**: Varje övning har en skräddarsydd system-prompt som följer Skolverkets centrala innehåll

### AI Dialog

- [ ] **CHAT-01**: Elev kan starta ett övningspass genom att välja ämne → ämnesområde → övning
- [ ] **CHAT-02**: AI-läraren ställer ledande frågor och ger aldrig direkta svar
- [ ] **CHAT-03**: AI-dialogen streamar svar i realtid via SSE
- [ ] **CHAT-04**: AI:n anpassar nivån baserat på elevens svar
- [ ] **CHAT-05**: AI:n håller sig strikt till ämnet och styr tillbaka vid avvikelser
- [ ] **CHAT-06**: Elev kan avsluta ett pass manuellt
- [ ] **CHAT-07**: Alla meddelanden sparas i databasen per session
- [ ] **CHAT-08**: Konversationshistoriken trunkeras för att hantera token-kostnader

### Spaced Repetition

- [ ] **SRS-01**: Varje avslutad session får en score (1-5) via AI-bedömning
- [ ] **SRS-02**: SM-2-algoritmen beräknar nästa repetitionsdatum baserat på score
- [ ] **SRS-03**: Startsidan visar "Dags att repetera"-kort för ämnesområden som passerat review-datum
- [ ] **SRS-04**: Ease factor har ett minimumgolv (1.3) för att undvika "ease hell"

### Progress

- [ ] **PROG-01**: Elev kan se en översikt av sin progress per ämne
- [ ] **PROG-02**: Elev kan se styrkor och svagheter baserat på sessionshistorik
- [ ] **PROG-03**: Förälder kan se sitt barns progress och ämnesöversikt

### Frontend

- [ ] **UI-01**: Mobilanpassad (mobile-first) responsiv design
- [ ] **UI-02**: Rent, lugnt UI utan distraktioner
- [ ] **UI-03**: Landningssida med info om tjänsten och registrering/inloggning
- [ ] **UI-04**: Chattvy med tydlig meddelandehistorik och input-fält

### Security

- [ ] **SEC-01**: Go-backend loggar aldrig API-nycklar eller Authorization-headers
- [ ] **SEC-02**: Invite-länkar är engångs och tidsbegränsade (atomisk invalidering)
- [ ] **SEC-03**: GDPR-samtycke samlas in explicit vid föräldraregistrering

## v2 Requirements

### Extended Auth

- **AUTH-V2-01**: Förälder kan återställa lösenord via e-post
- **AUTH-V2-02**: Förälder kan ha flera barn-profiler

### Extended Content

- **CONT-V2-01**: Fler ämnen (Kemi, Fysik, Geografi, Historia, Religion)
- **CONT-V2-02**: Elev kan föreslå egna övningsämnen (fritext)

### Notifications

- **NOTF-V2-01**: Påminnelse om dags att repetera (e-post eller push)

### Analytics

- **ANAL-V2-01**: Detaljerad statistik med diagram över progress över tid
- **ANAL-V2-02**: Förälder-dashboard med sammanfattning per vecka

## Out of Scope

| Feature | Reason |
|---------|--------|
| Betalning/prenumeration | BYOK-modell — användaren betalar Anthropic direkt |
| Fotouppladdning av läxor | Bara färdiga övningspass för kontrollerad kvalitet |
| Realtidschat mellan elever | Individuell AI-dialog, inte socialt |
| Mobilapp (native) | Web-first, mobilanpassat — native app i framtiden |
| Gamification (streaks, poäng, badges) | Anti-feature: skapar ångest, inte lärande |
| OAuth/social login | E-post/lösenord räcker för v1 |
| Mörkt/ljust tema-toggle | Följ systemets inställning, ingen manuell toggle i v1 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | — | Pending |
| AUTH-02 | — | Pending |
| AUTH-03 | — | Pending |
| AUTH-04 | — | Pending |
| AUTH-05 | — | Pending |
| AUTH-06 | — | Pending |
| AUTH-07 | — | Pending |
| AUTH-08 | — | Pending |
| KEY-01 | — | Pending |
| KEY-02 | — | Pending |
| KEY-03 | — | Pending |
| KEY-04 | — | Pending |
| CONT-01 | — | Pending |
| CONT-02 | — | Pending |
| CONT-03 | — | Pending |
| CONT-04 | — | Pending |
| CONT-05 | — | Pending |
| CHAT-01 | — | Pending |
| CHAT-02 | — | Pending |
| CHAT-03 | — | Pending |
| CHAT-04 | — | Pending |
| CHAT-05 | — | Pending |
| CHAT-06 | — | Pending |
| CHAT-07 | — | Pending |
| CHAT-08 | — | Pending |
| SRS-01 | — | Pending |
| SRS-02 | — | Pending |
| SRS-03 | — | Pending |
| SRS-04 | — | Pending |
| PROG-01 | — | Pending |
| PROG-02 | — | Pending |
| PROG-03 | — | Pending |
| UI-01 | — | Pending |
| UI-02 | — | Pending |
| UI-03 | — | Pending |
| UI-04 | — | Pending |
| SEC-01 | — | Pending |
| SEC-02 | — | Pending |
| SEC-03 | — | Pending |

**Coverage:**
- v1 requirements: 39 total
- Mapped to phases: 0
- Unmapped: 39 ⚠️

---
*Requirements defined: 2026-04-03*
*Last updated: 2026-04-03 after initial definition*
