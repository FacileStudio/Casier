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
	import {
		Alert,
		Badge,
		Button,
		Card,
		EmptyState,
		Field,
		IconButton,
		Input,
		REDACTED,
		SecretField,
		Skeleton,
		Spinner,
		Table,
		Textarea,
		icons,
		toast,
	} from '@facile/muse';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import PersonLabel from '$lib/components/PersonLabel.svelte';

	let {
		projectSlug,
		environment,
		canWrite,
	}: { projectSlug: string; environment: Environment; canWrite: boolean } = $props();

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
			const res = await api.get<SecretsResponse>(basePath);
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
			toast.danger(err instanceof Error ? err.message : `Failed to reveal ${secret.key}`);
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
			toast.danger(err instanceof Error ? err.message : 'Clipboard unavailable');
		}
	}

	async function openTrash() {
		trashOpen = true;
		trashLoading = true;
		trashError = '';
		try {
			const res = await api.get<TrashResponse>(`${basePath}/trash`);
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
			await api.post(`${basePath}/${encodeURIComponent(secret.key)}/restore`, {});
			trashed = trashed.filter((s) => s.id !== secret.id);
			toast.success(`Secret ${secret.key} restored`);
			await loadSecrets();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : `Failed to restore ${secret.key}`);
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
			const secret = await api.post<Secret>(basePath, { key, value: newValue, tags });
			secrets = [...secrets, secret];
			addOpen = false;
			toast.success(`Secret ${secret.key} added to ${environment.name}`);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to add secret');
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
			toast.danger(err instanceof Error ? err.message : `Failed to read ${secret.key}`);
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
			);
			secrets = secrets.map((s) => (s.id === updated.id ? updated : s));
			values = new Map(values).set(updated.id, editValue);
			toast.success(`Secret ${editTarget.key} updated`);
			editOpen = false;
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to update secret');
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
			await api.del(`${basePath}/${encodeURIComponent(deleteTarget.key)}`);
			secrets = secrets.filter((s) => s.id !== deleteTarget!.id);
			toast.success(`Secret ${deleteTarget.key} deleted`);
			deleteOpen = false;
			deleteTarget = null;
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to delete secret');
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
			);
			secrets = secrets.map((s) => (s.id === updated.id ? updated : s));
			toast.success(`Rolled ${historyTarget.key} back to v${version}`);
			historyOpen = false;
			historyTarget = null;
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Rollback failed');
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
			);
			toast.success(`Imported: ${res.created} created, ${res.updated} updated, ${res.skipped} skipped`);
			importOpen = false;
			importContent = '';
			await loadSecrets();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Import failed');
		} finally {
			importing = false;
		}
	}

	async function handleExport() {
		exporting = true;
		try {
			const res = await api.get<ExportResult>(`${basePath}/export`);
			const blob = new Blob([res.content], { type: 'text/plain' });
			const url = URL.createObjectURL(blob);
			const link = document.createElement('a');
			link.href = url;
			link.download = `${projectSlug}.${environment.slug}.env`;
			link.click();
			URL.revokeObjectURL(url);
			toast.success(`Exported ${environment.name} as .env`);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Export failed');
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
		<Badge tone="danger" class="whitespace-nowrap">
			Rotate{age === null ? '' : ` · ${age}d`}
		</Badge>
	{:else if secret.rotation_status === 'due_soon'}
		<Badge tone="warning" class="whitespace-nowrap">
			Due soon{age === null ? '' : ` · ${age}d`}
		</Badge>
	{/if}
{/snippet}

{#snippet valueField(secret: Secret)}
	{@const isRevealed = revealed.has(secret.id)}
	{@const isRevealing = revealing.has(secret.id)}
	<SecretField
		value={isRevealed ? (values.get(secret.id) ?? '') : REDACTED}
		sensitive={false}
		copyable={isRevealed}
		class="min-w-0"
		onCopy={(ok) =>
			ok ? toast.success(`Copied ${secret.key} to clipboard`) : toast.danger('Clipboard unavailable')}
	>
		{#snippet actions()}
			{#if !isRevealed}
				<IconButton
					variant="ghost"
					aria-label="Copy {secret.key} to clipboard"
					title="Copy"
					onclick={() => copyValue(secret)}
				>
					<iconify-icon icon={icons.copy} width="18" height="18" class="block"></iconify-icon>
				</IconButton>
			{/if}
			<IconButton
				variant="ghost"
				disabled={isRevealing}
				aria-label={isRevealed ? `Hide ${secret.key}` : `Reveal ${secret.key}`}
				title={isRevealed ? 'Hide' : 'Reveal'}
				aria-pressed={isRevealed}
				onclick={() => toggleReveal(secret)}
			>
				{#if isRevealing}
					<Spinner label="Loading" size="sm" />
				{:else}
					<iconify-icon
						icon={isRevealed ? icons.eyeClosed : icons.eye}
						width="18"
						height="18"
						class="block"
					></iconify-icon>
				{/if}
			</IconButton>
		{/snippet}
	</SecretField>
{/snippet}

{#snippet secretActions(secret: Secret)}
	<Button
		variant="ghost"
		icon={icons.history}
		aria-label="History of {secret.key}"
		title="History"
		onclick={() => openHistory(secret)}
	/>
	{#if canWrite}
		<Button
			variant="ghost"
			icon={icons.edit}
			aria-label="Edit {secret.key}"
			title="Edit"
			onclick={() => openEdit(secret)}
		/>
		<Button
			variant="ghost-danger"
			icon={icons.remove}
			aria-label="Delete {secret.key}"
			title="Delete"
			onclick={() => openDelete(secret)}
		/>
	{/if}
{/snippet}

<section class="flex flex-col gap-4" aria-label="Secrets">
	{#if environment.protected}
		<Alert tone="info">
			{environment.name} is protected — only an owner or an admin can write to it.
		</Alert>
	{/if}

	{#if !loading && !loadError && rotationDays > 0 && needsAttention > 0}
		<Alert tone="warning" class="flex flex-wrap items-center justify-between gap-3">
			<span class="min-w-0">
				{#if overdue.length > 0}
					<span class="font-medium">
						{overdue.length}
						{overdue.length === 1 ? 'secret has' : 'secrets have'} gone more than {rotationDays} days
						without being rotated.
					</span>
				{/if}
				{#if dueSoon.length > 0}
					<span>
						{dueSoon.length}
						{dueSoon.length === 1 ? 'is' : 'are'} due for rotation shortly.
					</span>
				{/if}
			</span>
			<Button
				variant="outline"
				size="sm"
				icon={staleOnly ? icons.refresh : icons.filter}
				class="shrink-0"
				onclick={() => (staleOnly = !staleOnly)}
			>
				{staleOnly ? 'Show all secrets' : 'Review them'}
			</Button>
		</Alert>
	{/if}

	<div class="flex flex-col gap-3 sm:flex-row sm:items-center">
		<div class="min-w-0 flex-1">
			<Input
				type="search"
				placeholder="Filter by key or tag"
				bind:value={search}
				aria-label="Filter secrets"
			/>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<Button variant="outline" icon={icons.remove} onclick={openTrash}>Trash</Button>
			<Button
				variant="outline"
				icon={icons.download}
				disabled={exporting || secrets.length === 0}
				onclick={handleExport}
			>
				{exporting ? 'Exporting…' : 'Export'}
			</Button>
			{#if canWrite}
				<Button variant="outline" icon={icons.upload} onclick={() => (importOpen = true)}>
					Import
				</Button>
				<Button icon={icons.plus} onclick={openAdd}>Add secret</Button>
			{/if}
		</div>
	</div>

	{#if loading}
		<Card class="flex flex-col gap-2">
			{#each Array(5) as _, i (i)}
				<Skeleton class="h-11 w-full" />
			{/each}
		</Card>
	{:else if loadError}
		<ErrorState message={loadError} onretry={loadSecrets} />
	{:else if secrets.length === 0}
		<EmptyState
			icon={icons.key}
			title="No secrets in {environment.name}"
			description="Secrets are encrypted at rest and only visible to members of this project."
		>
			{#if canWrite}
				<Button icon={icons.plus} onclick={openAdd}>Add your first secret</Button>
			{/if}
		</EmptyState>
	{:else if filtered.length === 0}
		<EmptyState
			icon={icons.search}
			title="No matches"
			description={staleOnly && !search.trim()
				? 'Every secret in this environment is within the rotation policy.'
				: `No secret matches "${search}".`}
		>
			<Button
				variant="outline"
				icon={icons.refresh}
				onclick={() => {
					search = '';
					staleOnly = false;
				}}
			>
				Clear filters
			</Button>
		</EmptyState>
	{:else}
		<div class="hidden lg:block">
			<Table>
				<thead>
					<tr>
						<th scope="col">Key</th>
						<th scope="col">Value</th>
						<th scope="col">Version</th>
						<th scope="col">Last rotated</th>
						<th scope="col" class="text-right">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each filtered as secret (secret.id)}
						<tr>
							<td>
								<div class="flex flex-wrap items-center gap-1.5">
									<span class="font-fc-mono text-fc-xs text-fc-fg">{secret.key}</span>
									{#each secret.tags ?? [] as tag (tag)}
										<Badge tone="neutral">{tag}</Badge>
									{/each}
								</div>
							</td>
							<td>
								<div class="max-w-sm min-w-56">{@render valueField(secret)}</div>
							</td>
							<td class="whitespace-nowrap text-fc-fg-muted">v{secret.version}</td>
							<td>
								<div class="flex flex-wrap items-center gap-2">
									<span class="whitespace-nowrap text-fc-fg-muted">
										{formatDate(secret.rotated_at)}
									</span>
									{@render rotationBadge(secret)}
								</div>
							</td>
							<td>
								<div class="flex items-center justify-end gap-1">
									{@render secretActions(secret)}
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		</div>

		<div class="flex flex-col gap-2 lg:hidden">
			{#each filtered as secret (secret.id)}
				<Card class="flex flex-col gap-4">
					<div class="flex flex-wrap items-center gap-1.5">
						<span class="min-w-0 truncate font-fc-mono text-fc-sm font-medium text-fc-fg">
							{secret.key}
						</span>
						{#each secret.tags ?? [] as tag (tag)}
							<Badge tone="neutral">{tag}</Badge>
						{/each}
					</div>

					{@render valueField(secret)}

					<div class="flex flex-wrap items-center gap-2 text-fc-xs text-fc-fg-muted">
						<span class="whitespace-nowrap">v{secret.version}</span>
						<span class="whitespace-nowrap">rotated {formatDate(secret.rotated_at)}</span>
						{@render rotationBadge(secret)}
					</div>

					<div class="flex items-center justify-end gap-1">
						{@render secretActions(secret)}
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</section>

<Drawer bind:open={addOpen} title="Add secret to {environment.name}">
	<form onsubmit={handleAdd} class="flex flex-col gap-4">
		<Field label="Key" error={newKeyError}>
			<Input
				placeholder="DATABASE_URL"
				bind:value={newKey}
				required
				disabled={adding}
				class="font-fc-mono"
			/>
		</Field>
		<Field label="Value">
			<Textarea
				rows={3}
				placeholder="Secret value"
				bind:value={newValue}
				disabled={adding}
				class="font-fc-mono"
			/>
		</Field>
		<Field label="Tags" helper="Comma-separated, optional">
			<Input placeholder="backend, database" bind:value={newTags} disabled={adding} />
		</Field>
		<div class="flex flex-col-reverse gap-2 pt-1 sm:flex-row sm:justify-end">
			<Button
				type="button"
				variant="ghost"
				class="w-full sm:w-auto"
				disabled={adding}
				onclick={() => (addOpen = false)}
			>
				Cancel
			</Button>
			<Button
				type="submit"
				icon={icons.plus}
				class="w-full sm:w-auto"
				disabled={adding || !newKey.trim() || Boolean(newKeyError)}
			>
				{#if adding}<Spinner label="Loading" size="sm" />{/if}
				{adding ? 'Adding…' : 'Add secret'}
			</Button>
		</div>
	</form>
</Drawer>

<Drawer bind:open={trashOpen} title="Trash — {environment.name}">
	{#if trashLoading}
		<div class="flex flex-col gap-2">
			{#each Array(3) as _, i (i)}
				<Skeleton class="h-11 w-full" />
			{/each}
		</div>
	{:else if trashError}
		<ErrorState message={trashError} onretry={openTrash} />
	{:else if trashed.length === 0}
		<EmptyState
			bare
			icon={icons.remove}
			title="Nothing in the trash"
			description="Deleted secrets stay here for 30 days before they are destroyed."
		/>
	{:else}
		<div class="flex flex-col gap-3">
			<p class="text-fc-sm text-fc-fg-muted">
				A deleted key stays reserved until it is restored or purged, so it cannot be recreated
				from scratch while it sits here.
			</p>
			<ul class="flex flex-col divide-y divide-fc-border">
				{#each trashed as secret (secret.id)}
					<li class="flex items-center justify-between gap-3 py-3">
						<div class="flex min-w-0 flex-col gap-1">
							<span class="truncate font-fc-mono text-fc-sm text-fc-fg">{secret.key}</span>
							<span class="flex min-w-0 items-center gap-1 text-fc-xs text-fc-fg-muted">
								<PersonLabel person={secret.actor} />
								<span class="shrink-0">
									deleted {formatDateTime(secret.deleted_at)} · purged {formatDate(secret.purge_at)}
								</span>
							</span>
						</div>
						{#if canWrite}
							<Button
								variant="outline"
								icon={icons.refresh}
								class="shrink-0"
								disabled={restoring === secret.id}
								onclick={() => handleRestore(secret)}
							>
								{#if restoring === secret.id}<Spinner label="Loading" size="sm" />{/if}
								Restore
							</Button>
						{/if}
					</li>
				{/each}
			</ul>
		</div>
	{/if}
</Drawer>

<Drawer bind:open={editOpen} title="Edit {editTarget?.key ?? 'secret'}">
	<form onsubmit={handleEdit} class="flex flex-col gap-4">
		<Field label="Value" helper="Saving creates a new version">
			<Textarea
				rows={4}
				bind:value={editValue}
				disabled={saving || editLoading}
				class="font-fc-mono"
			/>
		</Field>
		<div class="flex flex-col-reverse gap-2 pt-1 sm:flex-row sm:justify-end">
			<Button
				type="button"
				variant="ghost"
				class="w-full sm:w-auto"
				disabled={saving}
				onclick={() => (editOpen = false)}
			>
				Cancel
			</Button>
			<Button
				type="submit"
				icon={icons.check}
				class="w-full sm:w-auto"
				disabled={saving || editLoading}
			>
				{#if saving || editLoading}<Spinner label="Loading" size="sm" />{/if}
				{editLoading ? 'Loading…' : saving ? 'Saving…' : 'Save'}
			</Button>
		</div>
	</form>
</Drawer>

<Drawer bind:open={historyOpen} title="History of {historyTarget?.key ?? 'secret'}">
	{#if versionsLoading}
		<div class="flex justify-center py-6">
			<Spinner label="Loading" size="md" />
		</div>
	{:else if versionsError}
		<Alert tone="danger">{versionsError}</Alert>
	{:else if versions.length === 0}
		<EmptyState bare icon={icons.history} title="No previous versions" />
	{:else}
		<ul class="flex flex-col divide-y divide-fc-border">
			{#each versions as version (version.version)}
				<li class="flex items-center justify-between gap-3 py-3">
					<div class="min-w-0">
						<p class="flex items-center gap-2 text-fc-sm text-fc-fg">
							v{version.version}
							{#if historyTarget && version.version === historyTarget.version}
								<Badge tone="accent">current</Badge>
							{/if}
						</p>
						<p class="flex min-w-0 items-center gap-1 text-fc-xs text-fc-fg-muted">
							<PersonLabel person={version.actor} />
							<span class="shrink-0">— {formatDateTime(version.created_at)}</span>
						</p>
					</div>
					{#if canWrite && historyTarget && version.version !== historyTarget.version}
						<Button
							variant="outline"
							icon={icons.refresh}
							class="shrink-0"
							disabled={rollingBack !== null}
							onclick={() => handleRollback(version.version)}
						>
							{#if rollingBack === version.version}<Spinner label="Loading" size="sm" />{/if}
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
				class="font-fc-mono"
			/>
		</Field>
		<div class="flex flex-col-reverse gap-2 pt-1 sm:flex-row sm:justify-end">
			<Button
				type="button"
				variant="ghost"
				class="w-full sm:w-auto"
				disabled={importing}
				onclick={() => (importOpen = false)}
			>
				Cancel
			</Button>
			<Button
				type="submit"
				icon={icons.upload}
				class="w-full sm:w-auto"
				disabled={importing || !importContent.trim()}
			>
				{#if importing}<Spinner label="Loading" size="sm" />{/if}
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
