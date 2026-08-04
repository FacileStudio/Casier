<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { api, type ApiToken, type ApiTokensResponse } from '$lib/backend';
	import { formatDate, formatDateTime } from '$lib/format';
	import { Alert, Badge, Button, Field, Input, Modal, Select, Skeleton, Spinner, Table } from '@facile/lib';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { toast } from 'svelte-sonner';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import KeySquareIcon from '@lucide/svelte/icons/key-square';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';

	let { spaceSlug }: { spaceSlug: string } = $props();

	let tokens = $state<ApiToken[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	let createOpen = $state(false);
	let newName = $state('');
	let newPermissions = $state('read');
	let newEnvironments = $state('*');
	let creating = $state(false);
	let createdToken = $state<string | null>(null);

	let revokeTarget = $state<ApiToken | null>(null);
	let revokeOpen = $state(false);
	let revoking = $state(false);

	async function loadTokens() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<ApiTokensResponse>(`/spaces/${spaceSlug}/tokens`, auth.token);
			tokens = res.tokens ?? [];
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load tokens';
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		newName = '';
		newPermissions = 'read';
		newEnvironments = '*';
		createdToken = null;
		createOpen = true;
	}

	async function handleCreate(e: Event) {
		e.preventDefault();
		if (!newName.trim()) return;
		creating = true;
		try {
			const token = await api.post<ApiToken>(
				`/spaces/${spaceSlug}/tokens`,
				{
					name: newName.trim(),
					permissions: newPermissions,
					environments: newEnvironments.trim() || '*',
				},
				auth.token
			);
			createdToken = token.token ?? null;
			toast.success(`Token "${token.name}" created`);
			await loadTokens();
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to create token');
		} finally {
			creating = false;
		}
	}

	async function copyCreatedToken() {
		if (!createdToken) return;
		try {
			await navigator.clipboard.writeText(createdToken);
			toast.success('Token copied to clipboard');
		} catch {
			toast.error('Clipboard unavailable');
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
			await api.del(`/spaces/${spaceSlug}/tokens/${revokeTarget.id}`, auth.token);
			tokens = tokens.filter((t) => t.id !== revokeTarget!.id);
			toast.success(`Token "${revokeTarget.name}" revoked`);
			revokeOpen = false;
			revokeTarget = null;
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to revoke token');
		} finally {
			revoking = false;
		}
	}

	$effect(() => {
		spaceSlug;
		loadTokens();
	});
</script>

<section class="flex flex-col gap-4" aria-label="API tokens">
	<div class="flex flex-wrap items-center justify-between gap-3">
		<p class="text-fc-md text-fc-fg-muted">
			Tokens let CI pipelines and the Casier CLI read secrets without a user session.
		</p>
		<Button onclick={openCreate}>
			<PlusIcon class="size-4" aria-hidden="true" />
			New token
		</Button>
	</div>

	{#if loading}
		<div class="flex flex-col gap-2 rounded-fc-md bg-fc-component px-6 py-5">
			{#each Array(3) as _, i (i)}
				<Skeleton class="h-9 w-full" />
			{/each}
		</div>
	{:else if loadError}
		<ErrorState message={loadError} onretry={loadTokens} />
	{:else if tokens.length === 0}
		<EmptyState
			title="No API tokens"
			description="Create a token to give machines scoped access to this space."
		>
			{#snippet icon()}
				<KeySquareIcon class="size-10" aria-hidden="true" />
			{/snippet}
			<Button onclick={openCreate}>
				<PlusIcon class="size-4" aria-hidden="true" />
				Create a token
			</Button>
		</EmptyState>
	{:else}
		<Table>
			<thead>
				<tr>
					<th>Name</th>
					<th>Permissions</th>
					<th>Environments</th>
					<th>Last used</th>
					<th>Created</th>
					<th class="w-1 text-right"><span class="sr-only">Actions</span></th>
				</tr>
			</thead>
			<tbody>
				{#each tokens as token (token.id)}
					<tr>
						<td class="font-medium text-fc-fg">{token.name}</td>
						<td>
							<Badge tone={token.permissions.includes('write') ? 'accent' : 'neutral'}>
								{token.permissions}
							</Badge>
						</td>
						<td><code class="font-mono text-fc-sm">{token.environments || '*'}</code></td>
						<td class="whitespace-nowrap text-fc-fg-muted">
							{formatDateTime(token.last_used_at) === '—' ? 'Never' : formatDateTime(token.last_used_at)}
						</td>
						<td class="whitespace-nowrap text-fc-fg-muted">{formatDate(token.created_at)}</td>
						<td>
							<div class="flex justify-end">
								<button
									type="button"
									class="inline-flex h-8 w-8 items-center justify-center rounded-fc-sm text-fc-fg-muted transition-colors hover:bg-fc-surface hover:text-fc-danger focus-visible:outline-2 focus-visible:outline-fc-accent"
									aria-label={`Revoke ${token.name}`}
									onclick={() => openRevoke(token)}
								>
									<Trash2Icon class="size-4" aria-hidden="true" />
								</button>
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</Table>
	{/if}
</section>

<Modal bind:open={createOpen} title="Create API token">
	{#if createdToken}
		<div class="flex flex-col gap-4">
			<Alert tone="success" title="Token created">
				Copy it now — it will not be shown again.
			</Alert>
			<div class="flex items-center gap-2 rounded-fc-sm bg-fc-surface p-3">
				<code class="min-w-0 flex-1 break-all font-mono text-fc-sm text-fc-fg">{createdToken}</code>
				<button
					type="button"
					class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-fc-sm text-fc-fg-muted transition-colors hover:bg-fc-fg/10 hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-fc-accent"
					aria-label="Copy token"
					onclick={copyCreatedToken}
				>
					<CopyIcon class="size-4" aria-hidden="true" />
				</button>
			</div>
			<div class="flex justify-end">
				<Button onclick={() => (createOpen = false)}>Done</Button>
			</div>
		</div>
	{:else}
		<form onsubmit={handleCreate} class="flex flex-col gap-4">
			<Field label="Name">
				<Input placeholder="ci-deploy" bind:value={newName} required disabled={creating} />
			</Field>
			<Field label="Permissions">
				<Select bind:value={newPermissions} disabled={creating}>
					<option value="read">Read — fetch secrets</option>
					<option value="read,write">Read and write — fetch and update secrets</option>
				</Select>
			</Field>
			<Field label="Environments" helper={'Comma-separated environment slugs, or * for all'}>
				<Input placeholder="*" bind:value={newEnvironments} disabled={creating} class="font-mono" />
			</Field>
			<div class="flex justify-end gap-2 pt-1">
				<Button type="button" variant="ghost" disabled={creating} onclick={() => (createOpen = false)}>
					Cancel
				</Button>
				<Button type="submit" disabled={creating || !newName.trim()}>
					{#if creating}<Spinner size="sm" />{/if}
					{creating ? 'Creating…' : 'Create token'}
				</Button>
			</div>
		</form>
	{/if}
</Modal>

<ConfirmDialog
	bind:open={revokeOpen}
	title="Revoke token"
	message={`This immediately revokes the token "${revokeTarget?.name}". Anything using it will lose access.`}
	confirmLabel="Revoke token"
	loading={revoking}
	onconfirm={handleRevoke}
/>
