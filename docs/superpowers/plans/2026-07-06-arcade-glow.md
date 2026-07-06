# Arcade Glow Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Hela frontenden får identiteten "Arcade Glow" — mörk gaming-look med neonaccenter, Baloo 2/Nunito-typografi och lekfulla mikroanimationer. Spec: `docs/superpowers/specs/2026-07-06-arcade-glow-design.md`.

**Architecture:** Tokens-först: hela paletten byts i `frontend/src/app.css` `@theme` så alla shadcn-komponenter följer med. Tre nya presentationskomponenter i `src/lib/components/arcade/`. Därefter restylas vyerna sida för sida — endast styling/markup, ALDRIG logik (API-anrop, state, redirects, event-handlers lämnas orörda).

**Tech Stack:** SvelteKit (Svelte 5 runes), Tailwind CSS v4 (`@theme`-tokens), shadcn-svelte, @fontsource.

## Global Constraints

- ENDAST styling/markup + nya presentationskomponenter. Ingen ändring av logik, API-anrop, stores, routing eller flöden. Om en fils logik måste röras för att nå målet: rapportera DONE_WITH_CONCERNS istället för att röra den.
- Mörk-först, EN identitet: `.dark`-blocket i app.css tas bort; mörka värden blir default.
- Bindande palettriktning (exakta decimaler får finjusteras för kontrast): bakgrund `oklch(0.16 0.03 285)`, kort `oklch(0.21 0.035 285)`, primary neon-cyan `oklch(0.85 0.15 195)`, accent magenta `oklch(0.70 0.25 330)`, XP/success lime `oklch(0.85 0.22 135)`, stjärngul `oklch(0.85 0.17 90)`, border `oklch(0.30 0.03 285)`.
- Typografi: Baloo 2 (display/rubriker), Nunito (brödtext) via @fontsource — inga externa font-anrop.
- Micro-transitions 150–250 ms; allt dekorativt bakom `@media (prefers-reduced-motion: no-preference)` eller motsvarande; inga nya JS-animationsbibliotek.
- Brödtext/muted mot bakgrund ska klara WCAG AA (kontrast ≥ 4.5:1); verifiera med oklch-omdöme + stickprov i devtools/beräkning.
- All text svenska. Svelte 5 runes-idiom. Mobil 390px ska se bra ut på barnvyerna.
- Verifiering per task: `cd frontend && npx svelte-check --threshold error` (0 fel) `&& npm run build` (grön).
- Deploya INTE (starta inte om tjänster) — det gör controllern i sista tasken.
- Commit efter varje grönt steg, trailer: `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`.

---

### Task 1: Fundament — tokens, fonter, global bas

**Files:**
- Modify: `frontend/src/app.css` (hela `@theme` + ta bort `.dark`-blocket)
- Modify: `frontend/package.json` (+@fontsource/baloo-2, @fontsource/nunito)
- Modify: `frontend/src/routes/+layout.svelte` (font-importer om de inte läggs i app.css)
- Modify: `frontend/src/routes/+page.svelte` (mörk bakgrund på redirect-sidan)

**Interfaces:**
- Produces (alla senare tasks): CSS-tokens ovan; utility-klasser `font-display` (Baloo 2); CSS-variabler `--glow-cyan`, `--glow-magenta`, `--glow-lime`, `--glow-gold` för box-shadows.

- [ ] **Step 1: Installera fonter**

```bash
cd /home/admin/projekt/nioplugget/frontend
npm install @fontsource/baloo-2 @fontsource/nunito
```

- [ ] **Step 2: Skriv om `src/app.css`**

Ersätt HELA nuvarande `@theme`-block + `.dark`-block med (behåll `@import "tailwindcss";`, `@import "tw-animate-css";` och övriga befintliga regler under dem — läs filen till slutet först):

