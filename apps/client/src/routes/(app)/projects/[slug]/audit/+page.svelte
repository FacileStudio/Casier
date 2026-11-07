<script lang="ts">
	import { getContext } from 'svelte';
	import {
		Badge,
		Button,
		Card,
		EmptyState,
		SettingsSection,
		Skeleton,
		Table,
		icons,
	} from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		DELETED_USER,
		type AuditLogEntry,
		type AuditLogsResponse,
		type Environment,
		type Project,
	} from '$lib/backend';
	import { formatDateTime } from '$lib/format';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import PersonLabel from '$lib/components/PersonLabel.svelte';
	import { t } from '$lib/i18n/index.svelte';

	type ProjectContext = {
		readonly project: Project | null;
		readonly environments: Environment[];
		readonly canWrite: boolean;
		readonly canAdmin: boolean;
		readonly slug: string;
		reloadEnvironments: () => Promise<void>;
		reloadProject: () => Promise<void>;
	};

	const ctx = getContext<ProjectContext>('project');

	let logs = $state<AuditLogEntry[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	function actionTone(action: string): 'danger' | 'warning' | 'success' | 'neutral' {
		if (action.includes('delete') || action.includes('purge') || action.includes('revoke')) {
			return 'danger';
		}
		if (action.includes('reveal') || action.includes('export')) return 'warning';
		if (action.includes('create') || action.includes('restore')) return 'success';
		return 'neutral';
	}

	async function loadLogs() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<AuditLogsResponse>(
				`/projects/${ctx.slug}/audit?limit=50`,
			);
			logs = res.logs ?? [];
		} catch (err) {
			loadError = err instanceof Error ? err.message : t('audit.error.load');
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		ctx.slug;
		loadLogs();
	});
</script>

<svelte:head>
	<title>{t('audit.head', { project: ctx.project?.name ?? t('project.fallbackName') })}</title>
</svelte:head>

<div class="flex flex-col gap-10 px-4 py-6 md:px-8">
	<SettingsSection
		title={t('project.tab.audit')}
		description={t('audit.description')}
		bare
	>
		{#snippet actions()}
			<Button variant="secondary" icon={icons.refresh} disabled={loading} onclick={loadLogs}>
				{t('common.refresh')}
			</Button>
		{/snippet}

		{#if loading}
			<Card class="flex flex-col gap-3">
				{#each Array(6) as _, index (index)}
					<Skeleton class="h-8 w-full" />
				{/each}
			</Card>
		{:else if loadError}
			<ErrorState message={loadError} onretry={loadLogs} />
		{:else if logs.length === 0}
			<EmptyState
				icon={icons.history}
				title={t('audit.empty.title')}
				description={t('audit.empty.description')}
			/>
		{:else}
			<Table>
				<thead>
					<tr>
						<th>{t('audit.col.when')}</th>
						<th>{t('audit.col.who')}</th>
						<th>{t('audit.col.action')}</th>
						<th>{t('audit.col.resource')}</th>
						<th>{t('audit.col.detail')}</th>
					</tr>
				</thead>
				<tbody>
					{#each logs as log (log.id)}
						<tr>
							<td class="whitespace-nowrap text-fc-fg-muted">{formatDateTime(log.created_at)}</td>
							<td class="max-w-48 whitespace-nowrap">
								<PersonLabel person={log.actor} fallback={log.user_email && log.user_email !== DELETED_USER ? log.user_email : undefined} />
							</td>
							<td><Badge tone={actionTone(log.action)}>{log.action}</Badge></td>
							<td class="font-fc-mono text-fc-xs text-fc-fg-muted">{log.resource || '—'}</td>
							<td class="max-w-64 truncate text-fc-fg-muted" title={log.detail}>
								{log.detail || '—'}
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
			<p class="text-fc-xs text-fc-fg-muted">{t('audit.limit')}</p>
		{/if}
	</SettingsSection>
</div>
