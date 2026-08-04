<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, type Space, type SpacesResponse } from '$lib/backend';
	import { formatDate } from '$lib/format';
	import { Badge, Button, Field, Input, Modal, Skeleton, Spinner, Textarea } from '@facile/lib';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { toast } from 'svelte-sonner';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import FolderLockIcon from '@lucide/svelte/icons/folder-lock';

	let spaces = $state<Space[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	let createOpen = $state(false);
	let createName = $state('');
	let createDescription = $state('');
	let creating = $state(false);

	let nameError = $derived(
		createName.length > 0 && createName.trim().length < 2 ? 'Name must be at least 2 characters' : undefined
	);

	const roleTones = {
		owner: 'accent',
		admin: 'success',
		member: 'neutral',
		viewer: 'neutral',
	} as const;

	async function loadSpaces() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<SpacesResponse>('/spaces', auth.token);
			spaces = (res.spaces ?? []).toSorted((a, b) => a.name.localeCompare(b.name));
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load spaces';
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		createName = '';
		createDescription = '';
		createOpen = true;
	}

	async function handleCreate(e: Event) {
		e.preventDefault();
		if (createName.trim().length < 2) return;
		creating = true;
		try {
			const space = await api.post<Space>(
				'/spaces',
				{ name: createName.trim(), description: createDescription.trim() },
				auth.token
			);
			createOpen = false;
			toast.success(`Space "${space.name}" created`);
			goto(`/spaces/${space.slug}`);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to create space');
		} finally {
			creating = false;
		}
	}

	$effect(() => {
		loadSpaces();
	});
</script>

<svelte:head>
	<title>Casier — Spaces</title>
</svelte:head>

<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
	<div>
		<h1 class="text-fc-2xl font-semibold text-fc-fg">Spaces</h1>
		<p class="text-fc-md text-fc-fg-muted">Each space groups environments and their secrets</p>
	</div>
	<Button onclick={openCreate}>
		<PlusIcon class="size-4" aria-hidden="true" />
		New space
	</Button>
</div>

{#if loading}
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		{#each Array(6) as _, i (i)}
			<div class="flex flex-col gap-3 rounded-fc-md bg-fc-component px-8 py-6">
				<Skeleton class="h-5 w-2/5" />
				<Skeleton class="h-4 w-4/5" />
				<Skeleton class="h-3 w-1/3" />
			</div>
		{/each}
	</div>
{:else if loadError}
	<ErrorState message={loadError} onretry={loadSpaces} />
{:else if spaces.length === 0}
	<EmptyState
		title="No spaces yet"
		description="A space holds your environments and secrets for one project or team. Create your first one to get started."
	>
		{#snippet icon()}
			<FolderLockIcon class="size-10" aria-hidden="true" />
		{/snippet}
		<Button onclick={openCreate}>
			<PlusIcon class="size-4" aria-hidden="true" />
			Create a space
		</Button>
	</EmptyState>
{:else}
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		{#each spaces as space (space.id)}
			<a
				href="/spaces/{space.slug}"
				class="group flex flex-col gap-2 rounded-fc-md border border-transparent bg-fc-component px-8 py-6 transition-colors hover:border-fc-accent/40 focus-visible:outline-2 focus-visible:outline-fc-accent"
			>
				<div class="flex items-start justify-between gap-2">
					<span class="truncate text-fc-lg font-semibold text-fc-fg group-hover:text-fc-accent">
						{space.name}
					</span>
					<Badge tone={roleTones[space.role] ?? 'neutral'}>{space.role}</Badge>
				</div>
				{#if space.description}
					<p class="line-clamp-2 text-fc-md text-fc-fg-muted">{space.description}</p>
				{/if}
				<p class="mt-auto pt-2 text-fc-xs text-fc-fg-muted">
					Updated {formatDate(space.updated_at)}
				</p>
			</a>
		{/each}
	</div>
{/if}

<Modal bind:open={createOpen} title="Create space">
	<form onsubmit={handleCreate} class="flex flex-col gap-4">
		<Field label="Name" error={nameError}>
			<Input placeholder="my-project" bind:value={createName} required disabled={creating} />
		</Field>
		<Field label="Description" helper="Optional">
			<Textarea rows={2} placeholder="What lives in this space?" bind:value={createDescription} disabled={creating} />
		</Field>
		<div class="flex justify-end gap-2 pt-1">
			<Button type="button" variant="ghost" disabled={creating} onclick={() => (createOpen = false)}>
				Cancel
			</Button>
			<Button type="submit" disabled={creating || createName.trim().length < 2}>
				{#if creating}<Spinner size="sm" />{/if}
				{creating ? 'Creating…' : 'Create space'}
			</Button>
		</div>
	</form>
</Modal>
