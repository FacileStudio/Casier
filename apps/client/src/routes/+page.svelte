<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		FeatureSplit,
		LandingCta,
		LandingFooter,
		LandingHeader,
		SignInWithFacile,
		WorksWith,
		icons,
		type FeatureItem,
		type WorksWithApp
	} from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { API_BASE } from '$lib/backend';
	import LandingOffers from '$lib/components/LandingOffers.svelte';
	import DemoTiles from '$lib/components/landing/DemoTiles.svelte';
	import LandingDevelopers from '$lib/components/landing/LandingDevelopers.svelte';
	import RunDemo from '$lib/components/landing/RunDemo.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	const APP_ICON = 'solar:key-minimalistic-square-bold-duotone';

	let ready = $state(false);

	const signInHref = $derived(`${API_BASE}/auth/oidc`);
	const links = $derived([
		{ href: '#features', label: t('landing.nav.features') },
		{ href: '#pricing', label: t('landing.nav.pricing') },
		{ href: '#developers', label: t('landing.nav.developers') }
	]);
	const features = $derived<FeatureItem[]>([
		{ icon: 'solar:lock-keyhole-linear', title: t('landing.feature.encrypted.title'), text: t('landing.feature.encrypted.body') },
		{ icon: 'solar:code-square-linear', title: t('landing.feature.ci.title'), text: t('landing.feature.ci.body') },
		{ icon: icons.usersGroup, title: t('landing.feature.roles.title'), text: t('landing.feature.roles.body') },
		{ icon: icons.history, title: t('landing.feature.history.title'), text: t('landing.feature.history.body') }
	]);
	const apps = $derived<WorksWithApp[]>([
		{ name: 'Bulle', icon: 'solar:smile-circle-bold-duotone', line: t('landing.with.bulle'), href: 'https://bulle.facile.studio' },
		{ name: 'Boîte', icon: 'solar:box-minimalistic-bold-duotone', line: t('landing.with.boite'), href: 'https://boite.facile.studio' }
	]);

	onMount(async () => {
		if (!auth.loaded) {
			await auth.check();
		}

		if (auth.isAuthenticated) {
			goto('/projects', { replaceState: true });
			return;
		}

		ready = true;
	});
</script>

<svelte:head>
	<title>{t('landing.title')}</title>
	<meta name="description" content={t('landing.metaDescription')} />
</svelte:head>

{#if ready}
	<div class="flex min-h-dvh flex-col overflow-x-clip bg-fc-page text-fc-fg">
		<LandingHeader appName="Casier" icon={APP_ICON} {links} skipHref="#main">
			{#snippet actions()}
				<SignInWithFacile href={signInHref} />
			{/snippet}
		</LandingHeader>

		<main id="main" class="flex-1">
			<section class="mx-auto max-w-5xl px-4 pt-20 pb-12 text-center sm:px-6 sm:pt-24">
				<h1 class="mx-auto max-w-3xl font-fc-title text-fc-3xl font-bold tracking-tight text-balance sm:text-5xl lg:text-6xl">
					{t('landing.heroTitle')}
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted">{t('landing.description')}</p>
				<LandingCta {signInHref} align="center" class="mx-auto mt-10 max-w-md sm:max-w-none" />
			</section>

			<section class="mx-auto max-w-5xl px-4 pb-8 sm:px-6">
				<RunDemo />
			</section>

			<DemoTiles />

			<WorksWith class="mx-auto w-full max-w-5xl px-4 py-16 sm:px-6" {apps} />

			<FeatureSplit
				class="mx-auto w-full max-w-5xl px-4 py-16 sm:px-6"
				title={t('landing.features.title')}
				description={t('landing.features.line')}
				{features}
			/>

			<LandingOffers />

			<LandingDevelopers />

			<section class="mx-auto max-w-5xl px-4 py-20 text-center sm:px-6">
				<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight sm:text-fc-3xl">{t('landing.ctaTitle')}</h2>
				<p class="mt-4 text-fc-sm text-fc-fg-muted">{t('landing.ctaText')}</p>
				<LandingCta {signInHref} align="center" size="md" class="mx-auto mt-8 max-w-md sm:max-w-none" />
			</section>
		</main>

		<LandingFooter appName="Casier" icon={APP_ICON} tagline={t('login.tagline')} legal={{ onLocaleChange: rememberLocale }} />
	</div>
{/if}
