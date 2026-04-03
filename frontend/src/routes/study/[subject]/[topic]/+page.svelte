<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Card, CardHeader, CardTitle, CardDescription } from '$lib/components/ui/card';
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

	function difficultyDots(level: number): string {
		return '●'.repeat(level) + '○'.repeat(5 - level);
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

	<h1 class="mb-2 text-2xl font-bold">{topicName || 'Välj övning'}</h1>
	<p class="text-muted-foreground mb-6 text-sm">{subjectName}</p>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<p class="text-red-500">{error}</p>
	{:else}
		<div class="space-y-4">
			{#each exercises as exercise}
				<Card class="hover:shadow-md">
					<CardHeader class="flex flex-row items-center justify-between gap-4">
						<div class="min-w-0 flex-1">
							<CardTitle class="text-lg">{exercise.title}</CardTitle>
							<CardDescription class="mt-1">{exercise.description}</CardDescription>
							<p class="text-muted-foreground mt-2 text-xs tracking-wider">
								{difficultyDots(exercise.difficultyOrder)}
							</p>
						</div>
						<Button
							onclick={() => startExercise(exercise.id)}
							disabled={starting !== ''}
							class="min-h-[44px] shrink-0"
						>
							{starting === exercise.id ? 'Startar...' : 'Starta övning'}
						</Button>
					</CardHeader>
				</Card>
			{/each}
		</div>
	{/if}
</div>
