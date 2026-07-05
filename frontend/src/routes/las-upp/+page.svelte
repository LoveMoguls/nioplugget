<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
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

<div class="flex min-h-[calc(100vh-4rem)] items-center justify-center px-4 py-12">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle class="text-center text-2xl">Nioplugget</CardTitle>
			<CardDescription class="text-center">
				Ange familjekoden för att låsa upp den här enheten.
			</CardDescription>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit} class="flex flex-col gap-4">
				{#if errorMsg}
					<Alert variant="destructive">
						<AlertDescription>{errorMsg}</AlertDescription>
					</Alert>
				{/if}

				<div class="flex flex-col gap-1.5">
					<Label for="code">Familjekod</Label>
					<Input
						id="code"
						type="password"
						bind:value={code}
						autocomplete="off"
						required
						autofocus
					/>
				</div>

				<Button type="submit" class="w-full" disabled={loading || !code}>
					{loading ? 'Låser upp...' : 'Lås upp'}
				</Button>

				<p class="text-center text-sm text-muted-foreground">
					<a href="/login" class="text-muted-foreground underline-offset-4 hover:underline">
						Logga in med lösenord
					</a>
				</p>
			</form>
		</CardContent>
	</Card>
</div>
