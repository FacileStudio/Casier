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
	let error = $state('');
	let loading = $state(false);
	let ssoOnly = $state(false);
	let oidcEnabled = $state(false);

	$effect(() => {
		api.get<{ sso_only: boolean; oidc_enabled: boolean }>('/auth/config')
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
			const res = await api.post<{ user_id: string; token: string }>(
				'/auth/login',
				{ email, password }
			);
			auth.login(res.token, res.user_id, email);
			goto('/spaces');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed';
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
			<CardTitle class="text-xl">Casier</CardTitle>
			<CardDescription>Sign in to manage your secrets</CardDescription>
		</CardHeader>
		<CardContent>
			{#if oidcEnabled}
				<div class="flex flex-col gap-4">
					<Button class="w-full" onclick={() => window.location.href = `${API_BASE}/auth/oidc`}>
						Login with SSO
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

				<form onsubmit={handleLogin} class="flex flex-col gap-4">
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
							autocomplete="current-password"
						/>
					</div>

					<Button type="submit" class="w-full" disabled={loading}>
						{loading ? 'Signing in...' : 'Sign in'}
					</Button>

					<div class="flex items-center justify-between text-sm text-muted-foreground">
						<a href="/forgot-password" class="text-primary underline underline-offset-4 hover:text-primary/80">
							Forgot password?
						</a>
						<span>
							No account?
							<a href="/register" class="text-primary underline underline-offset-4 hover:text-primary/80">
								Register
							</a>
						</span>
					</div>
				</form>
			{/if}
		</CardContent>
	</Card>
</div>
