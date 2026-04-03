<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { apiKey as apiKeyApi, children as childrenApi, getErrorMessage } from '$lib/api';
	import { user, isLoggedIn, isParent } from '$lib/stores/auth';

	// API Key state
	interface ApiKeyData {
		masked: string;
		hasKey: boolean;
	}
	let apiKeyData = $state<ApiKeyData | null>(null);
	let apiKeyInput = $state('');
	let apiKeyLoading = $state(false);
	let apiKeyError = $state('');
	let apiKeySuccess = $state('');
	let showUpdateForm = $state(false);

	// Children state
	interface Child {
		id: string;
		name: string;
		activated: boolean;
		inviteToken?: string;
	}
	let childList = $state<Child[]>([]);
	let newChildName = $state('');
	let childLoading = $state(false);
	let childError = $state('');
	let inviteLinks = $state<Record<string, string>>({});

	// Auth check on mount
	onMount(async () => {
		await user.checkAuth();
		if (!$isLoggedIn || !$isParent) {
			goto('/login');
			return;
		}
		await loadData();
	});

	async function loadData() {
		await Promise.all([loadApiKey(), loadChildren()]);
	}

	async function loadApiKey() {
		try {
			const data = (await apiKeyApi.get()) as { masked: string } | null;
			if (data && data.masked) {
				apiKeyData = { masked: data.masked, hasKey: true };
			} else {
				apiKeyData = { masked: '', hasKey: false };
			}
		} catch (err: unknown) {
			const e = err as { status?: number };
			if (e.status === 404 || e.status === 401) {
				apiKeyData = { masked: '', hasKey: false };
			} else {
				apiKeyError = getErrorMessage(err, 'Kunde inte hämta API-nyckel.');
			}
		}
	}

	async function loadChildren() {
		try {
			const data = (await childrenApi.list()) as Child[] | null;
			childList = data || [];
		} catch {
			childList = [];
		}
	}

	async function handleStoreApiKey(e: Event) {
		e.preventDefault();
		apiKeyError = '';
		apiKeySuccess = '';
		if (!apiKeyInput.trim()) {
			apiKeyError = 'Ange en API-nyckel.';
			return;
		}
		apiKeyLoading = true;
		try {
			const data = (await apiKeyApi.store(apiKeyInput.trim())) as { masked: string };
			apiKeyData = { masked: data.masked, hasKey: true };
			apiKeyInput = '';
			apiKeySuccess = 'API-nyckeln har sparats.';
		} catch (err) {
			apiKeyError = getErrorMessage(err, 'Kunde inte spara API-nyckeln. Kontrollera att nyckeln är giltig.');
		} finally {
			apiKeyLoading = false;
		}
	}

	async function handleUpdateApiKey(e: Event) {
		e.preventDefault();
		apiKeyError = '';
		apiKeySuccess = '';
		if (!apiKeyInput.trim()) {
			apiKeyError = 'Ange en ny API-nyckel.';
			return;
		}
		apiKeyLoading = true;
		try {
			const data = (await apiKeyApi.update(apiKeyInput.trim())) as { masked: string };
			apiKeyData = { masked: data.masked, hasKey: true };
			apiKeyInput = '';
			apiKeySuccess = 'API-nyckeln har uppdaterats.';
			showUpdateForm = false;
		} catch (err) {
			apiKeyError = getErrorMessage(err, 'Kunde inte uppdatera API-nyckeln. Kontrollera att nyckeln är giltig.');
		} finally {
			apiKeyLoading = false;
		}
	}

	async function handleDeleteApiKey() {
		if (!confirm('Är du säker på att du vill ta bort API-nyckeln? Ditt barn kan inte logga in utan en giltig nyckel.')) return;
		apiKeyError = '';
		apiKeySuccess = '';
		apiKeyLoading = true;
		try {
			await apiKeyApi.delete();
			apiKeyData = { masked: '', hasKey: false };
			apiKeySuccess = 'API-nyckeln har tagits bort.';
		} catch (err) {
			apiKeyError = getErrorMessage(err, 'Kunde inte ta bort API-nyckeln.');
		} finally {
			apiKeyLoading = false;
		}
	}

	async function handleCreateChild(e: Event) {
		e.preventDefault();
		childError = '';
		if (!newChildName.trim()) {
			childError = 'Ange ett namn för barnet.';
			return;
		}
		childLoading = true;
		try {
			const child = (await childrenApi.create(newChildName.trim())) as Child;
			childList = [...childList, child];
			// Generate invite link immediately
			await handleGenerateInvite(child.id);
			newChildName = '';
		} catch (err) {
			childError = getErrorMessage(err, 'Kunde inte skapa barnprofilen.');
		} finally {
			childLoading = false;
		}
	}

	async function handleGenerateInvite(childId: string) {
		try {
			const data = (await childrenApi.generateInvite(childId)) as { inviteToken: string; inviteUrl?: string };
			const token = data.inviteToken || data.inviteUrl || '';
			const link = token.startsWith('http') ? token : `${window.location.origin}/invite/${token}`;
			inviteLinks = { ...inviteLinks, [childId]: link };
		} catch (err) {
			childError = getErrorMessage(err, 'Kunde inte skapa inbjudningslänk.');
		}
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text).catch(() => {
			// Fallback for older browsers
			const el = document.createElement('textarea');
			el.value = text;
			document.body.appendChild(el);
			el.select();
			document.execCommand('copy');
			document.body.removeChild(el);
		});
	}
