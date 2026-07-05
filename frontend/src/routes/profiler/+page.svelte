<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { device, getErrorMessage } from '$lib/api';

	interface Profile {
		id: string;
		name: string;
		role: 'parent' | 'child';
	}

	const AVATAR_COLORS = ['#e11d48', '#7c3aed', '#0891b2', '#16a34a', '#d97706', '#db2777'];
	function avatarColor(name: string): string {
		let hash = 0;
		for (const ch of name) hash = (hash * 31 + ch.charCodeAt(0)) | 0;
		return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
	}

	let profiles = $state<Profile[]>([]);
	let loading = $state(true);
	let errorMsg = $state('');
	let selectingId = $state('');

	onMount(async () => {
		try {
			profiles = await device.profiles();
		} catch (err: unknown) {
			const e = err as { status?: number };
			if (e.status === 401) {
				goto('/las-upp');
				return;
			}
			errorMsg = getErrorMessage(err, 'Kunde inte hämta profiler.');
		} finally {
			loading = false;
		}
	});

	async function selectProfile(profile: Profile) {
		errorMsg = '';
		selectingId = profile.id;
		try {
			const data = await device.profileLogin(profile.id, profile.role);
			goto(data.role === 'parent' ? '/dashboard' : '/study');
		} catch (err) {
			errorMsg = getErrorMessage(err, 'Kunde inte logga in. Försök igen.');
		} finally {
			selectingId = '';
		}
	}
</script>

<svelte:head>
	<title>Vem pluggar? — Nioplugget</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-4rem)] flex-col items-center justify-center px-4 py-12">
	<h1 class="mb-8 text-2xl font-bold text-foreground sm:text-3xl">Vem pluggar?</h1>

	{#if errorMsg}
		<p class="mb-6 text-sm text-destructive">{errorMsg}</p>
	{/if}

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else}
		<div class="flex flex-wrap items-start justify-center gap-6">
			{#each profiles as profile (profile.id)}
				<button
					type="button"
					onclick={() => selectProfile(profile)}
					disabled={selectingId !== ''}
					class="flex w-24 flex-col items-center gap-2 rounded-lg p-2 text-center hover:bg-muted/50 disabled:opacity-50"
				>
					<span
						class="flex h-20 w-20 items-center justify-center rounded-full text-2xl font-bold text-white"
						style="background-color: {avatarColor(profile.name)}"
					>
						{profile.name.charAt(0).toUpperCase()}
					</span>
					<span class="text-sm font-medium text-foreground">{profile.name}</span>
				</button>
			{/each}
		</div>
	{/if}
</div>
