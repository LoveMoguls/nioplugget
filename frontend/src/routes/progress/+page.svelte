<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card';
	import { progress, getErrorMessage } from '$lib/api';
	import { user, isLoggedIn, isChild } from '$lib/stores/auth';

	interface TopicProgress {
		id: string;
		name: string;
		slug: string;
		totalSessions: number;
		avgScore: number;
		uniqueExercises: number;
	}

	interface SubjectProgress {
		id: string;
		name: string;
		slug: string;
		totalSessions: number;
		avgScore: number;
		uniqueExercises: number;
		topics: TopicProgress[];
	}

	interface ProgressData {
		subjects: SubjectProgress[];
	}

	let data = $state<ProgressData | null>(null);
	let loading = $state(true);
	let error = $state('');

	// Derived: strengths and weaknesses
	let strengths = $derived.by(() => {
		if (!data) return [];
		return data.subjects
			.flatMap((s) =>
				s.topics
					.filter((t) => t.avgScore >= 4)
					.map((t) => ({ ...t, subjectName: s.name }))
			)
			.sort((a, b) => b.avgScore - a.avgScore);
	});

	let weaknesses = $derived.by(() => {
		if (!data) return [];
		return data.subjects
			.flatMap((s) =>
				s.topics
					.filter((t) => t.avgScore > 0 && t.avgScore < 3)
					.map((t) => ({ ...t, subjectName: s.name }))
			)
			.sort((a, b) => a.avgScore - b.avgScore);
	});

	let hasAnySessions = $derived(
		data ? data.subjects.some((s) => s.totalSessions > 0) : false
	);

	function scoreColor(score: number): string {
		if (score === 0) return 'bg-gray-200';
		if (score >= 4) return 'bg-emerald-200';
		if (score >= 3) return 'bg-amber-200';
		return 'bg-rose-200';
	}

	onMount(async () => {
		await user.checkAuth();
		if (!$isLoggedIn || !$isChild) {
			goto('/child');
			return;
		}

		try {
			data = (await progress.mine()) as ProgressData;
		} catch (err) {
			error = getErrorMessage(err, 'Kunde inte hämta progress');
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Min progress — Nioplugget</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8">
	<div class="mb-6 flex items-center justify-between">
		<h1 class="text-2xl font-bold text-foreground">Min progress</h1>
		<a
			href="/study"
			class="text-sm text-muted-foreground transition-colors hover:text-foreground"
		>
			Tillbaka till ämnen
		</a>
	</div>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<p class="text-red-500">{error}</p>
	{:else if !hasAnySessions}
		<Card>
			<CardContent class="py-8 text-center">
				<p class="text-muted-foreground">
					Inga genomförda pass ännu. Börja studera för att se din progress!
				</p>
				<a
					href="/study"
					class="mt-4 inline-block text-sm text-foreground underline-offset-4 hover:underline"
				>
					Börja studera →
				</a>
			</CardContent>
		</Card>
	{:else if data}
		<!-- Per-subject cards -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each data.subjects as subject (subject.id)}
				<Card class="h-full">
					<CardHeader class="pb-2">
						<CardTitle class="text-lg">{subject.name}</CardTitle>
						<CardDescription>
							{subject.totalSessions} pass genomförda · Snittbetyg: {subject.avgScore > 0
								? subject.avgScore.toFixed(1)
								: '–'}
						</CardDescription>
					</CardHeader>
					<CardContent>
						{#if subject.topics.length > 0}
							<div class="space-y-2">
								{#each subject.topics as topic (topic.id)}
									<div class="flex items-center gap-2">
										<span
											class="w-24 truncate text-sm text-muted-foreground"
											title={topic.name}
										>
											{topic.name}
										</span>
										<div class="h-4 flex-1 overflow-hidden rounded-full bg-gray-100">
											{#if topic.avgScore > 0}
												<div
													class="{scoreColor(topic.avgScore)} h-full rounded-full transition-all"
													style="width: {(topic.avgScore / 5) * 100}%"
												></div>
											{/if}
										</div>
										<span class="w-8 text-right text-sm text-muted-foreground">
											{topic.avgScore > 0 ? topic.avgScore.toFixed(1) : '–'}
										</span>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-sm text-muted-foreground">Inga ämnesområden</p>
						{/if}
					</CardContent>
				</Card>
			{/each}
		</div>

		<!-- Strengths and weaknesses -->
		{#if strengths.length > 0 || weaknesses.length > 0}
			<div class="mt-8">
				<h2 class="mb-4 text-xl font-semibold text-foreground">Styrkor och svagheter</h2>

				{#if strengths.length > 0}
					<div class="mb-4">
						<h3 class="mb-2 text-sm font-medium text-muted-foreground">Starka områden</h3>
						<div class="space-y-1">
							{#each strengths as topic}
								<p class="text-sm text-foreground">
									<span class="text-emerald-600">✓</span>
									{topic.name} ({topic.avgScore.toFixed(1)}) — {topic.subjectName}
								</p>
							{/each}
						</div>
					</div>
				{/if}

				{#if weaknesses.length > 0}
					<div>
						<h3 class="mb-2 text-sm font-medium text-muted-foreground">Behöver övning</h3>
						<div class="space-y-1">
							{#each weaknesses as topic}
								<p class="text-sm text-foreground">
									<span class="text-rose-500">○</span>
									{topic.name} ({topic.avgScore.toFixed(1)}) — {topic.subjectName}
								</p>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{:else}
			<div class="mt-8">
				<p class="text-sm text-muted-foreground">
					Genomför fler pass för att se styrkor och svagheter.
				</p>
			</div>
		{/if}
	{/if}
</div>
