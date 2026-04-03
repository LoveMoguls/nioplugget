<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { page } from '$app/stores';
	import { sessions } from '$lib/api';
	import ChatBubble from '$lib/components/chat/ChatBubble.svelte';
	import ChatInput from '$lib/components/chat/ChatInput.svelte';
	import TypingIndicator from '$lib/components/chat/TypingIndicator.svelte';
	import {
		messages,
		isStreaming,
		sessionEnded,
		sendMessage,
		loadMessages,
		resetChat,
	} from '$lib/stores/chat';

	let sessionId = $state('');
	let scrollContainer: HTMLDivElement | undefined = $state();
	let endAnchor: HTMLDivElement | undefined = $state();
	let showEndConfirm = $state(false);
	let ending = $state(false);
	let scoreResult: { score: number; summary: string; feedback: string } | null = $state(null);

	onMount(async () => {
		resetChat();
		sessionId = ($page.params as Record<string, string>).sessionId;

		// Load existing messages
		await loadMessages(sessionId);

		// Check if session is already ended
		try {
			const session = (await sessions.get(sessionId)) as {
				endedAt?: string;
				score?: number;
				summary?: string;
			};
			if (session.endedAt) {
				sessionEnded.set(true);
				if (session.score) {
					scoreResult = {
						score: session.score,
						summary: session.summary || '',
						feedback: '',
					};
				}
			}
		} catch {
			// Ignore
		}

		scrollToBottom();
	});

	// Auto-scroll on new messages
	$effect(() => {
		// Subscribe to messages to trigger scroll
		const _ = $messages;
		tick().then(scrollToBottom);
	});

	function scrollToBottom() {
		endAnchor?.scrollIntoView({ behavior: 'smooth' });
	}

	async function handleSend(message: string) {
		await sendMessage(sessionId, message);
	}

	async function handleEndSession() {
		ending = true;
		try {
			const result = (await sessions.end(sessionId)) as {
				score: number;
				summary: string;
				feedback: string;
				endedAt: string;
			};
			scoreResult = result;
			sessionEnded.set(true);
			showEndConfirm = false;
		} catch {
			showEndConfirm = false;
		} finally {
			ending = false;
		}
	}

	function scoreStars(score: number): string {
		return '★'.repeat(score) + '☆'.repeat(5 - score);
	}
</script>

<svelte:head>
	<title>Övningspass — Nioplugget</title>
</svelte:head>

<div class="flex h-[calc(100vh-4rem)] flex-col">
	<!-- Header -->
	<div class="flex min-h-[44px] items-center justify-between border-b border-border px-4 py-3">
		<a href="/study" class="min-h-[44px] flex items-center text-sm text-muted-foreground hover:underline">&larr; Övningar</a>
		{#if !$sessionEnded}
			<button
				onclick={() => (showEndConfirm = true)}
				disabled={$isStreaming}
				class="min-h-[44px] rounded-lg bg-destructive/10 px-3 py-1.5 text-sm font-medium text-destructive
					hover:bg-destructive/20 disabled:opacity-50"
			>
				Avsluta pass
			</button>
		{/if}
	</div>

	<!-- End session confirmation -->
	{#if showEndConfirm}
		<div class="border-b border-border bg-accent px-4 py-3 text-sm">
			<p class="mb-2 font-medium text-accent-foreground">Är du säker på att du vill avsluta passet?</p>
			<div class="flex gap-2">
				<button
					onclick={handleEndSession}
					disabled={ending}
					class="min-h-[44px] rounded-lg bg-destructive px-3 py-1.5 text-sm text-primary-foreground hover:bg-destructive/90 disabled:opacity-50"
				>
					{ending ? 'Avslutar...' : 'Ja, avsluta'}
				</button>
				<button
					onclick={() => (showEndConfirm = false)}
					class="min-h-[44px] rounded-lg bg-secondary px-3 py-1.5 text-sm text-secondary-foreground hover:bg-secondary/80"
				>
					Nej, fortsätt
				</button>
			</div>
		</div>
	{/if}

	<!-- Messages -->
	<div bind:this={scrollContainer} class="flex-1 overflow-y-auto bg-background px-4 py-6">
		{#if $messages.length === 0 && !$isStreaming}
			<p class="text-center text-sm text-muted-foreground">
				Skriv ett meddelande för att börja din övning.
			</p>
		{/if}

		{#each $messages as msg}
			<ChatBubble role={msg.role} content={msg.content} />
		{/each}

		{#if $isStreaming && ($messages.length === 0 || $messages[$messages.length - 1]?.content === '')}
			<TypingIndicator />
		{/if}

		<!-- Score card -->
		{#if scoreResult}
			<div class="mx-auto mt-6 max-w-sm rounded-2xl border border-border bg-card p-6 text-center shadow-sm">
				<p class="text-3xl tracking-wider text-primary">{scoreStars(scoreResult.score)}</p>
				<p class="mt-2 text-lg font-semibold text-card-foreground">{scoreResult.score} av 5</p>
				{#if scoreResult.feedback}
					<p class="mt-2 text-sm text-muted-foreground">{scoreResult.feedback}</p>
				{/if}
				<a
					href="/study"
					class="mt-4 inline-block min-h-[44px] rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
				>
					Tillbaka till övningar
				</a>
			</div>
		{/if}

		<div bind:this={endAnchor}></div>
	</div>

	<!-- Input -->
	{#if !$sessionEnded}
		<div class="border-t border-border bg-card px-4 py-3" style="padding-bottom: max(env(safe-area-inset-bottom, 0px), 0.75rem);">
			<ChatInput onSend={handleSend} disabled={$isStreaming} />
		</div>
	{/if}
</div>
