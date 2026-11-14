<script lang="ts">
	import { Badge, maskSecret } from '@facile/muse';
	import { Roll, Stage, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';

	const ENVS = ['dev', 'staging', 'prod'] as const;
	const KEYS = [
		{ key: 'DATABASE_URL', value: maskSecret('postgres://shop@db:5432/shop'), versions: [7, 3, 2] },
		{ key: 'STRIPE_SECRET_KEY', value: maskSecret('sk_test_4eC39HqLyjWD'), versions: [4, 2, 1] },
		{ key: 'REDIS_URL', value: maskSecret('redis://cache:6379'), versions: [2, 1, 1] }
	];
	const TAB = 84;

	let active = $state(0);

	function reset() {
		active = 0;
	}

	function poster() {
		active = 2;
	}

	async function script(act: Act) {
		await act.wait(1200);
		for (let i = 1; i < ENVS.length; i++) {
			active = i;
			await act.wait(1600);
		}
		await act.wait(800);
		reset();
		await act.wait(900);
	}
</script>

<Stage width={400} height={260} label={t('landing.tile.env.label')} {script} {reset} {poster}>
	{#snippet children()}
		<div class="ev-panel">
			<div class="ev-tabs">
				<span class="ev-pill" style="transform: translateX({active * TAB}px)"></span>
				{#each ENVS as env, i (env)}
					<span class="ev-tab" class:ev-tab-on={i === active}>{env}</span>
				{/each}
			</div>
			<div class="ev-list">
				{#each KEYS as row (row.key)}
					<div class="ev-row">
						<span class="ev-key">{row.key}</span>
						<span class="ev-value">{row.value}</span>
						<Badge class="shrink-0 whitespace-nowrap"><Roll value="v{row.versions[active]}" /></Badge>
					</div>
				{/each}
			</div>
		</div>
	{/snippet}
</Stage>

<style>
	.ev-panel {
		position: absolute;
		inset: 16px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 16px;
		border-radius: var(--radius-fc-lg);
		background: var(--color-fc-bg);
		box-shadow: 0 8px 24px -12px rgb(0 0 0 / 0.25);
		font-family: var(--font-fc-body);
		color: var(--color-fc-fg);
	}
	.ev-tabs {
		position: relative;
		display: flex;
		align-self: flex-start;
		padding: 4px;
		border-radius: var(--radius-fc-pill);
		background: var(--color-fc-surface);
	}
	.ev-pill {
		position: absolute;
		top: 4px;
		left: 4px;
		width: 84px;
		height: 30px;
		border-radius: var(--radius-fc-pill);
		background: var(--color-fc-bg);
		box-shadow: 0 2px 8px -2px rgb(0 0 0 / 0.15);
		transition: transform 380ms cubic-bezier(0.2, 0.8, 0.2, 1);
	}
	.ev-tab {
		position: relative;
		width: 84px;
		height: 30px;
		line-height: 30px;
		text-align: center;
		font-size: 13px;
		color: var(--color-fc-fg-muted);
		transition: color 300ms ease;
	}
	.ev-tab-on {
		color: var(--color-fc-fg);
		font-weight: 700;
	}
	.ev-list {
		display: flex;
		flex-direction: column;
		padding: 4px 0;
		border-radius: var(--radius-fc-lg);
		background: var(--color-fc-surface);
	}
	.ev-row {
		display: flex;
		align-items: center;
		gap: 12px;
		height: 44px;
		padding: 0 14px;
		font-family: var(--font-fc-mono);
		font-size: 12px;
		white-space: nowrap;
	}
	.ev-key {
		flex: none;
		width: 136px;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.ev-value {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		color: var(--color-fc-fg-muted);
	}
</style>
