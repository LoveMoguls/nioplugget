<script lang="ts">
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
</script>

<svelte:head>
	<title>Nioplugget — Övningar för nationella provet</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-4rem)] items-center justify-center px-4 py-12">
	<div class="w-full max-w-3xl">
		<div class="mb-10 text-center">
			<h1 class="mb-3 text-4xl font-bold text-foreground">Nioplugget</h1>
			<p class="text-lg text-muted-foreground">
				Övningsapp för nationella provet i åk 9 — AI-lärare som ställer ledande frågor
			</p>
		</div>

		<div class="grid gap-6 sm:grid-cols-2">
			<Card class="cursor-pointer transition-shadow hover:shadow-md">
				<CardHeader>
					<CardTitle>Förälder</CardTitle>
					<CardDescription>
						Logga in eller registrera ett konto. Hantera din Claude API-nyckel och ditt barns profil.
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="flex flex-col gap-2">
						<Button href="/login" class="w-full">Logga in</Button>
						<Button href="/register" variant="outline" class="w-full">Registrera dig</Button>
					</div>
				</CardContent>
			</Card>

			<Card class="cursor-pointer transition-shadow hover:shadow-md">
				<CardHeader>
					<CardTitle>Elev</CardTitle>
					<CardDescription>
						Logga in med ditt namn och PIN-kod. Din förälder har skickat en inbjudningslänk.
					</CardDescription>
				</CardHeader>
				<CardContent>
					<Button href="/child/login" class="w-full">Logga in som elev</Button>
				</CardContent>
			</Card>
		</div>
	</div>
</div>
