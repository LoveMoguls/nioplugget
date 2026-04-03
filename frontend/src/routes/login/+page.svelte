<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { auth as authApi, getErrorMessage } from '$lib/api';
	import { user } from '$lib/stores/auth';

	let email = $state('');
	let password = $state('');
	let loading = $state(false);
	let errorMsg = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		errorMsg = '';
		loading = true;
		try {
			const data = (await authApi.login(email, password)) as { id: string; email: string };
			user.setUser({ id: data.id, email: data.email, role: 'parent' });
			goto('/dashboard');
		} catch (err) {
			errorMsg = getErrorMessage(err, 'Inloggningen misslyckades. Försök igen.');
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Logga in — Nioplugget</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-4rem)] items-center justify-center px-4 py-12">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Logga in</CardTitle>
			<CardDescription>Logga in på ditt föräldrakonto.</CardDescription>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit} class="flex flex-col gap-4">
				{#if errorMsg}
					<Alert variant="destructive">
						<AlertDescription>{errorMsg}</AlertDescription>
					</Alert>
				{/if}

				<div class="flex flex-col gap-1.5">
					<Label for="email">E-postadress</Label>
					<Input
						id="email"
						type="email"
						bind:value={email}
						placeholder="din@epost.se"
						autocomplete="email"
						required
					/>
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="password">Lösenord</Label>
					<Input
						id="password"
						type="password"
						bind:value={password}
						autocomplete="current-password"
						required
					/>
				</div>

				<Button type="submit" class="w-full" disabled={loading}>
					{loading ? 'Loggar in...' : 'Logga in'}
				</Button>

				<p class="text-center text-sm text-muted-foreground">
					Inget konto ännu?
					<a href="/register" class="text-foreground underline-offset-4 hover:underline">
						Registrera dig
					</a>
				</p>
			</form>
		</CardContent>
	</Card>
</div>
