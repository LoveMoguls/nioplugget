<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { user, isLoggedIn, isParent } from '$lib/stores/auth';
	import { Menu, X } from '@lucide/svelte';

	let { children } = $props();
	let menuOpen = $state(false);

	onMount(() => {
		user.checkAuth();
	});

	// Close menu on navigation
	$effect(() => {
		// Track page changes to close menu
		$page.url;
		menuOpen = false;
	});
</script>

<div class="min-h-screen bg-background font-sans antialiased">
	{#if $isLoggedIn}
		<nav class="relative bg-gradient-to-r from-violet-600 to-indigo-600 shadow-md" style="padding-left: env(safe-area-inset-left, 0px); padding-right: env(safe-area-inset-right, 0px);">
			<div class="mx-auto flex max-w-4xl items-center justify-between px-4 py-3">
				<a href="/" class="text-lg font-semibold text-white tracking-wide">Nioplugget</a>

				<!-- Desktop nav -->
				<div class="hidden items-center gap-4 md:flex">
					{#if $isParent}
						<a href="/dashboard" class="min-h-[44px] flex items-center text-sm text-white/80 hover:text-white">
							Dashboard
						</a>
					{:else}
						<a href="/study" class="min-h-[44px] flex items-center text-sm text-white/80 hover:text-white">
							Plugga
						</a>
					{/if}
					<button
						onclick={async () => {
							await user.logout();
							goto('/profiler');
						}}
						class="min-h-[44px] flex items-center text-sm text-white/80 hover:text-white"
					>
						Byt profil
					</button>
				</div>

				<!-- Mobile hamburger -->
				<button
					class="flex min-h-[44px] min-w-[44px] items-center justify-center text-white md:hidden"
					onclick={() => (menuOpen = !menuOpen)}
					aria-expanded={menuOpen}
					aria-label="Meny"
				>
					{#if menuOpen}
						<X size={24} />
					{:else}
						<Menu size={24} />
					{/if}
				</button>
			</div>

			<!-- Mobile menu overlay -->
			{#if menuOpen}
				<div class="border-t border-white/20 bg-indigo-700 px-4 pb-4 md:hidden" style="padding-left: env(safe-area-inset-left, 0px); padding-right: env(safe-area-inset-right, 0px);">
					<div class="mx-auto flex max-w-4xl flex-col">
						{#if $isParent}
							<a
								href="/dashboard"
								class="flex min-h-[44px] items-center text-sm text-white/80 hover:text-white"
							>
								Dashboard
							</a>
						{:else}
							<a
								href="/study"
								class="flex min-h-[44px] items-center text-sm text-white/80 hover:text-white"
							>
								Plugga
							</a>
						{/if}
						<button
							onclick={async () => {
								menuOpen = false;
								await user.logout();
								goto('/profiler');
							}}
							class="flex min-h-[44px] items-center text-sm text-white/80 hover:text-white"
						>
							Byt profil
						</button>
					</div>
				</div>
			{/if}
		</nav>
	{/if}

	<main style="padding-bottom: env(safe-area-inset-bottom, 0px);">
		{@render children()}
	</main>
</div>
