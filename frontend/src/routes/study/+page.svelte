<script lang="ts">
	import { onMount } from 'svelte';
	import { Card, CardHeader, CardTitle, CardDescription } from '$lib/components/ui/card';
	import { content, reviews } from '$lib/api';

	interface Subject {
		id: string;
		name: string;
		slug: string;
		displayOrder: number;
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
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			const [subjectsData, reviewsData] = await Promise.all([
				content.subjects(),
				reviews.due().catch(() => [])
			]);
			subjects = subjectsData as Subject[];
			dueReviews = (reviewsData as DueReview[]) || [];
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
		<h1 class="text-2xl font-bold">Välj ämne</h1>
		<a
			href="/progress"
			class="text-sm text-muted-foreground transition-colors hover:text-foreground"
		>
			Min progress
		</a>
	</div>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<p class="text-red-500">{error}</p>
	{:else}
		{#if dueReviews.length > 0}
			<div class="mb-8">
				<h2 class="mb-4 text-xl font-semibold">Dags att repetera</h2>
				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
					{#each dueReviews as review}
						<a
							href="/study/{review.subjectSlug}/{review.topicSlug}?exercise={review.exerciseId}"
							class="block"
						>
							<Card class="border-amber-200 bg-amber-50 transition-shadow hover:shadow-md">
								<CardHeader class="pb-3">
									<CardTitle class="text-base">{review.exerciseTitle}</CardTitle>
									<CardDescription>
										{review.subjectName} · {review.topicName}
										<span class="mt-1 block text-amber-700">
											{review.daysOverdue === 0 ? 'Dags idag' : `${review.daysOverdue} dagar sedan`}
										</span>
									</CardDescription>
								</CardHeader>
							</Card>
						</a>
					{/each}
				</div>
			</div>
		{/if}

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
