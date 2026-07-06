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
				goto('/las-upp', { replaceState: true });
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
			if ((err as { status?: number })?.status === 401) {
				goto('/las-upp');
				return;
			}
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
	<h1 class="font-display mb-10 text-4xl font-extrabold tracking-wider text-primary uppercase">
		Vem pluggar?
	</h1>

	{#if errorMsg}
		<p class="mb-6 text-sm text-destructive">{errorMsg}</p>
	{/if}

	{#if loading}
		<p class="text-muted-foreground">Laddar...</p>
	{:else}
		<div
			class="mx-auto grid w-full max-w-md grid-cols-2 place-items-center gap-6 sm:max-w-lg sm:grid-cols-3 md:max-w-2xl md:grid-cols-4"
		>
			{#each profiles as profile (profile.id)}
				<button
					type="button"
					onclick={() => selectProfile(profile)}
					disabled={selectingId !== ''}
					class="group flex w-24 flex-col items-center gap-2 rounded-lg p-2 text-center disabled:opacity-50"
				>
					<div
						class="avatar-ring rounded-full p-[3px] transition-transform duration-200 group-hover:scale-[1.08] group-hover:[box-shadow:var(--glow-cyan)] group-focus-visible:scale-[1.08] group-focus-visible:[box-shadow:var(--glow-cyan)]"
					>
						<div class="rounded-full bg-card p-[3px]">
							<span
								class="flex h-24 w-24 items-center justify-center rounded-full text-2xl font-bold text-white"
								style="background-color: {avatarColor(profile.name)}"
							>
								{profile.name.charAt(0).toUpperCase()}
							</span>
						</div>
					</div>
					<span class="font-display text-sm font-bold text-foreground">{profile.name}</span>
				</button>
			{/each}
		</div>
	{/if}
</div>

<style>
	.avatar-ring {
		background: conic-gradient(
			from var(--ring-angle, 0deg),
			oklch(0.85 0.15 195),
			oklch(0.7 0.25 330),
			oklch(0.85 0.22 135),
			oklch(0.85 0.15 195)
		);
	}

	@property --ring-angle {
		syntax: '<angle>';
		initial-value: 0deg;
		inherits: false;
	}

	@media (prefers-reduced-motion: no-preference) {
		.avatar-ring {
			animation: spin-ring 6s linear infinite;
		}

		@keyframes spin-ring {
			to {
				--ring-angle: 360deg;
			}
		}
	}
</style>
