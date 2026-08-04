<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, COOKIE_SESSION, type AuthConfig, type Me } from '$lib/backend';
	import Logo from '$lib/components/Logo.svelte';

	let ready = $state(false);
	let ssoOnly = $state(false);

	const ctaHref = $derived(ssoOnly ? '/login' : '/register');
	const ctaLabel = $derived(ssoOnly ? 'Continue with SSO' : 'Get started');
	const heroCtaLabel = $derived(ssoOnly ? 'Continue with SSO' : 'Start managing secrets');

	onMount(async () => {
		if (auth.isAuthenticated) {
			goto('/projects', { replaceState: true });
			return;
		}

		try {
			const me = await api.get<Me>('/auth/me');
			auth.login(COOKIE_SESSION, me.id, me.email);
			goto('/projects', { replaceState: true });
			return;
		} catch {}

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
		<header class="border-b border-border">
			<div class="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
				<a href="/" class="text-2xl">
					<Logo size={28} wordmark />
				</a>
				<div class="flex items-center gap-2">
					<a
						href="/login"
						class="hidden h-9 items-center justify-center gap-2 rounded-md px-4 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground sm:inline-flex"
					>
						Log in
					</a>
					<a
						href={ctaHref}
						class="inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
					>
						{ctaLabel}
					</a>
				</div>
			</div>
		</header>

		<main class="flex-1">
			<section class="mx-auto max-w-5xl px-6 py-24 text-center">
				<h1 class="text-5xl font-bold font-heading tracking-tight leading-[1.1] sm:text-6xl">
					Your secrets.<br />Your server.
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-lg text-muted-foreground">
					Casier is a self-hosted secrets and environment manager for teams. Encrypted at rest,
					versioned, audited — and one command away from your app.
				</p>
				<div class="mt-10 flex flex-wrap justify-center gap-3">
					<a
						href={ctaHref}
						class="inline-flex h-11 items-center justify-center gap-2 rounded-md bg-primary px-6 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
					>
						{heroCtaLabel}
						<iconify-icon icon="solar:alt-arrow-right-linear" width="16" aria-hidden="true"
						></iconify-icon>
					</a>
					<a
						href="/login"
						class="inline-flex h-11 items-center justify-center gap-2 rounded-md border border-border bg-background px-6 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
					>
						Log in
					</a>
				</div>
			</section>

			<div class="mx-auto max-w-5xl"><div class="h-px bg-border"></div></div>

			<section class="mx-auto max-w-5xl px-6 py-20">
				<div class="grid gap-6 md:grid-cols-3">
					<div class="rounded-lg border border-border p-6">
						<div
							class="mb-3 flex h-10 w-10 items-center justify-center rounded-md border border-border"
						>
							<iconify-icon icon="solar:lock-keyhole-linear" width="20" aria-hidden="true"
							></iconify-icon>
						</div>
						<h3 class="text-base font-semibold">Encrypted at rest</h3>
						<p class="mt-1.5 text-sm text-muted-foreground">
							Every project gets its own AES-256-GCM data key. Values are encrypted before they ever
							reach the database.
						</p>
					</div>

					<div class="rounded-lg border border-border p-6">
						<div
							class="mb-3 flex h-10 w-10 items-center justify-center rounded-md border border-border"
						>
							<iconify-icon icon="solar:layers-linear" width="20" aria-hidden="true"></iconify-icon>
						</div>
						<h3 class="text-base font-semibold">An environment per deploy</h3>
						<p class="mt-1.5 text-sm text-muted-foreground">
							dev, staging and prod out of the box. Diff them, import a .env, roll any secret back to
							an earlier version.
						</p>
					</div>

					<div class="rounded-lg border border-border p-6">
						<div
							class="mb-3 flex h-10 w-10 items-center justify-center rounded-md border border-border"
						>
							<iconify-icon icon="solar:code-square-linear" width="20" aria-hidden="true"
							></iconify-icon>
						</div>
						<h3 class="text-base font-semibold">Made for CI</h3>
						<p class="mt-1.5 text-sm text-muted-foreground">
							<code class="rounded bg-muted px-2 py-1 font-mono text-xs">casier run</code> injects
							secrets as env vars. Scoped API tokens for machines, a full audit log for humans.
						</p>
					</div>
				</div>
			</section>

			<div class="mx-auto max-w-5xl"><div class="h-px bg-border"></div></div>

			<section class="mx-auto max-w-5xl px-6 py-20">
				<h2 class="text-center text-3xl font-bold font-heading tracking-tight">
					Three commands and you're done.
				</h2>
				<div
					class="mx-auto mt-8 max-w-2xl overflow-x-auto rounded-lg border border-border bg-muted p-6 font-mono text-sm"
				>
					<div><span class="text-muted-foreground">$</span> <span class="text-foreground">casier login</span></div>
					<div><span class="text-muted-foreground">$</span> <span class="text-foreground">casier init</span></div>
					<div><span class="text-muted-foreground">$</span> <span class="text-foreground">casier run -- npm start</span></div>
				</div>
				<p class="mt-4 text-center text-sm text-muted-foreground">
					Secrets never touch your shell history or a .env file on disk.
				</p>
			</section>

			<div class="mx-auto max-w-5xl"><div class="h-px bg-border"></div></div>

			<section class="mx-auto max-w-5xl px-6 py-20 text-center">
				<h2 class="text-3xl font-bold font-heading tracking-tight">
					{ssoOnly ? 'Ready to sign in?' : 'Ready to lock things down?'}
				</h2>
				<p class="mt-4 text-muted-foreground">
					{ssoOnly
						? 'Use your organization SSO to access Casier.'
						: 'Self-hosted, open by design. Bring your own server and keep your keys.'}
				</p>
				<a
					href={ctaHref}
					class="mt-8 inline-flex h-11 items-center justify-center gap-2 rounded-md bg-primary px-6 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
				>
					{ssoOnly ? 'Continue with SSO' : 'Create an account'}
				</a>
			</section>
		</main>

		<footer class="border-t border-border">
			<div
				class="mx-auto flex max-w-5xl flex-col items-center justify-between gap-2 px-6 py-6 sm:flex-row"
			>
				<p class="text-xs text-muted-foreground">
					© {new Date().getFullYear()} Casier by <a
						href="https://facile.studio"
						class="underline underline-offset-2 hover:text-foreground transition-colors"
						target="_blank"
						rel="noopener">Facile</a
					>.
				</p>
				<p class="text-xs text-muted-foreground">Part of the Facile Suite.</p>
			</div>
		</footer>
	</div>
{/if}
