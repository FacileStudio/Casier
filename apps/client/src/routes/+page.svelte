<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Card, LandingFooter, LandingHeader, SignInWithFacile, icons } from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { API_BASE } from '$lib/backend';
	import LandingOffers from '$lib/components/LandingOffers.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	const features = [
		{ icon: 'solar:lock-keyhole-linear', id: 'encrypted' },
		{ icon: 'solar:layers-linear', id: 'environments' },
		{ icon: 'solar:code-square-linear', id: 'ci' }
	] as const;

	const commands = ['casier login', 'casier init', 'casier run -- npm start'];

	let ready = $state(false);

	const signInHref = $derived(`${API_BASE}/auth/oidc`);
	const heroLines = $derived(t('landing.heroTitle').split('|'));

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
	<div class="flex min-h-dvh flex-col bg-background text-foreground">
		<LandingHeader appName="Casier" icon="solar:key-minimalistic-square-bold-duotone" skipHref="#main">
			{#snippet actions()}
				<SignInWithFacile href={signInHref} />
			{/snippet}
		</LandingHeader>

		<main id="main" class="flex-1">
			<section class="mx-auto max-w-5xl px-6 py-20 text-center sm:py-24">
				<h1
					class="font-heading text-4xl leading-[1.1] font-semibold tracking-tight sm:text-5xl lg:text-6xl"
				>
					{heroLines[0]}<br />{heroLines[1]}
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-base text-muted-foreground sm:text-lg">
					{t('landing.description')}
				</p>
				<div class="mt-10 flex justify-center">
					<SignInWithFacile href={signInHref} size="lg" />
				</div>
			</section>

			<section class="mx-auto max-w-5xl px-6 py-16 sm:py-20">
				<div class="grid gap-4 md:grid-cols-3">
					{#each features as feature (feature.id)}
						<Card class="flex flex-col gap-4">
							<div
								class="flex size-10 items-center justify-center rounded-md bg-muted"
								aria-hidden="true"
							>
								<iconify-icon icon={feature.icon} width="20" height="20"></iconify-icon>
							</div>
							<div class="flex flex-col gap-1">
								<h3 class="text-base font-semibold">{t(`landing.feature.${feature.id}.title`)}</h3>
								<p class="text-sm text-muted-foreground">{t(`landing.feature.${feature.id}.body`)}</p>
							</div>
						</Card>
					{/each}
				</div>
			</section>

			<section class="mx-auto max-w-5xl px-6 py-16 sm:py-20">
				<h2 class="text-center font-heading text-2xl font-semibold tracking-tight sm:text-3xl">
					{t('landing.commandsTitle')}
				</h2>
				<Card class="mx-auto mt-8 max-w-2xl overflow-x-auto">
					<div class="flex w-max min-w-full flex-col gap-1 font-mono text-sm">
						{#each commands as command (command)}
							<div><span class="text-muted-foreground">$</span> {command}</div>
						{/each}
					</div>
				</Card>
				<p class="mt-4 text-center text-sm text-muted-foreground">{t('landing.commandsNote')}</p>
			</section>

			<section class="mx-auto max-w-5xl px-6 py-16 text-center sm:py-20">
				<h2 class="font-heading text-2xl font-semibold tracking-tight sm:text-3xl">
					{t('landing.ctaTitle')}
				</h2>
				<p class="mx-auto mt-4 max-w-xl text-muted-foreground">{t('landing.ctaText')}</p>
				<div class="mt-8 flex justify-center">
					<SignInWithFacile href={signInHref} size="lg" />
				</div>
			</section>
			<LandingOffers />
		</main>

		<LandingFooter
			appName="Casier"
			icon="solar:key-minimalistic-square-bold-duotone"
			tagline={t('login.tagline')}
			legal={{ onLocaleChange: rememberLocale }}
		/>
	</div>
{/if}