</script>

<svelte:head>
	<title>Dashboard — Nioplugget</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8">
	<h1 class="mb-6 text-2xl font-bold text-foreground">Dashboard</h1>

	<!-- API Key section -->
	<Card class="mb-6">
		<CardHeader>
			<CardTitle>Claude API-nyckel</CardTitle>
			<CardDescription>
				För att ditt barn ska kunna använda Nioplugget behöver du koppla din Claude API-nyckel.
			</CardDescription>
		</CardHeader>
		<CardContent>
			{#if apiKeyError}
				<Alert variant="destructive" class="mb-4">
					<AlertDescription>{apiKeyError}</AlertDescription>
				</Alert>
			{/if}
			{#if apiKeySuccess}
				<Alert class="mb-4">
					<AlertDescription>{apiKeySuccess}</AlertDescription>
				</Alert>
			{/if}

			{#if apiKeyData === null}
				<p class="text-sm text-muted-foreground">Laddar...</p>
			{:else if !apiKeyData.hasKey}
				<!-- No API key yet -->
				<div class="mb-4 rounded-lg bg-muted p-4">
					<p class="mb-2 text-sm text-muted-foreground">
						Du har ingen API-nyckel sparad. Hämta en nyckel från
						<a
							href="https://console.anthropic.com"
							target="_blank"
							rel="noopener noreferrer"
							class="text-foreground underline-offset-4 hover:underline"
						>
							console.anthropic.com
						</a>
						och klistra in den här.
					</p>
					<p class="text-xs text-muted-foreground">
						Nyckeln krypteras och lagras säkert. Den används enbart för att kommunicera med Claude.
					</p>
				</div>
				<form onsubmit={handleStoreApiKey} class="flex flex-col gap-3">
					<div class="flex flex-col gap-1.5">
						<Label for="apikey">API-nyckel</Label>
						<Input
							id="apikey"
							type="password"
							bind:value={apiKeyInput}
							placeholder="sk-ant-..."
							autocomplete="off"
						/>
					</div>
					<Button type="submit" disabled={apiKeyLoading}>
						{apiKeyLoading ? 'Sparar och validerar...' : 'Spara API-nyckel'}
					</Button>
				</form>
			{:else if showUpdateForm}
				<!-- Update form -->
				<div class="mb-3 flex items-center gap-2">
					<span class="text-sm font-mono text-muted-foreground">{apiKeyData.masked}</span>
				</div>
				<form onsubmit={handleUpdateApiKey} class="flex flex-col gap-3">
					<div class="flex flex-col gap-1.5">
						<Label for="apikey-update">Ny API-nyckel</Label>
						<Input
							id="apikey-update"
							type="password"
							bind:value={apiKeyInput}
							placeholder="sk-ant-..."
							autocomplete="off"
						/>
					</div>
					<div class="flex gap-2">
						<Button type="submit" disabled={apiKeyLoading}>
							{apiKeyLoading ? 'Uppdaterar...' : 'Uppdatera'}
						</Button>
						<Button
							type="button"
							variant="outline"
							onclick={() => {
								showUpdateForm = false;
								apiKeyInput = '';
								apiKeyError = '';
							}}
						>
							Avbryt
						</Button>
					</div>
				</form>
			{:else}
				<!-- Has key — show masked + actions -->
				<div class="flex flex-col gap-3 rounded-lg border border-border bg-muted/30 p-3 sm:flex-row sm:items-center sm:justify-between">
					<div class="min-w-0">
						<p class="text-xs text-muted-foreground">Sparad nyckel</p>
						<p class="truncate font-mono text-sm text-foreground">{apiKeyData.masked}</p>
					</div>
					<div class="flex gap-2">
						<Button
							variant="outline"
							size="sm"
							class="min-h-[44px]"
							onclick={() => {
								showUpdateForm = true;
								apiKeyError = '';
								apiKeySuccess = '';
							}}
						>
							Uppdatera
						</Button>
						<Button
							variant="outline"
							size="sm"
							class="min-h-[44px] text-destructive hover:text-destructive"
							onclick={handleDeleteApiKey}
							disabled={apiKeyLoading}
						>
							Ta bort
						</Button>
					</div>
				</div>
			{/if}
		</CardContent>
	</Card>

	<!-- Children section -->
	<Card>
		<CardHeader>
			<CardTitle>Barn</CardTitle>
			<CardDescription>
				Lägg till ditt barn och skicka en inbjudningslänk.
			</CardDescription>
		</CardHeader>
		<CardContent>
			{#if childError}
				<Alert variant="destructive" class="mb-4">
					<AlertDescription>{childError}</AlertDescription>
				</Alert>
			{/if}

			<!-- Children list -->
			{#if childList.length > 0}
				<div class="mb-4 flex flex-col gap-3">
					{#each childList as child (child.id)}
						<div class="rounded-lg border border-border p-3">
							<div class="mb-2 flex items-center justify-between">
								<div>
									<p class="font-medium text-foreground">{child.name}</p>
									<p class="text-xs text-muted-foreground">
										{child.activated ? 'Aktiverad' : 'Väntar på aktivering'}
									</p>
								</div>
								<div class="flex items-center gap-2">
									{#if child.activated}
										<a
											href="/dashboard/child/{child.id}"
											class="text-sm text-muted-foreground transition-colors hover:text-foreground"
										>
											Se progress →
										</a>
									{:else}
										<Button
											variant="outline"
											size="sm"
											onclick={() => handleGenerateInvite(child.id)}
										>
											Skapa ny länk
										</Button>
									{/if}
								</div>
							</div>
							{#if inviteLinks[child.id]}
								<div class="mt-2 rounded bg-muted p-2">
									<p class="mb-1 text-xs text-muted-foreground">Inbjudningslänk:</p>
									<div class="flex items-center gap-2">
										<code class="flex-1 overflow-hidden text-ellipsis whitespace-nowrap text-xs">
											{inviteLinks[child.id]}
										</code>
										<Button
											variant="outline"
											size="sm"
											onclick={() => copyToClipboard(inviteLinks[child.id])}
										>
											Kopiera
										</Button>
									</div>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{:else}
				<p class="mb-4 text-sm text-muted-foreground">Inga barn tillagda ännu.</p>
			{/if}

			<!-- Add child form -->
			<form onsubmit={handleCreateChild} class="flex gap-2">
				<Input
					bind:value={newChildName}
					placeholder="Barnets namn"
					class="flex-1"
				/>
				<Button type="submit" disabled={childLoading}>
					{childLoading ? 'Lägger till...' : 'Lägg till barn'}
				</Button>
			</form>
		</CardContent>
	</Card>
</div>
