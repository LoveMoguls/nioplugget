<!-- frontend/src/lib/components/challenges/ChallengeUpload.svelte -->
<!-- One friendly box that takes anything: drag & drop, camera, file picker,
     Ctrl+V-pasted screenshots or text. Used by both parent and child. -->
<script lang="ts">
	import { onDestroy } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { challenges, getErrorMessage } from '$lib/api';

	export interface CreatedChallenge {
		id: string;
		title: string;
		description: string;
		coverEmoji: string;
		published: boolean;
		createdBy: string;
		createdAt: string;
	}

	interface Props {
		onCreated: (challenge: CreatedChallenge) => void;
	}

	let { onCreated }: Props = $props();

	let files = $state<File[]>([]);
	let previews = $state<(string | null)[]>([]);
	let text = $state('');
	let loading = $state(false);
	let error = $state('');
	let step = $state('');
	let elapsed = $state(0);
	let dragOver = $state(false);
	let interval: ReturnType<typeof setInterval> | null = null;

	let cameraInput: HTMLInputElement | undefined = $state();
	let fileInput: HTMLInputElement | undefined = $state();

	const MAX_FILES = 6;

	function fileKind(f: File): 'image' | 'pdf' | 'text' | null {
		const name = f.name.toLowerCase();
		if (f.type.startsWith('image/')) return 'image';
		if (f.type === 'application/pdf' || name.endsWith('.pdf')) return 'pdf';
		if (
			f.type === 'text/plain' ||
			f.type === 'text/markdown' ||
			name.endsWith('.txt') ||
			name.endsWith('.md')
		)
			return 'text';
		return null;
	}

	function addFiles(incoming: File[]) {
		error = '';
		for (const f of incoming) {
			if (files.length >= MAX_FILES) {
				error = `Max ${MAX_FILES} filer.`;
				break;
			}
			const kind = fileKind(f);
			if (!kind) {
				error = `${f.name}: filtypen stöds inte — använd bilder, PDF eller text.`;
				continue;
			}
			const limit = kind === 'pdf' ? 10 << 20 : kind === 'text' ? 1 << 20 : 5 << 20;
			if (f.size > limit) {
				error = `${f.name} är för stor (max ${Math.round(limit / (1 << 20))} MB).`;
				continue;
			}
			files = [...files, f];
			previews = [...previews, kind === 'image' ? URL.createObjectURL(f) : null];
		}
	}

	function removeFile(i: number) {
		const url = previews[i];
		if (url) URL.revokeObjectURL(url);
		files = files.filter((_, j) => j !== i);
		previews = previews.filter((_, j) => j !== i);
	}

	function handlePaste(e: ClipboardEvent) {
		if (loading) return;
		const pastedFiles = Array.from(e.clipboardData?.files ?? []);
		if (pastedFiles.length > 0) {
			e.preventDefault();
			addFiles(pastedFiles);
			return;
		}
		// Pasted text lands in the textarea by default; only catch it elsewhere
		const target = e.target as HTMLElement | null;
		if (target?.tagName === 'TEXTAREA' || target?.tagName === 'INPUT') return;
		const pasted = e.clipboardData?.getData('text') ?? '';
		if (pasted.trim()) {
			e.preventDefault();
			text = text ? text + '\n' + pasted : pasted;
		}
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragOver = false;
		if (loading) return;
		addFiles(Array.from(e.dataTransfer?.files ?? []));
	}

	async function handleCreate() {
		if (files.length === 0 && !text.trim()) {
			error = 'Lägg till en bild, fil eller lite text först.';
			return;
		}
		loading = true;
		error = '';
		elapsed = 0;
		step = 'Skickar läxan...';
		interval = setInterval(() => {
			elapsed += 1;
			if (elapsed === 3) step = 'Claude läser läxan...';
			else if (elapsed === 10) step = 'Analyserar innehållet...';
			else if (elapsed === 25) step = 'Skapar övningar...';
			else if (elapsed === 45) step = 'Nästan klart...';
		}, 1000);
		try {
			const created = (await challenges.create(files, text)) as CreatedChallenge;
			previews.forEach((url) => url && URL.revokeObjectURL(url));
			files = [];
			previews = [];
			text = '';
			onCreated(created);
		} catch (err) {
			error = getErrorMessage(err, 'Kunde inte skapa utmaningen. Försök med tydligare material.');
		} finally {
			loading = false;
			step = '';
			if (interval) {
				clearInterval(interval);
				interval = null;
			}
		}
	}

	onDestroy(() => {
		if (interval) clearInterval(interval);
		previews.forEach((url) => url && URL.revokeObjectURL(url));
	});
