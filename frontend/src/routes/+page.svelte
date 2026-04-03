<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card';
	import { isLoggedIn, isParent } from '$lib/stores/auth';
	import { MessageCircleQuestion, CalendarClock, BookOpen, ShieldCheck, ChevronDown } from '@lucide/svelte';

	// Redirect logged-in users
	onMount(() => {
		if ($isLoggedIn) {
			if ($isParent) {
				goto('/dashboard');
			} else {
				goto('/study');
			}
		}
	});

	// FAQ toggle state
	let openFaq = $state<number | null>(null);

	function toggleFaq(index: number) {
		openFaq = openFaq === index ? null : index;
	}

	const faqs = [
		{
			q: 'Vad kostar det?',
			a: 'Nioplugget är gratis att använda. Du betalar bara för din Claude API-användning direkt till Anthropic. En typisk studiesession kostar några ören.'
		},
		{
			q: 'Vad är en API-nyckel?',
			a: 'En API-nyckel är ditt personliga lösenord till Claudes AI-tjänst. Du skapar en gratis på console.anthropic.com och kopplar den till ditt Nioplugget-konto.'
		},
		{
			q: 'Är det säkert?',
			a: 'Ja. Din API-nyckel krypteras med AES-256-GCM och lagras säkert. Vi loggar aldrig känslig data och delar ingenting med tredje part.'
		},
		{
			q: 'Vilka ämnen finns?',
			a: 'Biologi, Samhällskunskap och Matematik — alla anpassade efter Skolverkets centrala innehåll för åk 9. Varje ämne har fyra områden med 3-5 övningar.'
		},
		{
			q: 'Ger AI:n direkta svar?',
			a: 'Nej. AI-läraren använder den sokratiska metoden och ställer ledande frågor så att eleven lär sig genom eget tänkande. Det ger djupare förståelse.'
		}
	];
</script>

<svelte:head>
	<title>Nioplugget — Plugga smartare inför nationella provet</title>
	<meta name="description" content="AI-lärare som guidar med frågor, aldrig ger direkta svar. Övningar anpassade efter Skolverkets centrala innehåll för åk 9." />
</svelte:head>

