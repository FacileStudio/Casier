<script lang="ts">
	import { tick } from 'svelte';
	import { fade } from 'svelte/transition';
	import { Badge, Button, icons, maskSecret } from '@facile/muse';
	import { Bar, Caret, Cursor, Stage, Terminal, Window, cursorAt, grow, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';

	type Line = { kind: 'cmd' | 'out' | 'var' | 'ok'; text: string; value?: string };

	const ADDED = { key: 'STRIPE_SECRET_KEY', value: maskSecret('sk_live_51Hq0c9f3a') };
	const STORED = [
		{ key: 'DATABASE_URL', value: maskSecret('postgres://shop@db:5432/shop') },
		{ key: 'REDIS_URL', value: maskSecret('redis://cache:6379') },
		{ key: 'SESSION_SECRET', value: maskSecret('k8Zq2vTn') }
	];
	const COMMAND = 'casier run -e prod -- bun start';
	const OUTPUT: Line[] = [
		{ kind: 'out', text: '$ bun run server.ts' },
		...[STORED[0], ADDED, STORED[1]].map((row) => ({ kind: 'var' as const, text: row.key, value: row.value })),
		{ kind: 'ok', text: 'listening on http://localhost:3000' }
	];
	const ENVS = ['dev', 'staging', 'prod'];

	let added = $state(false);
	let lines = $state<Line[]>([]);
	let cursor = $state(cursorAt());
	let addButton = $state<HTMLElement>();
	let terminal = $state<HTMLElement>();

	function reset() {
		added = false;
		lines = [];
		cursor.visible = false;
	}

	function poster() {
		added = true;
		lines = [{ kind: 'cmd', text: COMMAND }, ...OUTPUT];
	}

	async function script(act: Act) {
		cursor.x = 420;
		cursor.y = 300;
		await act.wait(700);
		cursor.visible = true;
		await act.move(cursor, act.point(addButton, 0.5, 0.62));
		await act.click(cursor);
		added = true;
		await tick();
		await act.wait(700);
		await act.move(cursor, act.point(terminal, 0.42, 0.3));
		cursor.visible = false;
		lines.push({ kind: 'cmd', text: '' });
		await act.type(COMMAND, (v) => (lines[0].text = v), 55);
		await act.wait(450);
		for (const line of OUTPUT) {
			lines.push(line);
			await act.wait(line.kind === 'out' ? 600 : 380);
		}
		await act.wait(2200);
		reset();
		await act.wait(600);
	}
</script>

{#snippet nav()}
	<div class="flex items-center gap-2 px-5 pt-3 pb-5 font-fc-title text-[15px] font-bold text-fc-fg">
		<iconify-icon icon="solar:key-minimalistic-square-bold-duotone" width="20" height="20"></iconify-icon>
		<span>Casier</span>
	</div>
	<div class="grid gap-1 px-3">
		{#each [58, 44, 64, 50] as w, i (i)}
			<div class="flex h-[34px] items-center gap-2.5 rounded-fc-md px-2.5 {i === 0 ? 'bg-fc-fg/5' : ''}">
				<span class="size-4 shrink-0 rounded-[5px] bg-fc-fg/20"></span><Bar w="{w}%" />
			</div>
		{/each}
	</div>
{/snippet}

{#snippet row(key: string, value: string, fresh: boolean, compact: boolean)}
	<div class="rd-key" class:font-bold={fresh}>{key}</div>
	{#if !compact || !fresh}
		<div class="text-fc-fg-muted">{value}</div>
	{/if}
	{#if fresh}
		<Badge tone="success" class="whitespace-nowrap">v1</Badge>
	{/if}
{/snippet}

<Stage
	width={960}
	height={560}
	compact={{ below: 640, width: 358, height: 496 }}
	label={t('landing.demo.label')}
	{script}
	{reset}
	{poster}
>
	{#snippet children({ compact })}
		<Window
			style={compact
				? 'top:12px;left:12px;width:334px;height:300px'
				: 'top:20px;left:236px;width:700px;height:360px'}
			sidebar={compact ? undefined : nav}
		>
			<div class="rd-body" class:rd-compact={compact}>
				<div class="flex items-center justify-between gap-3">
					<div class="flex min-w-0 items-center gap-2">
						<span class="font-fc-title text-fc-lg font-bold">shop</span>
						{#if !compact}<Badge tone="info">{t('role.owner')}</Badge>{/if}
					</div>
					<span bind:this={addButton}>
						<Button icon={icons.plus} size={compact ? 'sm' : 'md'}>{t('sec.add')}</Button>
					</span>
				</div>

				<div class="rd-tabs">
					{#each ENVS as env (env)}
						<span class="rd-tab" class:rd-tab-on={env === 'prod'}>{env}</span>
					{/each}
				</div>

				<div class="rd-list">
					{#if added}
						<div class="rd-row" in:grow out:fade={{ duration: 200 }}>
							{@render row(ADDED.key, ADDED.value, true, compact)}
						</div>
					{/if}
					{#each compact ? STORED.slice(0, 2) : STORED as secret (secret.key)}
						<div class="rd-row">{@render row(secret.key, secret.value, false, compact)}</div>
					{/each}
				</div>
			</div>
		</Window>

		<div
			bind:this={terminal}
			class="absolute"
			style={compact
				? 'top:272px;left:12px;width:334px;height:212px'
				: 'top:336px;left:24px;width:520px;height:204px'}
		>
			<Terminal style="inset:0">
				<div class="mb-3 flex items-center gap-2 text-fc-on-inverse-muted">
					<span class="size-2 rounded-fc-pill bg-fc-success"></span>~/shop
				</div>
				{#each lines as line, i (i)}
					<div class="whitespace-nowrap" in:fade={{ duration: 160 }}>
						{#if line.kind === 'cmd'}
							<span class="rd-dim">$</span>
							{line.text}{#if i === lines.length - 1}<Caret />{/if}
						{:else if line.kind === 'var'}
							<span class="rd-dim">env</span>
							<span class="rd-var">{line.text}</span>
							<span class="rd-dim">{line.value}</span>
						{:else}
							<span class:rd-ok={line.kind === 'ok'} class:rd-dim={line.kind === 'out'}>{line.text}</span>
						{/if}
					</div>
				{/each}
				{#if lines.length === 0}
					<div class="whitespace-nowrap"><span class="rd-dim">$</span> <Caret /></div>
				{/if}
			</Terminal>
		</div>
		<Cursor {cursor} />
	{/snippet}
</Stage>

<style>
	.rd-body {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 18px;
		min-height: 0;
		padding: 18px 28px 24px;
		font-family: var(--font-fc-body);
		color: var(--color-fc-fg);
	}
	.rd-compact {
		gap: 14px;
		padding: 12px 16px 16px;
	}
	.rd-tabs {
		display: flex;
		align-self: flex-start;
		gap: 4px;
		padding: 4px;
		border-radius: var(--radius-fc-pill);
		background: var(--color-fc-surface);
	}
	.rd-tab {
		padding: 6px 14px;
		border-radius: var(--radius-fc-pill);
		font-size: 13px;
		color: var(--color-fc-fg-muted);
	}
	.rd-tab-on {
		background: var(--color-fc-bg);
		color: var(--color-fc-fg);
		font-weight: 700;
		box-shadow: 0 2px 8px -2px rgb(0 0 0 / 0.15);
	}
	.rd-list {
		display: flex;
		flex-direction: column;
		padding: 4px 0;
		border-radius: var(--radius-fc-lg);
		background: var(--color-fc-surface);
	}
	.rd-row {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 9px 18px;
		font-family: var(--font-fc-mono);
		font-size: 13px;
		white-space: nowrap;
	}
	.rd-compact .rd-row {
		gap: 10px;
		padding: 10px 14px;
		font-size: 12px;
	}
	.rd-key {
		flex: none;
		width: 168px;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.rd-compact .rd-key {
		width: auto;
		flex: 1;
	}
	.rd-dim {
		color: var(--color-fc-on-inverse-muted);
	}
	.rd-var {
		display: inline-block;
		min-width: 18ch;
	}
	.rd-ok {
		color: color-mix(in oklab, var(--color-fc-success), var(--color-fc-on-inverse) 35%);
	}
</style>
