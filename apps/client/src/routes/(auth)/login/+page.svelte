<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/auth.svelte';
	import { api, API_BASE, type AuthConfig, type AuthResponse } from '$lib/backend';

	const inputClass =
		'flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50';
	const labelClass = 'text-sm font-medium leading-none';
	const primaryButtonClass =
		'inline-flex h-10 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50';
	const outlineButtonClass =
		'inline-flex h-10 w-full items-center justify-center rounded-md border border-border bg-background px-4 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);
	let ssoOnly = $state(false);
	let oidcEnabled = $state(false);
	let configLoaded = $state(false);

	onMount(async () => {
		const ssoError = $page.url.searchParams.get('error');
		if (ssoError) {
			error = ssoError;
		}

		try {
			const cfg = await fetch(`${API_BASE}/auth/config`, { credentials: 'include' }).then((r) => r.json());
			ssoOnly = cfg.sso_only ?? false;
			oidcEnabled = cfg.oidc_enabled ?? false;
		} catch {
			ssoOnly = false;
		}
		configLoaded = true;
	});

	async function handleLogin(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;

		try {
			const res = await api.post<AuthResponse>('/auth/login', { email, password });
			await auth.check();
			goto('/projects');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Log in — Casier</title>
</svelte:head>

<div class="mb-8">
	<h1 class="font-heading text-2xl font-bold tracking-tight text-foreground">Welcome back</h1>
	<p class="mt-1.5 text-sm text-muted-foreground">
		{ssoOnly
			? 'Sign in with your organization account to access Casier.'
			: 'Log in to your Casier account.'}
	</p>
</div>

{#if !configLoaded}
	<div class="h-40"></div>
{:else}
	{#if !ssoOnly}
		<div class="mb-6 flex rounded-lg border border-border bg-muted p-1 gap-1">
			<a
				href="/login"
				aria-current="page"
				class="flex-1 rounded-md py-1.5 text-center text-sm font-medium transition-colors bg-background text-foreground shadow-sm"
			>
				Log in
			</a>
			<a
				href="/register"
				class="flex-1 rounded-md py-1.5 text-center text-sm font-medium transition-colors text-muted-foreground hover:text-foreground"
			>
				Register
			</a>
		</div>

		<form onsubmit={handleLogin} class="space-y-4">
			<div class="space-y-1.5">
				<label for="email" class={labelClass}>Email</label>
				<input
					id="email"
					name="email"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
					autocomplete="email"
					disabled={loading}
					class={inputClass}
				/>
			</div>

			<div class="space-y-1.5">
				<label for="password" class={labelClass}>Password</label>
				<input
					id="password"
					name="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
					autocomplete="current-password"
					disabled={loading}
					class={inputClass}
				/>
			</div>

			{#if error}
				<p
					role="alert"
					class="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive"
				>
					{error}
				</p>
			{/if}

			<button type="submit" disabled={loading} aria-busy={loading} class={primaryButtonClass}>
				{loading ? 'Logging in…' : 'Log in'}
			</button>
		</form>
	{/if}

	{#if oidcEnabled}
		{#if !ssoOnly}
			<div class="my-5 flex items-center gap-3">
				<div class="h-px flex-1 bg-border"></div>
				<span class="text-xs text-muted-foreground">or</span>
				<div class="h-px flex-1 bg-border"></div>
			</div>
		{/if}

		<button
			type="button"
			onclick={() => (window.location.href = `${API_BASE}/auth/oidc`)}
			class={outlineButtonClass}
		>
			Continue with SSO
		</button>
	{/if}

	{#if ssoOnly && !oidcEnabled}
		<p role="alert" class="text-sm text-destructive">
			SSO is not configured. Contact your administrator.
		</p>
	{/if}

	{#if !ssoOnly}
		<p class="mt-6 text-center text-sm text-muted-foreground">
			<a
				href="/forgot-password"
				class="underline underline-offset-2 transition-colors hover:text-foreground"
			>
				Forgot your password?
			</a>
		</p>
	{/if}
{/if}
