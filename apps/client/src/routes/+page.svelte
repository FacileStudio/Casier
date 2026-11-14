<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button, DownloadSection, LandingFooter, LandingHeader, SignInWithFacile, icons } from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { API_BASE } from '$lib/backend';
	import { CHECKSUMS_URL, INSTALL_METHODS, RELEASES_URL } from '$lib/downloads';
	import LandingOffers from '$lib/components/LandingOffers.svelte';
	import RunDemo from '$lib/components/landing/RunDemo.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	const features = [
		{ icon: 'solar:lock-keyhole-linear', id: 'encrypted' },
		{ icon: 'solar:layers-linear', id: 'environments' },
		{ icon: 'solar:code-square-linear', id: 'ci' },
		{ icon: icons.history, id: 'audit' }
	] as const;

	let ready = $state(false);

	const signInHref = $derived(`${API_BASE}/auth/oidc`);
	const heroLines = $derived(t('landing.heroTitle').split('|'));
	const links = $derived([
		{ href: '#features', label: t('landing.nav.features') },
		{ href: '#download', label: t('landing.nav.download') }
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

{#if ready}
	<div class="flex min-h-dvh flex-col overflow-x-clip bg-fc-page text-fc-fg">
		<LandingHeader appName="Casier" icon="solar:key-minimalistic-square-bold-duotone" {links} skipHref="#main">
			{#snippet actions()}
				<SignInWithFacile href={signInHref} />
			{/snippet}
		</LandingHeader>

		<main id="main" class="flex-1">
			<section class="mx-auto max-w-5xl px-4 pt-20 pb-12 text-center sm:px-6 sm:pt-24">
				<h1 class="font-fc-title text-fc-3xl font-bold tracking-tight sm:text-5xl lg:text-6xl">
					<span class="block">{heroLines[0]}</span>
					<span class="block">{heroLines[1]}</span>
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted">{t('landing.description')}</p>
				<div class="mt-10 flex flex-wrap items-center justify-center gap-3">
					<SignInWithFacile href={signInHref} size="lg" />
					<Button href="#download" variant="secondary" size="lg" icon={icons.download}>
						{t('downloads.cta')}
					</Button>
				</div>
			</section>

			<section class="mx-auto max-w-5xl px-4 pb-8 sm:px-6">
				<RunDemo />
			</section>

			<section id="features" class="mx-auto max-w-5xl scroll-mt-20 px-4 py-12 sm:px-6">
				<ul class="grid gap-4 sm:grid-cols-2">
					{#each features as feature (feature.id)}
						<li class="rounded-fc-lg bg-fc-component p-6">
							<div class="mb-4 flex size-10 items-center justify-center rounded-fc-md bg-fc-surface" aria-hidden="true">
								<iconify-icon icon={feature.icon} width="20" height="20" class="block shrink-0"></iconify-icon>
							</div>
							<h2 class="text-fc-lg font-bold">{t(`landing.feature.${feature.id}.title`)}</h2>
							<p class="mt-1.5 text-fc-sm text-fc-fg-muted">{t(`landing.feature.${feature.id}.body`)}</p>
						</li>
					{/each}
				</ul>
			</section>

			<section id="download" class="mx-auto flex max-w-5xl scroll-mt-20 flex-col gap-8 px-4 py-20 sm:px-6">
				<div class="text-center">
					<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight">{t('downloads.heading')}</h2>
					<p class="mx-auto mt-3 max-w-md text-fc-sm text-fc-fg-muted">{t('downloads.intro')}</p>
				</div>
				<DownloadSection
					mode="device"
					allHref="/downloads"
					commands={INSTALL_METHODS}
					releasesHref={RELEASES_URL}
					checksumsHref={CHECKSUMS_URL}
					headingLevel={3}
				/>
			</section>

			<LandingOffers />

			<section class="mx-auto max-w-5xl px-4 py-20 text-center sm:px-6">
				<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight">{t('landing.ctaTitle')}</h2>
				<p class="mt-4 text-fc-sm text-fc-fg-muted">{t('landing.ctaText')}</p>
				<div class="mt-8 flex justify-center">
					<SignInWithFacile href={signInHref} size="lg" />
				</div>
			</section>
		</main>

		<LandingFooter
			appName="Casier"
			icon="solar:key-minimalistic-square-bold-duotone"
			tagline={t('login.tagline')}
			legal={{ onLocaleChange: rememberLocale }}
		/>
	</div>
{/if}