</script>

<svelte:window onpaste={handlePaste} />

<div class="flex flex-col gap-3">
	{#if error}
		<p class="text-sm text-destructive">{error}</p>
	{/if}

	{#if loading}
		<div class="rounded-xl border border-border bg-muted/30 p-4">
			<div class="mb-3 flex items-center gap-3">
				<svg class="h-5 w-5 animate-spin text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
					<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
				</svg>
				<span class="text-sm font-medium">{step}</span>
			</div>
			<div class="mb-1 h-1.5 w-full overflow-hidden rounded-full bg-muted">
				<div
					class="h-full rounded-full bg-primary transition-all duration-1000"
					style="width: {Math.min(95, elapsed * 1.8)}%"
				></div>
			</div>
			<p class="text-right text-xs text-muted-foreground">{elapsed}s</p>
		</div>
	{:else}
		<!-- Drop zone -->
		<div
			class="cursor-pointer rounded-xl border-2 border-dashed p-5 text-center transition-colors
				{dragOver ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/60'}"
			role="button"
			tabindex="0"
			ondrop={handleDrop}
			ondragover={(e) => {
				e.preventDefault();
				dragOver = true;
			}}
			ondragleave={() => (dragOver = false)}
			onclick={() => fileInput?.click()}
			onkeydown={(e) => e.key === 'Enter' && fileInput?.click()}
		>
			<p class="text-3xl">📚</p>
			<p class="mt-1 text-sm font-medium">Släpp läxan här</p>
			<p class="mt-0.5 text-xs text-muted-foreground">
				Foto, PDF eller textfil — eller klistra in direkt med Ctrl+V
			</p>
		</div>

		<div class="flex gap-2">
			<Button variant="outline" class="flex-1" onclick={() => cameraInput?.click()}>
				📷 Ta foto
			</Button>
			<Button variant="outline" class="flex-1" onclick={() => fileInput?.click()}>
				📁 Välj filer
			</Button>
		</div>

		<input
			bind:this={cameraInput}
			type="file"
			accept="image/*"
			capture="environment"
			class="hidden"
			onchange={(e) => {
				const input = e.currentTarget as HTMLInputElement;
				addFiles(Array.from(input.files ?? []));
				input.value = '';
			}}
		/>
		<input
			bind:this={fileInput}
			type="file"
			accept="image/*,.pdf,.txt,.md,application/pdf,text/plain"
			multiple
			class="hidden"
			onchange={(e) => {
				const input = e.currentTarget as HTMLInputElement;
				addFiles(Array.from(input.files ?? []));
				input.value = '';
			}}
		/>

		<!-- Previews -->
		{#if files.length > 0}
			<div class="flex flex-wrap gap-2">
				{#each files as f, i}
					<div class="relative">
						{#if previews[i]}
							<img
								src={previews[i]}
								alt={f.name}
								class="h-20 w-20 rounded-lg border border-border object-cover"
							/>
						{:else}
							<div
								class="flex h-20 w-20 flex-col items-center justify-center rounded-lg border border-border bg-muted/40 px-1"
							>
								<span class="text-2xl">{fileKind(f) === 'pdf' ? '📄' : '📝'}</span>
								<span class="w-full truncate text-center text-[10px] text-muted-foreground">{f.name}</span>
							</div>
						{/if}
						<button
							type="button"
							onclick={() => removeFile(i)}
							class="absolute -right-1.5 -top-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-destructive text-xs text-destructive-foreground"
							aria-label="Ta bort {f.name}"
						>✕</button>
					</div>
				{/each}
			</div>
		{/if}

		<textarea
			bind:value={text}
			rows={text ? 5 : 2}
			placeholder="...eller skriv/klistra in läxtexten här"
			class="w-full resize-y rounded-xl border border-input bg-background px-3 py-2 text-sm
				focus:border-primary focus:outline-none focus:ring-1 focus:ring-ring"
		></textarea>

		<Button onclick={handleCreate} disabled={files.length === 0 && !text.trim()}>
			Skapa utmaning ✨
		</Button>
	{/if}
</div>
