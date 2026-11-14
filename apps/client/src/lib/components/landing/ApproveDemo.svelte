<script lang="ts">
	import { tick } from 'svelte';
	import { Badge, Button, Icon, icons } from '@facile/muse';
	import { Cursor, Stage, cursorAt, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';

	let allowed = $state(false);
	let pressed = $state(false);
	let cursor = $state(cursorAt());
	let allow = $state<HTMLElement>();

	function reset() {
		allowed = false;
		pressed = false;
		cursor.visible = false;
	}

	function poster() {
		allowed = false;
	}

	async function script(act: Act) {
		cursor.x = 120;
		cursor.y = 250;
		await act.wait(1000);
		cursor.visible = true;
		await act.move(cursor, act.point(allow, 0.5, 0.6));
		await act.click(cursor, () => (pressed = true));
		allowed = true;
		await tick();
		await act.wait(500);
		cursor.visible = false;
		await act.wait(2400);
		reset();
		await act.wait(800);
	}
</script>

<Stage width={400} height={260} label={t('landing.tile.bulle.label')} {script} {reset} {poster}>
	{#snippet children()}
		<div class="ap-card">
			<div class="flex items-center justify-between gap-2">
				<Badge tone="warning" class="shrink-0 whitespace-nowrap">{t('pending.kind.reveal')}</Badge>
				<span class="text-[12px] text-fc-fg-muted">{t('pending.askedBy', { client: 'Bulle' })}</span>
			</div>
			<div class="ap-box">
				<span class="text-[12px] text-fc-fg-muted">shop / prod</span>
				<div class="flex items-center gap-2.5">
					<Icon icon={icons.key} size={16} class="text-fc-fg-muted" />
					<span class="flex-1 font-fc-mono text-[13px]">STRIPE_SECRET_KEY</span>
					<Badge class="shrink-0 whitespace-nowrap">v1</Badge>
				</div>
			</div>
			<div class="ap-foot">
				<div class="ap-layer ap-actions" class:ap-hidden={allowed}>
					<Button variant="ghost" size="sm">{t('pending.dismiss')}</Button>
					<span bind:this={allow} class:ap-pressed={pressed}>
						<Button size="sm" icon={icons.eye}>{t('pending.confirmReveal')}</Button>
					</span>
				</div>
				<div class="ap-layer ap-done" class:ap-hidden={!allowed}>
					<Icon icon={icons.check} size={18} class="text-fc-success" />
					<span>{t('landing.tile.bulle.allowed')}</span>
				</div>
			</div>
		</div>
		<Cursor {cursor} />
	{/snippet}
</Stage>

<style>
	.ap-card {
		position: absolute;
		top: 34px;
		right: 16px;
		left: 16px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 16px;
		border-radius: var(--radius-fc-lg);
		background: var(--color-fc-bg);
		box-shadow: 0 8px 24px -12px rgb(0 0 0 / 0.25);
		font-family: var(--font-fc-body);
		font-size: 13px;
		color: var(--color-fc-fg);
	}
	.ap-box {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 12px 14px;
		border-radius: var(--radius-fc-md);
		background: var(--color-fc-surface);
	}
	.ap-foot {
		position: relative;
		height: 40px;
	}
	.ap-layer {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		gap: 8px;
		transition:
			opacity 260ms ease,
			transform 260ms ease;
	}
	.ap-layer:not(.ap-hidden) {
		transition-delay: 220ms;
	}
	.ap-actions {
		justify-content: flex-end;
	}
	.ap-done {
		font-weight: 700;
	}
	.ap-hidden {
		opacity: 0;
		transform: translateY(4px);
		pointer-events: none;
	}
	.ap-pressed {
		transform: scale(0.97);
		transition: transform 120ms ease;
	}
</style>
