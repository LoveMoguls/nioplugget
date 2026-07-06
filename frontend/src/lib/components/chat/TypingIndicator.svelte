<div class="mb-3 flex justify-start">
	<div class="flex items-center gap-1 rounded-2xl rounded-bl-md border border-border bg-card px-4 py-3 shadow-sm">
		<span class="typing-dot h-2 w-2 rounded-full bg-primary"></span>
		<span class="typing-dot h-2 w-2 rounded-full bg-primary"></span>
		<span class="typing-dot h-2 w-2 rounded-full bg-primary"></span>
	</div>
</div>

<style>
	@media (prefers-reduced-motion: no-preference) {
		.typing-dot {
			animation: bounce-dot 1s infinite;
		}

		.typing-dot:nth-child(2) {
			animation-delay: 0.15s;
		}

		.typing-dot:nth-child(3) {
			animation-delay: 0.3s;
		}

		@keyframes bounce-dot {
			0%,
			60%,
			100% {
				transform: translateY(0);
			}
			30% {
				transform: translateY(-6px);
			}
		}
	}
</style>
