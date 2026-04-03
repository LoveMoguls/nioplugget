<script lang="ts">
	interface Props {
		onSend: (message: string) => void;
		disabled?: boolean;
	}

	let { onSend, disabled = false }: Props = $props();
	let message = $state('');

	function handleSubmit() {
		const trimmed = message.trim();
		if (trimmed && !disabled) {
			onSend(trimmed);
			message = '';
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSubmit();
		}
	}
</script>

<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="flex gap-2">
	<textarea
		bind:value={message}
		onkeydown={handleKeydown}
		{disabled}
		placeholder="Skriv ditt svar..."
		rows={1}
		class="flex-1 resize-none rounded-xl border border-input bg-background px-4 py-3 text-sm
			focus:border-primary focus:outline-none focus:ring-1 focus:ring-ring
			disabled:cursor-not-allowed disabled:opacity-50"
	></textarea>
	<button
		type="submit"
		disabled={disabled || !message.trim()}
		class="min-h-[44px] shrink-0 rounded-xl bg-primary px-4 py-3 text-sm font-medium text-primary-foreground
			hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
	>
		Skicka
	</button>
</form>
