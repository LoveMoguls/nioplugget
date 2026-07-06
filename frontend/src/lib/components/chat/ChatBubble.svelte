<script lang="ts">
	interface Props {
		role: 'user' | 'assistant';
		content: string;
	}

	let { role, content }: Props = $props();
</script>

<div class="flex {role === 'user' ? 'justify-end' : 'justify-start'} mb-3">
	<div
		class="max-w-[85%] rounded-2xl px-4 py-3 text-sm leading-relaxed shadow-sm
		{role === 'user'
			? 'user-bubble rounded-br-md'
			: 'rounded-bl-md border border-border bg-card text-foreground'}"
	>
		{#each content.split('\n') as line, i}
			{#if i > 0}<br />{/if}
			{line}
		{/each}
	</div>
</div>

<style>
	.user-bubble {
		background: linear-gradient(135deg, oklch(0.45 0.1 195), oklch(0.38 0.08 220));
		color: oklch(0.97 0.005 285);
	}
</style>
