<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { apiKey as apiKeyApi, children as childrenApi, challenges as challengesApi, getErrorMessage } from '$lib/api';
	import { user, isLoggedIn, isParent } from '$lib/stores/auth';
	import { browser } from '$app/environment';

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

	// Challenges state
	interface ChallengeItem {
		id: string;
		title: string;
		description: string;
		coverEmoji: string;
		createdAt: string;
	}
	interface ChallengeDraft {
		id: string;
		title: string;
		description: string;
		coverEmoji: string;
	}
	let challengeList = $state<ChallengeItem[]>([]);
	let challengeFiles = $state<File[]>([]);
	let challengePreviews = $state<string[]>([]);
	let challengeLoading = $state(false);
	let challengeError = $state('');
	let challengeSuccess = $state('');
	let challengeStep = $state('');
	let challengeElapsed = $state(0);
	let elapsedInterval = $state<ReturnType<typeof setInterval> | null>(null);
	let draft = $state<ChallengeDraft | null>(null);
	let draftTitle = $state('');
	let publishLoading = $state(false);

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
		await Promise.all([loadApiKey(), loadChildren(), loadChallenges()]);
	}

	async function loadChallenges() {
		try {
			const data = (await challengesApi.list()) as ChallengeItem[] | null;
			challengeList = data || [];
		} catch {
			challengeList = [];
		}
	}

	async function handleCreateChallenge(e: Event) {
		e.preventDefault();
		challengeError = '';
		challengeSuccess = '';
		if (challengeFiles.length === 0) {
			challengeError = 'Välj minst en bild.';
			return;
		}
		challengeLoading = true;
		challengeElapsed = 0;
		challengeStep = 'Skickar bilder...';
		elapsedInterval = setInterval(() => {
			challengeElapsed += 1;
			if (challengeElapsed === 3) challengeStep = 'Claude läser bilderna...';
			else if (challengeElapsed === 10) challengeStep = 'Analyserar innehållet...';
			else if (challengeElapsed === 25) challengeStep = 'Skapar övningar...';
			else if (challengeElapsed === 45) challengeStep = 'Nästan klart...';
		}, 1000);
		try {
			const created = (await challengesApi.create(challengeFiles)) as ChallengeItem;
			draft = { id: created.id, title: created.title, description: created.description, coverEmoji: created.coverEmoji };
			draftTitle = created.title;
			challengeFiles = [];
			challengePreviews.forEach(URL.revokeObjectURL);
			challengePreviews = [];
			const input = document.getElementById('challenge-files') as HTMLInputElement;
			if (input) input.value = '';
		} catch (err) {
			challengeError = getErrorMessage(err, 'Kunde inte skapa utmaningen. Försök med tydligare foton.');
		} finally {
			challengeLoading = false;
			challengeStep = '';
			if (elapsedInterval) { clearInterval(elapsedInterval); elapsedInterval = null; }
		}
	}

	async function handlePublish() {
		if (!draft) return;
		publishLoading = true;
		challengeError = '';
		try {
			const published = (await challengesApi.publish(draft.id, draftTitle.trim() || draft.title)) as ChallengeItem;
			challengeList = [published, ...challengeList];
			challengeSuccess = `✓ Publicerad: ${published.title}`;
			draft = null;
		} catch (err) {
			challengeError = getErrorMessage(err, 'Kunde inte publicera utmaningen.');
		} finally {
			publishLoading = false;
		}
	}

	async function handleDiscardDraft() {
		if (!draft) return;
		try { await challengesApi.delete(draft.id); } catch { /* ignore */ }
		draft = null;
	}

	async function handleDeleteChallenge(id: string) {
		if (!confirm('Ta bort utmaningen?')) return;
		try {
			await challengesApi.delete(id);
			challengeList = challengeList.filter((c) => c.id !== id);
		} catch {
			challengeError = 'Kunde inte ta bort utmaningen.';
		}
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
			const data = (await childrenApi.generateInvite(childId)) as { inviteURL?: string; inviteUrl?: string; inviteToken?: string };
			const link = data.inviteURL || data.inviteUrl || data.inviteToken || '';
			inviteLinks = { ...inviteLinks, [childId]: link };
		} catch (err) {
			childError = getErrorMessage(err, 'Kunde inte skapa inbjudningslänk.');
		}
	}

	async function handleLoginAs(childId: string) {
		try {
			const data = (await childrenApi.loginAs(childId)) as {
				id: string;
				name: string;
				parentEmail?: string;
			};
			if (browser && data?.parentEmail) {
				localStorage.setItem('childParentEmail', data.parentEmail);
			}
			user.setUser({ id: data.id, email: '', role: 'child' });
			goto('/study');
		} catch (err) {
			childError = getErrorMessage(err, 'Kunde inte starta session.');
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
										<div class="flex items-center gap-2">
											<Button
												variant="default"
												size="sm"
												onclick={() => handleLoginAs(child.id)}
											>
												Starta session
											</Button>
											<a
												href="/dashboard/child/{child.id}"
												class="text-sm text-muted-foreground transition-colors hover:text-foreground"
											>
												Se progress →
											</a>
										</div>
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

	<!-- Challenges section -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle>Utmaningar</CardTitle>
			<CardDescription>
				Fotografera prov, läxor eller övningsuppgifter så skapar AI:n en interaktiv utmaning.
			</CardDescription>
		</CardHeader>
		<CardContent>
			{#if challengeError}
				<Alert variant="destructive" class="mb-4">
					<AlertDescription>{challengeError}</AlertDescription>
				</Alert>
			{/if}
			{#if challengeSuccess}
				<Alert class="mb-4">
					<AlertDescription>{challengeSuccess}</AlertDescription>
				</Alert>
			{/if}

			<!-- Existing challenges -->
			{#if challengeList.length > 0}
				<div class="mb-4 flex flex-col gap-2">
					{#each challengeList as challenge}
						<div class="flex items-center justify-between rounded-lg border border-border p-3">
							<div class="min-w-0 flex-1">
								<p class="font-medium">{challenge.coverEmoji} {challenge.title}</p>
								<p class="text-xs text-muted-foreground">{challenge.description}</p>
							</div>
							<Button
								variant="outline"
								size="sm"
								class="ml-3 text-destructive hover:text-destructive"
								onclick={() => handleDeleteChallenge(challenge.id)}
							>
								Ta bort
							</Button>
						</div>
					{/each}
				</div>
			{:else}
				<p class="mb-4 text-sm text-muted-foreground">Inga utmaningar än.</p>
			{/if}

			<!-- Draft review step -->
			{#if draft}
				<div class="mb-4 rounded-lg border-2 border-primary/30 bg-primary/5 p-4">
					<p class="mb-3 text-sm font-medium">Granska och namnge utmaningen</p>
					<div class="mb-2 flex items-center gap-2 text-2xl">{draft.coverEmoji}</div>
					<div class="mb-3 flex flex-col gap-1.5">
						<Label for="draft-title">Namn på utmaningen</Label>
						<Input id="draft-title" bind:value={draftTitle} placeholder={draft.title} />
						<p class="text-xs text-muted-foreground">{draft.description}</p>
					</div>
					<div class="flex gap-2">
						<Button onclick={handlePublish} disabled={publishLoading}>
							{publishLoading ? 'Publicerar...' : 'Publicera läxa'}
						</Button>
						<Button variant="outline" onclick={handleDiscardDraft} disabled={publishLoading}>
							Kasta bort
						</Button>
					</div>
				</div>
			{/if}

			<!-- Upload form -->
			{#if !draft}
			<form onsubmit={handleCreateChallenge} class="flex flex-col gap-3">
				<div class="flex flex-col gap-1.5">
					<Label for="challenge-files">Bilder (1–6 st, max 5 MB/bild)</Label>
					<input
						id="challenge-files"
						type="file"
						accept="image/*"
						multiple
						disabled={challengeLoading}
						class="block w-full cursor-pointer rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground file:mr-3 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:opacity-50"
						onchange={(e) => {
							const input = e.currentTarget as HTMLInputElement;
							challengeFiles = input.files ? Array.from(input.files).slice(0, 6) : [];
							challengePreviews.forEach(URL.revokeObjectURL);
							challengePreviews = challengeFiles.map((f) => URL.createObjectURL(f));
						}}
					/>
					{#if challengePreviews.length > 0 && !challengeLoading}
						<div class="flex flex-wrap gap-2 pt-1">
							{#each challengePreviews as src, i}
								<div class="relative">
									<img
										{src}
										alt="Bild {i + 1}"
										class="h-24 w-24 rounded-md border border-border object-cover"
									/>
									<button
										type="button"
										onclick={() => {
											URL.revokeObjectURL(challengePreviews[i]);
											challengeFiles = challengeFiles.filter((_, j) => j !== i);
											challengePreviews = challengePreviews.filter((_, j) => j !== i);
										}}
										class="absolute -right-1.5 -top-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-destructive text-xs text-destructive-foreground"
									>✕</button>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				{#if challengeLoading}
					<div class="rounded-lg border border-border bg-muted/30 p-4">
						<div class="mb-3 flex items-center gap-3">
							<svg class="h-5 w-5 animate-spin text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
							</svg>
							<span class="text-sm font-medium">{challengeStep}</span>
						</div>
						<div class="mb-1 h-1.5 w-full overflow-hidden rounded-full bg-muted">
							<div
								class="h-full rounded-full bg-primary transition-all duration-1000"
								style="width: {Math.min(95, challengeElapsed * 1.8)}%"
							></div>
						</div>
						<p class="text-right text-xs text-muted-foreground">{challengeElapsed}s</p>
					</div>
				{:else}
					<Button type="submit" disabled={challengeFiles.length === 0}>
						Skapa utmaning
					</Button>
				{/if}
			</form>
			{/if}
		</CardContent>
	</Card>
</div>
