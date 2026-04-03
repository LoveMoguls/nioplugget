<script lang="ts">
	interface Props {
		role: 'user' | 'assistant';
		content: string;
	}

	let { role, content }: Props = $props();
</script>

<div class="flex {role === 'user' ? 'justify-end' : 'justify-start'} mb-3">
	<div
		class="max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed shadow-sm
		{role === 'user'
			? 'rounded-br-md bg-indigo-600 text-white'
			: 'rounded-bl-md bg-slate-100 text-slate-900'}"
	>
		{#each content.split('\n') as line, i}
			{#if i > 0}<br />{/if}
			{line}
		{/each}
	</div>
</div>
