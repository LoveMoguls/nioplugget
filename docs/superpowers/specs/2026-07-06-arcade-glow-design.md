# Arcade Glow — omdesign av nioplugget-frontend

Datum: 2026-07-06
Status: Godkänd av användaren

## Syfte

Appen ska kännas som en modern, populär app en designbyrå gjort 2026: rolig,
gaming, lekfull att plugga i. Nuvarande look (lugn blå/grön, systemfont,
standard-shadcn) är funktionell men anonym. Hela appen görs om; barnens vyer
får mest personlighet, föräldravyerna samma tema men dämpat.

## Identitet

**Arcade Glow** — mörk-först, en enda identitet (ljusa läget tas bort;
`.dark`-varianten i app.css utgår, mörka värdena blir default).

### Färgtokens (byts i `frontend/src/app.css` `@theme`)

- Bakgrund: djup mörkblå→lila, `oklch(0.16 0.03 285)`; kort `oklch(0.21 0.035 285)`
- Primary: neon-cyan `oklch(0.85 0.15 195)` (text på mörk: foreground mörk)
- Accent: magenta `oklch(0.70 0.25 330)`
- XP/success: limegrön `oklch(0.85 0.22 135)`
- Stjärnor: gul `oklch(0.85 0.17 90)`
- Destructive behålls röd; borders subtila `oklch(0.30 0.03 285)`
- Muted-foreground ljusgrå-lila för läsbarhet (WCAG AA mot bakgrunden)

Exakta värden får justeras för kontrast under implementation — riktningen
ovan är bindande, inte decimalerna.

### Typografi

- Rubriker/display: **Baloo 2** (vikt 600–800)
- Brödtext: **Nunito** (400/600/700)
- Via `@fontsource/baloo-2` + `@fontsource/nunito` (npm, self-hostade,
  importeras i app.css eller +layout) — inga externa font-anrop
- Barnvyer: stora rubriker, gärna versaler + lätt letter-spacing

### Effekter

- Glow: färgade box-shadows (t.ex. `0 0 24px oklch(... / 0.35)`)
- Gradient-borders på framhävda kort (pseudo-element eller border-image)
- Tryck-ner-effekt på primära knappar (translateY + skugg-reduktion vid
  :active)
- Micro-transitions 150–250 ms på interaktiva element
- `prefers-reduced-motion: reduce` stänger av dekorativa animationer
- Inga nya JS-animationsbibliotek — CSS + Svelte transitions

## Signaturkomponenter (nya, i `src/lib/components/arcade/`)

- **XPBar.svelte** `{current, max, label?}` — limegrön glödande mätare med
  animerad fill (CSS transition på width)
- **GlowCard.svelte** `{gradient?}` (slot) — kort med gradient-kant, hover-
  lyft + glow
- **PressButton.svelte** (slot, extends button-props) — stor CTA med
  3D-tryck och neonglow
- Profilväljarens avatar får roterande gradient-ring (CSS
  `@keyframes` conic-gradient) + bounce på hover

## Per vy

- **/las-upp**: minimal mörk scen, pulserande "NIOPLUGGET"-logotyp
  (text i Baloo, cyan-glow), ett glödande kodfält, PressButton "LÅS UPP".
  Diskreta /login-länken kvar.
- **/profiler**: rubrik "VEM PLUGGAR?" i neon-cyan, avatarer 96px med
  gradient-ring, namn i Baloo, bounce på hover/fokus.
- **/study**: XPBar överst (total XP från progress-datat om det finns
  lätt åtkomligt — annars visas stjärnsumma; ingen ny backend). Ämneskort
  = GlowCard med djup gradient per ämne + stor emoji. Repetera-sektionen
  rubriceras "⚡ DAGENS UPPDRAG". Utmaningssektionen behåller sin
  funktionalitet med ny styling.
- **/study/[subject] (+ övningslistor)**: övningskort med stjärnrad;
  klarade övningar får grön glow-kant; lås-ikon för olåsta behålls där
  logiken finns.
- **Chat**: mörk chattyta; elevens bubblor cyan-gradient, AI:ns kort
  mörkgrå med subtil border; studsande typing-prickar; befintlig
  StarAnimation behålls (får gärna större konfetti).
- **/challenges + challenge-vyer**: emoji som hero, progress som ring eller
  XPBar, GlowCard-lista.
- **/progress**: stat-tiles (XP, stjärnor, antal pass) med stor siffra i
  Baloo + glow; ämnesframsteg som neon-staplar (befintlig data, ingen ny).
- **/dashboard + /setup + /login + /register**: samma tokens och typografi
  men dämpat — ingen/svag glow, mer luft. Familjekod- och API-nyckelkorten
  följer nya kortstilen.
- **+layout (nav)**: mörk glasig topbar (backdrop-blur), logotyp i Baloo,
  "Byt profil" som tydlig knapp.
- **/ (rot)**: oförändrad logik (redirect-sida), bara mörk bakgrund så
  det inte blinkar vitt.

## Ramar

- Ingen logik-, API- eller routingändring. Endast styling/markup +
  nya presentationskomponenter.
- shadcn-komponenterna behålls; ny look via tokens + varianter, inte forks,
  där det räcker.
- Svelte 5 runes-idiom som resten av kodbasen.
- All text fortsatt svenska; inga nya beroenden utöver @fontsource-paketen.
- Mobilen är förstklassig: barnvyerna testas i 390px-bredd.

## Verifiering

- `npx svelte-check --threshold error` (0 fel) + `npm run build`
- Screenshots (Playwright eller manuellt) av: las-upp, profiler, study,
  övningslista, chat, challenges, progress, dashboard — mot dev-servern
- Kontrastkontroll på brödtext/muted mot bakgrund (AA)

## Utanför scope

Ljudeffekter, streaks/nya gamification-mekaniker (bara visualisering av
befintlig data), light mode, logotyp/varumärkesarbete utanför appen,
backend-ändringar.