{#if !$isLoggedIn}
<!-- Hero Section -->
<section class="px-4 py-16 sm:py-24">
	<div class="mx-auto max-w-3xl text-center">
		<h1 class="mb-4 text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
			Plugga smartare inför nationella provet
		</h1>
		<p class="mb-8 text-lg text-muted-foreground sm:text-xl">
			En AI-lärare som guidar med frågor — aldrig ger direkta svar. Anpassade övningar efter Skolverkets centrala innehåll för åk 9.
		</p>
		<div class="flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
			<Button href="/register" class="min-h-[44px] w-full px-8 text-base sm:w-auto">
				Kom igång gratis
			</Button>
			<a
				href="/login"
				class="flex min-h-[44px] items-center text-sm text-muted-foreground hover:text-foreground"
			>
				Redan medlem? Logga in
			</a>
		</div>
	</div>
</section>

<!-- Features Section -->
<section class="bg-muted/30 px-4 py-16">
	<div class="mx-auto max-w-4xl">
		<h2 class="mb-10 text-center text-2xl font-bold text-foreground sm:text-3xl">
			Byggt för lärande, inte för poäng
		</h2>
		<div class="grid gap-6 sm:grid-cols-3">
			<Card class="border-0 bg-card shadow-sm">
				<CardHeader class="items-center text-center">
					<div class="mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10">
						<MessageCircleQuestion class="h-6 w-6 text-primary" />
					</div>
					<CardTitle class="text-lg">Sokratisk dialog</CardTitle>
					<CardDescription>
						AI:n ställer frågor istället för att ge svar. Eleven lär sig genom att tänka själv.
					</CardDescription>
				</CardHeader>
			</Card>

			<Card class="border-0 bg-card shadow-sm">
				<CardHeader class="items-center text-center">
					<div class="mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10">
						<CalendarClock class="h-6 w-6 text-primary" />
					</div>
					<CardTitle class="text-lg">Anpassad repetition</CardTitle>
					<CardDescription>
						SM-2-algoritmen beräknar när det är dags att repetera. Ingen övning glöms bort.
					</CardDescription>
				</CardHeader>
			</Card>

			<Card class="border-0 bg-card shadow-sm">
				<CardHeader class="items-center text-center">
					<div class="mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10">
						<BookOpen class="h-6 w-6 text-primary" />
					</div>
					<CardTitle class="text-lg">Tre ämnen, tolv områden</CardTitle>
					<CardDescription>
						Biologi, Samhällskunskap och Matematik med övningar anpassade efter Skolverkets centrala innehåll.
					</CardDescription>
				</CardHeader>
			</Card>
		</div>
	</div>
</section>

<!-- How It Works Section -->
<section class="px-4 py-16">
	<div class="mx-auto max-w-3xl">
		<h2 class="mb-10 text-center text-2xl font-bold text-foreground sm:text-3xl">
			Så funkar det
		</h2>
		<div class="space-y-8">
			<div class="flex gap-4">
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary text-lg font-bold text-primary-foreground">
					1
				</div>
				<div>
					<h3 class="mb-1 text-lg font-semibold text-foreground">Skapa konto</h3>
					<p class="text-muted-foreground">
						Registrera dig som förälder och koppla din Claude API-nyckel. Det tar bara en minut.
					</p>
				</div>
			</div>

			<div class="flex gap-4">
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary text-lg font-bold text-primary-foreground">
					2
				</div>
				<div>
					<h3 class="mb-1 text-lg font-semibold text-foreground">Bjud in ditt barn</h3>
					<p class="text-muted-foreground">
						Skapa en barnprofil och skicka en inbjudningslänk. Barnet väljer en egen PIN-kod.
					</p>
				</div>
			</div>

			<div class="flex gap-4">
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary text-lg font-bold text-primary-foreground">
					3
				</div>
				<div>
					<h3 class="mb-1 text-lg font-semibold text-foreground">Barnet börjar plugga</h3>
					<p class="text-muted-foreground">
						Ditt barn loggar in med namn och PIN-kod och börjar öva direkt med AI-läraren.
					</p>
				</div>
			</div>
		</div>
	</div>
</section>

<!-- BYOK Section -->
<section class="bg-muted/30 px-4 py-16">
	<div class="mx-auto max-w-3xl text-center">
		<div class="mb-4 flex justify-center">
			<div class="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10">
				<ShieldCheck class="h-6 w-6 text-primary" />
			</div>
		</div>
		<h2 class="mb-4 text-2xl font-bold text-foreground sm:text-3xl">
			Du använder din egen Claude API-nyckel
		</h2>
		<p class="mb-4 text-muted-foreground">
			Nioplugget kostar inget — du betalar bara Anthropic direkt för din AI-användning. En typisk studiesession kostar några ören.
		</p>
		<p class="mb-6 text-sm text-muted-foreground">
			Din nyckel krypteras med AES-256-GCM och lagras säkert. Vi har aldrig tillgång till din nyckel i klartext.
		</p>
		<a
			href="https://console.anthropic.com"
			target="_blank"
			rel="noopener noreferrer"
			class="inline-flex min-h-[44px] items-center text-sm text-primary hover:underline"
		>
			Skapa en API-nyckel på console.anthropic.com &rarr;
		</a>
	</div>
</section>

<!-- FAQ Section -->
<section class="px-4 py-16">
	<div class="mx-auto max-w-3xl">
		<h2 class="mb-10 text-center text-2xl font-bold text-foreground sm:text-3xl">
			Vanliga frågor
		</h2>
		<div class="space-y-2">
			{#each faqs as faq, i}
				<button
					onclick={() => toggleFaq(i)}
					class="flex w-full min-h-[44px] items-center justify-between rounded-lg border border-border bg-card px-4 py-3 text-left text-sm font-medium text-foreground hover:bg-muted/50"
					aria-expanded={openFaq === i}
				>
					{faq.q}
					<ChevronDown class="h-4 w-4 shrink-0 text-muted-foreground transition-transform {openFaq === i ? 'rotate-180' : ''}" />
				</button>
				{#if openFaq === i}
					<div class="rounded-lg border border-border bg-card px-4 py-3 text-sm text-muted-foreground">
						{faq.a}
					</div>
				{/if}
			{/each}
		</div>
	</div>
</section>

<!-- Footer CTA Section -->
<section class="bg-muted/30 px-4 py-16">
	<div class="mx-auto max-w-3xl text-center">
		<h2 class="mb-4 text-2xl font-bold text-foreground">
			Redo att börja?
		</h2>
		<p class="mb-6 text-muted-foreground">
			Skapa ett gratiskonto och låt ditt barn plugga smartare.
		</p>
		<div class="flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
			<Button href="/register" class="min-h-[44px] w-full px-8 text-base sm:w-auto">
				Kom igång gratis
			</Button>
		</div>
		<p class="mt-6 text-sm text-muted-foreground">
			Elev?
			<a href="/child/login" class="text-primary hover:underline">Logga in här</a>
		</p>
	</div>
</section>
{/if}
