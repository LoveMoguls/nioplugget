<!-- frontend/src/lib/components/challenges/StarAnimation.svelte -->
<script lang="ts">
	interface Props {
		stars: number; // 1-3
		xp: number;
		visible: boolean;
	}

	let { stars, xp, visible }: Props = $props();
</script>

{#if visible}
	<div class="star-overlay fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
		<div class="flex flex-col items-center rounded-3xl bg-card px-10 py-8 text-center shadow-2xl">
			<div class="flex gap-2 text-6xl">
				{#each [1, 2, 3] as n}
					<span
						class="star {n <= stars ? 'earned' : 'unearned'}"
						style="animation-delay: {n * 0.25}s"
					>⭐</span>
				{/each}
			</div>
			<p class="xp-pop mt-4 text-3xl font-extrabold text-primary">+{xp} XP!</p>
			<p class="mt-2 text-sm text-muted-foreground">
				{#if stars === 3}Fantastiskt! Du kan det här! 🎉
				{:else if stars === 2}Bra jobbat! Fortsätt så! 💪
				{:else}Du är på väg! Kör igen om du vill. 🚀
				{/if}
			</p>
		</div>
	</div>
{/if}

<style>
	.star-overlay {
		animation: fade-in 0.25s ease-out;
	}

	.star {
		display: inline-block;
		opacity: 0;
		transform: scale(0) rotate(-30deg);
		animation: star-pop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
	}

	.star.earned {
		--star-opacity: 1;
	}

	.star.unearned {
		--star-opacity: 0.25;
		filter: grayscale(1);
	}

	.xp-pop {
		opacity: 0;
		animation: xp-pop 0.4s ease-out 1s forwards;
	}

	@keyframes fade-in {
		from {
			opacity: 0;
		}
	}

	@keyframes star-pop {
		60% {
			transform: scale(1.35) rotate(8deg);
			opacity: var(--star-opacity, 1);
		}
		100% {
			transform: scale(1) rotate(0deg);
			opacity: var(--star-opacity, 1);
		}
	}

	@keyframes xp-pop {
		from {
			opacity: 0;
			transform: translateY(10px) scale(0.8);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
</style>
