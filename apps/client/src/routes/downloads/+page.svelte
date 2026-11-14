<script lang="ts">
	import { DownloadSection, LandingFooter, LandingHeader, SignInWithFacile } from '@facile/muse';
	import { API_BASE } from '$lib/backend';
	import { CHECKSUMS_URL, INSTALL_METHODS, RELEASES_URL } from '$lib/downloads';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	const signInHref = $derived(`${API_BASE}/auth/oidc`);
</script>

<svelte:head>
	<title>{t('downloads.title')} · Casier</title>
</svelte:head>

<div class="flex min-h-dvh flex-col overflow-x-clip bg-fc-page text-fc-fg">
	<LandingHeader appName="Casier" icon="solar:key-minimalistic-square-bold-duotone" homeHref="/" skipHref="#main">
		{#snippet actions()}
			<SignInWithFacile href={signInHref} />
		{/snippet}
	</LandingHeader>

	<main id="main" class="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-8 px-4 py-16 sm:px-6 sm:py-20">
		<div class="text-center">
			<h1 class="font-fc-title text-fc-3xl font-bold tracking-tight">{t('downloads.title')}</h1>
			<p class="mx-auto mt-3 max-w-md text-fc-sm text-fc-fg-muted">{t('downloads.intro')}</p>
		</div>
		<DownloadSection
			commands={INSTALL_METHODS}
			releasesHref={RELEASES_URL}
			checksumsHref={CHECKSUMS_URL}
			headingLevel={2}
		/>
	</main>

	<LandingFooter
		appName="Casier"
		icon="solar:key-minimalistic-square-bold-duotone"
		tagline={t('login.tagline')}
		legal={{ onLocaleChange: rememberLocale }}
	/>
</div>
