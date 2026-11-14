<script lang="ts">
	import { LANDING_HEADING, PricingPlans } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';
</script>

<PricingPlans
	id="pricing"
	class="mx-auto w-full max-w-5xl scroll-mt-20 px-4 py-16 sm:px-6"
	title={t('landing.pricing')}
	headingClass={LANDING_HEADING}
	highlight={[]}
	highlightOnly
	notes={{ '*': t('landing.plan.included') }}
	recommended="pro"
	suiteHref="https://suite.facile.studio/#tarifs"
/>
