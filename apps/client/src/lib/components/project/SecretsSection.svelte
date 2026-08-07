<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		type Environment,
		type ExportResult,
		type ImportResult,
		type Secret,
		type SecretsResponse,
		type SecretVersion,
		type SecretVersionsResponse,
		type TrashedSecret,
		type TrashResponse,
	} from '$lib/backend';
	import { daysSince, formatDate, formatDateTime } from '$lib/format';
	import { Alert, Badge, Button, Field, Input, Skeleton, Spinner, Table, Textarea } from '@facile/muse';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { toast } from 'svelte-sonner';

	let {
		projectSlug,
		environment,
		canWrite,
	}: { projectSlug: string; environment: Environment; canWrite: boolean } = $props();

	const MASK = '••••••••';
	const actionButton =
		'inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground';

	let secrets = $state<Secret[]>([]);
	let rotationDays = $state(0);
	let loading = $state(true);
	let loadError = $state('');
	let search = $state('');
	let staleOnly = $state(false);
	let revealed = $state<Set<number>>(new Set());
	let revealing = $state<Set<number>>(new Set());
	let values = $state<Map<number, string>>(new Map());

	let addOpen = $state(false);
	let newKey = $state('');
	let newValue = $state('');
	let newTags = $state('');
	let adding = $state(false);

	let editTarget = $state<Secret | null>(null);
	let editOpen = $state(false);
	let editValue = $state('');
	let editLoading = $state(false);
	let saving = $state(false);

	let deleteTarget = $state<Secret | null>(null);
	let deleteOpen = $state(false);
	let deleting = $state(false);

	let historyTarget = $state<Secret | null>(null);
	let historyOpen = $state(false);
	let versions = $state<SecretVersion[]>([]);
	let versionsLoading = $state(false);
	let versionsError = $state('');
	let rollingBack = $state<number | null>(null);

	let trashOpen = $state(false);
	let trashed = $state<TrashedSecret[]>([]);
	let trashLoading = $state(false);
	let trashError = $state('');
	let restoring = $state<number | null>(null);

	let importOpen = $state(false);
	let importContent = $state('');
	let importing = $state(false);
	let exporting = $state(false);

	let basePath = $derived(
		`/projects/${projectSlug}/environments/${environment.slug}/secrets`
	);

	let overdue = $derived(secrets.filter((s) => s.rotation_status === 'overdue'));
	let dueSoon = $derived(secrets.filter((s) => s.rotation_status === 'due_soon'));
	let needsAttention = $derived(overdue.length + dueSoon.length);

	let filtered = $derived(
		secrets
			.filter((s) => {
				if (staleOnly && s.rotation_status !== 'overdue' && s.rotation_status !== 'due_soon') {
					return false;
				}
				const q = search.trim().toLowerCase();
				if (!q) return true;
				return (
					s.key.toLowerCase().includes(q) ||
					(s.tags ?? []).some((t) => t.toLowerCase().includes(q))
				);
			})
			.toSorted((a, b) => a.key.localeCompare(b.key))
	);

	let newKeyError = $derived.by(() => {
		const key = newKey.trim();
		if (!key) return undefined;
		if (/[\s=]/.test(key)) return 'Keys cannot contain whitespace or "="';
		if (secrets.some((s) => s.key === key)) return 'This key already exists in this environment';
		return undefined;
	});

	async function loadSecrets() {
		loading = true;
		loadError = '';
		revealed = new Set();
		values = new Map();
		staleOnly = false;
		try {
			const res = await api.get<SecretsResponse>(basePath, auth.token);
			secrets = res.secrets ?? [];
			rotationDays = res.rotation_days ?? 0;
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load secrets';
		} finally {
			loading = false;
		}
	}

	// Listing no longer carries plaintext, so a value is fetched the moment it is
	// actually needed and each fetch is audited on its own. Values are cached per
	// key only for the life of this view, and dropped whenever the list reloads.
	async function fetchValue(secret: Secret): Promise<string> {
		const cached = values.get(secret.id);
		if (cached !== undefined) return cached;

		const revealedSecret = await api.get<Secret>(
			`${basePath}/${encodeURIComponent(secret.key)}/reveal`,
			auth.token
		);
		const value = revealedSecret.value ?? '';
		values = new Map(values).set(secret.id, value);
		return value;
	}

	async function toggleReveal(secret: Secret) {
		if (revealed.has(secret.id)) {
			const next = new Set(revealed);
			next.delete(secret.id);
			revealed = next;
			return;
		}

		revealing = new Set(revealing).add(secret.id);
		try {
			await fetchValue(secret);
			revealed = new Set(revealed).add(secret.id);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : `Failed to reveal ${secret.key}`);
		} finally {
			const next = new Set(revealing);
			next.delete(secret.id);
			revealing = next;
		}
	}

	async function copyValue(secret: Secret) {
		try {
			const value = await fetchValue(secret);
			await navigator.clipboard.writeText(value);
			toast.success(`Copied ${secret.key} to clipboard`);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Clipboard unavailable');
		}
	}

	async function openTrash() {
		trashOpen = true;
		trashLoading = true;
		trashError = '';
		try {
			const res = await api.get<TrashResponse>(`${basePath}/trash`, auth.token);
			trashed = res.secrets ?? [];
		} catch (err) {
			trashError = err instanceof Error ? err.message : 'Failed to load the trash';
		} finally {
			trashLoading = false;
		}
	}

	async function handleRestore(secret: TrashedSecret) {
		restoring = secret.id;
		try {
			await api.post(`${basePath}/${encodeURIComponent(secret.key)}/restore`, {}, auth.token);
			trashed = trashed.filter((s) => s.id !== secret.id);
			toast.success(`Secret ${secret.key} restored`);
			await loadSecrets();
		} catch (err) {
			toast.error(err instanceof Error ? err.message : `Failed to restore ${secret.key}`);
		} finally {
			restoring = null;
		}
	}

	function openAdd() {
		newKey = '';
		newValue = '';
		newTags = '';
		addOpen = true;
	}

	async function handleAdd(e: Event) {
		e.preventDefault();
		const key = newKey.trim();
		if (!key || newKeyError) return;
		adding = true;
		try {
			const tags = newTags
				.split(',')
				.map((t) => t.trim())
				.filter(Boolean);
			const secret = await api.post<Secret>(basePath, { key, value: newValue, tags }, auth.token);
			secrets = [...secrets, secret];
			addOpen = false;
			toast.success(`Secret ${secret.key} added to ${environment.name}`);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to add secret');
		} finally {
			adding = false;
		}
	}

	// The editor has to start from the current value, and the listing no longer
	// carries it. Opening the drawer before the fetch resolves would let someone
	// save an empty textarea over a live secret.
	async function openEdit(secret: Secret) {
		editTarget = secret;
		editValue = '';
		editLoading = true;
		editOpen = true;
		try {
			editValue = await fetchValue(secret);
		} catch (err) {
			editOpen = false;
			editTarget = null;
			toast.error(err instanceof Error ? err.message : `Failed to read ${secret.key}`);
		} finally {
			editLoading = false;
		}
	}

	async function handleEdit(e: Event) {
		e.preventDefault();
		if (!editTarget) return;
		saving = true;
		try {
			const updated = await api.put<Secret>(
				`${basePath}/${encodeURIComponent(editTarget.key)}`,
				{ value: editValue },
				auth.token
			);
			secrets = secrets.map((s) => (s.id === updated.id ? updated : s));
			values = new Map(values).set(updated.id, editValue);
			toast.success(`Secret ${editTarget.key} updated`);
			editOpen = false;
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to update secret');
		} finally {
			saving = false;
		}
	}

	function openDelete(secret: Secret) {
		deleteTarget = secret;
		deleteOpen = true;
	}

	async function handleDelete() {
		if (!deleteTarget) return;
		deleting = true;
		try {
			await api.del(`${basePath}/${encodeURIComponent(deleteTarget.key)}`, auth.token);
			secrets = secrets.filter((s) => s.id !== deleteTarget!.id);
			toast.success(`Secret ${deleteTarget.key} deleted`);
			deleteOpen = false;
			deleteTarget = null;
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to delete secret');
		} finally {
			deleting = false;
		}
	}

	async function openHistory(secret: Secret) {
		historyTarget = secret;
		historyOpen = true;
		versionsLoading = true;
		versionsError = '';
		versions = [];
		try {
			const res = await api.get<SecretVersionsResponse>(
				`${basePath}/${encodeURIComponent(secret.key)}/versions`,
				auth.token
			);
			versions = (res.versions ?? []).toSorted((a, b) => b.version - a.version);
		} catch (err) {
			versionsError = err instanceof Error ? err.message : 'Failed to load versions';
		} finally {
			versionsLoading = false;
		}
	}

	async function handleRollback(version: number) {
		if (!historyTarget) return;
		rollingBack = version;
		try {
			const updated = await api.post<Secret>(
				`${basePath}/${encodeURIComponent(historyTarget.key)}/rollback`,
				{ version },
				auth.token
			);
			secrets = secrets.map((s) => (s.id === updated.id ? updated : s));
			toast.success(`Rolled ${historyTarget.key} back to v${version}`);
			historyOpen = false;
			historyTarget = null;
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Rollback failed');
		} finally {
			rollingBack = null;
		}
	}

	async function handleImport(e: Event) {
		e.preventDefault();
		if (!importContent.trim()) return;
		importing = true;
		try {
			const res = await api.post<ImportResult>(
				`${basePath}/import`,
				{ content: importContent },
				auth.token
			);
			toast.success(`Imported: ${res.created} created, ${res.updated} updated, ${res.skipped} skipped`);
			importOpen = false;
			importContent = '';
			await loadSecrets();
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Import failed');
		} finally {
			importing = false;
		}
	}

	async function handleExport() {
		exporting = true;
		try {
			const res = await api.get<ExportResult>(`${basePath}/export`, auth.token);
			const blob = new Blob([res.content], { type: 'text/plain' });
			const url = URL.createObjectURL(blob);
			const link = document.createElement('a');
			link.href = url;
			link.download = `${projectSlug}.${environment.slug}.env`;
			link.click();
			URL.revokeObjectURL(url);
			toast.success(`Exported ${environment.name} as .env`);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Export failed');
		} finally {
			exporting = false;
		}
	}

	$effect(() => {
		basePath;
		loadSecrets();
	});
