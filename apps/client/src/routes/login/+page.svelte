<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Alert, AuthScreen } from '@facile/muse';
	import { API_BASE } from '$lib/backend';
	import { auth } from '$lib/auth.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	let failed = $state(false);

	onMount(async () => {
		failed = page.url.searchParams.has('error');
		await auth.check();
		if (auth.isAuthenticated) goto('/projects', { replaceState: true });
	});
</script>

<svelte:head>
	<title>{t('login.title')}</title>
</svelte:head>

<AuthScreen appName="Casier" href="{API_BASE}/auth/oidc" onLocaleChange={rememberLocale}>
	{#snippet logo()}
		<Logo size={40} />
	{/snippet}
	{#snippet tagline()}
		{t('login.tagline')}
	{/snippet}
	{#snippet notice()}
		{#if failed}
			<Alert tone="danger">{t('login.failed')}</Alert>
		{/if}
	{/snippet}
</AuthScreen>
