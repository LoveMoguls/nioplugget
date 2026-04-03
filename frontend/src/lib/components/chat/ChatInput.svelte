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
		class="flex-1 resize-none rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm
			focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500
			disabled:cursor-not-allowed disabled:opacity-50"
	></textarea>
	<button
		type="submit"
		disabled={disabled || !message.trim()}
		class="shrink-0 rounded-xl bg-indigo-600 px-4 py-3 text-sm font-medium text-white
			hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-50"
	>
		Skicka
	</button>
</form>