```css
@import "tailwindcss";
@import "tw-animate-css";
@import "@fontsource/baloo-2/600.css";
@import "@fontsource/baloo-2/700.css";
@import "@fontsource/baloo-2/800.css";
@import "@fontsource/nunito/400.css";
@import "@fontsource/nunito/600.css";
@import "@fontsource/nunito/700.css";

@theme {
  --font-sans: "Nunito", ui-sans-serif, system-ui, sans-serif,
    "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji";
  --font-display: "Baloo 2", var(--font-sans);

  /* Arcade Glow — mörk gaming-bas med neonaccenter */
  --color-background: oklch(0.16 0.03 285);
  --color-foreground: oklch(0.95 0.01 285);
  --color-card: oklch(0.21 0.035 285);
  --color-card-foreground: oklch(0.95 0.01 285);
  --color-popover: oklch(0.21 0.035 285);
  --color-popover-foreground: oklch(0.95 0.01 285);
  --color-primary: oklch(0.85 0.15 195);
  --color-primary-foreground: oklch(0.18 0.03 285);
  --color-secondary: oklch(0.27 0.04 285);
  --color-secondary-foreground: oklch(0.95 0.01 285);
  --color-muted: oklch(0.25 0.03 285);
  --color-muted-foreground: oklch(0.72 0.02 285);
  --color-accent: oklch(0.70 0.25 330);
  --color-accent-foreground: oklch(0.98 0.005 285);
  --color-destructive: oklch(0.65 0.24 25);
  --color-border: oklch(0.30 0.03 285);
  --color-input: oklch(0.30 0.03 285);
  --color-ring: oklch(0.85 0.15 195);
  --color-success: oklch(0.85 0.22 135);
  --color-gold: oklch(0.85 0.17 90);
  --color-chart-1: oklch(0.85 0.15 195);
  --color-chart-2: oklch(0.85 0.22 135);
  --color-chart-3: oklch(0.70 0.25 330);
  --color-chart-4: oklch(0.85 0.17 90);
  --color-chart-5: oklch(0.75 0.12 250);
  --color-sidebar: oklch(0.19 0.03 285);
  --color-sidebar-foreground: oklch(0.95 0.01 285);
  --color-sidebar-primary: oklch(0.85 0.15 195);
  --color-sidebar-primary-foreground: oklch(0.18 0.03 285);
  --color-sidebar-accent: oklch(0.70 0.25 330);
  --color-sidebar-accent-foreground: oklch(0.98 0.005 285);
  --color-sidebar-border: oklch(0.30 0.03 285);
  --color-sidebar-ring: oklch(0.85 0.15 195);

  --radius-sm: calc(1rem - 6px);
  --radius-md: calc(1rem - 3px);
  --radius-lg: 1rem;
  --radius-xl: calc(1rem + 6px);
}

:root {
  --radius: 1rem;
  --glow-cyan: 0 0 24px oklch(0.85 0.15 195 / 0.35);
  --glow-magenta: 0 0 24px oklch(0.70 0.25 330 / 0.35);
  --glow-lime: 0 0 20px oklch(0.85 0.22 135 / 0.4);
  --glow-gold: 0 0 20px oklch(0.85 0.17 90 / 0.4);
}

.font-display {
  font-family: var(--font-display);
}

body {
  background: linear-gradient(160deg, oklch(0.17 0.035 300) 0%, oklch(0.15 0.03 270) 100%);
  min-height: 100vh;
}
```

(OBS: om Tailwind v4 redan genererar `font-display`-utility ur `--font-display`-tokenen räcker tokenen — verifiera med en testklass; behåll annars den manuella klassen. Ta bort `.dark`-blocket och `@custom-variant dark` HELT. Läs resten av filen — behåll ev. `@layer base`-regler men uppdatera färghårdkodningar som krockar.)

- [ ] **Step 3: Rot-sidan mörk** — `src/routes/+page.svelte`: redirect-sidan ska bara visa mörk bakgrund (ingen vit blink); se till att dess wrapper inte sätter egen ljus bakgrund.

- [ ] **Step 4: Svep för hårdkodade ljusa färger** — `grep -rn "bg-white\|text-black\|bg-gray-\|bg-slate-\|text-gray-\|text-slate-" src/routes src/lib/components --include='*.svelte' | grep -v node_modules`. Byt träffar mot token-klasser (`bg-card`, `text-foreground`, `text-muted-foreground` osv). Gradient-klasser på ämneskorten (`bg-gradient-to-br ...`) lämnas — de görs om i Task 4.

