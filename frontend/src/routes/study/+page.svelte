<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Card, CardHeader, CardTitle, CardDescription } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { content, reviews, challenges, telegram, type ApiError } from '$lib/api';
	import { user, isChild } from '$lib/stores/auth';
	import ChallengeUpload from '$lib/components/challenges/ChallengeUpload.svelte';
	import { GlowCard } from '$lib/components/arcade';

	const LIME_GRADIENT = 'linear-gradient(135deg, oklch(0.85 0.22 135), oklch(0.85 0.15 195))';
	const GOLD_GRADIENT = 'linear-gradient(135deg, oklch(0.85 0.17 90), oklch(0.85 0.15 195))';

	const SUBJECT_GRADIENTS: Record<string, string> = {
		biologi: 'linear-gradient(135deg, oklch(0.35 0.10 155), oklch(0.25 0.06 200))',
		samhalle: 'linear-gradient(135deg, oklch(0.35 0.10 265), oklch(0.25 0.08 310))',
		samhallskunskap: 'linear-gradient(135deg, oklch(0.35 0.10 265), oklch(0.25 0.08 310))',
		matematik: 'linear-gradient(135deg, oklch(0.35 0.12 25), oklch(0.28 0.08 330))',
		matte: 'linear-gradient(135deg, oklch(0.35 0.12 25), oklch(0.28 0.08 330))'
	};
	const SUBJECT_FALLBACK_GRADIENT = 'linear-gradient(135deg, oklch(0.30 0.03 285), oklch(0.22 0.03 285))';

	const SUBJECT_EMOJI: Record<string, string> = {
		biologi: '🧬',
		samhalle: '🏛️',
		samhallskunskap: '🏛️',
		matematik: '➗',
		matte: '➗'
	};
	const SUBJECT_FALLBACK_EMOJI = '📚';

	function subjectGradient(slug: string): string {
		return SUBJECT_GRADIENTS[slug] ?? SUBJECT_FALLBACK_GRADIENT;
	}

	function subjectEmoji(slug: string): string {
		return SUBJECT_EMOJI[slug] ?? SUBJECT_FALLBACK_EMOJI;
	}

	const childName = $derived(($user as { name?: string } | null)?.name);

	interface Subject {
		id: string;
		name: string;
		slug: string;
		displayOrder: number;
	}

	interface Challenge {
		id: string;
		title: string;
		description: string;
		coverEmoji: string;
		createdBy: string;
		createdAt: string;
	}

	function formatCreated(iso: string): string {
		const d = new Date(iso);
		if (isNaN(d.getTime())) return '';
		return d.toLocaleDateString('sv-SE', { day: 'numeric', month: 'short' });
	}

	interface DueReview {
		id: string;
		exerciseId: string;
		exerciseTitle: string;
		topicName: string;
		subjectName: string;
		subjectSlug: string;
		topicSlug: string;
		daysOverdue: number;
	}

	let subjects: Subject[] = $state([]);
	let dueReviews: DueReview[] = $state([]);
	let challengeList: Challenge[] = $state([]);
	let loading = $state(true);
	let error = $state('');
	let showCreate = $state(false);
	let telegramLink: string | null = $state(null);
	let telegramError = $state('');
	let telegramAvailable = $state(true);

	async function connectTelegram() {
		telegramError = '';
		try {
			const res = await telegram.createLinkCode();
			telegramLink = res.link;
		} catch (err) {
			if ((err as ApiError)?.status === 404) {
				telegramAvailable = false;
			} else {
				telegramError = 'Kunde inte skapa kopplingslänk. Försök igen.';
			}
		}
	}

	onMount(async () => {
		await user.checkAuth();
		if (!$isChild) {
			goto('/profiler');
			return;
		}
		try {
			const [subjectsData, reviewsData, challengesData] = await Promise.all([
				content.subjects(),
				reviews.due().catch(() => []),
				challenges.list().catch(() => [])
			]);
			subjects = subjectsData as Subject[];
			dueReviews = (reviewsData as DueReview[]) || [];
			challengeList = (challengesData as Challenge[]) || [];
		} catch {
			error = 'Kunde inte ladda ämnen';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Välj ämne — Nioplugget</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8">
	<div class="mb-6 flex items-center justify-between">
		<h1 class="font-display text-2xl font-bold text-foreground sm:text-3xl">
			{childName ? `HEJ ${childName.toUpperCase()}!` : 'DAGS ATT PLUGGA!'}
		</h1>
		<a
			href="/progress"
			class="flex min-h-[44px] items-center text-sm text-muted-foreground transition-colors hover:text-foreground"
		>
			Min progress
		</a>
	</div>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<p class="text-destructive">{error}</p>
	{:else}
		{#if dueReviews.length > 0}
			<div class="mb-8">
				<GlowCard gradient={LIME_GRADIENT}>
					<h2 class="font-display mb-4 text-lg font-bold tracking-wide text-foreground uppercase">
						⚡ Dagens uppdrag
					</h2>
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
						{#each dueReviews as review}
							<a
								href="/study/{review.subjectSlug}/{review.topicSlug}?exercise={review.exerciseId}"
								class="block"
							>
								<div
									class="h-full rounded-lg border border-border bg-secondary/40 p-4 transition-colors hover:bg-secondary/70"
								>
									<p class="font-display font-bold text-foreground">{review.exerciseTitle}</p>
									<p class="mt-1 text-sm text-muted-foreground">
										{review.subjectName} · {review.topicName}
									</p>
									<span class="mt-2 block text-sm font-semibold text-success">
										{review.daysOverdue === 0 ? 'Dags idag' : `${review.daysOverdue} dagar sedan`}
									</span>
								</div>
							</a>
						{/each}
					</div>
				</GlowCard>
			</div>
		{/if}

		<div class="mb-10 mx-auto max-w-xl text-center">
		<p class="text-lg font-semibold text-foreground mb-4">Till Isabella och Vanessa</p>
		<div class="italic text-muted-foreground leading-relaxed space-y-4 text-sm">
			<p>Det här läsåret har ni ett gäng nationella prov framför er. Det kan kännas som mycket, och det är det ibland. Men vi vill att ni ska veta en sak innan ni ens sätter er vid det första provet:</p>
			<p class="text-foreground font-medium not-italic">Vi är redan stolta.</p>
			<p>Det vi firar är inte ett visst betyg. Det vi firar är att ni förbereder er, dyker upp och gör ert bästa. Det är allt som krävs. Och varje gång ni gör det — varje prov, ett i taget — så firar vi det tillsammans.</p>
			<p>Det kan vara ett par Adidas ni länge velat ha. En middag hemma med kompisarna där vi fixar maten och ni fixar stämningen. En hel vecka där ni bestämmer vad vi äter varje kväll, utan ett enda veto. En biofilm, en escape room, en shoppingrunda. Ni väljer.</p>
			<p class="text-foreground font-medium not-italic">Det är vår deal.</p>
			<p>Inte för att proven är ett hinder att ta sig igenom. Utan för att vi vill att ni ska minnas det här läsåret som ett där ni faktiskt gav det en chans — och fick något tillbaka för det.</p>
			<p>Ett prov i taget. Vi är med hela vägen.</p>
			<p class="not-italic text-muted-foreground">Med kärlek ❤️</p>
		</div>
	</div>

	<div class="mb-8 rounded-xl border border-border bg-card p-5">
		<div class="mb-4 flex flex-wrap items-center gap-3">
			<h2 class="font-display text-lg font-bold text-foreground">Utmaningar</h2>
			<span
				class="font-display inline-flex items-center rounded-full border border-gold/40 bg-gold/10 px-3 py-1 text-sm font-bold text-gold"
			>
				🏆 {challengeList.length} {challengeList.length === 1 ? 'utmaning' : 'utmaningar'}
			</span>
		</div>
		<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each challengeList as challenge}
				<a href="/challenges/{challenge.id}" class="block transition-transform hover:scale-[1.02]">
					<GlowCard gradient={GOLD_GRADIENT}>
						<div class="mb-1 text-5xl">{challenge.coverEmoji}</div>
						<p class="font-display text-lg font-bold text-foreground">{challenge.title}</p>
						<p class="mt-1 text-sm text-muted-foreground">{challenge.description}</p>
						<p class="mt-2 text-xs text-muted-foreground/80">
							{challenge.createdBy} · {formatCreated(challenge.createdAt)}
						</p>
					</GlowCard>
				</a>
			{/each}

			<!-- Create-your-own card -->
			{#if !showCreate}
				<button
					onclick={() => (showCreate = true)}
					class="flex h-full min-h-[120px] flex-col items-center justify-center rounded-xl border-2 border-dashed border-border p-4 text-center transition-colors hover:border-primary/60 hover:bg-primary/5"
				>
					<span class="text-3xl">➕</span>
					<span class="mt-1 text-sm font-medium text-foreground">Skapa egen utmaning</span>
					<span class="mt-0.5 text-xs text-muted-foreground">Fota eller klistra in vilken läxa som helst</span>
				</button>
			{/if}
		</div>

		{#if showCreate}
			<Card class="mt-3">
				<CardHeader class="pb-2">
					<div class="flex items-center justify-between">
						<CardTitle class="text-base">Skapa egen utmaning</CardTitle>
						<button
							onclick={() => (showCreate = false)}
							class="rounded-lg p-2 text-muted-foreground hover:bg-muted hover:text-foreground"
							aria-label="Stäng"
						>✕</button>
					</div>
					<CardDescription>Funkar med alla ämnen — inte bara nationella prov!</CardDescription>
				</CardHeader>
				<div class="px-6 pb-6">
					<ChallengeUpload onCreated={(c) => goto(`/challenges/${c.id}`)} />
				</div>
			</Card>
		{/if}
	</div>

	<h2 class="font-display mb-4 text-lg font-bold text-foreground">Ämnen</h2>
	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each subjects as subject}
				<a href="/study/{subject.slug}" class="block transition-transform hover:scale-[1.02]">
					<GlowCard gradient={subjectGradient(subject.slug)}>
						<div class="relative -m-5 overflow-hidden rounded-[calc(1rem+4px)] p-5">
							<div
								class="pointer-events-none absolute inset-0 opacity-30"
								style="background: {subjectGradient(subject.slug)};"
							></div>
							<div class="relative z-10">
								<div class="text-3xl">{subjectEmoji(subject.slug)}</div>
								<p class="font-display mt-2 text-lg font-bold text-foreground">{subject.name}</p>
								<p class="mt-1 text-sm text-foreground/70">
									Utforska övningar i {subject.name.toLowerCase()}
								</p>
							</div>
						</div>
					</GlowCard>
				</a>
			{/each}
		</div>

		{#if telegramAvailable}
			<section class="mt-8 rounded-xl border border-border bg-card p-5">
				{#if telegramLink}
					<p>
						<a href={telegramLink} target="_blank" rel="noopener" class="underline">
							📲 Öppna Telegram och tryck Start →
						</a>
					</p>
					<p class="text-sm text-muted-foreground">Länken gäller i 15 minuter.</p>
				{:else}
					<Button type="button" variant="outline" onclick={connectTelegram}>Koppla Telegram</Button>
					{#if telegramError}<p class="text-sm text-destructive">{telegramError}</p>{/if}
				{/if}
			</section>
		{/if}
	{/if}
</div>
