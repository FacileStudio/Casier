<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, API_BASE, type AuthConfig } from '$lib/backend';
	import { theme, type ThemePreference } from '$lib/theme.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';

	type Tab = 'account' | 'appearance' | 'about';

	const tabItems: { id: Tab; label: string; icon: string }[] = [
		{ id: 'account', label: 'Account', icon: 'solar:user-circle-linear' },
		{ id: 'appearance', label: 'Appearance', icon: 'solar:pallete-2-linear' },
		{ id: 'about', label: 'About', icon: 'solar:info-circle-linear' },
	];

	const themeOptions: { id: ThemePreference; label: string }[] = [
		{ id: 'system', label: 'System' },
		{ id: 'light', label: 'Light' },
		{ id: 'dark', label: 'Dark' },
	];

	const cliCommands = ['casier login', 'casier init', 'casier run -- npm start'];

	let activeTab = $state<Tab>('account');
	let ssoOnly = $state(false);
	let oidcEnabled = $state(false);

	onMount(async () => {
		try {
			const config = await api.get<AuthConfig>('/auth/config');
			ssoOnly = config.sso_only;
			oidcEnabled = config.oidc_enabled;
		} catch {}
	});

	async function handleLogout() {
		try {
			await api.post('/auth/logout', {}, auth.token);
		} catch {}
		auth.logout();
		goto('/login');
	}
</script>

<svelte:head>
	<title>Settings — Casier</title>
</svelte:head>

{#snippet tabs()}
	{#each tabItems as tab (tab.id)}
		<button
			type="button"
			aria-pressed={activeTab === tab.id}
			onclick={() => (activeTab = tab.id)}
			class="inline-flex h-9 shrink-0 items-center gap-2 rounded-md px-4 text-sm font-medium transition-colors {activeTab ===
			tab.id
				? 'bg-foreground text-background'
				: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
		>
			<iconify-icon icon={tab.icon} width="18" aria-hidden="true"></iconify-icon>
			{tab.label}
		</button>
	{/each}
{/snippet}

<PageHeader title="Settings" {tabs} />

<div class="px-4 py-6 md:px-8">
	<div class="max-w-xl space-y-8">
		{#if activeTab === 'account'}
			<div class="space-y-4">
				<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Account</h2>
				<div>
					<label for="account-email" class="mb-1.5 block text-sm font-medium">Email</label>
					<input
						id="account-email"
						type="email"
						value={auth.email ?? ''}
						disabled
						class="flex h-10 w-full rounded-md border border-input bg-muted px-3 py-2 text-sm text-muted-foreground"
					/>
					<p class="mt-1.5 text-xs text-muted-foreground">
						Your email comes from your Casier account and can't be changed here.
					</p>
				</div>
			</div>

			<div class="space-y-4">
				<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
					Password
				</h2>
				{#if ssoOnly}
					<p class="text-sm text-muted-foreground">
						Authentication is handled by your identity provider. Manage your password there.
					</p>
				{:else}
					<p class="text-sm text-muted-foreground">Password changes go through the reset flow.</p>
					<a
						href="/forgot-password"
						class="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-border bg-background px-4 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
					>
						Reset password
					</a>
					{#if oidcEnabled}
						<p class="text-xs text-muted-foreground">
							SSO sign-in is also available on this instance.
						</p>
					{/if}
				{/if}
			</div>

			<div class="space-y-4">
				<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Session</h2>
				<button
					type="button"
					onclick={handleLogout}
					class="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-border bg-background px-4 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
				>
					<iconify-icon icon="solar:logout-2-linear" width="16" aria-hidden="true"></iconify-icon>
					Log out
				</button>
			</div>
		{:else if activeTab === 'appearance'}
			<div class="space-y-4">
				<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Theme</h2>
				<div class="flex gap-1 rounded-lg border border-border bg-muted p-1">
					{#each themeOptions as option (option.id)}
						<button
							type="button"
							aria-pressed={theme.preference === option.id}
							onclick={() => theme.set(option.id)}
							class="flex-1 rounded-md py-1.5 text-center text-sm font-medium transition-colors {theme.preference ===
							option.id
								? 'bg-background text-foreground shadow-sm'
								: 'text-muted-foreground hover:text-foreground'}"
						>
							{option.label}
						</button>
					{/each}
				</div>
				<p class="text-xs text-muted-foreground">System follows your operating system setting.</p>
			</div>
		{:else}
			<div class="space-y-4">
				<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Casier</h2>
				<p class="text-sm text-muted-foreground">
					A self-hosted secrets and environment manager for teams. Values are encrypted at rest,
					versioned, and audited.
				</p>
				<dl class="space-y-2 text-sm">
					<div class="flex items-center justify-between gap-3">
						<dt class="text-muted-foreground">Interface</dt>
						<dd>Casier web</dd>
					</div>
					<div class="flex items-center justify-between gap-3">
						<dt class="text-muted-foreground">API</dt>
						<dd class="min-w-0 truncate rounded bg-muted px-2 py-1 font-mono text-xs">{API_BASE}</dd>
					</div>
				</dl>
			</div>

			<div class="space-y-4">
				<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
					Command line
				</h2>
				<p class="text-sm text-muted-foreground">
					The <span class="rounded bg-muted px-2 py-1 font-mono text-xs">casier</span> CLI pulls the
					same secrets into your shell, your scripts, and your CI.
				</p>
				<div class="overflow-x-auto rounded-md bg-muted px-3 py-2 font-mono text-xs">
					{#each cliCommands as command (command)}
						<div class="flex gap-2 py-0.5">
							<span class="text-muted-foreground select-none">$</span>
							<span class="whitespace-pre">{command}</span>
						</div>
					{/each}
				</div>
			</div>

			<div class="space-y-4">
				<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
					Facile Suite
				</h2>
				<p class="text-sm text-muted-foreground">
					Casier is part of the <a
						href="https://facile.studio"
						target="_blank"
						rel="noopener"
						class="underline underline-offset-2 transition-colors hover:text-foreground">Facile Suite</a
					>.
				</p>
			</div>
		{/if}
	</div>
</div>
