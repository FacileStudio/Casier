<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Alert,
		Badge,
		Button,
		Card,
		EmptyState,
		Field,
		Input,
		SecretField,
		Select,
		SettingsSection,
		Skeleton,
		Spinner,
		Table,
		icons,
		toast,
	} from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		type ApiToken,
		type ApiTokensResponse,
		type Environment,
		type Project,
	} from '$lib/backend';
	import { formatDate, formatDateTime } from '$lib/format';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
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

	let tokens = $state<ApiToken[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	let createOpen = $state(false);
	let newName = $state('');
	let newPermissions = $state('read');
	let newEnvironments = $state('*');
	let newExpiresInDays = $state('90');
	let creating = $state(false);
	let createdToken = $state('');

	let revokeTarget = $state<ApiToken | null>(null);
	let revokeOpen = $state(false);
	let revoking = $state(false);

	const environmentHelper = $derived(
		ctx.environments.length > 0
			? t('tokens.environments.helperList', { slugs: ctx.environments.map((env) => env.slug).join(', ') })
			: t('tokens.environments.helperAll')
	);

	function isRevoked(token: ApiToken) {
		return Boolean(token.revoked_at);
	}

	function isExpired(token: ApiToken) {
		return !isRevoked(token) && new Date(token.expires_at).getTime() < Date.now();
	}

	async function loadTokens() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<ApiTokensResponse>(`/projects/${ctx.slug}/tokens`);
			tokens = res.tokens ?? [];
		} catch (err) {
			loadError = err instanceof Error ? err.message : t('tokens.error.load');
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		createdToken = '';
		newName = '';
		newPermissions = 'read';
		newEnvironments = '*';
		newExpiresInDays = '90';
		createOpen = true;
	}

	async function handleCreate(event: Event) {
		event.preventDefault();
		if (!newName.trim()) return;
		creating = true;
		try {
			const token = await api.post<ApiToken>(
				`/projects/${ctx.slug}/tokens`,
				{
					name: newName.trim(),
					permissions: newPermissions,
					environments: newEnvironments.trim() || '*',
					expires_in_days: Number(newExpiresInDays),
				},
			);
			createdToken = token.token ?? '';
			if (!createdToken) {
				toast.warning(t('tokens.createdNoValue', { name: token.name }));
			} else {
				toast.success(t('tokens.created', { name: token.name }));
			}
			await loadTokens();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('tokens.error.create'));
		} finally {
			creating = false;
		}
	}

	function openRevoke(token: ApiToken) {
		revokeTarget = token;
		revokeOpen = true;
	}

	async function handleRevoke() {
		if (!revokeTarget) return;
		revoking = true;
		try {
			await api.del(`/projects/${ctx.slug}/tokens/${revokeTarget.id}`);
			toast.success(t('tokens.revoked', { name: revokeTarget.name }));
			revokeOpen = false;
			revokeTarget = null;
			await loadTokens();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('tokens.error.revoke'));
		} finally {
			revoking = false;
		}
	}

	$effect(() => {
		if (ctx.project && !ctx.canAdmin) {
			goto(`/projects/${ctx.slug}`, { replaceState: true });
		}
	});

	$effect(() => {
		ctx.slug;
		loadTokens();
	});
</script>

<svelte:head>
	<title>{t('tokens.head', { project: ctx.project?.name ?? t('project.fallbackName') })}</title>
</svelte:head>

