<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Card, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { content } from '$lib/api';

	interface Topic {
		id: string;
		name: string;
		slug: string;
		displayOrder: number;
	}

	interface TopicsResponse {
		subject: { id: string; name: string; slug: string };
		topics: Topic[];
	}

	let subjectName = $state('');
	let subjectSlug = $state('');
	let topics: Topic[] = $state([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		subjectSlug = ($page.params as Record<string, string>).subject;
		try {
			const data = (await content.topics(subjectSlug)) as TopicsResponse;
			subjectName = data.subject.name;
			topics = data.topics;
		} catch {
			error = 'Kunde inte ladda områden';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>{subjectName || 'Områden'} — Nioplugget</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8">
	<a href="/study" class="text-muted-foreground mb-4 inline-block text-sm hover:underline">
		&larr; Tillbaka till ämnen
	</a>

	<h1 class="mb-6 text-2xl font-bold">{subjectName || 'Välj område'}</h1>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<p class="text-red-500">{error}</p>
	{:else}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
			{#each topics as topic}
				<a
					href="/study/{subjectSlug}/{topic.slug}"
					class="block transition-transform hover:scale-[1.02]"
				>
					<Card class="h-full cursor-pointer hover:shadow-md">
						<CardHeader>
							<CardTitle>{topic.name}</CardTitle>
						</CardHeader>
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</div>
