<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, device } from '$lib/api';

	// Root redirect flow: logged-in users go to their home, otherwise fall
	// back to the device profile picker, or the unlock screen if the device
	// itself hasn't been unlocked yet.
	onMount(async () => {
		try {
			const me = (await auth.me()) as { role: string };
			goto(me.role === 'parent' ? '/dashboard' : '/study', { replaceState: true });
			return;
		} catch {}
		try {
			await device.profiles();
			goto('/profiler', { replaceState: true });
		} catch {
			goto('/las-upp', { replaceState: true });
		}
	});
</script>

<svelte:head>
	<title>Nioplugget — Plugga smartare inför nationella provet</title>
	<meta name="description" content="AI-lärare som guidar med frågor, aldrig ger direkta svar. Övningar anpassade efter Skolverkets centrala innehåll för åk 9." />
</svelte:head>