{#if ctx.canAdmin}
	<div class="flex flex-col gap-10 px-4 py-6 md:px-8">
		<SettingsSection
			title={t('project.tab.tokens')}
			description={t('tokens.description')}
			bare
		>
			{#snippet actions()}
				<Button icon={icons.plus} onclick={openCreate}>{t('tokens.new')}</Button>
			{/snippet}

			{#if loading}
				<Card class="flex flex-col gap-3">
					{#each Array(3) as _, index (index)}
						<Skeleton class="h-10 w-full" />
					{/each}
				</Card>
			{:else if loadError}
				<ErrorState message={loadError} onretry={loadTokens} />
			{:else if tokens.length === 0}
				<EmptyState
					icon={icons.key}
					title={t('tokens.empty.title')}
					description={t('tokens.empty.description')}
				>
					<Button icon={icons.plus} onclick={openCreate}>{t('tokens.create')}</Button>
				</EmptyState>
			{:else}
				<Table>
					<thead>
						<tr>
							<th>{t('tokens.col.name')}</th>
							<th>{t('tokens.col.owner')}</th>
							<th>{t('tokens.col.status')}</th>
							<th>{t('tokens.col.permissions')}</th>
							<th>{t('tokens.col.environments')}</th>
							<th>{t('tokens.col.lastUsed')}</th>
							<th>{t('tokens.col.expires')}</th>
							<th class="w-1 text-right" aria-label={t('common.actions')}></th>
						</tr>
					</thead>
					<tbody>
						{#each tokens as token (token.id)}
							<tr class={isRevoked(token) || isExpired(token) ? 'opacity-55' : ''}>
								<td class="font-medium text-fc-fg">{token.name}</td>
								<td class="max-w-48 whitespace-nowrap"><PersonLabel person={token.owner} /></td>
								<td>
									{#if isRevoked(token)}
										<Badge tone="neutral">{t('tokens.status.revoked', { date: formatDate(token.revoked_at ?? '') })}</Badge>
									{:else if isExpired(token)}
										<Badge tone="warning">{t('tokens.status.expired')}</Badge>
									{:else}
										<Badge tone="success">{t('tokens.status.active')}</Badge>
									{/if}
								</td>
								<td>
									<Badge tone={token.permissions.includes('write') ? 'accent' : 'neutral'}>
										{token.permissions}
									</Badge>
								</td>
								<td class="font-fc-mono text-fc-xs text-fc-fg-muted">
									{token.environments || t('common.all')}
								</td>
								<td class="whitespace-nowrap text-fc-fg-muted">
									{formatDateTime(token.last_used_at) === '—'
										? t('common.never')
										: formatDateTime(token.last_used_at)}
								</td>
								<td class="whitespace-nowrap text-fc-fg-muted">{formatDate(token.expires_at)}</td>
								<td>
									<div class="flex justify-end">
										{#if !isRevoked(token)}
											<Button
												variant="ghost-danger"
												size="sm"
												icon={icons.revoke}
												onclick={() => openRevoke(token)}
											>
												{t('tokens.revoke')}
											</Button>
										{/if}
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</Table>
			{/if}
		</SettingsSection>
	</div>

	<Drawer bind:open={createOpen} title={t('tokens.drawer.title')}>
		{#if createdToken}
			<div class="flex flex-col gap-4">
				<Alert tone="warning" title={t('tokens.copy.title')}>{t('tokens.copy.body')}</Alert>

				<SecretField
					value={createdToken}
					visible
					autoHideMs={0}
					label={t('tokens.copy.label')}
					helper={t('tokens.copy.helper')}
				/>

				<div class="flex justify-end">
					<Button size="lg" onclick={() => (createOpen = false)}>{t('common.done')}</Button>
				</div>
			</div>
		{:else}
			<form onsubmit={handleCreate} class="flex flex-col gap-4">
				<Field label={t('common.name')} helper={t('tokens.name.helper')}>
					<Input placeholder="ci-deploy" bind:value={newName} required disabled={creating} />
				</Field>
				<Field label={t('tokens.col.permissions')}>
					<Select bind:value={newPermissions} disabled={creating}>
						<option value="read">{t('tokens.permissions.read')}</option>
						<option value="read,write">{t('tokens.permissions.write')}</option>
					</Select>
				</Field>
				<Field label={t('tokens.col.environments')} helper={environmentHelper}>
					<Input
						placeholder="*"
						bind:value={newEnvironments}
						disabled={creating}
						class="font-fc-mono"
					/>
				</Field>
				<Field
					label={t('tokens.expires.label')}
					helper={t('tokens.expires.helper')}
				>
					<Select bind:value={newExpiresInDays} disabled={creating}>
						<option value="30">{t('tokens.expires.30')}</option>
						<option value="90">{t('tokens.expires.90')}</option>
						<option value="365">{t('tokens.expires.365')}</option>
						<option value="730">{t('tokens.expires.730')}</option>
					</Select>
				</Field>
				<div class="flex justify-end gap-2 pt-1">
					<Button
						type="button"
						variant="ghost"
						size="lg"
						disabled={creating}
						onclick={() => (createOpen = false)}
					>
						{t('common.cancel')}
					</Button>
					<Button type="submit" size="lg" disabled={creating || !newName.trim()}>
						{#if creating}<Spinner label={t('common.loading')} size="sm" />{/if}
						{creating ? t('common.creating') : t('tokens.createSubmit')}
					</Button>
				</div>
			</form>
		{/if}
	</Drawer>

	<ConfirmDialog
		bind:open={revokeOpen}
		title={t('tokens.revoke.title')}
		message={t('tokens.revoke.message', { name: revokeTarget?.name ?? '' })}
		confirmLabel={t('tokens.revoke.title')}
		loading={revoking}
		onconfirm={handleRevoke}
	/>
{/if}
