<script lang="ts">
	import { goto } from '$app/navigation';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { PressButton } from '$lib/components/arcade';
	import { device, getErrorMessage } from '$lib/api';

	let code = $state('');
	let loading = $state(false);
	let errorMsg = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		errorMsg = '';
		loading = true;
		try {
			await device.unlock(code);
			goto('/profiler');
		} catch (err) {
			errorMsg = getErrorMessage(err, 'Kunde inte låsa upp enheten. Försök igen.');
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Lås upp — Nioplugget</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-4rem)] flex-col items-center justify-center gap-10 px-4 py-12 text-center">
	<div class="flex flex-col items-center gap-3">
		<h1 class="unlock-title font-display text-5xl font-extrabold text-primary">NIOPLUGGET</h1>
		<p class="max-w-xs text-sm text-muted-foreground">
			Ange familjekoden för att låsa upp den här enheten.
		</p>
	</div>

	<form onsubmit={handleSubmit} class="flex w-full max-w-xs flex-col items-center gap-4">
		{#if errorMsg}
			<Alert variant="destructive" class="w-full">
				<AlertDescription>{errorMsg}</AlertDescription>
			</Alert>
		{/if}

		<div class="flex w-full flex-col items-center gap-2">
			<Label for="code" class="font-display text-xs tracking-widest text-muted-foreground uppercase"
				>Familjekod</Label
			>
			<Input
				id="code"
				type="password"
				bind:value={code}
				autocomplete="off"
				required
				autofocus
				class="h-14 w-full rounded-xl border-2 bg-input/40 text-center text-2xl tracking-[0.3em] focus-visible:[box-shadow:var(--glow-cyan)]"
			/>
		</div>

		<PressButton variant="cyan" type="submit" class="w-full" disabled={loading || !code}>
			{loading ? 'Låser upp...' : 'Lås upp'}
		</PressButton>

		<a href="/login" class="text-xs text-muted-foreground underline-offset-4 hover:underline">
			Logga in med lösenord
		</a>
	</form>
</div>

<style>
	@keyframes pulse-glow {
		0%,
		100% {
			text-shadow: 0 0 20px oklch(0.85 0.15 195 / 0.5);
		}
		50% {
			text-shadow: 0 0 45px oklch(0.85 0.15 195 / 0.9);
		}
	}

	@media (prefers-reduced-motion: no-preference) {
		.unlock-title {
			animation: pulse-glow 2.4s ease-in-out infinite;
		}
	}
</style>
