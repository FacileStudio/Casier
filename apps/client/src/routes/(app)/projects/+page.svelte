<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, type Project, type ProjectRole, type ProjectsResponse } from '$lib/backend';
	import { formatDate } from '$lib/format';
	import { t } from '$lib/i18n/index.svelte';
	import {
		Badge,
		Button,
		Card,
		EmptyState,
		Field,
		Input,
		Skeleton,
		Spinner,
		Textarea,
		icons,
		toast,
	} from '@facile/muse';
	import Drawer from '$lib/components/Drawer.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';

	const app = getContext<{ projects: Project[]; refreshProjects: () => Promise<void> }>('app');

	const roleTones = {
		owner: 'owner',
		admin: 'admin',
		member: 'neutral',
		viewer: 'neutral',
	} as const satisfies Record<ProjectRole, string>;

	let projects = $state<Project[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	let createOpen = $state(false);
	let createName = $state('');
	let createDescription = $state('');
	let creating = $state(false);

	let nameError = $derived(
		createName.length > 0 && createName.trim().length < 2 ? t('projects.nameTooShort') : undefined
	);

	async function loadProjects() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<ProjectsResponse>('/projects');
			projects = (res.projects ?? []).toSorted((a, b) => a.name.localeCompare(b.name));
		} catch (err) {
			loadError = err instanceof Error ? err.message : t('projects.error.load');
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
			);
			createOpen = false;
			toast.success(t('projects.created', { name: project.name }));
			await app.refreshProjects();
			goto(`/projects/${project.slug}`);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('projects.error.create'));
		} finally {
			creating = false;
		}
	}

	$effect(() => {
		loadProjects();
	});
</script>

<svelte:head>
	<title>{t('projects.head')}</title>
</svelte:head>

{#snippet newProjectButton()}
	<Button icon={icons.plus} onclick={openCreate}>{t('projects.new')}</Button>
{/snippet}

<PageHeader
	title={t('nav.projects')}
	description={t('projects.description')}
	actions={newProjectButton}
/>

<div class="px-4 py-6 md:px-8">
	{#if loading}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array(6) as _, index (index)}
				<Skeleton class="h-40 rounded-fc-md" />
			{/each}
		</div>
	{:else if loadError}
		<ErrorState message={loadError} onretry={loadProjects} />
	{:else if projects.length === 0}
		<EmptyState
			icon={icons.folder}
			title={t('projects.empty.title')}
			description={t('projects.empty.description')}
		>
			{@render newProjectButton()}
		</EmptyState>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each projects as project (project.id)}
				<Card href="/projects/{project.slug}" class="flex flex-col gap-4">
					<div class="flex items-start justify-between gap-3">
						<div class="flex min-w-0 items-center gap-3">
							<span
								class="flex size-9 shrink-0 items-center justify-center rounded-fc-sm bg-fc-surface text-fc-fg transition-colors group-hover:bg-fc-component"
							>
								<iconify-icon icon={icons.folder} width="18" height="18" class="block"
								></iconify-icon>
							</span>
							<span class="truncate text-fc-sm font-medium text-fc-fg">{project.name}</span>
						</div>
						<Badge tone={roleTones[project.role] ?? 'neutral'}>{t(`role.${project.role}`)}</Badge>
					</div>
					<p class="line-clamp-2 text-fc-sm text-fc-fg-muted">
						{project.description || t('projects.noDescription')}
					</p>
					<div class="mt-auto flex items-center gap-2 text-fc-xs text-fc-fg-muted">
						<span class="truncate">{t('projects.updated', { date: formatDate(project.updated_at) })}</span>
						<iconify-icon
							icon={icons.arrow}
							width="16"
							height="16"
							class="ml-auto block shrink-0 transition-transform group-hover:translate-x-0.5"
						></iconify-icon>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>

<Drawer bind:open={createOpen} title={t('projects.new')}>
	<form onsubmit={handleCreate} class="flex flex-col gap-4">
		<Field label={t('common.name')} error={nameError}>
			<Input placeholder="my-project" bind:value={createName} required disabled={creating} />
		</Field>
		<Field label={t('projects.description.label')} helper={t('common.optional')}>
			<Textarea
				rows={2}
				placeholder={t('projects.description.placeholder')}
				bind:value={createDescription}
				disabled={creating}
			/>
		</Field>
		<div class="flex flex-col-reverse gap-2 pt-1 sm:flex-row sm:justify-end">
			<Button
				type="button"
				variant="ghost"
				class="w-full sm:w-auto"
				disabled={creating}
				onclick={() => (createOpen = false)}
			>
				{t('common.cancel')}
			</Button>
			<Button
				type="submit"
				icon={icons.plus}
				class="w-full sm:w-auto"
				disabled={creating || createName.trim().length < 2 || Boolean(nameError)}
			>
				{#if creating}<Spinner label={t('common.loading')} size="sm" />{/if}
				{creating ? t('common.creating') : t('common.create')}
			</Button>
		</div>
	</form>
</Drawer>