- [ ] **Step 5: Verifiera** — `npx svelte-check --threshold error && npm run build` grönt. Starta `npm run preview -- --port 5199` och `curl -s localhost:5199/las-upp | grep -io 'baloo' | head -1` (fonten refereras i CSS-bundlen: kontrollera istället att build-output innehåller baloo-woff2: `ls .svelte-kit/output/client/_app/immutable/assets/ | grep -i baloo | head -2`). Döda preview-processen.

- [ ] **Step 6: Commit** — `git commit -m "feat(design): arcade glow tokens, fonts and dark-first base"`

---

### Task 2: Arcade-komponenter + topbar

**Files:**
- Create: `frontend/src/lib/components/arcade/XPBar.svelte`
- Create: `frontend/src/lib/components/arcade/GlowCard.svelte`
- Create: `frontend/src/lib/components/arcade/PressButton.svelte`
- Create: `frontend/src/lib/components/arcade/index.ts`
- Modify: `frontend/src/routes/+layout.svelte` (topbar)

**Interfaces:**
- Produces: `import { XPBar, GlowCard, PressButton } from '$lib/components/arcade';`
  - `XPBar` props: `{ current: number; max: number; label?: string }`
  - `GlowCard` props: `{ gradient?: string }` (CSS-gradient för kanten; default cyan→magenta) + default-slot (`children` snippet)
  - `PressButton` props: standard button-attrs + `{ variant?: 'cyan' | 'magenta' | 'lime' }` + default-slot

- [ ] **Step 1: XPBar.svelte**

```svelte
<script lang="ts">
	let { current, max, label = 'XP' }: { current: number; max: number; label?: string } = $props();
	const pct = $derived(max > 0 ? Math.min(100, Math.round((current / max) * 100)) : 0);
</script>

<div class="w-full">
	<div class="mb-1 flex items-baseline justify-between">
		<span class="font-display text-sm font-bold tracking-wide text-success uppercase">{label}</span>
		<span class="text-sm font-bold text-success">{current}</span>
	</div>
	<div class="h-3 overflow-hidden rounded-full bg-secondary">
		<div
			class="h-full rounded-full bg-success transition-[width] duration-500 ease-out"
			style="width: {pct}%; box-shadow: var(--glow-lime);"
		></div>
	</div>
</div>
```

- [ ] **Step 2: GlowCard.svelte**

```svelte
<script lang="ts">
	import type { Snippet } from 'svelte';
	let {
		gradient = 'linear-gradient(135deg, oklch(0.85 0.15 195), oklch(0.70 0.25 330))',
		children
	}: { gradient?: string; children: Snippet } = $props();
</script>

<div class="glow-card rounded-xl p-[2px] transition-transform duration-200" style="background: {gradient};">
	<div class="rounded-[calc(1rem+4px)] bg-card p-5">
		{@render children()}
	</div>
</div>

<style>
	@media (prefers-reduced-motion: no-preference) {
		.glow-card:hover {
			transform: translateY(-3px);
		}
	}
	.glow-card:hover {
		box-shadow: var(--glow-cyan);
	}
</style>
```

- [ ] **Step 3: PressButton.svelte**

```svelte
<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	let {
		variant = 'cyan',
		children,
		...rest
	}: { variant?: 'cyan' | 'magenta' | 'lime'; children: Snippet } & HTMLButtonAttributes = $props();

	const styles = {
		cyan: 'background: oklch(0.85 0.15 195); color: oklch(0.18 0.03 285); --press-glow: var(--glow-cyan); --press-edge: oklch(0.60 0.12 195);',
		magenta: 'background: oklch(0.70 0.25 330); color: oklch(0.98 0.005 285); --press-glow: var(--glow-magenta); --press-edge: oklch(0.48 0.20 330);',
		lime: 'background: oklch(0.85 0.22 135); color: oklch(0.18 0.03 285); --press-glow: var(--glow-lime); --press-edge: oklch(0.60 0.17 135);'
	} as const;
</script>

<button
	{...rest}
	class="press-btn font-display rounded-xl px-8 py-3 text-lg font-bold tracking-wide uppercase disabled:cursor-not-allowed disabled:opacity-50 {rest.class ?? ''}"
	style="{styles[variant]} {rest.style ?? ''}"
>
	{@render children()}
</button>

<style>
	.press-btn {
		box-shadow:
			0 5px 0 var(--press-edge),
			var(--press-glow);
		transition:
			transform 0.15s ease,
			box-shadow 0.15s ease;
	}
	.press-btn:active:not(:disabled) {
		transform: translateY(4px);
		box-shadow:
			0 1px 0 var(--press-edge),
			var(--press-glow);
	}
</style>
```

