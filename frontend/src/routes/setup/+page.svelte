<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { apiKey as apiKeyApi, getErrorMessage } from '$lib/api';
	import { user, isLoggedIn, isParent } from '$lib/stores/auth';

	let apiKeyInput = $state('');
	let loading = $state(false);
	let errorMsg = $state('');
	let successMsg = $state('');

	onMount(async () => {
		await user.checkAuth();
		if (!$isLoggedIn || !$isParent) {
			goto('/login');
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		errorMsg = '';
		successMsg = '';
		if (!apiKeyInput.trim()) {
			errorMsg = 'Ange en API-nyckel.';
			return;
		}
		loading = true;
		try {
			await apiKeyApi.store(apiKeyInput.trim());
			successMsg = 'API-nyckeln är sparad!';
			setTimeout(() => goto('/dashboard'), 1500);
		} catch (err) {
			errorMsg = getErrorMessage(err, 'Kunde inte spara API-nyckeln. Kontrollera att nyckeln är giltig.');
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Lägg till API-nyckel — Nioplugget</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-4rem)] items-center justify-center px-4 py-12">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Koppla Claude API-nyckel</CardTitle>
			<CardDescription>
				För att ditt barn ska kunna använda Nioplugget behöver du koppla din Claude API-nyckel.
				Hämta en nyckel från
				<a
					href="https://console.anthropic.com"
					target="_blank"
					rel="noopener noreferrer"
					class="text-foreground underline-offset-4 hover:underline"
				>
					console.anthropic.com
				</a>.
			</CardDescription>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit} class="flex flex-col gap-4">
				{#if errorMsg}
					<Alert variant="destructive">
						<AlertDescription>{errorMsg}</AlertDescription>
					</Alert>
				{/if}
				{#if successMsg}
					<Alert>
						<AlertDescription>{successMsg}</AlertDescription>
					</Alert>
				{/if}

				<div class="flex flex-col gap-1.5">
					<Label for="apikey">API-nyckel</Label>
					<Input
						id="apikey"
						type="password"
						bind:value={apiKeyInput}
						placeholder="sk-ant-..."
						autocomplete="off"
					/>
					<p class="text-xs text-muted-foreground">
						Nyckeln krypteras och lagras säkert. Den används enbart för att kommunicera med Claude.
					</p>
				</div>

				<Button type="submit" class="w-full" disabled={loading}>
					{loading ? 'Sparar och validerar...' : 'Spara API-nyckel'}
				</Button>

				<Button variant="outline" href="/dashboard" class="w-full">Tillbaka till dashboard</Button>
			</form>
		</CardContent>
	</Card>
</div>
