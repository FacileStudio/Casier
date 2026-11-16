<script lang="ts">
	import { Badge, PendingCard, icons } from '@facile/muse';
	import type { PendingAction } from '$lib/backend';
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
</script>

<PendingCard
	kind={t(reveal ? 'pending.kind.reveal' : 'pending.kind.delete')}
	tone={reveal ? 'warning' : 'danger'}
	client={action.client_id}
	createdAt={action.created_at}
	state={action.state}
	{busy}
	confirmLabel={t(reveal ? 'pending.confirmReveal' : 'pending.confirmDelete')}
	confirmIcon={reveal ? icons.eye : icons.remove}
	confirmVariant={reveal ? 'primary' : 'danger'}
	onConfirm={() => onconfirm(action)}
	onDismiss={() => ondismiss(action)}
	dismissLabel={t('pending.dismiss')}
	changedLabel={t('pending.changed')}
	missingLabel={t('pending.missing')}
>
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
</PendingCard>
