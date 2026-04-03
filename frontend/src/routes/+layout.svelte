<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { user, isLoggedIn, isParent } from '$lib/stores/auth';

	let { children } = $props();

	onMount(async () => {
		await user.checkAuth();
	});
</script>

<div class="min-h-screen bg-background font-sans antialiased">
	{#if $isLoggedIn}
		<nav class="border-b border-border bg-card px-4 py-3">
			<div class="mx-auto flex max-w-4xl items-center justify-between">
				<a href="/" class="text-lg font-semibold text-foreground">Nioplugget</a>
				<div class="flex items-center gap-4">
					{#if $isParent}
						<a href="/dashboard" class="text-sm text-muted-foreground hover:text-foreground">
							Dashboard
						</a>
					{/if}
					<button
						onclick={async () => {
							await user.logout();
							goto('/');
						}}
						class="text-sm text-muted-foreground hover:text-foreground"
					>
						Logga ut
					</button>
				</div>
			</div>
		</nav>
	{/if}

	<main>
		{@render children()}
	</main>
</div>
