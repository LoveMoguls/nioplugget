<div class="mb-3 flex justify-start">
	<div class="flex items-center gap-1 rounded-2xl rounded-bl-md bg-slate-100 px-4 py-3 shadow-sm">
		<span class="typing-dot h-2 w-2 rounded-full bg-slate-400"></span>
		<span class="typing-dot animation-delay-200 h-2 w-2 rounded-full bg-slate-400"></span>
		<span class="typing-dot animation-delay-400 h-2 w-2 rounded-full bg-slate-400"></span>
	</div>
</div>

<style>
	.typing-dot {
		animation: bounce 1.4s infinite ease-in-out;
	}

	.animation-delay-200 {
		animation-delay: 0.2s;
	}

	.animation-delay-400 {
		animation-delay: 0.4s;
	}

	@keyframes bounce {
		0%,
		80%,
		100% {
			transform: translateY(0);
		}
		40% {
			transform: translateY(-6px);
		}
	}
</style>
