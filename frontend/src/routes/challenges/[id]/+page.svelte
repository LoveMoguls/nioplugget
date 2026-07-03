<script lang="ts">
	import { onMount } from 'svelte';
	import { goto, replaceState } from '$app/navigation';
	import { page } from '$app/stores';
	import { Card, CardHeader, CardTitle, CardDescription } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { challenges, sessions } from '$lib/api';
	import { user, isChild } from '$lib/stores/auth';
	import StarAnimation from '$lib/components/challenges/StarAnimation.svelte';

	interface ChallengeExercise {
		id: string;
		title: string;
		description: string;
		displayOrder: number;
		completed?: boolean;
		stars?: number;
		xp?: number;
		sessionId?: string;
	}

	interface Challenge {
		id: string;
		title: string;
		description: string;
		coverEmoji: string;
		exercises: ChallengeExercise[];
	}

	let challenge = $state<Challenge | null>(null);
	let loading = $state(true);
	let starting = $state('');
	let error = $state('');
	let showStars = $state(false);
	let lastStars = $state(0);
	let lastXp = $state(0);

	onMount(async () => {
		await user.checkAuth();
		if (!$isChild) {
			goto('/child/login');
			return;
		}
		const id = ($page.params as Record<string, string>).id;
		try {
			challenge = (await challenges.get(id)) as Challenge;
		} catch {
			error = 'Kunde inte ladda utmaningen';
		} finally {
			loading = false;
		}

		// Coming back from a finished exercise? Celebrate, then clean the URL
		const starsParam = $page.url.searchParams.get('stars');
		const xpParam = $page.url.searchParams.get('xp');
		if (starsParam && xpParam) {
			lastStars = Math.min(3, Math.max(1, parseInt(starsParam) || 1));
			lastXp = parseInt(xpParam) || 0;
			showStars = true;
			replaceState(`/challenges/${id}`, {});
			setTimeout(() => {
				showStars = false;
			}, 3000);
		}
	});

	// An exercise is unlocked if it's the first, or the previous one is completed
	function isUnlocked(index: number): boolean {
		if (!challenge) return false;
		if (index === 0) return true;
		return !!challenge.exercises[index - 1].completed;
	}

	async function startExercise(exerciseId: string) {
		if (!challenge) return;
		starting = exerciseId;
		try {
			const session = (await sessions.createChallenge(exerciseId)) as { id: string };
			goto(`/chat/${session.id}?returnTo=/challenges/${challenge.id}`);
		} catch {
			error = 'Kunde inte starta övningen';
			starting = '';
		}
	}

	function starDisplay(stars: number): string {
		return '★'.repeat(stars) + '☆'.repeat(3 - stars);
	}

	function completedCount(): number {
		return challenge?.exercises.filter((e) => e.completed).length ?? 0;
	}

	function totalXp(): number {
		return challenge?.exercises.reduce((sum, e) => sum + (e.xp ?? 0), 0) ?? 0;
	}
</script>

<svelte:head>
	<title>{challenge?.title || 'Utmaning'} — Nioplugget</title>
</svelte:head>

<StarAnimation stars={lastStars} xp={lastXp} visible={showStars} />

<div class="mx-auto max-w-4xl px-4 py-8">
	<a
		href="/study"
		class="text-muted-foreground mb-4 inline-flex min-h-[44px] items-center text-sm hover:underline"
	>
		&larr; Tillbaka
	</a>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<p class="text-red-500">{error}</p>
	{:else if challenge}
		<div class="mb-6">
			<div class="mb-2 text-4xl">{challenge.coverEmoji}</div>
			<h1 class="mb-1 text-2xl font-bold">{challenge.title}</h1>
			<p class="text-muted-foreground text-sm">{challenge.description}</p>
			<p class="mt-2 text-sm font-medium text-primary">
				{completedCount()} av {challenge.exercises.length} klara · {totalXp()} XP
			</p>
		</div>

		<div class="space-y-4">
			{#each challenge.exercises as exercise, i}
				{@const unlocked = isUnlocked(i)}
				<Card
					class="hover:shadow-md {exercise.completed
						? 'border-emerald-200 bg-emerald-50/40'
						: ''} {!unlocked ? 'opacity-60' : ''}"
				>
					<CardHeader class="flex flex-row items-center justify-between gap-4">
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2">
								<CardTitle class="text-lg">{exercise.title}</CardTitle>
								{#if exercise.completed && exercise.stars}
									<span class="text-sm text-amber-500">{starDisplay(exercise.stars)}</span>
								{/if}
							</div>
							<CardDescription class="mt-1">{exercise.description}</CardDescription>
						</div>
						{#if exercise.completed && exercise.sessionId}
							<Button
								variant="outline"
								onclick={() => goto(`/chat/${exercise.sessionId}?returnTo=/challenges/${challenge!.id}`)}
								class="min-h-[44px] shrink-0"
							>
								Fortsätt
							</Button>
						{:else if unlocked}
							<Button
								onclick={() => startExercise(exercise.id)}
								disabled={starting !== ''}
								class="min-h-[44px] shrink-0"
							>
								{starting === exercise.id ? 'Startar...' : 'Starta'}
							</Button>
						{:else}
							<Button variant="outline" disabled class="min-h-[44px] shrink-0">🔒 Låst</Button>
						{/if}
					</CardHeader>
				</Card>
			{/each}
		</div>
	{/if}
</div>
