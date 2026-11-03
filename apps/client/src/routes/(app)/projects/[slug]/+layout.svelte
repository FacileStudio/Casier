<script lang="ts">
	import { setContext } from 'svelte';
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		type Environment,
		type EnvironmentsResponse,
		type Project,
		type ProjectRole,
	} from '$lib/backend';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { Badge, Skeleton, Tabs, icons } from '@facile/muse';

	let { children } = $props();

	const roleTones = {
		owner: 'owner',
		admin: 'admin',
		member: 'neutral',
		viewer: 'neutral',
	} as const satisfies Record<ProjectRole, string>;

	let slug = $derived(page.params.slug ?? '');

	let project = $state<Project | null>(null);
	let environments = $state<Environment[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	let canWrite = $derived(
		project !== null && ['owner', 'admin', 'member'].includes(project.role)
	);
	let canAdmin = $derived(project !== null && ['owner', 'admin'].includes(project.role));

	let navTabs = $derived(
		[
			{
				id: 'secrets',
				href: `/projects/${slug}`,
				label: 'Secrets',
				icon: icons.key,
				exact: true,
				visible: true,
			},
			{
				id: 'members',
				href: `/projects/${slug}/members`,
				label: 'Members',
				icon: icons.usersGroup,
				exact: false,
				visible: true,
			},
			{
				id: 'tokens',
				href: `/projects/${slug}/tokens`,
				label: 'API tokens',
				icon: icons.code,
				exact: false,
				visible: canAdmin,
			},
			{
				id: 'audit',
				href: `/projects/${slug}/audit`,
				label: 'Audit log',
				icon: icons.history,
				exact: false,
				visible: true,
			},
			{
				id: 'settings',
				href: `/projects/${slug}/settings`,
				label: 'Settings',
				icon: icons.settings,
				exact: false,
				visible: canAdmin,
			},
		].filter((tab) => tab.visible)
	);

	function isActive(href: string, exact: boolean) {
		if (exact) return page.url.pathname === href;
		return page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);
	}

	let activeTab = $derived(navTabs.find((tab) => isActive(tab.href, tab.exact))?.id ?? 'secrets');

	async function reloadProject() {
		project = await api.get<Project>(`/projects/${slug}`);
	}

	async function reloadEnvironments() {
		const res = await api.get<EnvironmentsResponse>(
			`/projects/${slug}/environments`,
		);
		environments = (res.environments ?? []).toSorted((a, b) => a.position - b.position);
	}

	async function load() {
		loading = true;
		loadError = '';
		try {
			await Promise.all([reloadProject(), reloadEnvironments()]);
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load project';
		} finally {
			loading = false;
		}
	}

	setContext('project', {
		get project() {
			return project;
		},
		get environments() {
			return environments;
		},
		get canWrite() {
			return canWrite;
		},
		get canAdmin() {
			return canAdmin;
		},
		get slug() {
			return slug;
		},
		reloadEnvironments,
		reloadProject,
	});

	$effect(() => {
		slug;
		load();
	});
</script>

{#if loading}
	<div class="flex flex-col gap-4 px-4 py-4 md:px-8 md:py-5">
		<Skeleton class="h-3 w-24" />
		<Skeleton class="h-6 w-56 max-w-full" />
		<Skeleton class="h-4 w-80 max-w-full" />
		<div class="flex gap-1 overflow-hidden">
			{#each Array(4) as _, i (i)}
				<Skeleton class="h-11 w-28 shrink-0 rounded-fc-pill" />
			{/each}
		</div>
	</div>
{:else if loadError}
	<div class="px-4 py-6 md:px-8">
		<ErrorState message={loadError} onretry={load} />
	</div>
{:else if project}
	<PageHeader
		title={project.name}
		description={project.description || undefined}
		back={{ href: '/projects', label: 'All projects' }}
	>
		{#snippet actions()}
			<Badge tone={project ? (roleTones[project.role] ?? 'neutral') : 'neutral'}>
				{project?.role}
			</Badge>
		{/snippet}
		{#snippet tabs()}
			<Tabs items={navTabs} value={activeTab} label="Project sections" />
		{/snippet}
	</PageHeader>

	{@render children?.()}
{/if}
