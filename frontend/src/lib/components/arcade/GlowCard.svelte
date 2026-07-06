<script lang="ts">
	import type { Snippet } from 'svelte';
	let {
		gradient = 'linear-gradient(135deg, oklch(0.85 0.15 195), oklch(0.70 0.25 330))',
		children
	}: { gradient?: string; children: Snippet } = $props();
</script>

<div class="glow-card rounded-xl p-[2px] transition-transform duration-200" style="background: {gradient};">
	<div class="rounded-[calc(1rem+4px)] bg-card p-5">
		{@render children()}
	</div>
</div>

<style>
	@media (prefers-reduced-motion: no-preference) {
		.glow-card:hover {
			transform: translateY(-3px);
		}
	}
	.glow-card:hover {
		box-shadow: var(--glow-cyan);
	}
</style>