- [ ] **Step 4: index.ts**

```ts
export { default as XPBar } from './XPBar.svelte';
export { default as GlowCard } from './GlowCard.svelte';
export { default as PressButton } from './PressButton.svelte';
```

- [ ] **Step 5: Topbar** — läs `+layout.svelte`; gör navfältet mörkt glasigt: wrapper-klasser i stil med `sticky top-0 z-50 border-b border-border/50 bg-background/70 backdrop-blur-md`; logotypen/appnamnet i `font-display text-xl font-bold` med cyan färg (`text-primary`); "Byt profil"-knappen behåller sin logik men får `variant="outline"`-stil som matchar mörkret. Ändra INTE navigations-/logoutlogik.

- [ ] **Step 6: Verifiera** — svelte-check + build grönt.

- [ ] **Step 7: Commit** — `git commit -m "feat(design): arcade components (XPBar, GlowCard, PressButton) and glassy topbar"`

---

### Task 3: /las-upp + /profiler

**Files:**
- Modify: `frontend/src/routes/las-upp/+page.svelte`
- Modify: `frontend/src/routes/profiler/+page.svelte`

**Interfaces:**
- Consumes: `PressButton` från Task 2, tokens från Task 1. LOGIKEN (unlock-anrop, redirects, felhantering, 401-hantering) får inte ändras.

- [ ] **Step 1: /las-upp** — läs filen. Ny presentation: helskärmscentrerad mörk scen; "NIOPLUGGET" i `font-display text-5xl font-extrabold text-primary` med pulserande glow:

```css
@keyframes pulse-glow {
	0%, 100% { text-shadow: 0 0 20px oklch(0.85 0.15 195 / 0.5); }
	50% { text-shadow: 0 0 45px oklch(0.85 0.15 195 / 0.9); }
}
```
(applicera bara under `@media (prefers-reduced-motion: no-preference)`). Undertext "Ange familjekoden för att låsa upp den här enheten." i muted. Kodfältet: stort (`h-14 text-center text-2xl tracking-[0.3em]`), mörk input med cyan focus-ring och `box-shadow: var(--glow-cyan)` vid fokus. Knappen ersätts med `<PressButton variant="cyan" type="submit">LÅS UPP</PressButton>`. Felmeddelanden i destructive som idag. `/login`-länken kvar längst ner, diskret muted.

- [ ] **Step 2: /profiler** — läs filen. Rubrik "VEM PLUGGAR?" i `font-display text-4xl font-extrabold uppercase tracking-wider text-primary`. Avatarerna: 96px cirklar (behåll `avatarColor`-logiken för fyllningen) omgivna av en roterande gradient-ring:

```svelte
<div class="avatar-ring rounded-full p-[3px]">
	<!-- befintlig avatar-cirkel inuti, med bg-card-gap: extra p-[3px] mörk ring emellan -->
</div>
```
```css
.avatar-ring {
	background: conic-gradient(from var(--ring-angle, 0deg), oklch(0.85 0.15 195), oklch(0.70 0.25 330), oklch(0.85 0.22 135), oklch(0.85 0.15 195));
}
@property --ring-angle { syntax: '<angle>'; initial-value: 0deg; inherits: false; }
@media (prefers-reduced-motion: no-preference) {
	.avatar-ring { animation: spin-ring 6s linear infinite; }
	@keyframes spin-ring { to { --ring-angle: 360deg; } }
}
```
(Om `@property`-varianten strular i build: fallback = statisk gradientring utan rotation — acceptabelt, notera i rapporten.) Hover/fokus: skala upp 1.08 med `transition-transform duration-200` + glow. Namn under i `font-display font-bold`. Grid centrerad, fungerar på 390px (2 kolumner). Rör INTE selectProfile/redirect-logiken.

