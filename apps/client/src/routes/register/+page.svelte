<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, API_BASE } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
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
		api.get<{ sso_only: boolean; oidc_enabled: boolean }>('/auth/config')
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
			const res = await api.post<{ user_id: string; token: string }>(
				'/auth/register',
				{ email, password }
			);
			auth.login(res.token, res.user_id, email);
			goto('/spaces');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Registration failed';
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center px-4">
	<Card class="w-full max-w-sm">
		<CardHeader class="text-center">
			<div class="mx-auto mb-2 flex h-10 w-10 items-center justify-center rounded-lg bg-primary text-primary-foreground">
				<KeyRoundIcon class="size-5" />
			</div>
			<CardTitle class="text-xl">Create account</CardTitle>
			<CardDescription>Get started with Clef</CardDescription>
		</CardHeader>
		<CardContent>
			{#if oidcEnabled}
				<div class="flex flex-col gap-4">
					<Button class="w-full" onclick={() => window.location.href = `${API_BASE}/auth/oidc`}>
						Sign up with SSO
					</Button>
				</div>
			{/if}

			{#if !ssoOnly}
				{#if oidcEnabled}
					<div class="relative my-4">
						<div class="absolute inset-0 flex items-center"><span class="w-full border-t border-border"></span></div>
						<div class="relative flex justify-center text-xs uppercase"><span class="bg-card px-2 text-muted-foreground">or</span></div>
					</div>
				{/if}

				<form onsubmit={handleRegister} class="flex flex-col gap-4">
					{#if error}
						<div class="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
							{error}
						</div>
					{/if}

					<div class="flex flex-col gap-1.5">
						<Label for="email">Email</Label>
						<Input
							id="email"
							type="email"
							placeholder="you@example.com"
							bind:value={email}
							required
							autocomplete="email"
						/>
					</div>

					<div class="flex flex-col gap-1.5">
						<Label for="password">Password</Label>
						<Input
							id="password"
							type="password"
							placeholder="Password"
							bind:value={password}
							required
							minlength={8}
							autocomplete="new-password"
						/>
					</div>

					<div class="flex flex-col gap-1.5">
						<Label for="confirm-password">Confirm password</Label>
						<Input
							id="confirm-password"
							type="password"
							placeholder="Confirm password"
							bind:value={confirmPassword}
							required
							minlength={8}
							autocomplete="new-password"
						/>
						{#if passwordMismatch}
							<p class="text-xs text-destructive">Passwords do not match</p>
						{/if}
					</div>

					<Button type="submit" class="w-full" disabled={loading || passwordMismatch}>
						{loading ? 'Creating account...' : 'Create account'}
					</Button>

					<p class="text-center text-sm text-muted-foreground">
						Already have an account?
						<a href="/login" class="text-primary underline underline-offset-4 hover:text-primary/80">
							Sign in
						</a>
					</p>
				</form>
			{/if}
		</CardContent>
	</Card>
</div>
