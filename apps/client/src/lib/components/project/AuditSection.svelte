<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { api, type AuditLogEntry, type AuditLogsResponse } from '$lib/backend';
	import { formatDateTime } from '$lib/format';
	import { Badge, Skeleton, Table } from '@facile/muse';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';

	let { projectSlug }: { projectSlug: string } = $props();

	let logs = $state<AuditLogEntry[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	async function loadLogs() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<AuditLogsResponse>(`/projects/${projectSlug}/audit?limit=50`, auth.token);
			logs = res.logs ?? [];
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load audit log';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		projectSlug;
		loadLogs();
	});
</script>

<section class="flex flex-col gap-4" aria-label="Audit log">
	{#if loading}
		<div class="flex flex-col gap-2 rounded-lg border border-border bg-card p-4">
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
				<iconify-icon icon="solar:history-linear" width="32" aria-hidden="true"></iconify-icon>
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
						<td class="whitespace-nowrap text-muted-foreground">{formatDateTime(log.created_at)}</td>
						<td class="whitespace-nowrap">{log.user_email || '—'}</td>
						<td><Badge tone="neutral" class="text-xs">{log.action}</Badge></td>
						<td>
							<code class="rounded bg-muted px-2 py-1 font-mono text-xs text-foreground">
								{log.resource || '—'}
							</code>
						</td>
						<td class="max-w-64 truncate text-muted-foreground" title={log.detail}>{log.detail || '—'}</td>
					</tr>
				{/each}
			</tbody>
		</Table>
		<p class="text-xs text-muted-foreground">Showing the 50 most recent events.</p>
	{/if}
</section>
