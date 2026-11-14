<script lang="ts">
	import { Badge } from '@facile/muse';
	import { Tile } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import ApproveDemo from './ApproveDemo.svelte';
	import AuditDemo from './AuditDemo.svelte';
	import EnvDemo from './EnvDemo.svelte';
</script>

{#snippet pro()}
	<Badge tone="info" class="shrink-0 whitespace-nowrap">{t('landing.tile.bulle.badge')}</Badge>
{/snippet}

<section id="features" class="mx-auto w-full max-w-5xl scroll-mt-20 px-4 py-16 sm:px-6">
	<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight sm:text-fc-3xl">{t('landing.tiles.title')}</h2>
	<div class="mt-8 grid gap-4 lg:grid-cols-3">
		<Tile title={t('landing.tile.audit.title')} body={t('landing.tile.audit.body')}>
			<div class="mx-auto w-full max-w-md"><AuditDemo /></div>
		</Tile>
		<Tile title={t('landing.tile.env.title')} body={t('landing.tile.env.body')}>
			<div class="mx-auto w-full max-w-md"><EnvDemo /></div>
		</Tile>
		<Tile title={t('landing.tile.bulle.title')} body={t('landing.tile.bulle.body')} badge={pro}>
			<div class="mx-auto w-full max-w-md"><ApproveDemo /></div>
		</Tile>
	</div>
</section>
