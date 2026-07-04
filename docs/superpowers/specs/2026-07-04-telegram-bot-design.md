# Telegram-plugg för nioplugget — design (familjeprototyp)

Datum: 2026-07-04
Status: Godkänd av användaren

## Syfte och hypotes

Barnet ska kunna plugga via en Telegram-bot — hela pluggflödet (välja övning,
sokratisk dialog, stjärnor/XP, SRS) sker i chatten. Uppladdning av läxor,
granskning och föräldrapanel stannar i web-GUI:t. Hypotesen: chattkanalen +
push-notiser gör att barnet faktiskt repeterar, eftersom Telegram bor i fickan
medan webbappen kräver att man öppnar den.

Detta är en prototyp för en familj. En central bot skapas via BotFather och
konfigureras i `.env`. Datamodellen (länktabell, inte hårdkodning) ska dock
generalisera till multi-familj senare utan omskrivning.

## Arkitektur

Nytt paket `backend/internal/telegram` som körs i samma process som servern:

- **Long polling** mot Telegram Bot API (ingen webhook, ingen publik URL —
  fungerar bakom NAT, t.ex. på MacBooken dit projektet flyttar).
- Anropar befintliga paket `chat`, `srs`, `challenges`, `progress` som vanliga
  Go-funktioner. Inga nya HTTP-endpoints förutom länkkodsflödet.
- `TELEGRAM_BOT_TOKEN` i `backend/.env`. Tom/saknad token ⇒ Telegram-lagret
  startar inte alls; noll påverkan på befintlig drift.
- Polling-loop med exponentiell backoff vid API-fel; fel loggas.

## Koppling barn ↔ Telegram

1. Barnets studysida i web-GUI:t får knappen **"Koppla Telegram"**.
2. Backend genererar en engångskod (giltig 15 min) via ny endpoint
   (autentiserad som barnet), frontend visar deep link
   `https://t.me/<botnamn>?start=<kod>`.
3. Barnet trycker på länken; boten tar emot `/start <kod>`, verifierar koden
   och sparar kopplingen.
4. Ny migration: tabell `telegram_links (child_id, telegram_user_id, chat_id,
   linked_at)` med unik `telegram_user_id`.
5. Meddelanden från okopplade `telegram_user_id` besvaras endast med en
   hänvisning: "be din förälder om en kopplingslänk". Boten är publikt nåbar
   på Telegram, så detta är säkerhetsgränsen.

## Konversationsflöde

- Kopplat barn möts av meny med inline-knappar:
  **📚 Plugga** (ämne → övning, med lås-/stjärnstatus),
  **⭐ Repetera (N)** (due SRS-övningar),
  **🏆 Utmaningar** (publicerade läxutmaningar),
  **📊 Progress** (kort sammanfattning).
- Vald övning startar en session i befintliga chat-motorn. Fri text därefter
  är sokratisk dialog mot samma motor, med förälderns BYOK-nyckel.
- **Ingen streaming:** Claudes svar buffras helt; boten skickar
  `sendChatAction: typing` under generering. Svar > 4096 tecken delas i flera
  meddelanden.
- Avslutad övning ger samma bedömning som webben: stjärnor/XP-meddelande
  ("⭐⭐⭐ 3 stjärnor! +50 XP") + SRS-schemaläggning. `/avsluta` eller
  avsluta-knapp avbryter en session och återgår till menyn.
- **Matte:** Telegram-kanalen lägger en extra systemprompt-instruktion:
  formler i enkel text/unicode (`x² + 3x = 10`), aldrig LaTeX.

## Sessionstate

Tabell `telegram_sessions (chat_id, child_id, state, active_chat_session_id,
updated_at)` — pekar på befintliga chat-sessioner och håller menyläge, så
state överlever omstart av backenden.

## Pushar

- Goroutine-ticker i backenden: inom fönstret 15:00–20:00 lokal tid, kolla
  due SRS-övningar per kopplat barn; skicka max **en** påminnelse per barn och
  dag, med "Kör nu"-knapp som startar repetitionsflödet.
- Vid publicering av läxutmaning (förälderflödet i webben): omedelbar notis
  till barnets chat med knapp till utmaningen.

## Felhantering och test

- Bot-API-fel: logga + backoff; sessioner får inte korrumperas av tappade
  uppdateringar (state i DB, idempotent hantering av update-ID).
- Enhetstester: kommandoparsning, länkkodsflödet (generering, TTL, engångs-
  användning), meddelandedelning vid 4096 tecken.
- Dialogmotorn är redan testad; Telegram-lagret testas mot en mockad Bot API-
  klient.

## Utanför scope (v1)

Uppladdning via Telegram, föräldranotiser, gruppchatt, flera barn per
Telegram-konto, generaliserad multi-familj-drift (BYOB eller central bot för
alla användare), webhook-läge.
