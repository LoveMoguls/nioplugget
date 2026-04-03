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
	let gdprConsent = $state(false);
	let loading = $state(false);
	let errorMsg = $state('');

	function validate(): string {
		if (!email || !email.includes('@')) return 'Ange en giltig e-postadress.';
		if (password.length < 8) return 'Lösenordet måste vara minst 8 tecken.';
		if (!gdprConsent) return 'Du måste godkänna GDPR-villkoren för att registrera dig.';
		return '';
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		errorMsg = '';
		const validationError = validate();
		if (validationError) {
			errorMsg = validationError;
			return;
		}
		loading = true;
		try {
			const data = (await authApi.register(email, password, gdprConsent)) as {
				id: string;
				email: string;
			};
			user.setUser({ id: data.id, email: data.email, role: 'parent' });
			goto('/dashboard');
		} catch (err) {
			errorMsg = getErrorMessage(err, 'Registreringen misslyckades. Försök igen.');
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Registrera dig — Nioplugget</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-4rem)] items-center justify-center px-4 py-12">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Registrera dig</CardTitle>
			<CardDescription>Skapa ett föräldrakonto för att komma igång.</CardDescription>
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
						placeholder="Minst 8 tecken"
						autocomplete="new-password"
						required
					/>
				</div>

				<div class="flex items-start gap-2 rounded-lg border border-border p-3">
					<input
						id="gdpr"
						type="checkbox"
						bind:checked={gdprConsent}
						class="mt-0.5 h-4 w-4 rounded border-border"
					/>
					<Label for="gdpr" class="cursor-pointer text-sm leading-snug">
						Jag samtycker till att mina uppgifter behandlas enligt GDPR. Dina uppgifter används
						endast för att tillhandahålla tjänsten och delas inte med tredje part.
					</Label>
				</div>

				<Button type="submit" class="w-full" disabled={loading}>
					{loading ? 'Registrerar...' : 'Registrera dig'}
				</Button>

				<p class="text-center text-sm text-muted-foreground">
					Har du redan ett konto?
					<a href="/login" class="text-foreground underline-offset-4 hover:underline">Logga in</a>
				</p>
			</form>
		</CardContent>
	</Card>
</div>
