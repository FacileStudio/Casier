<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, API_BASE, type AuthConfig, type AuthResponse } from '$lib/backend';
	import { Alert, Button, Field, Input, Spinner } from '@facile/lib';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';

	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let loading = $state(false);
	let ssoOnly = $state(false);
	let oidcEnabled = $state(false);

	let passwordMismatch = $derived(confirmPassword.length > 0 && password !== confirmPassword);

	$effect(() => {
		api.get<AuthConfig>('/auth/config')
			.then((config) => {
				ssoOnly = config.sso_only;
				oidcEnabled = config.oidc_enabled;
			})
			.catch(() => {});
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
			goto('/spaces');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Registration failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Casier — Create account</title>
</svelte:head>

<div class="flex flex-col gap-5">
	<div class="flex flex-col gap-1 text-center">
		<h1 class="text-fc-xl font-semibold text-fc-fg">Create account</h1>
		<p class="text-fc-md text-fc-fg-muted">Get started with Casier</p>
	</div>

	{#if oidcEnabled}
		<Button class="w-full" onclick={() => (window.location.href = `${API_BASE}/auth/oidc`)}>
			<KeyRoundIcon class="size-4" aria-hidden="true" />
			Sign up with SSO
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
		<form onsubmit={handleRegister} class="flex flex-col gap-4">
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

			<Field label="Password" helper="At least 8 characters">
				<Input
					type="password"
					name="password"
					placeholder="Password"
					bind:value={password}
					required
					minlength={8}
					autocomplete="new-password"
				/>
			</Field>

			<Field label="Confirm password" error={passwordMismatch ? 'Passwords do not match' : undefined}>
				<Input
					type="password"
					name="confirm-password"
					placeholder="Confirm password"
					bind:value={confirmPassword}
					required
					minlength={8}
					autocomplete="new-password"
				/>
			</Field>

			<Button
				type="submit"
				class="w-full"
				variant={oidcEnabled ? 'outline' : 'primary'}
				disabled={loading || passwordMismatch}
			>
				{#if loading}<Spinner size="sm" />{/if}
				{loading ? 'Creating account…' : 'Create account'}
			</Button>

			<p class="text-center text-fc-sm text-fc-fg-muted">
				Already have an account?
				<a href="/login" class="text-fc-accent hover:underline">Sign in</a>
			</p>
		</form>
	{/if}
</div>
