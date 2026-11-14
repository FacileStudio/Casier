<script lang="ts">
	import { ForDevelopers, type DeveloperRow } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	const SOURCE = 'https://github.com/FacileStudio/Casier';
	const SELF_HOST = `${SOURCE}/blob/main/docs/deployment.md`;

	const rows = $derived<DeveloperRow[]>([
		{ kind: 'cli', text: t('landing.dev.cli'), href: '/downloads', command: 'brew install FacileStudio/tap/casier' },
		{ kind: 'api', text: t('landing.dev.api'), href: '/docs' },
		{ kind: 'mcp', text: t('landing.dev.mcp'), href: `${SOURCE}/blob/main/docs/mcp.md` },
		{ kind: 'source', text: t('landing.dev.source'), href: SOURCE }
	]);
</script>

<div id="developers" class="mx-auto w-full max-w-5xl scroll-mt-20 px-4 py-16 sm:px-6">
	<ForDevelopers {rows} description={t('landing.dev.line')} selfHostHref={SELF_HOST} />
</div>
