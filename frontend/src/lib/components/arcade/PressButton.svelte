<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	let {
		variant = 'cyan',
		children,
		...rest
	}: { variant?: 'cyan' | 'magenta' | 'lime'; children: Snippet } & HTMLButtonAttributes = $props();

	const styles = {
		cyan: 'background: oklch(0.85 0.15 195); color: oklch(0.18 0.03 285); --press-glow: var(--glow-cyan); --press-edge: oklch(0.60 0.12 195);',
		magenta: 'background: oklch(0.70 0.25 330); color: oklch(0.98 0.005 285); --press-glow: var(--glow-magenta); --press-edge: oklch(0.48 0.20 330);',
		lime: 'background: oklch(0.85 0.22 135); color: oklch(0.18 0.03 285); --press-glow: var(--glow-lime); --press-edge: oklch(0.60 0.17 135);'
	} as const;
</script>

<button
	{...rest}
	class="press-btn font-display rounded-xl px-8 py-3 text-lg font-bold tracking-wide uppercase disabled:cursor-not-allowed disabled:opacity-50 {rest.class ?? ''}"
	style="{styles[variant]} {rest.style ?? ''}"
>
	{@render children()}
</button>

<style>
	.press-btn {
		box-shadow:
			0 5px 0 var(--press-edge),
			var(--press-glow);
		transition:
			transform 0.15s ease,
			box-shadow 0.15s ease;
	}
	.press-btn:active:not(:disabled) {
		transform: translateY(4px);
		box-shadow:
			0 1px 0 var(--press-edge),
			var(--press-glow);
	}
</style>
