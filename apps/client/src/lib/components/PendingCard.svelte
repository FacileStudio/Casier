<script lang="ts">
	import { Badge, Button, Card, icons } from '@facile/muse';
	import type { PendingAction } from '$lib/backend';
	import { formatDateTime } from '$lib/format';
	import { t } from '$lib/i18n/index.svelte';

	let {
		action,
		busy,
		onconfirm,
		ondismiss
	}: {
		action: PendingAction;
		busy: boolean;
		onconfirm: (action: PendingAction) => void;
		ondismiss: (action: PendingAction) => void;
	} = $props();

	const reveal = $derived(action.kind === 'reveal_secrets');
	const secrets = $derived(action.secrets ?? []);

	/** Names the client that asked: the host of a client URL, else its id. */
	function clientName(clientId: string): string {
		if (clientId === 'bulle') return 'Bulle';
		try {
			return new URL(clientId).host;
		} catch {
			return clientId || '?';
		}
	}
</script>

<Card class="flex flex-col gap-4">
	<div class="flex flex-wrap items-center justify-between gap-2">
		<Badge tone={reveal ? 'warning' : 'danger'} class="shrink-0 whitespace-nowrap">
			{t(reveal ? 'pending.kind.reveal' : 'pending.kind.delete')}
		</Badge>
		<span class="text-fc-xs text-fc-fg-muted">
			{t('pending.askedBy', { client: clientName(action.client_id) })} · {formatDateTime(action.created_at)}
		</span>
	</div>

	{#if secrets.length > 0}
		<div class="flex flex-col gap-2 rounded-fc-md bg-fc-surface p-4">
			<p class="break-words text-fc-sm text-fc-fg-muted">{action.project} / {action.environment}</p>
			<ul class="flex flex-col gap-2">
				{#each secrets as secret (secret.key)}
					<li class="flex items-center gap-3">
						<iconify-icon icon={icons.key} width="18" height="18" class="block shrink-0 text-fc-fg-muted"></iconify-icon>
						<span class="min-w-0 flex-1 break-all font-mono text-fc-sm text-fc-fg">{secret.key}</span>
						<Badge class="shrink-0 whitespace-nowrap">{t('pending.version', { version: secret.version })}</Badge>
					</li>
				{/each}
			</ul>
		</div>
		<p class="text-fc-sm text-fc-fg-muted">{t(reveal ? 'pending.revealNote' : 'pending.deleteNote')}</p>
	{:else}
		<p class="break-words text-fc-sm text-fc-fg-muted">{action.summary}</p>
	{/if}

	{#if action.state === 'changed'}
		<p class="text-fc-sm text-fc-warning">{t('pending.changed')}</p>
	{:else if action.state === 'missing'}
		<p class="text-fc-sm text-fc-warning">{t('pending.missing')}</p>
	{/if}

	<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
		<Button variant="ghost" class="w-full sm:w-auto [@media(pointer:coarse)]:h-11" disabled={busy} onclick={() => ondismiss(action)}>
			{t('pending.dismiss')}
		</Button>
		{#if action.state === 'ready'}
			<Button
				variant={reveal ? 'primary' : 'danger'}
				class="w-full sm:w-auto [@media(pointer:coarse)]:h-11"
				disabled={busy}
				icon={reveal ? icons.eye : icons.remove}
				onclick={() => onconfirm(action)}
			>
				{t(reveal ? 'pending.confirmReveal' : 'pending.confirmDelete')}
			</Button>
		{/if}
	</div>
</Card>
