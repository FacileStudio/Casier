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
	} from '$lib/backend';
	import { formatDate, formatDateTime } from '$lib/format';
	import { Badge, Button, Field, Input, Modal, Skeleton, Spinner, Table, Textarea } from '@facile/lib';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { toast } from 'svelte-sonner';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import HistoryIcon from '@lucide/svelte/icons/history';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import SearchIcon from '@lucide/svelte/icons/search';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import UploadIcon from '@lucide/svelte/icons/upload';

	let {
		spaceSlug,
		environment,
		canWrite,
	}: { spaceSlug: string; environment: Environment; canWrite: boolean } = $props();

	const MASK = '••••••••';
	const actionButton =
		'inline-flex h-8 w-8 items-center justify-center rounded-fc-sm text-fc-fg-muted transition-colors hover:bg-fc-surface hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-fc-accent';

	let secrets = $state<Secret[]>([]);
	let loading = $state(true);
	let loadError = $state('');
	let search = $state('');
	let revealed = $state<Set<number>>(new Set());

	let addOpen = $state(false);
	let newKey = $state('');
	let newValue = $state('');
	let newTags = $state('');
	let adding = $state(false);

	let editTarget = $state<Secret | null>(null);
	let editOpen = $state(false);
	let editValue = $state('');
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

	let importOpen = $state(false);
	let importContent = $state('');
	let importing = $state(false);
	let exporting = $state(false);

	let basePath = $derived(
		`/spaces/${spaceSlug}/environments/${environment.slug}/secrets`
	);

	let filtered = $derived(
		secrets
			.filter((s) => {
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
		if (/[\s=]/.test(key)) return 'Keys cannot contain spaces or "="';
		if (secrets.some((s) => s.key === key)) return 'This key already exists in this environment';
		return undefined;
	});

	async function loadSecrets() {
		loading = true;
		loadError = '';
		revealed = new Set();
		try {
			const res = await api.get<SecretsResponse>(basePath, auth.token);
			secrets = res.secrets ?? [];
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load secrets';
		} finally {
			loading = false;
		}
	}

	function toggleReveal(id: number) {
		const next = new Set(revealed);
		if (next.has(id)) {
			next.delete(id);
		} else {
			next.add(id);
		}
		revealed = next;
	}

	async function copyValue(secret: Secret) {
		try {
			await navigator.clipboard.writeText(secret.value);
			toast.success(`Copied ${secret.key} to clipboard`);
		} catch {
			toast.error('Clipboard unavailable');
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

	function openEdit(secret: Secret) {
		editTarget = secret;
		editValue = secret.value;
		editOpen = true;
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
			link.download = `${spaceSlug}.${environment.slug}.env`;
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

<section class="flex flex-col gap-4" aria-label="Secrets">
	<div class="flex flex-wrap items-center gap-2">
		<div class="relative min-w-0 flex-1 sm:max-w-xs">
			<SearchIcon
				class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-fc-fg-muted"
				aria-hidden="true"
			/>
			<Input
				type="search"
				placeholder="Filter by key or tag"
				bind:value={search}
				class="pl-9"
				aria-label="Filter secrets"
			/>
		</div>
		<div class="ml-auto flex items-center gap-2">
			<Button variant="outline" disabled={exporting || secrets.length === 0} onclick={handleExport}>
				<DownloadIcon class="size-4" aria-hidden="true" />
				Export
			</Button>
			{#if canWrite}
				<Button variant="outline" onclick={() => (importOpen = true)}>
					<UploadIcon class="size-4" aria-hidden="true" />
					Import
				</Button>
				<Button onclick={openAdd}>
					<PlusIcon class="size-4" aria-hidden="true" />
					Add secret
				</Button>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="flex flex-col gap-2 rounded-fc-md bg-fc-component px-6 py-5">
			{#each Array(5) as _, i (i)}
				<Skeleton class="h-9 w-full" />
			{/each}
		</div>
	{:else if loadError}
		<ErrorState message={loadError} onretry={loadSecrets} />
	{:else if secrets.length === 0}
		<EmptyState
			title="No secrets in {environment.name}"
			description="Secrets are encrypted at rest and only visible to members of this space."
		>
			{#snippet icon()}
				<KeyRoundIcon class="size-10" aria-hidden="true" />
			{/snippet}
			{#if canWrite}
				<Button onclick={openAdd}>
					<PlusIcon class="size-4" aria-hidden="true" />
					Add your first secret
				</Button>
			{/if}
		</EmptyState>
	{:else if filtered.length === 0}
		<EmptyState title="No matches" description={`No secret matches "${search}".`} />
	{:else}
		<Table>
			<thead>
				<tr>
					<th>Key</th>
					<th>Value</th>
					<th>Version</th>
					<th>Updated</th>
					<th class="w-1 text-right"><span class="sr-only">Actions</span></th>
				</tr>
			</thead>
			<tbody>
				{#each filtered as secret (secret.id)}
					<tr>
						<td>
							<div class="flex flex-wrap items-center gap-1.5">
								<code class="font-mono text-fc-sm text-fc-fg">{secret.key}</code>
								{#each secret.tags ?? [] as tag (tag)}
									<Badge tone="neutral">{tag}</Badge>
								{/each}
							</div>
						</td>
						<td>
							{#if revealed.has(secret.id)}
								<code class="block max-w-64 truncate font-mono text-fc-sm text-fc-fg" title={secret.value}>
									{secret.value}
								</code>
							{:else}
								<span class="select-none text-fc-fg-muted" aria-label="Hidden value">{MASK}</span>
							{/if}
						</td>
						<td class="whitespace-nowrap text-fc-fg-muted">v{secret.version}</td>
						<td class="whitespace-nowrap text-fc-fg-muted">{formatDate(secret.updated_at)}</td>
						<td>
							<div class="flex items-center justify-end gap-0.5">
								<button
									type="button"
									class={actionButton}
									aria-label={revealed.has(secret.id) ? `Hide ${secret.key}` : `Reveal ${secret.key}`}
									onclick={() => toggleReveal(secret.id)}
								>
									{#if revealed.has(secret.id)}
										<EyeOffIcon class="size-4" aria-hidden="true" />
									{:else}
										<EyeIcon class="size-4" aria-hidden="true" />
									{/if}
								</button>
								<button
									type="button"
									class={actionButton}
									aria-label={`Copy ${secret.key}`}
									onclick={() => copyValue(secret)}
								>
									<CopyIcon class="size-4" aria-hidden="true" />
								</button>
								<button
									type="button"
									class={actionButton}
									aria-label={`History of ${secret.key}`}
									onclick={() => openHistory(secret)}
								>
									<HistoryIcon class="size-4" aria-hidden="true" />
								</button>
								{#if canWrite}
									<button
										type="button"
										class={actionButton}
										aria-label={`Edit ${secret.key}`}
										onclick={() => openEdit(secret)}
									>
										<PencilIcon class="size-4" aria-hidden="true" />
									</button>
									<button
										type="button"
										class="{actionButton} hover:text-fc-danger"
										aria-label={`Delete ${secret.key}`}
										onclick={() => openDelete(secret)}
									>
										<Trash2Icon class="size-4" aria-hidden="true" />
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

<Modal bind:open={addOpen} title="Add secret to {environment.name}">
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
			<Button type="button" variant="ghost" disabled={adding} onclick={() => (addOpen = false)}>
				Cancel
			</Button>
			<Button type="submit" disabled={adding || !newKey.trim() || !!newKeyError}>
				{#if adding}<Spinner size="sm" />{/if}
				{adding ? 'Adding…' : 'Add secret'}
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:open={editOpen} title="Edit {editTarget?.key}">
	<form onsubmit={handleEdit} class="flex flex-col gap-4">
		<Field label="Value" helper="Saving creates a new version">
			<Textarea rows={4} bind:value={editValue} disabled={saving} class="font-mono" />
		</Field>
		<div class="flex justify-end gap-2 pt-1">
			<Button type="button" variant="ghost" disabled={saving} onclick={() => (editOpen = false)}>
				Cancel
			</Button>
			<Button type="submit" disabled={saving}>
				{#if saving}<Spinner size="sm" />{/if}
				{saving ? 'Saving…' : 'Save'}
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:open={historyOpen} title="History of {historyTarget?.key}">
	{#if versionsLoading}
		<div class="flex justify-center py-6">
			<Spinner size="md" />
		</div>
	{:else if versionsError}
		<p class="py-4 text-center text-fc-md text-fc-danger">{versionsError}</p>
	{:else if versions.length === 0}
		<p class="py-4 text-center text-fc-md text-fc-fg-muted">No previous versions.</p>
	{:else}
		<ul class="flex max-h-72 flex-col gap-1 overflow-y-auto">
			{#each versions as version (version.version)}
				<li class="flex items-center justify-between gap-3 rounded-fc-sm px-2 py-2 hover:bg-fc-surface">
					<div class="min-w-0">
						<p class="text-fc-md text-fc-fg">
							v{version.version}
							{#if historyTarget && version.version === historyTarget.version}
								<Badge tone="accent">current</Badge>
							{/if}
						</p>
						<p class="truncate text-fc-xs text-fc-fg-muted">
							{version.changed_by || 'unknown'} — {formatDateTime(version.created_at)}
						</p>
					</div>
					{#if canWrite && historyTarget && version.version !== historyTarget.version}
						<Button
							variant="outline"
							size="sm"
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
</Modal>

<Modal bind:open={importOpen} title="Import into {environment.name}">
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
			<Button type="button" variant="ghost" disabled={importing} onclick={() => (importOpen = false)}>
				Cancel
			</Button>
			<Button type="submit" disabled={importing || !importContent.trim()}>
				{#if importing}<Spinner size="sm" />{/if}
				{importing ? 'Importing…' : 'Import'}
			</Button>
		</div>
	</form>
</Modal>

<ConfirmDialog
	bind:open={deleteOpen}
	title="Delete secret"
	message={`This permanently deletes ${deleteTarget?.key} from the ${environment.name} environment. This cannot be undone.`}
	confirmLabel="Delete secret"
	loading={deleting}
	onconfirm={handleDelete}
/>
