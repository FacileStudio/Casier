<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		type Environment,
		type EnvironmentsResponse,
		type Space,
	} from '$lib/backend';
	import { nav } from '$lib/nav.svelte';
	import { Badge, Button, Field, Input, Modal, Skeleton, Spinner } from '@facile/lib';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import SecretsSection from '$lib/components/space/SecretsSection.svelte';
	import MembersSection from '$lib/components/space/MembersSection.svelte';
	import TokensSection from '$lib/components/space/TokensSection.svelte';
	import AuditSection from '$lib/components/space/AuditSection.svelte';
	import { toast } from 'svelte-sonner';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';

	let slug = $derived(page.params.slug ?? '');

	let space = $state<Space | null>(null);
	let environments = $state<Environment[]>([]);
	let selectedEnvSlug = $state<string | null>(null);
	let loading = $state(true);
	let loadError = $state('');

	let tab = $state<'secrets' | 'members' | 'tokens' | 'audit'>('secrets');

	let createEnvOpen = $state(false);
	let newEnvName = $state('');
	let creatingEnv = $state(false);

	let deleteEnvOpen = $state(false);
	let deletingEnv = $state(false);

	let deleteSpaceOpen = $state(false);
	let deletingSpace = $state(false);

	let canWrite = $derived(
		space !== null && ['owner', 'admin', 'member'].includes(space.role)
	);
	let canAdmin = $derived(space !== null && ['owner', 'admin'].includes(space.role));
	let selectedEnv = $derived(environments.find((e) => e.slug === selectedEnvSlug) ?? null);

	const roleTones = {
		owner: 'accent',
		admin: 'success',
		member: 'neutral',
		viewer: 'neutral',
	} as const;

	async function loadSpace() {
		loading = true;
		loadError = '';
		try {
			const [spaceRes, envsRes] = await Promise.all([
				api.get<Space>(`/spaces/${slug}`, auth.token),
				api.get<EnvironmentsResponse>(`/spaces/${slug}/environments`, auth.token),
			]);
			space = spaceRes;
			nav.spaceName = spaceRes.name;
			environments = (envsRes.environments ?? []).toSorted((a, b) => a.position - b.position);
			if (!environments.some((e) => e.slug === selectedEnvSlug)) {
				selectedEnvSlug = environments[0]?.slug ?? null;
			}
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load space';
		} finally {
			loading = false;
		}
	}

	async function reloadEnvironments() {
		const envsRes = await api.get<EnvironmentsResponse>(`/spaces/${slug}/environments`, auth.token);
		environments = (envsRes.environments ?? []).toSorted((a, b) => a.position - b.position);
		if (!environments.some((e) => e.slug === selectedEnvSlug)) {
			selectedEnvSlug = environments[0]?.slug ?? null;
		}
	}

	async function handleCreateEnv(e: Event) {
		e.preventDefault();
		if (!newEnvName.trim()) return;
		creatingEnv = true;
		try {
			const env = await api.post<Environment>(
				`/spaces/${slug}/environments`,
				{ name: newEnvName.trim() },
				auth.token
			);
			createEnvOpen = false;
			newEnvName = '';
			await reloadEnvironments();
			selectedEnvSlug = env.slug;
			toast.success(`Environment "${env.name}" created`);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to create environment');
		} finally {
			creatingEnv = false;
		}
	}

	async function handleDeleteEnv() {
		if (!selectedEnv) return;
		deletingEnv = true;
		try {
			await api.del(`/spaces/${slug}/environments/${selectedEnv.slug}`, auth.token);
			toast.success(`Environment "${selectedEnv.name}" deleted`);
			deleteEnvOpen = false;
			await reloadEnvironments();
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to delete environment');
		} finally {
			deletingEnv = false;
		}
	}

	async function handleDeleteSpace() {
		if (!space) return;
		deletingSpace = true;
		try {
			await api.del(`/spaces/${slug}`, auth.token);
			toast.success(`Space "${space.name}" deleted`);
			goto('/spaces');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to delete space');
			deletingSpace = false;
		}
	}

	$effect(() => {
		slug;
		loadSpace();
		return () => {
			nav.spaceName = null;
		};
	});
</script>

<svelte:head>
	<title>{space ? `Casier — ${space.name}` : 'Casier'}</title>
</svelte:head>

