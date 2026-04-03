<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Card, CardHeader, CardTitle, CardContent } from '$lib/components/ui/card';
	import { progress as progressApi, getErrorMessage } from '$lib/api';
	import { user, isLoggedIn, isParent } from '$lib/stores/auth';

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

	interface SessionEntry {
		id: string;
		score: number;
		startedAt: string;
		endedAt: string;
		exerciseTitle: string;
		topicName: string;
		subjectName: string;
	}

	interface SessionsData {
		sessions: SessionEntry[];
	}

	let { data: pageData } = $props();
	let progressData = $state<ProgressData | null>(null);
	let sessions = $state<SessionEntry[]>([]);
	let loading = $state(true);
	let error = $state('');

	let totalSessions = $derived(
		progressData ? progressData.subjects.reduce((sum, s) => sum + s.totalSessions, 0) : 0
	);

	let recentSessionCount = $derived.by(() => {
		const weekAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);
		return sessions.filter((s) => new Date(s.endedAt) >= weekAgo).length;
	});

	function scoreColor(score: number): string {
		if (score === 0) return 'bg-gray-200';
		if (score >= 4) return 'bg-emerald-200';
		if (score >= 3) return 'bg-amber-200';
		return 'bg-rose-200';
	}

	function formatDate(iso: string): string {
		const d = new Date(iso);
		return d.toLocaleDateString('sv-SE', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	onMount(async () => {
		await user.checkAuth();
		if (!$isLoggedIn || !$isParent) {
			goto('/login');
			return;
		}

		try {
			const [progResp, sessResp] = await Promise.all([
				progressApi.child(pageData.studentId),
				progressApi.childSessions(pageData.studentId)
			]);
			progressData = progResp as ProgressData;
			sessions = (sessResp as SessionsData).sessions || [];
		} catch (err: unknown) {
			const apiErr = err as { status?: number };
			if (apiErr.status === 403) {
				error = 'Du har inte behörighet att se denna profil';
			} else if (apiErr.status === 404) {
				error = 'Eleven hittades inte';
			} else {
				error = getErrorMessage(err, 'Kunde inte hämta progress');
			}
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Barnets progress — Nioplugget</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8">
	<a
		href="/dashboard"
		class="mb-4 inline-block text-sm text-muted-foreground transition-colors hover:text-foreground"
	>
		← Tillbaka till dashboard
	</a>

	<h1 class="mb-6 text-2xl font-bold text-foreground">Progress</h1>

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else if error}
		<Card>
			<CardContent class="py-8 text-center">
				<p class="text-red-500">{error}</p>
			</CardContent>
		</Card>
	{:else if progressData}
		<!-- Summary stats -->
		<div class="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
			<Card>
				<CardHeader class="pb-2">
					<CardTitle class="text-sm font-medium text-muted-foreground">
						Pass senaste veckan
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-2xl font-bold text-foreground">{recentSessionCount}</p>
				</CardContent>
			</Card>

			<Card>
				<CardHeader class="pb-2">
					<CardTitle class="text-sm font-medium text-muted-foreground">
						Snittbetyg per ämne
					</CardTitle>
				</CardHeader>
				<CardContent>
					{#each progressData.subjects as subject (subject.id)}
						{#if subject.totalSessions > 0}
							<p class="text-sm text-foreground">
								{subject.name}: {subject.avgScore.toFixed(1)}
							</p>
						{:else}
							<p class="text-sm text-muted-foreground">{subject.name}: –</p>
						{/if}
					{/each}
					{#if !progressData.subjects.some((s) => s.totalSessions > 0)}
						<p class="text-sm text-muted-foreground">Inga pass ännu</p>
					{/if}
				</CardContent>
			</Card>

			<Card>
				<CardHeader class="pb-2">
					<CardTitle class="text-sm font-medium text-muted-foreground">
						Totalt antal pass
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-2xl font-bold text-foreground">{totalSessions}</p>
				</CardContent>
			</Card>
		</div>

		<!-- Per-subject topic breakdowns -->
		{#each progressData.subjects.filter((s) => s.totalSessions > 0) as subject (subject.id)}
			<div class="mb-6">
				<h3 class="mb-3 text-lg font-semibold text-foreground">{subject.name}</h3>
				<div class="space-y-2">
					{#each subject.topics as topic (topic.id)}
						<div class="flex items-center gap-2">
							<span
								class="w-28 truncate text-sm text-muted-foreground"
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
			</div>
		{/each}

		<!-- Session history -->
		<h2 class="mb-4 mt-8 text-xl font-semibold text-foreground">Sessionshistorik</h2>

		{#if sessions.length === 0}
			<p class="text-sm text-muted-foreground">Inga genomförda pass ännu.</p>
		{:else}
			<div class="space-y-2">
				{#each sessions as session (session.id)}
					<div
						class="flex items-center justify-between rounded-lg border border-border p-3"
					>
						<div>
							<p class="font-medium text-foreground">{session.exerciseTitle}</p>
							<p class="text-sm text-muted-foreground">
								{session.subjectName} · {session.topicName}
							</p>
						</div>
						<div class="text-right">
							<p class="font-medium text-foreground">{session.score}/5</p>
							<p class="text-xs text-muted-foreground">
								{formatDate(session.endedAt)}
							</p>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	{/if}
</div>