</script>

{#snippet rotationBadge(secret: Secret)}
	{@const age = daysSince(secret.rotated_at)}
	{#if secret.rotation_status === 'overdue'}
		<Badge class="border-foreground bg-foreground whitespace-nowrap text-background">
			<iconify-icon icon="solar:danger-triangle-linear" width="12" aria-hidden="true"></iconify-icon>
			Rotate{age === null ? '' : ` · ${age}d`}
		</Badge>
	{:else if secret.rotation_status === 'due_soon'}
		<Badge class="border-border bg-transparent whitespace-nowrap text-muted-foreground">
			<iconify-icon icon="solar:clock-circle-linear" width="12" aria-hidden="true"></iconify-icon>
			Due soon{age === null ? '' : ` · ${age}d`}
		</Badge>
	{/if}
{/snippet}

<section class="flex flex-col gap-4" aria-label="Secrets">
	{#if !loading && !loadError && rotationDays > 0 && needsAttention > 0}
		<Alert tone="warning" class="flex flex-wrap items-center justify-between gap-3">
			<p class="min-w-0">
				{#if overdue.length > 0}
					<span class="font-medium">
						{overdue.length}
						{overdue.length === 1 ? 'secret has' : 'secrets have'} gone more than {rotationDays} days
						without being rotated.
					</span>
				{/if}
				{#if dueSoon.length > 0}
					<span class="text-muted-foreground">
						{dueSoon.length}
						{dueSoon.length === 1 ? 'is' : 'are'} due for rotation shortly.
					</span>
				{/if}
			</p>
			<Button
				variant="outline"
				size="sm"
				class="h-8 shrink-0"
				onclick={() => (staleOnly = !staleOnly)}
			>
				{staleOnly ? 'Show all secrets' : 'Review them'}
			</Button>
		</Alert>
	{/if}

	<div class="flex flex-wrap items-center gap-2">
		<div class="relative min-w-0 flex-1 sm:max-w-xs">
			<iconify-icon
				icon="solar:magnifer-linear"
				width="16"
				class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
				aria-hidden="true"
			></iconify-icon>
			<Input
				type="search"
				placeholder="Filter by key or tag"
				bind:value={search}
				class="pl-9"
				aria-label="Filter secrets"
			/>
		</div>
		<div class="ml-auto flex items-center gap-2">
			<Button variant="outline" class="h-9" onclick={openTrash}>
				<iconify-icon icon="solar:trash-bin-trash-linear" width="16" aria-hidden="true"></iconify-icon>
				Trash
			</Button>
			<Button
				variant="outline"
				class="h-9"
				disabled={exporting || secrets.length === 0}
				onclick={handleExport}
			>
				<iconify-icon icon="solar:export-linear" width="16" aria-hidden="true"></iconify-icon>
				Export
			</Button>
			{#if canWrite}
				<Button variant="outline" class="h-9" onclick={() => (importOpen = true)}>
					<iconify-icon icon="solar:import-linear" width="16" aria-hidden="true"></iconify-icon>
					Import
				</Button>
				<Button class="h-9" onclick={openAdd}>
					<iconify-icon icon="solar:add-circle-linear" width="16" aria-hidden="true"></iconify-icon>
					Add secret
				</Button>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="flex flex-col gap-2 rounded-lg border border-border bg-card p-4">
			{#each Array(5) as _, i (i)}
				<Skeleton class="h-9 w-full" />
			{/each}
		</div>
	{:else if loadError}
		<ErrorState message={loadError} onretry={loadSecrets} />
	{:else if secrets.length === 0}
		<EmptyState
			title="No secrets in {environment.name}"
			description="Secrets are encrypted at rest and only visible to members of this project."
		>
			{#snippet icon()}
				<iconify-icon icon="solar:key-linear" width="32" aria-hidden="true"></iconify-icon>
			{/snippet}
			{#if canWrite}
				<Button class="h-9" onclick={openAdd}>
					<iconify-icon icon="solar:add-circle-linear" width="16" aria-hidden="true"></iconify-icon>
					Add your first secret
				</Button>
			{/if}
		</EmptyState>
	{:else if filtered.length === 0}
		<EmptyState
			title="No matches"
			description={staleOnly && !search.trim()
				? 'Every secret in this environment is within the rotation policy.'
				: `No secret matches "${search}".`}
		/>
	{:else}
		<Table>
			<thead>
				<tr>
					<th>Key</th>
					<th>Value</th>
					<th>Version</th>
					<th>Last rotated</th>
					<th class="w-1 text-right" aria-label="Actions"></th>
				</tr>
			</thead>
			<tbody>
				{#each filtered as secret (secret.id)}
					<tr>
						<td>
							<div class="flex flex-wrap items-center gap-1.5">
								<code class="rounded bg-muted px-2 py-1 font-mono text-xs text-foreground">{secret.key}</code>
								{#each secret.tags ?? [] as tag (tag)}
									<Badge tone="neutral" class="text-xs">{tag}</Badge>
								{/each}
							</div>
						</td>
						<td>
							{#if revealed.has(secret.id)}
								<code
									class="block max-w-64 truncate rounded bg-muted px-2 py-1 font-mono text-xs text-foreground"
									title={values.get(secret.id) ?? ''}
								>
									{values.get(secret.id) ?? ''}
								</code>
							{:else if revealing.has(secret.id)}
								<Spinner size="sm" />
							{:else}
								<span class="select-none font-mono text-sm text-muted-foreground" aria-label="Hidden value">
									{MASK}
								</span>
							{/if}
						</td>
						<td class="whitespace-nowrap text-muted-foreground">v{secret.version}</td>
						<td>
							<div class="flex flex-wrap items-center gap-2">
								<span class="whitespace-nowrap text-muted-foreground">
									{formatDate(secret.rotated_at)}
								</span>
								{@render rotationBadge(secret)}
							</div>
						</td>
						<td>
							<div class="flex items-center justify-end gap-0.5">
								<button
									type="button"
									class={actionButton}
									aria-label={revealed.has(secret.id) ? `Hide ${secret.key}` : `Reveal ${secret.key}`}
									onclick={() => toggleReveal(secret)}
								>
									{#if revealed.has(secret.id)}
										<iconify-icon icon="solar:eye-closed-linear" width="16" aria-hidden="true"></iconify-icon>
									{:else}
										<iconify-icon icon="solar:eye-linear" width="16" aria-hidden="true"></iconify-icon>
									{/if}
								</button>
								<button
									type="button"
									class={actionButton}
									aria-label={`Copy ${secret.key}`}
									onclick={() => copyValue(secret)}
								>
									<iconify-icon icon="solar:copy-linear" width="16" aria-hidden="true"></iconify-icon>
								</button>
								<button
									type="button"
									class={actionButton}
									aria-label={`History of ${secret.key}`}
									onclick={() => openHistory(secret)}
								>
									<iconify-icon icon="solar:history-linear" width="16" aria-hidden="true"></iconify-icon>
								</button>
								{#if canWrite}
									<button
										type="button"
										class={actionButton}
										aria-label={`Edit ${secret.key}`}
										onclick={() => openEdit(secret)}
									>
										<iconify-icon icon="solar:pen-new-square-linear" width="16" aria-hidden="true"></iconify-icon>
									</button>
									<button
										type="button"
										class={actionButton}
										aria-label={`Delete ${secret.key}`}
										onclick={() => openDelete(secret)}
									>
										<iconify-icon icon="solar:trash-bin-2-linear" width="16" aria-hidden="true"></iconify-icon>
									</button>
								{/if}
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</Table>
	{/if}
</section>

<Drawer bind:open={addOpen} title="Add secret to {environment.name}">
	<form onsubmit={handleAdd} class="flex flex-col gap-4">
		<Field label="Key" error={newKeyError}>
			<Input placeholder="DATABASE_URL" bind:value={newKey} required disabled={adding} class="font-mono" />
		</Field>
		<Field label="Value">
			<Textarea rows={3} placeholder="Secret value" bind:value={newValue} disabled={adding} class="font-mono" />
		</Field>
		<Field label="Tags" helper="Comma-separated, optional">
			<Input placeholder="backend, database" bind:value={newTags} disabled={adding} />
		</Field>
		<div class="flex justify-end gap-2 pt-1">
			<Button type="button" variant="ghost" class="h-9" disabled={adding} onclick={() => (addOpen = false)}>
				Cancel
			</Button>
			<Button type="submit" class="h-9" disabled={adding || !newKey.trim() || !!newKeyError}>
				{#if adding}<Spinner size="sm" />{/if}
				{adding ? 'Adding…' : 'Add secret'}
			</Button>
		</div>
	</form>
</Drawer>

<Drawer bind:open={trashOpen} title="Trash — {environment.name}">
	{#if trashLoading}
		<div class="flex flex-col gap-2">
			{#each Array(3) as _, i (i)}
				<Skeleton class="h-9 w-full" />
			{/each}
		</div>
	{:else if trashError}
		<ErrorState message={trashError} onretry={openTrash} />
	{:else if trashed.length === 0}
		<EmptyState
			title="Nothing in the trash"
			description="Deleted secrets stay here for 30 days before they are destroyed."
		>
			{#snippet icon()}
				<iconify-icon icon="solar:trash-bin-trash-linear" width="32" aria-hidden="true"></iconify-icon>
			{/snippet}
		</EmptyState>
	{:else}
		<div class="flex flex-col gap-3">
			<p class="text-sm text-muted-foreground">
				A deleted key stays reserved until it is restored or purged, so it cannot be recreated
				from scratch while it sits here.
			</p>
			{#each trashed as secret (secret.id)}
				<div class="flex items-center justify-between gap-3 rounded-lg border border-border p-3">
					<div class="flex min-w-0 flex-col gap-1">
						<code class="w-fit rounded bg-muted px-2 py-1 font-mono text-xs text-foreground">
							{secret.key}
						</code>
						<span class="text-xs text-muted-foreground">
							Deleted {formatDateTime(secret.deleted_at)} · purged {formatDate(secret.purge_at)}
						</span>
					</div>
					{#if canWrite}
						<Button
							variant="outline"
							class="h-8 shrink-0"
							disabled={restoring === secret.id}
							onclick={() => handleRestore(secret)}
						>
							{#if restoring === secret.id}<Spinner size="sm" />{/if}
							Restore
						</Button>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</Drawer>

<Drawer bind:open={editOpen} title="Edit {editTarget?.key}">
	<form onsubmit={handleEdit} class="flex flex-col gap-4">
		<Field label="Value" helper="Saving creates a new version">
			<Textarea rows={4} bind:value={editValue} disabled={saving || editLoading} class="font-mono" />
		</Field>
		<div class="flex justify-end gap-2 pt-1">
			<Button type="button" variant="ghost" class="h-9" disabled={saving} onclick={() => (editOpen = false)}>
				Cancel
			</Button>
			<Button type="submit" class="h-9" disabled={saving || editLoading}>
				{#if saving || editLoading}<Spinner size="sm" />{/if}
				{editLoading ? 'Loading…' : saving ? 'Saving…' : 'Save'}
			</Button>
		</div>
	</form>
</Drawer>

<Drawer bind:open={historyOpen} title="History of {historyTarget?.key}">
	{#if versionsLoading}
		<div class="flex justify-center py-6">
			<Spinner size="md" />
		</div>
	{:else if versionsError}
		<p class="py-4 text-center text-sm text-destructive">{versionsError}</p>
	{:else if versions.length === 0}
		<p class="py-4 text-center text-sm text-muted-foreground">No previous versions.</p>
	{:else}
		<ul class="flex flex-col gap-1">
			{#each versions as version (version.version)}
				<li class="flex items-center justify-between gap-3 rounded-md px-2 py-2 hover:bg-muted">
					<div class="min-w-0">
						<p class="flex items-center gap-2 text-sm text-foreground">
							v{version.version}
							{#if historyTarget && version.version === historyTarget.version}
								<Badge tone="accent" class="text-xs">current</Badge>
							{/if}
						</p>
						<p class="truncate text-xs text-muted-foreground">
							{version.changed_by || 'unknown'} — {formatDateTime(version.created_at)}
						</p>
					</div>
					{#if canWrite && historyTarget && version.version !== historyTarget.version}
						<Button
							variant="outline"
							size="sm"
							class="h-9"
							disabled={rollingBack !== null}
							onclick={() => handleRollback(version.version)}
						>
							{rollingBack === version.version ? 'Rolling back…' : 'Roll back'}
						</Button>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</Drawer>

<Drawer bind:open={importOpen} title="Import into {environment.name}">
	<form onsubmit={handleImport} class="flex flex-col gap-4">
		<Field label=".env content" helper="Paste KEY=value lines; existing keys are updated">
			<Textarea
				rows={8}
				placeholder={'DATABASE_URL=postgres://…\nAPI_KEY=…'}
				bind:value={importContent}
				disabled={importing}
				class="font-mono"
			/>
		</Field>
		<div class="flex justify-end gap-2 pt-1">
			<Button type="button" variant="ghost" class="h-9" disabled={importing} onclick={() => (importOpen = false)}>
				Cancel
			</Button>
			<Button type="submit" class="h-9" disabled={importing || !importContent.trim()}>
				{#if importing}<Spinner size="sm" />{/if}
				{importing ? 'Importing…' : 'Import'}
			</Button>
		</div>
	</form>
</Drawer>

<ConfirmDialog
	bind:open={deleteOpen}
	title="Delete secret"
	message={`This moves ${deleteTarget?.key} out of the ${environment.name} environment and into the trash, where it can be restored for 30 days before it is destroyed for good. Anything reading it will break immediately. The key stays reserved until it is restored or purged.`}
	confirmLabel="Delete secret"
	loading={deleting}
	onconfirm={handleDelete}
/>
