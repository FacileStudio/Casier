<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { api, type AuditLogEntry, type AuditLogsResponse } from '$lib/backend';
	import { formatDateTime } from '$lib/format';
	import { Badge, Skeleton, Table } from '@facile/lib';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import ScrollTextIcon from '@lucide/svelte/icons/scroll-text';

	let { spaceSlug }: { spaceSlug: string } = $props();

	let logs = $state<AuditLogEntry[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	async function loadLogs() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<AuditLogsResponse>(`/spaces/${spaceSlug}/audit?limit=50`, auth.token);
			logs = res.logs ?? [];
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load audit log';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		spaceSlug;
		loadLogs();
	});
</script>

<section class="flex flex-col gap-4" aria-label="Audit log">
	{#if loading}
		<div class="flex flex-col gap-2 rounded-fc-md bg-fc-component px-6 py-5">
			{#each Array(6) as _, i (i)}
				<Skeleton class="h-8 w-full" />
			{/each}
		</div>
	{:else if loadError}
		<ErrorState message={loadError} onretry={loadLogs} />
	{:else if logs.length === 0}
		<EmptyState
			title="No activity yet"
			description="Changes to secrets, members and settings will show up here."
		>
			{#snippet icon()}
				<ScrollTextIcon class="size-10" aria-hidden="true" />
			{/snippet}
		</EmptyState>
	{:else}
		<Table>
			<thead>
				<tr>
					<th>When</th>
					<th>Who</th>
					<th>Action</th>
					<th>Resource</th>
					<th>Detail</th>
				</tr>
			</thead>
			<tbody>
				{#each logs as log (log.id)}
					<tr>
						<td class="whitespace-nowrap text-fc-fg-muted">{formatDateTime(log.created_at)}</td>
						<td class="whitespace-nowrap">{log.user_email || '—'}</td>
						<td><Badge tone="neutral">{log.action}</Badge></td>
						<td><code class="font-mono text-fc-sm">{log.resource || '—'}</code></td>
						<td class="max-w-64 truncate text-fc-fg-muted" title={log.detail}>{log.detail || '—'}</td>
					</tr>
				{/each}
			</tbody>
		</Table>
		<p class="text-fc-xs text-fc-fg-muted">Showing the 50 most recent events.</p>
	{/if}
</section>