{#if loading}
	<div class="flex flex-col gap-6">
		<Skeleton class="h-4 w-24" />
		<div class="flex flex-col gap-2">
			<Skeleton class="h-8 w-64" />
			<Skeleton class="h-4 w-96 max-w-full" />
		</div>
		<Skeleton class="h-10 w-72 max-w-full rounded-fc-pill" />
		<div class="rounded-fc-md bg-fc-component px-8 py-6">
			<div class="flex flex-col gap-3">
				{#each Array(5) as _, i (i)}
					<Skeleton class="h-9 w-full" />
				{/each}
			</div>
		</div>
	</div>
{:else if loadError}
	<ErrorState message={loadError} onretry={loadSpace} />
{:else if space}
	<div class="flex flex-col gap-6">
		<div>
			<a
				href="/spaces"
				class="mb-3 inline-flex items-center gap-1.5 rounded-fc-sm text-fc-sm text-fc-fg-muted transition-colors hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-fc-accent"
			>
				<ArrowLeftIcon class="size-3.5" aria-hidden="true" />
				All spaces
			</a>
			<div class="flex flex-wrap items-start justify-between gap-3">
				<div class="min-w-0">
					<div class="flex items-center gap-3">
						<h1 class="truncate text-fc-2xl font-semibold text-fc-fg">{space.name}</h1>
						<Badge tone={roleTones[space.role] ?? 'neutral'}>{space.role}</Badge>
					</div>
					{#if space.description}
						<p class="mt-1 text-fc-md text-fc-fg-muted">{space.description}</p>
					{/if}
				</div>
				{#if canAdmin}
					<Button variant="ghost" class="text-fc-danger" onclick={() => (deleteSpaceOpen = true)}>
						<Trash2Icon class="size-4" aria-hidden="true" />
						Delete space
					</Button>
				{/if}
			</div>
		</div>

		<div class="flex flex-wrap items-center gap-2">
			<div
				class="inline-flex max-w-full items-center gap-1 overflow-x-auto rounded-fc-pill bg-fc-surface p-1"
				role="group"
				aria-label="Environments"
			>
				{#each environments as env (env.id)}
					<button
						type="button"
						class="shrink-0 rounded-fc-pill px-3 py-1.5 text-fc-sm font-medium transition-colors {selectedEnvSlug ===
						env.slug
							? 'bg-fc-accent text-fc-accent-fg'
							: 'text-fc-fg-muted hover:text-fc-fg'}"
						aria-pressed={selectedEnvSlug === env.slug}
						onclick={() => (selectedEnvSlug = env.slug)}
					>
						{env.name}
					</button>
				{/each}
			</div>
			{#if canAdmin}
				<button
					type="button"
					class="inline-flex h-9 w-9 items-center justify-center rounded-fc-pill bg-fc-surface text-fc-fg-muted transition-colors hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-fc-accent"
					aria-label="Add environment"
					onclick={() => (createEnvOpen = true)}
				>
					<PlusIcon class="size-4" aria-hidden="true" />
				</button>
				{#if selectedEnv && environments.length > 1}
					<button
						type="button"
						class="inline-flex h-9 w-9 items-center justify-center rounded-fc-pill bg-fc-surface text-fc-fg-muted transition-colors hover:text-fc-danger focus-visible:outline-2 focus-visible:outline-fc-accent"
						aria-label={`Delete environment ${selectedEnv.name}`}
						onclick={() => (deleteEnvOpen = true)}
					>
						<Trash2Icon class="size-4" aria-hidden="true" />
					</button>
				{/if}
			{/if}
		</div>

		<div class="flex gap-1 border-b border-fc-fg/10" role="tablist" aria-label="Space sections">
			{#each [['secrets', 'Secrets'], ['members', 'Members'], ...(canAdmin ? [['tokens', 'API tokens']] : []), ['audit', 'Audit log']] as [id, label] (id)}
				<button
					type="button"
					role="tab"
					aria-selected={tab === id}
					class="-mb-px border-b-2 px-3 py-2 text-fc-md font-medium transition-colors {tab === id
						? 'border-fc-accent text-fc-fg'
						: 'border-transparent text-fc-fg-muted hover:text-fc-fg'}"
					onclick={() => (tab = id as typeof tab)}
				>
					{label}
				</button>
			{/each}
		</div>

		{#if tab === 'secrets'}
			{#if selectedEnv}
				<SecretsSection spaceSlug={slug} environment={selectedEnv} {canWrite} />
			{:else}
				<p class="py-8 text-center text-fc-md text-fc-fg-muted">
					No environment yet. Create one to start adding secrets.
				</p>
			{/if}
		{:else if tab === 'members'}
			<MembersSection spaceSlug={slug} {canAdmin} />
		{:else if tab === 'tokens' && canAdmin}
			<TokensSection spaceSlug={slug} />
		{:else if tab === 'audit'}
			<AuditSection spaceSlug={slug} />
		{/if}
	</div>

	<Modal bind:open={createEnvOpen} title="Create environment">
		<form onsubmit={handleCreateEnv} class="flex flex-col gap-4">
			<Field label="Name" helper="e.g. Preview, QA">
				<Input placeholder="Environment name" bind:value={newEnvName} required disabled={creatingEnv} />
			</Field>
			<div class="flex justify-end gap-2 pt-1">
				<Button type="button" variant="ghost" disabled={creatingEnv} onclick={() => (createEnvOpen = false)}>
					Cancel
				</Button>
				<Button type="submit" disabled={creatingEnv || !newEnvName.trim()}>
					{#if creatingEnv}<Spinner size="sm" />{/if}
					{creatingEnv ? 'Creating…' : 'Create'}
				</Button>
			</div>
		</form>
	</Modal>

	<ConfirmDialog
		bind:open={deleteEnvOpen}
		title="Delete environment"
		message={`This permanently deletes the "${selectedEnv?.name}" environment and all of its secrets in ${space.name}. This cannot be undone.`}
		confirmLabel="Delete environment"
		loading={deletingEnv}
		onconfirm={handleDeleteEnv}
	/>

	<ConfirmDialog
		bind:open={deleteSpaceOpen}
		title="Delete space"
		message={`This permanently deletes the space "${space.name}" with all of its environments, secrets, members and tokens. This cannot be undone.`}
		confirmLabel="Delete space"
		loading={deletingSpace}
		onconfirm={handleDeleteSpace}
	/>
{/if}
