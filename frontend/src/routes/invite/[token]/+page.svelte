<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { invite as inviteApi, getErrorMessage } from '$lib/api';

	let token = $derived($page.params.token);
	let name = $state('');
	let pin = $state('');
	let pinConfirm = $state('');
	let loading = $state(false);
	let errorMsg = $state('');
	let successMsg = $state('');
	let activated = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		errorMsg = '';

		// Validate
		if (!name.trim()) {
			errorMsg = 'Ange ditt namn.';
			return;
		}
		if (!/^\d{4}$/.test(pin)) {
			errorMsg = 'PIN-koden måste vara 4 siffror.';
			return;
		}
		if (pin !== pinConfirm) {
			errorMsg = 'PIN-koderna stämmer inte överens.';
			return;
		}

		loading = true;
		try {
			await inviteApi.activate(token, name.trim(), pin);
			activated = true;
			successMsg = `Välkommen ${name.trim()}! Ditt konto är aktiverat.`;
			setTimeout(() => goto('/child/login'), 3000);
		} catch (err: unknown) {
			const e = err as { status?: number; data?: { error?: string } };
			if (e.status === 404 || e.status === 410) {
				errorMsg = 'Länken har gått ut. Be din förälder skapa en ny.';
			} else if (e.data?.error) {
				errorMsg = e.data.error;
			} else {
				errorMsg = getErrorMessage(err, 'Aktiveringen misslyckades. Länken kan ha gått ut.');
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Aktivera ditt konto — Nioplugget</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-4 py-12">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Aktivera ditt konto</CardTitle>
			<CardDescription>Välj ett namn och en PIN-kod för att komma igång.</CardDescription>
		</CardHeader>
		<CardContent>
			{#if activated}
				<Alert class="mb-4">
					<AlertDescription>{successMsg}</AlertDescription>
				</Alert>
				<p class="text-center text-sm text-muted-foreground">
					Du skickas vidare till inloggningen om några sekunder...
				</p>
			{:else}
				<form onsubmit={handleSubmit} class="flex flex-col gap-4">
					{#if errorMsg}
						<Alert variant="destructive">
							<AlertDescription>{errorMsg}</AlertDescription>
						</Alert>
					{/if}

					<div class="flex flex-col gap-1.5">
						<Label for="name">Ditt namn</Label>
						<Input
							id="name"
							type="text"
							bind:value={name}
							placeholder="Ditt förnamn"
							autocomplete="given-name"
							required
						/>
					</div>

					<div class="flex flex-col gap-1.5">
						<Label for="pin">PIN-kod (4 siffror)</Label>
						<Input
							id="pin"
							type="password"
							inputmode="numeric"
							pattern="[0-9]*"
							maxlength={4}
							bind:value={pin}
							placeholder="••••"
							autocomplete="new-password"
						/>
					</div>

					<div class="flex flex-col gap-1.5">
						<Label for="pin-confirm">Bekräfta PIN-kod</Label>
						<Input
							id="pin-confirm"
							type="password"
							inputmode="numeric"
							pattern="[0-9]*"
							maxlength={4}
							bind:value={pinConfirm}
							placeholder="••••"
							autocomplete="new-password"
						/>
					</div>

					<Button type="submit" class="w-full" disabled={loading}>
						{loading ? 'Aktiverar...' : 'Aktivera konto'}
					</Button>
				</form>
			{/if}
		</CardContent>
	</Card>
</div>
