<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, type Project, type ProjectsResponse } from '$lib/backend';
	import { formatDate } from '$lib/format';
	import { Badge, Field, Input, Modal, Skeleton, Textarea } from '@facile/lib';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import { toast } from 'svelte-sonner';

	const app = getContext<{ projects: Project[]; refreshProjects: () => Promise<void> }>('app');

	let projects = $state<Project[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	let createOpen = $state(false);
	let createName = $state('');
	let createDescription = $state('');
	let creating = $state(false);

	let nameError = $derived(
		createName.length > 0 && createName.trim().length < 2
			? 'Name must be at least 2 characters'
			: undefined
	);

	async function loadProjects() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<ProjectsResponse>('/projects', auth.token);
			projects = (res.projects ?? []).toSorted((a, b) => a.name.localeCompare(b.name));
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load projects';
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		createName = '';
		createDescription = '';
		createOpen = true;
	}

	async function handleCreate(event: Event) {
		event.preventDefault();
		if (createName.trim().length < 2) return;
		creating = true;
		try {
			const project = await api.post<Project>(
				'/projects',
				{ name: createName.trim(), description: createDescription.trim() },
				auth.token
			);
			createOpen = false;
			toast.success(`Project "${project.name}" created`);
			await app.refreshProjects();
			goto(`/projects/${project.slug}`);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to create project');
		} finally {
			creating = false;
		}
	}

	$effect(() => {
		loadProjects();
	});
</script>

<svelte:head>
	<title>Projects — Casier</title>
</svelte:head>

{#snippet newProjectButton()}
	<button
		type="button"
		onclick={openCreate}
		class="inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50"
	>
		<iconify-icon icon="solar:add-circle-linear" width="16" aria-hidden="true"></iconify-icon>
		New project
	</button>
{/snippet}

<PageHeader
	title="Projects"
	description="Every project has its own encryption key and its own environments."
	actions={newProjectButton}
/>

<div class="px-4 py-6 md:px-8">
	{#if loading}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array(6) as _, index (index)}
				<Skeleton class="h-32 rounded-lg" />
			{/each}
		</div>
	{:else if loadError}
		<ErrorState message={loadError} onretry={loadProjects} />
	{:else if projects.length === 0}
		<EmptyState
			title="No projects yet"
			description="A project groups the secrets for one app or service, across all of its environments."
		>
			{#snippet icon()}
				<iconify-icon icon="solar:widget-2-linear" width="40" aria-hidden="true"></iconify-icon>
			{/snippet}
			{@render newProjectButton()}
		</EmptyState>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each projects as project (project.id)}
				<a
					href="/projects/{project.slug}"
					class="group rounded-lg border border-border bg-card p-5 transition-colors hover:bg-muted"
				>
					<div class="flex items-start justify-between gap-3">
						<span class="truncate text-base font-semibold">{project.name}</span>
						<Badge class="text-xs">{project.role}</Badge>
					</div>
					{#if project.description}
						<p class="mt-1.5 line-clamp-2 text-sm text-muted-foreground">{project.description}</p>
					{:else}
						<p class="mt-1.5 line-clamp-2 text-sm text-muted-foreground">No description</p>
					{/if}
					<div class="mt-4 flex items-center gap-2 text-xs text-muted-foreground">
						<span class="truncate">Updated {formatDate(project.updated_at)}</span>
						<iconify-icon
							icon="solar:alt-arrow-right-linear"
							width="14"
							class="ml-auto shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
							aria-hidden="true"
						></iconify-icon>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>

<Modal bind:open={createOpen} title="New project">
	<form onsubmit={handleCreate} class="flex flex-col gap-4">
		<Field label="Name" error={nameError}>
			<Input placeholder="my-project" bind:value={createName} required disabled={creating} />
		</Field>
		<Field label="Description" helper="Optional">
			<Textarea
				rows={2}
				placeholder="What lives in this project?"
				bind:value={createDescription}
				disabled={creating}
			/>
		</Field>
		<div class="flex justify-end gap-2 pt-1">
			<button
				type="button"
				onclick={() => (createOpen = false)}
				disabled={creating}
				class="inline-flex h-9 items-center justify-center gap-2 rounded-md px-4 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-50"
			>
				Cancel
			</button>
			<button
				type="submit"
				disabled={creating || createName.trim().length < 2}
				class="inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50"
			>
				{creating ? 'Creating…' : 'Create'}
			</button>
		</div>
	</form>
</Modal>
