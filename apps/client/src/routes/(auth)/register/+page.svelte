<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, API_BASE, type AuthConfig, type AuthResponse } from '$lib/backend';

	const inputClass =
		'flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50';

	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let loading = $state(false);
	let ssoOnly = $state(false);
	let oidcEnabled = $state(false);
	let configLoaded = $state(false);

	let passwordMismatch = $derived(confirmPassword.length > 0 && password !== confirmPassword);

	onMount(async () => {
		try {
			const config = await api.get<AuthConfig>('/auth/config');
			ssoOnly = config.sso_only;
			oidcEnabled = config.oidc_enabled;
		} catch {
			ssoOnly = false;
		}
		configLoaded = true;
	});

	async function handleRegister(e: Event) {
		e.preventDefault();
		error = '';

		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}

		loading = true;
		try {
			const res = await api.post<AuthResponse>('/auth/register', { email, password });
			auth.login(res.token, res.user_id, email);
			goto('/projects');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Registration failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Create account — Casier</title>
</svelte:head>

<div class="mb-8">
	<h1 class="font-heading text-2xl font-bold tracking-tight text-foreground">Create account</h1>
	<p class="mt-1.5 text-sm text-muted-foreground">
		{ssoOnly
			? 'Sign in with your organization account to access Casier.'
			: 'Sign up to start managing secrets.'}
	</p>
</div>

{#if !configLoaded}
	<div class="h-40"></div>
{:else}
	{#if !ssoOnly}
		<div class="mb-6 flex gap-1 rounded-lg border border-border bg-muted p-1">
			<a
				href="/login"
				class="flex-1 rounded-md py-1.5 text-center text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
			>
				Log in
			</a>
			<a
				href="/register"
				aria-current="page"
				class="flex-1 rounded-md bg-background py-1.5 text-center text-sm font-medium text-foreground shadow-sm transition-colors"
			>
				Register
			</a>
		</div>

		<form onsubmit={handleRegister} class="space-y-4">
			<div class="space-y-1.5">
				<label for="email" class="text-sm font-medium leading-none">Email</label>
				<input
					id="email"
					name="email"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
					autocomplete="email"
					class={inputClass}
				/>
			</div>

			<div class="space-y-1.5">
				<label for="password" class="text-sm font-medium leading-none">Password</label>
				<input
					id="password"
					name="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
					minlength={8}
					autocomplete="new-password"
					class={inputClass}
				/>
				<p class="text-xs text-muted-foreground">At least 8 characters.</p>
			</div>

			<div class="space-y-1.5">
				<label for="confirm-password" class="text-sm font-medium leading-none">
					Confirm password
				</label>
				<input
					id="confirm-password"
					name="confirm-password"
					type="password"
					bind:value={confirmPassword}
					placeholder="••••••••"
					required
					minlength={8}
					autocomplete="new-password"
					class={inputClass}
				/>
				{#if passwordMismatch}
					<p class="text-sm text-destructive">Passwords do not match</p>
				{/if}
			</div>

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<button
				type="submit"
				disabled={loading || passwordMismatch}
				class="inline-flex h-10 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50"
			>
				{loading ? 'Creating account…' : 'Create account'}
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
			class="inline-flex h-10 w-full items-center justify-center rounded-md border border-border bg-background px-4 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
		>
			Continue with SSO
		</button>
	{/if}

	<p class="mt-6 text-center text-sm text-muted-foreground">
		Already have an account?
		<a href="/login" class="underline underline-offset-2 transition-colors hover:text-foreground">
			Log in
		</a>
	</p>
{/if}
