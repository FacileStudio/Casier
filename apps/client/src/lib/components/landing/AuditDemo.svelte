<script lang="ts">
	import { Badge } from '@facile/muse';
	import { Stage, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';

	type Event = { who: string; initial: string; action: string; tone: 'warning' | 'neutral'; target: string; when: string };

	const EVENTS: Event[] = [
		{ who: 'Léa', initial: 'L', action: 'secrets.reveal', tone: 'warning', target: 'prod / STRIPE_SECRET_KEY', when: '14:02' },
		{ who: 'ci-deploy', initial: 'C', action: 'secrets.export', tone: 'warning', target: 'prod', when: '14:05' },
		{ who: 'Hugo', initial: 'H', action: 'secrets.rollback', tone: 'neutral', target: 'staging / REDIS_URL', when: '14:11' }
	];

	let shown = $state(1);

	function reset() {
		shown = 1;
	}

	function poster() {
		shown = EVENTS.length;
	}

	async function script(act: Act) {
		await act.wait(900);
		for (let i = 2; i <= EVENTS.length; i++) {
			shown = i;
			await act.wait(1300);
		}
		await act.wait(2000);
		reset();
		await act.wait(700);
	}
</script>

<Stage width={400} height={260} label={t('landing.tile.audit.label')} {script} {reset} {poster}>
	{#snippet children()}
		<div class="ad-panel">
			<div class="ad-head">
				<span>{t('audit.col.who')}</span><span>{t('audit.col.action')}</span>
			</div>
			{#each EVENTS as event, i (event.action)}
				<div class="ad-row" class:ad-on={i < shown}>
					<span class="ad-avatar">{event.initial}</span>
					<div class="min-w-0 flex-1">
						<div class="flex items-center justify-between gap-2">
							<span class="font-bold">{event.who}</span>
							<span class="text-fc-fg-muted">{event.when}</span>
						</div>
						<div class="mt-1 flex items-center gap-2">
							<Badge tone={event.tone} class="shrink-0 whitespace-nowrap">{event.action}</Badge>
							<span class="ad-target">{event.target}</span>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/snippet}
</Stage>

<style>
	.ad-panel {
		position: absolute;
		inset: 16px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px;
		border-radius: var(--radius-fc-lg);
		background: var(--color-fc-bg);
		box-shadow: 0 8px 24px -12px rgb(0 0 0 / 0.25);
		font-family: var(--font-fc-body);
		font-size: 13px;
		color: var(--color-fc-fg);
	}
	.ad-head {
		display: flex;
		justify-content: space-between;
		padding: 0 4px 2px;
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--color-fc-fg-muted);
	}
	.ad-row {
		display: flex;
		align-items: center;
		gap: 10px;
		height: 56px;
		padding: 0 10px;
		border-radius: var(--radius-fc-md);
		background: var(--color-fc-surface);
		opacity: 0;
		transform: translateY(8px);
		transition:
			opacity 320ms ease,
			transform 320ms ease;
	}
	.ad-on {
		opacity: 1;
		transform: none;
	}
	.ad-avatar {
		display: flex;
		flex: none;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: var(--radius-fc-pill);
		background: var(--color-fc-component);
		font-weight: 700;
		font-size: 12px;
	}
	.ad-target {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-family: var(--font-fc-mono);
		font-size: 12px;
		color: var(--color-fc-fg-muted);
	}
</style>
