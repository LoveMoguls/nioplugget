<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { childAuth, getErrorMessage } from '$lib/api';
	import { user } from '$lib/stores/auth';

	// Login steps: 'email' | 'select' | 'pin'
	type Step = 'email' | 'select' | 'pin';
	let step = $state<Step>('email');

	let parentEmail = $state('');
	let selectedName = $state('');
	let pin = $state('');
	let childNames = $state<string[]>([]);
	let loading = $state(false);
	let errorMsg = $state('');

	async function handleEmailSubmit(e: Event) {
		e.preventDefault();
		errorMsg = '';
		if (!parentEmail || !parentEmail.includes('@')) {
			errorMsg = 'Ange en giltig e-postadress.';
			return;
		}
		loading = true;
		try {
			const data = (await childAuth.names(parentEmail)) as string[] | { names: string[] };
			const names = Array.isArray(data) ? data : data.names || [];
			if (names.length === 0) {
				errorMsg = 'Inga barn hittades för den e-postadressen.';
				return;
			}
			childNames = names;
			step = 'select';
		} catch (err) {
			errorMsg = getErrorMessage(err, 'Kunde inte hämta barn. Kontrollera e-postadressen.');
		} finally {
			loading = false;
		}
	}

	function selectName(name: string) {
		selectedName = name;
		pin = '';
		errorMsg = '';
		step = 'pin';
	}

	async function handlePinSubmit(e: Event) {
		e.preventDefault();
		errorMsg = '';
		if (pin.length !== 4 || !/^\d{4}$/.test(pin)) {
			errorMsg = 'PIN-koden måste vara 4 siffror.';
			return;
		}
		loading = true;
		try {
			const data = (await childAuth.login(parentEmail, selectedName, pin)) as {
				id: string;
				name: string;
			};
			user.setUser({ id: data.id, email: '', role: 'child' });
			goto('/study');
		} catch (err: unknown) {
			const e = err as { status?: number; data?: { error?: string; remaining?: number } };
			if (e.data?.error) {
				errorMsg = e.data.error;
			} else if (e.status === 429) {
				errorMsg = 'För många försök. Försök igen om 15 minuter.';
			} else {
				errorMsg = getErrorMessage(err, 'Inloggningen misslyckades.');
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Elevlogin — Nioplugget</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-4 py-12">
	<div class="w-full max-w-sm">
		{#if step === 'email'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-2xl">Logga in</CardTitle>
					<CardDescription class="text-center">Ange din förälders e-postadress</CardDescription>
				</CardHeader>
				<CardContent>
					<form onsubmit={handleEmailSubmit} class="flex flex-col gap-4">
						{#if errorMsg}
							<Alert variant="destructive">
								<AlertDescription>{errorMsg}</AlertDescription>
							</Alert>
						{/if}
						<div class="flex flex-col gap-1.5">
							<Label for="parent-email">Förälders e-post</Label>
							<Input
								id="parent-email"
								type="email"
								bind:value={parentEmail}
								placeholder="mamma@epost.se"
								autocomplete="email"
								required
							/>
						</div>
						<Button type="submit" class="w-full" disabled={loading}>
							{loading ? 'Söker...' : 'Nästa'}
						</Button>
					</form>
				</CardContent>
			</Card>

		{:else if step === 'select'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-2xl">Vem är du?</CardTitle>
					<CardDescription class="text-center">Välj ditt namn</CardDescription>
				</CardHeader>
				<CardContent>
					{#if errorMsg}
						<Alert variant="destructive" class="mb-4">
							<AlertDescription>{errorMsg}</AlertDescription>
						</Alert>
					{/if}
					<div class="flex flex-col gap-2">
						{#each childNames as name (name)}
							<Button
								variant="outline"
								class="h-12 w-full text-base"
								onclick={() => selectName(name)}
							>
								{name}
							</Button>
						{/each}
					</div>
					<Button
						variant="ghost"
						class="mt-4 w-full"
						onclick={() => {
							step = 'email';
							errorMsg = '';
						}}
					>
						Tillbaka
					</Button>
				</CardContent>
			</Card>

		{:else if step === 'pin'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-2xl">Hej, {selectedName}!</CardTitle>
					<CardDescription class="text-center">Ange din 4-siffriga PIN-kod</CardDescription>
				</CardHeader>
				<CardContent>
					<form onsubmit={handlePinSubmit} class="flex flex-col gap-4">
						{#if errorMsg}
							<Alert variant="destructive">
								<AlertDescription>{errorMsg}</AlertDescription>
							</Alert>
						{/if}
						<div class="flex flex-col items-center gap-3">
							<Input
								type="password"
								inputmode="numeric"
								pattern="[0-9]*"
								maxlength={4}
								bind:value={pin}
								placeholder="••••"
								autocomplete="current-password"
								class="h-14 w-32 text-center text-2xl tracking-widest"
								autofocus
							/>
						</div>
						<Button type="submit" class="w-full" disabled={loading || pin.length !== 4}>
							{loading ? 'Loggar in...' : 'Logga in'}
						</Button>
						<Button
							type="button"
							variant="ghost"
							class="w-full"
							onclick={() => {
								step = 'select';
								pin = '';
								errorMsg = '';
							}}
						>
							Tillbaka
						</Button>
					</form>
				</CardContent>
			</Card>
		{/if}
	</div>
</div>
