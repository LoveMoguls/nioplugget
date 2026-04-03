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
			? 'rounded-br-md bg-primary text-primary-foreground'
			: 'rounded-bl-md bg-muted text-foreground'}"
	>
		{#each content.split('\n') as line, i}
			{#if i > 0}<br />{/if}
			{line}
		{/each}
	</div>
</div>