- [ ] **Step 3: Verifiera** — svelte-check + build grönt.
- [ ] **Step 4: Commit** — `git commit -m "feat(design): arcade unlock scene and glowing profile picker"`

---

### Task 4: /study + ämnes-/övningsvyer

**Files:**
- Modify: `frontend/src/routes/study/+page.svelte`
- Modify: `frontend/src/routes/study/[subject]/**` (läs strukturen; alla list-/kortvyer där)

**Interfaces:**
- Consumes: `XPBar`, `GlowCard` från Task 2. Data/logik orörd — XP-summan för XPBar hämtas ur data som study-sidan REDAN har (t.ex. stjärnor/XP i utmanings-/progressdata som redan laddas). Finns ingen lämplig summa redan laddad: visa XPBar i utmaningssektionen med den datan, INTE nya API-anrop.

- [ ] **Step 1: /study** — läs filen. Layout uppifrån: (1) hälsning "HEJ {namn}!" i font-display om namnet finns i sidans data, annars "DAGS ATT PLUGGA!"; (2) XPBar (se Consumes); (3) sektion "⚡ DAGENS UPPDRAG" = befintliga repetera-sektionen (samma logik/länkar) som GlowCard med lime-gradient (`linear-gradient(135deg, oklch(0.85 0.22 135), oklch(0.85 0.15 195))`); (4) ämneskorten som GlowCard med djupa mörka gradienter + stor emoji (🧬 biologi, 🏛️ samhälle, ➗ matte — mappa på slug med fallback 📚):
  - biologi: `linear-gradient(135deg, oklch(0.35 0.10 155), oklch(0.25 0.06 200))`
  - samhalle/samhällskunskap: `linear-gradient(135deg, oklch(0.35 0.10 265), oklch(0.25 0.08 310))`
  - matematik/matte: `linear-gradient(135deg, oklch(0.35 0.12 25), oklch(0.28 0.08 330))`
  (gradient-prop på GlowCard används som kant; kortets innehållsyta kan därtill få svag inre gradient via egen klass — håll det läsbart); (5) utmaningssektion + Telegram-sektion behåller funktion, styling till GlowCard/mörka kort.
- [ ] **Step 2: Övningslistor** (`study/[subject]/...`) — läs strukturen (`ls -R src/routes/study`). Övningskort: mörka kort med titel i font-display, stjärnrad (⭐ i `text-gold`), beskrivning muted. Klarade övningar (om vyn redan vet det): grön glow-kant `box-shadow: var(--glow-lime); border-color: var(--color-success)`. Låsta (om lås-logik finns i vyn): sänkt opacity + 🔒. Behåll all befintlig navigering.
- [ ] **Step 3: Verifiera** — svelte-check + build grönt.
- [ ] **Step 4: Commit** — `git commit -m "feat(design): arcade study home and exercise lists"`

---

### Task 5: Chat + utmaningsvyer

**Files:**
- Modify: chatvyn (hitta med `ls src/routes/chat src/routes/study/chat 2>/dev/null` + läs; chatten kan ligga under study/chat)
- Modify: `frontend/src/routes/challenges/**` och `frontend/src/routes/study/challenges/**` (läs strukturen)

**Interfaces:**
- Consumes: GlowCard, XPBar, tokens. SSE-/chatlogik, ?returnTo-, stjärn-/XP-logik och StarAnimation-komponentens beteende får INTE ändras (endast utseende).

- [ ] **Step 1: Chat** — elevens bubblor: `background: linear-gradient(135deg, oklch(0.45 0.10 195), oklch(0.38 0.08 220)); color: oklch(0.97 0.005 285)`, rundade 1rem med en platt hörnradie mot avsändarsidan. AI:ns meddelanden: `bg-card border border-border`. Typing-indikator: tre prickar med studs:

