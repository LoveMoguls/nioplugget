<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { content, sessions } from '$lib/api';

	interface Exercise {
		id: string;
		title: string;
		description: string;
		difficultyOrder: number;
	}

	interface ExercisesResponse {
		subject: { id: string; name: string; slug: string };
		topic: { id: string; name: string; slug: string };
		exercises: Exercise[];
	}

	let subjectName = $state('');
	let subjectSlug = $state('');
	let topicName = $state('');
	let exercises: Exercise[] = $state([]);
	let loading = $state(true);
	let starting = $state('');
	let error = $state('');

	onMount(async () => {
		const params = $page.params as Record<string, string>;
		subjectSlug = params.subject;
		const topicSlug = params.topic;
		try {
			const data = (await content.exercises(subjectSlug, topicSlug)) as ExercisesResponse;
			subjectName = data.subject.name;
			topicName = data.topic.name;
			exercises = data.exercises;
		} catch {
			error = 'Kunde inte ladda övningar';
		} finally {
			loading = false;
		}
	});

	async function startExercise(exerciseId: string) {
		starting = exerciseId;
		try {
			const session = (await sessions.create(exerciseId)) as { id: string };
			goto(`/chat/${session.id}`);
		} catch {
			error = 'Kunde inte starta övningen';
			starting = '';
		}
	}

	function difficultyStars(level: number): string {
		return '⭐'.repeat(level);
	}
</script>

<svelte:head>
	<title>{topicName || 'Övningar'} — Nioplugget</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8">
	<a
		href="/study/{subjectSlug}"
		class="text-muted-foreground mb-4 inline-flex min-h-[44px] items-center text-sm hover:underline"
	>
		&larr; Tillbaka till {subjectName || 'områden'}
	</a>

	<h1 class="font-display mb-2 text-2xl font-bold">{topicName || 'Välj övning'}</h1>
	<p class="text-muted-foreground mb-6 text-sm">{subjectName}</p>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<p class="text-destructive">{error}</p>
	{:else}
		<div class="space-y-4">
			{#each exercises as exercise}
				<div
					class="flex flex-row items-center justify-between gap-4 rounded-xl border border-border bg-card p-5"
				>
					<div class="min-w-0 flex-1">
						<p class="font-display text-lg font-bold text-foreground">{exercise.title}</p>
						<p class="mt-1 text-sm text-muted-foreground">{exercise.description}</p>
						<p class="mt-2 text-xs tracking-wider text-gold">
							{difficultyStars(exercise.difficultyOrder)}
						</p>
					</div>
					<Button
						onclick={() => startExercise(exercise.id)}
						disabled={starting !== ''}
						class="min-h-[44px] shrink-0"
					>
						{starting === exercise.id ? 'Startar...' : 'Starta övning'}
					</Button>
				</div>
			{/each}
		</div>
	{/if}
</div>
