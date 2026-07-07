<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
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
	<a href="/study" class="text-muted-foreground mb-4 inline-flex min-h-[44px] items-center text-sm hover:underline">
		&larr; Tillbaka till ämnen
	</a>

	<h1 class="font-display mb-6 text-2xl font-bold">{subjectName || 'Välj område'}</h1>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<p class="text-destructive">{error}</p>
	{:else}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
			{#each topics as topic}
				<a
					href="/study/{subjectSlug}/{topic.slug}"
					class="block transition-transform hover:scale-[1.02]"
				>
					<div class="h-full rounded-xl border border-border bg-card p-5 transition-colors hover:bg-secondary/40">
						<p class="font-display text-lg font-bold text-foreground">{topic.name}</p>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
