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
			? `Comma-separated environment slugs, or * for every environment. This project has: ${ctx.environments.map((env) => env.slug).join(', ')}.`
			: 'Comma-separated environment slugs, or * for every environment.'
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
			const res = await api.get<ApiTokensResponse>(`/projects/${ctx.slug}/tokens`, auth.token);
			tokens = res.tokens ?? [];
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load tokens';
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
				auth.token
			);
			createdToken = token.token ?? '';
			if (!createdToken) {
				toast.warning(
					`Token "${token.name}" was created, but the server did not return its value. Revoke it and issue another.`
				);
			} else {
				toast.success(`Token "${token.name}" created`);
			}
			await loadTokens();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to create token');
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
			await api.del(`/projects/${ctx.slug}/tokens/${revokeTarget.id}`, auth.token);
			toast.success(`Token "${revokeTarget.name}" revoked`);
			revokeOpen = false;
			revokeTarget = null;
			await loadTokens();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to revoke token');
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
	<title>API tokens — {ctx.project?.name ?? 'Project'} — Casier</title>
</svelte:head>

{#if ctx.canAdmin}
	<div class="flex flex-col gap-10 px-4 py-6 md:px-8">
		<SettingsSection
			title="API tokens"
			description="Machine credentials for CI and the casier CLI. A token is bound to this project, to a permission set and to the environments it names — it can never read past them, and revoking one never touches the others."
			bare
		>
			{#snippet actions()}
				<Button icon={icons.plus} onclick={openCreate}>New token</Button>
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
					title="No API tokens"
					description="Create one to give a pipeline or a machine scoped access to this project."
				>
					<Button icon={icons.plus} onclick={openCreate}>Create a token</Button>
				</EmptyState>
			{:else}
				<Table>
					<thead>
						<tr>
							<th>Name</th>
							<th>Status</th>
							<th>Permissions</th>
							<th>Environments</th>
							<th>Last used</th>
							<th>Expires</th>
							<th class="w-1 text-right" aria-label="Actions"></th>
						</tr>
					</thead>
					<tbody>
						{#each tokens as token (token.id)}
							<tr class={isRevoked(token) || isExpired(token) ? 'opacity-55' : ''}>
								<td class="font-medium text-fc-fg">{token.name}</td>
								<td>
									{#if isRevoked(token)}
										<Badge tone="neutral">Revoked {formatDate(token.revoked_at ?? '')}</Badge>
									{:else if isExpired(token)}
										<Badge tone="warning">Expired</Badge>
									{:else}
										<Badge tone="success">Active</Badge>
									{/if}
								</td>
								<td>
									<Badge tone={token.permissions.includes('write') ? 'accent' : 'neutral'}>
										{token.permissions}
									</Badge>
								</td>
								<td class="font-fc-mono text-fc-xs text-fc-fg-muted">
									{token.environments || 'all'}
								</td>
								<td class="whitespace-nowrap text-fc-fg-muted">
									{formatDateTime(token.last_used_at) === '—'
										? 'Never'
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
												Revoke
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

	<Drawer bind:open={createOpen} title="New API token">
		{#if createdToken}
			<div class="flex flex-col gap-4">
				<Alert tone="warning" title="Copy it now">
					This is the only time the token is shown. Casier stores a hash of it, so it cannot be
					shown again — losing it means issuing another.
				</Alert>

				<SecretField
					value={createdToken}
					visible
					autoHideMs={0}
					label="Token"
					helper="Store it in your CI secret store, or hand it to the CLI as CASIER_TOKEN."
				/>

				<div class="flex justify-end">
					<Button size="lg" onclick={() => (createOpen = false)}>Done</Button>
				</div>
			</div>
		{:else}
			<form onsubmit={handleCreate} class="flex flex-col gap-4">
				<Field label="Name" helper="Where the token will live — a machine, a pipeline, a script.">
					<Input placeholder="ci-deploy" bind:value={newName} required disabled={creating} />
				</Field>
				<Field label="Permissions">
					<Select bind:value={newPermissions} disabled={creating}>
						<option value="read">Read — fetch secrets</option>
						<option value="read,write">Read and write — fetch and update secrets</option>
					</Select>
				</Field>
				<Field label="Environments" helper={environmentHelper}>
					<Input
						placeholder="*"
						bind:value={newEnvironments}
						disabled={creating}
						class="font-fc-mono"
					/>
				</Field>
				<Field
					label="Expires in"
					helper="A machine credential that never lapses is the one nobody remembers to rotate."
				>
					<Select bind:value={newExpiresInDays} disabled={creating}>
						<option value="30">30 days</option>
						<option value="90">90 days</option>
						<option value="365">1 year</option>
						<option value="730">2 years</option>
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
						Cancel
					</Button>
					<Button type="submit" size="lg" disabled={creating || !newName.trim()}>
						{#if creating}<Spinner size="sm" class="border-current border-t-transparent" />{/if}
						{creating ? 'Creating…' : 'Create token'}
					</Button>
				</div>
			</form>
		{/if}
	</Drawer>

	<ConfirmDialog
		bind:open={revokeOpen}
		title="Revoke token"
		message={`"${revokeTarget?.name ?? ''}" stops working immediately, and any CI pipeline or CLI still using it starts failing. It cannot be un-revoked. The row stays listed so the audit log still names what it changed.`}
		confirmLabel="Revoke token"
		loading={revoking}
		onconfirm={handleRevoke}
	/>
{/if}
