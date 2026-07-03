<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { childAuth, getErrorMessage } from '$lib/api';
	import { user } from '$lib/stores/auth';

	// Login steps: 'setup' | 'select' | 'pin'
	// 'setup' is only shown if no parent email is saved on device
	type Step = 'setup' | 'select' | 'pin';
	let step = $state<Step>('select');

	let parentEmail = $state('');
	let selectedName = $state('');
	let pin = $state('');
	let childNames = $state<string[]>([]);
	let loading = $state(false);
	let errorMsg = $state('');

	onMount(async () => {
		const saved = localStorage.getItem('childParentEmail');
		if (saved) {
			parentEmail = saved;
			await loadNames();
		} else {
			step = 'setup';
		}
	});

	async function loadNames() {
		loading = true;
		errorMsg = '';
		try {
			const data = (await childAuth.names(parentEmail)) as
				| string[]
				| { name: string }[]
				| { names: string[] };
			let names: string[];
			if (Array.isArray(data)) {
				names = data.map((d) => (typeof d === 'string' ? d : d.name));
			} else {
				names = data.names || [];
			}
			if (names.length === 0) {
				errorMsg = 'Inga barn hittades. Be din förälder skapa en inbjudningslänk.';
				step = 'setup';
				return;
			}
			childNames = names;
			step = 'select';
		} catch {
			errorMsg = 'Kunde inte hämta konton. Be din förälder kontrollera inbjudningslänken.';
			step = 'setup';
		} finally {
			loading = false;
		}
	}

	async function handleSetupSubmit(e: Event) {
		e.preventDefault();
		errorMsg = '';
		if (!parentEmail || !parentEmail.includes('@')) {
			errorMsg = 'Ange en giltig e-postadress.';
			return;
		}
		localStorage.setItem('childParentEmail', parentEmail);
		await loadNames();
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
		{#if step === 'setup'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-2xl">Kom igång</CardTitle>
					<CardDescription class="text-center">Be din förälder ange sin e-post för att ställa in den här enheten</CardDescription>
				</CardHeader>
				<CardContent>
					<form onsubmit={handleSetupSubmit} class="flex flex-col gap-4">
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
							{loading ? 'Söker...' : 'Fortsätt'}
						</Button>
					</form>
				</CardContent>
			</Card>

		{:else if step === 'select'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-2xl">Vem är du?</CardTitle>
				</CardHeader>
				<CardContent>
					{#if errorMsg}
						<Alert variant="destructive" class="mb-4">
							<AlertDescription>{errorMsg}</AlertDescription>
						</Alert>
					{/if}
					{#if loading}
						<p class="text-center text-muted-foreground">Laddar...</p>
					{:else}
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
							class="mt-4 w-full text-xs text-muted-foreground"
							onclick={() => {
								localStorage.removeItem('childParentEmail');
								parentEmail = '';
								childNames = [];
								errorMsg = '';
								step = 'setup';
							}}
						>
							Fel familj? Byt e-postadress
						</Button>
					{/if}
				</CardContent>
			</Card>

		{:else if step === 'pin'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-2xl">Hej, {selectedName}!</CardTitle>
					<CardDescription class="text-center">Ange din PIN-kod</CardDescription>
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
