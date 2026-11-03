<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button, Card } from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { api, type AuthConfig } from '$lib/backend';
	import Logo from '$lib/components/Logo.svelte';

	type Feature = { icon: string; title: string; body: string };

	const features: Feature[] = [
		{
			icon: 'solar:lock-keyhole-linear',
			title: 'Encrypted at rest',
			body: 'Every project gets its own AES-256-GCM data key. Values are encrypted before they ever reach the database.'
		},
		{
			icon: 'solar:layers-linear',
			title: 'An environment per deploy',
			body: 'dev, staging and prod out of the box. Diff them, import a .env, roll any secret back to an earlier version.'
		},
		{
			icon: 'solar:code-square-linear',
			title: 'Made for CI',
			body: 'casier run injects secrets as env vars. Scoped API tokens for machines, a full audit log for humans.'
		}
	];

	const commands = ['casier login', 'casier init', 'casier run -- npm start'];

	let ready = $state(false);
	let ssoOnly = $state(false);

	const ctaHref = $derived(ssoOnly ? '/login' : '/register');
	const ctaLabel = $derived(ssoOnly ? 'Sign in with Facile' : 'Get started');
	const heroCtaLabel = $derived(ssoOnly ? 'Sign in with Facile' : 'Start managing secrets');

	onMount(async () => {
		if (!auth.loaded) {
			await auth.check();
		}

		if (auth.isAuthenticated) {
			goto('/projects', { replaceState: true });
			return;
		}

		try {
			const config = await api.get<AuthConfig>('/auth/config');
			ssoOnly = config.sso_only;
		} catch {}

		ready = true;
	});
</script>

<svelte:head>
	<title>Casier — Self-hosted secrets manager</title>
	<meta
		name="description"
		content="Casier is a self-hosted secrets and environment manager for teams. Encrypted at rest, versioned, audited — and one command away from your app."
	/>
</svelte:head>

{#if ready}
	<div class="flex min-h-dvh flex-col bg-background text-foreground">
		<header>
			<div class="mx-auto flex max-w-5xl items-center justify-between gap-4 px-6 py-4">
				<a href="/" class="text-xl sm:text-2xl">
					<Logo size={28} wordmark />
				</a>
				<div class="flex items-center gap-2">
					<Button href="/login" variant="ghost" size="sm" class="hidden sm:inline-flex rounded-fc-md">Log in</Button>
					<Button href={ctaHref} size="sm" class="rounded-fc-md">{ctaLabel}</Button>
				</div>
			</div>
		</header>

		<main class="flex-1">
			<section class="mx-auto max-w-5xl px-6 py-20 text-center sm:py-24">
				<h1
					class="font-heading text-4xl leading-[1.1] font-semibold tracking-tight sm:text-5xl lg:text-6xl"
				>
					Your secrets.<br />Your server.
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-base text-muted-foreground sm:text-lg">
					Casier is a self-hosted secrets and environment manager for teams. Encrypted at rest,
					versioned, audited — and one command away from your app.
				</p>
				<div class="mt-10 flex flex-col items-stretch gap-3 sm:flex-row sm:justify-center">
					<Button href={ctaHref} size="lg" iconRight="solar:alt-arrow-right-linear" class="rounded-fc-md">
						{heroCtaLabel}
					</Button>
					<Button href="/login" variant="outline" size="lg" class="rounded-fc-md">Log in</Button>
				</div>
			</section>


			<section class="mx-auto max-w-5xl px-6 py-16 sm:py-20">
				<div class="grid gap-4 md:grid-cols-3">
					{#each features as feature (feature.title)}
						<Card class="flex flex-col gap-4">
							<div
								class="flex size-10 items-center justify-center rounded-md bg-muted"
								aria-hidden="true"
							>
								<iconify-icon icon={feature.icon} width="20" height="20"></iconify-icon>
							</div>
							<div class="flex flex-col gap-1">
								<h3 class="text-base font-semibold">{feature.title}</h3>
								<p class="text-sm text-muted-foreground">{feature.body}</p>
							</div>
						</Card>
					{/each}
				</div>
			</section>


			<section class="mx-auto max-w-5xl px-6 py-16 sm:py-20">
				<h2 class="text-center font-heading text-2xl font-semibold tracking-tight sm:text-3xl">
					Three commands and you're done.
				</h2>
				<Card class="mx-auto mt-8 max-w-2xl overflow-x-auto">
					<div class="flex w-max min-w-full flex-col gap-1 font-mono text-sm">
						{#each commands as command (command)}
							<div><span class="text-muted-foreground">$</span> {command}</div>
						{/each}
					</div>
				</Card>
				<p class="mt-4 text-center text-sm text-muted-foreground">
					Secrets never touch your shell history or a .env file on disk.
				</p>
			</section>


			<section class="mx-auto max-w-5xl px-6 py-16 text-center sm:py-20">
				<h2 class="font-heading text-2xl font-semibold tracking-tight sm:text-3xl">
					{ssoOnly ? 'Ready to sign in?' : 'Ready to lock things down?'}
				</h2>
				<p class="mx-auto mt-4 max-w-xl text-muted-foreground">
					{ssoOnly
						? 'Use your Facile SSO to access Casier.'
						: 'Self-hosted, open by design. Bring your own server and keep your keys.'}
				</p>
				<div class="mt-8 flex justify-center">
					<Button href={ctaHref} size="lg" class="rounded-fc-md">
						{ssoOnly ? 'Sign in with Facile' : 'Create an account'}
					</Button>
				</div>
			</section>
		</main>

		<footer>
			<div
				class="mx-auto flex max-w-5xl flex-col items-center justify-between gap-2 px-6 py-6 sm:flex-row"
			>
				<p class="text-xs text-muted-foreground">
					© {new Date().getFullYear()} Casier by <a
						href="https://facile.studio"
						class="underline underline-offset-2 transition-colors hover:text-foreground"
						target="_blank"
						rel="noopener">Facile</a
					>.
				</p>
				<p class="text-xs text-muted-foreground">Part of the Facile Suite.</p>
			</div>
		</footer>
	</div>
{/if}