```css
.typing-dot { animation: bounce-dot 1s infinite; }
.typing-dot:nth-child(2) { animation-delay: 0.15s; }
.typing-dot:nth-child(3) { animation-delay: 0.3s; }
@keyframes bounce-dot { 0%, 60%, 100% { transform: translateY(0); } 30% { transform: translateY(-6px); } }
```
(bakom prefers-reduced-motion-guard). Inputfältet mörkt med cyan-ring; skicka-knappen cyan. "Avsluta"-knappen behåller funktion. StarAnimation: om den använder färgkonstanter som försvinner med ljusa temat — uppdatera färgerna till guld/lime/cyan, inte beteendet.
- [ ] **Step 2: Utmaningsvyer** — utmaningslista: GlowCard per utmaning med stor cover-emoji + titel i font-display + progress (befintlig data) som XPBar eller stjärnrad. Detaljvy: emoji-hero överst (text-6xl, centrerad, svag glow), övningslista i samma stil som Task 4 Step 2.
- [ ] **Step 3: Verifiera** — svelte-check + build grönt.
- [ ] **Step 4: Commit** — `git commit -m "feat(design): arcade chat bubbles and challenge views"`

---

### Task 6: /progress + föräldravyer (dashboard, setup, login, register)

**Files:**
- Modify: `frontend/src/routes/progress/+page.svelte`
- Modify: `frontend/src/routes/dashboard/+page.svelte` (+ ev. `dashboard/child/**`)
- Modify: `frontend/src/routes/setup/+page.svelte`, `login/+page.svelte`, `register/+page.svelte`

**Interfaces:**
- Consumes: XPBar, GlowCard, tokens. Föräldravyer = SAMMA tokens/typografi men dämpat: inga glow-shadows, ingen animation utöver hover, mer whitespace.

- [ ] **Step 1: /progress** — stat-tiles överst (3 st: XP/stjärnor/pass — ur befintlig data): stor siffra i `font-display text-4xl font-extrabold` färgad (lime/guld/cyan) med matchande svag glow, etikett i muted under. Ämnesframsteg: befintliga staplar/listor får neonfärger ur chart-tokens; XPBar kan återanvändas för procentvisningar.
- [ ] **Step 2: Dashboard** — korten (barn, API-nyckel, Familjekod, utmaningar) får den nya mörka kortstilen (bg-card, rundning, border) UTAN gradient-kanter/glow; rubriker i font-display. Inga logikändringar (familjekod-/API-nyckelflödena från förra featuren är färska — rör bara klasser/markup).
- [ ] **Step 3: setup/login/register** — mörka centrerade kort, rubrik i font-display, inputs med cyan-ring; knappar standard-shadcn (INTE PressButton — vuxenvyer är lugna).
- [ ] **Step 4: Verifiera** — svelte-check + build grönt. Extra svep: `grep -rn "bg-white\|text-black" src/routes --include='*.svelte'` ska ge 0 träffar.
- [ ] **Step 5: Commit** — `git commit -m "feat(design): arcade progress stats and calm parent views"`

---

### Task 7: Visuell verifiering (controller + användare)

- [ ] Bygg + starta preview lokalt; om Playwright/chromium finns (`npx playwright --version` eller befintlig devDependency): ta screenshots av las-upp, profiler, study, övningslista, chat, challenges, progress, dashboard i 390px + 1280px och granska. Annars: manuell granskning i webbläsare av användaren.
- [ ] Kontrastkontroll: foreground/muted-foreground mot background/card (AA).
- [ ] Slutreview av hela branchen → ev. fixvåg → merge → deploy (bygg + restart tjänster) → användaren tittar live.

---

## Self-review (utförd vid planskrivning)

- Spec-täckning: tokens/fonter/bas (T1), signaturkomponenter+topbar (T2), las-upp/profiler (T3), study/övningar (T4), chat/utmaningar (T5), progress/föräldravyer/rot (T1+T6), verifiering/deploy (T7). Reduced-motion och AA i Global Constraints.
- Medvetna kontrollpunkter (ej placeholders): font-display-utility i Tailwind v4 (T1), @property-fallback (T3), chatvyns exakta sökväg (T5), XP-datakälla begränsad till redan laddad data (T4).
- Typkonsistens: XPBar/GlowCard/PressButton-props samma i T2-definition och T3–T6-användning; CSS-variabler --glow-* definieras i T1 och används i T2–T6.
