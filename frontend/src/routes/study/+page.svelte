<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Card, CardHeader, CardTitle, CardDescription } from '$lib/components/ui/card';
	import { content, reviews, challenges } from '$lib/api';
	import { user, isChild } from '$lib/stores/auth';

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
		createdAt: string;
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

	onMount(async () => {
		await user.checkAuth();
		if (!$isChild) {
			goto('/child/login');
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
		<h1 class="text-2xl font-bold">Välj ämne</h1>
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
							<Card class="border-accent bg-accent/30 transition-shadow hover:shadow-md">
								<CardHeader class="pb-3">
									<CardTitle class="text-base">{review.exerciseTitle}</CardTitle>
									<CardDescription>
										{review.subjectName} · {review.topicName}
										<span class="mt-1 block font-medium text-primary">
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

	{#if challengeList.length > 0}
		<div class="mb-8">
			<h2 class="mb-4 text-xl font-semibold">Utmaningar</h2>
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
				{#each challengeList as challenge}
					<a href="/challenges/{challenge.id}" class="block transition-transform hover:scale-[1.02]">
						<Card class="h-full hover:shadow-md">
							<CardHeader class="pb-3">
								<div class="mb-1 text-2xl">{challenge.coverEmoji}</div>
								<CardTitle class="text-base">{challenge.title}</CardTitle>
								<CardDescription>{challenge.description}</CardDescription>
							</CardHeader>
						</Card>
					</a>
				{/each}
			</div>
		</div>
	{/if}

	<h2 class="mb-4 text-xl font-semibold">Ämnen</h2>
	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each subjects as subject, i}
				{@const gradients = ['from-violet-500 to-indigo-600','from-emerald-500 to-teal-600','from-orange-400 to-rose-500','from-sky-500 to-blue-600','from-pink-500 to-fuchsia-600']}
				<a href="/study/{subject.slug}" class="block transition-transform hover:scale-[1.02]">
					<div class="h-full cursor-pointer rounded-xl bg-gradient-to-br {gradients[i % gradients.length]} p-5 text-white shadow-md hover:shadow-lg">
						<p class="text-lg font-semibold">{subject.name}</p>
						<p class="mt-1 text-sm text-white/75">Utforska övningar i {subject.name.toLowerCase()}</p>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
