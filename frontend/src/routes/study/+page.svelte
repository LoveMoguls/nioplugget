<script lang="ts">
	import { onMount } from 'svelte';
	import { Card, CardHeader, CardTitle, CardDescription } from '$lib/components/ui/card';
	import { content } from '$lib/api';

	interface Subject {
		id: string;
		name: string;
		slug: string;
		displayOrder: number;
	}

	let subjects: Subject[] = $state([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			subjects = (await content.subjects()) as Subject[];
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
	<h1 class="mb-6 text-2xl font-bold">Välj ämne</h1>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<p class="text-red-500">{error}</p>
	{:else}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each subjects as subject}
				<a href="/study/{subject.slug}" class="block transition-transform hover:scale-[1.02]">
					<Card class="h-full cursor-pointer hover:shadow-md">
						<CardHeader>
							<CardTitle>{subject.name}</CardTitle>
							<CardDescription>Utforska övningar i {subject.name.toLowerCase()}</CardDescription>
						</CardHeader>
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</div>
