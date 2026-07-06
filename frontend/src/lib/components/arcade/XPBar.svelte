<script lang="ts">
	let { current, max, label = 'XP' }: { current: number; max: number; label?: string } = $props();
	const pct = $derived(max > 0 ? Math.min(100, Math.round((current / max) * 100)) : 0);
</script>

<div class="w-full">
	<div class="mb-1 flex items-baseline justify-between">
		<span class="font-display text-sm font-bold tracking-wide text-success uppercase">{label}</span>
		<span class="text-sm font-bold text-success">{current}</span>
	</div>
	<div class="h-3 overflow-hidden rounded-full bg-secondary">
		<div
			class="h-full rounded-full bg-success transition-[width] duration-500 ease-out"
			style="width: {pct}%; box-shadow: var(--glow-lime);"
		></div>
	</div>
</div>
