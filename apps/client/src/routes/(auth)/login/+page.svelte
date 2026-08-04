<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, API_BASE, type AuthConfig, type AuthResponse } from '$lib/backend';
	import { Alert, Button, Field, Input, Spinner } from '@facile/lib';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);
	let ssoOnly = $state(false);
	let oidcEnabled = $state(false);

	$effect(() => {
		api.get<AuthConfig>('/auth/config')
			.then((config) => {
				ssoOnly = config.sso_only;
				oidcEnabled = config.oidc_enabled;
			})
			.catch(() => {});
	});

	async function handleLogin(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;

		try {
			const res = await api.post<AuthResponse>('/auth/login', { email, password });
			auth.login(res.token, res.user_id, email);
			goto('/spaces');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Casier — Sign in</title>
</svelte:head>

<div class="flex flex-col gap-5">
	<div class="flex flex-col gap-1 text-center">
		<h1 class="text-fc-xl font-semibold text-fc-fg">Welcome back</h1>
		<p class="text-fc-md text-fc-fg-muted">Sign in to manage your secrets</p>
	</div>

	{#if oidcEnabled}
		<Button class="w-full" onclick={() => (window.location.href = `${API_BASE}/auth/oidc`)}>
			<KeyRoundIcon class="size-4" aria-hidden="true" />
			Continue with SSO
		</Button>
		{#if !ssoOnly}
			<div class="flex items-center gap-3">
				<span class="h-px flex-1 bg-fc-fg/10"></span>
				<span class="text-fc-xs uppercase text-fc-fg-muted">or</span>
				<span class="h-px flex-1 bg-fc-fg/10"></span>
			</div>
		{/if}
	{/if}

	{#if !ssoOnly}
		<form onsubmit={handleLogin} class="flex flex-col gap-4">
			{#if error}
				<Alert tone="danger">{error}</Alert>
			{/if}

			<Field label="Email">
				<Input
					type="email"
					name="email"
					placeholder="you@example.com"
					bind:value={email}
					required
					autocomplete="email"
				/>
			</Field>

			<Field label="Password">
				<Input
					type="password"
					name="password"
					placeholder="Your password"
					bind:value={password}
					required
					minlength={8}
					autocomplete="current-password"
				/>
			</Field>

			<Button
				type="submit"
				class="w-full"
				variant={oidcEnabled ? 'outline' : 'primary'}
				disabled={loading}
			>
				{#if loading}<Spinner size="sm" />{/if}
				{loading ? 'Signing in…' : 'Sign in'}
			</Button>

			<div class="flex items-center justify-between text-fc-sm text-fc-fg-muted">
				<a href="/forgot-password" class="text-fc-accent hover:underline">Forgot password?</a>
				<span>
					No account?
					<a href="/register" class="text-fc-accent hover:underline">Register</a>
				</span>
			</div>
		</form>
	{/if}
</div>
