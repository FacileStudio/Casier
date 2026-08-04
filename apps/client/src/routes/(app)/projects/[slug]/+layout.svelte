<script lang="ts">
	import { setContext } from 'svelte';
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		type Environment,
		type EnvironmentsResponse,
		type Project,
	} from '$lib/backend';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { Badge, Skeleton } from '@facile/lib';

	let { children } = $props();

	let slug = $derived(page.params.slug ?? '');

	let project = $state<Project | null>(null);
	let environments = $state<Environment[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	let canWrite = $derived(
		project !== null && ['owner', 'admin', 'member'].includes(project.role)
	);
	let canAdmin = $derived(project !== null && ['owner', 'admin'].includes(project.role));

	let navTabs = $derived([
		{
			href: `/projects/${slug}`,
			label: 'Secrets',
			icon: 'solar:key-linear',
			exact: true,
			visible: true,
		},
		{
			href: `/projects/${slug}/members`,
			label: 'Members',
			icon: 'solar:users-group-rounded-linear',
			exact: false,
			visible: true,
		},
		{
			href: `/projects/${slug}/tokens`,
			label: 'API tokens',
			icon: 'solar:code-square-linear',
			exact: false,
			visible: canAdmin,
		},
		{
			href: `/projects/${slug}/audit`,
			label: 'Audit log',
			icon: 'solar:history-linear',
			exact: false,
			visible: true,
		},
		{
			href: `/projects/${slug}/settings`,
			label: 'Settings',
			icon: 'solar:settings-linear',
			exact: false,
			visible: canAdmin,
		},
	].filter((tab) => tab.visible));

	function isActive(href: string, exact: boolean) {
		if (exact) return page.url.pathname === href;
		return page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);
	}

	async function reloadProject() {
		project = await api.get<Project>(`/projects/${slug}`, auth.token);
	}

	async function reloadEnvironments() {
		const res = await api.get<EnvironmentsResponse>(
			`/projects/${slug}/environments`,
			auth.token
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
	<div class="border-b border-border px-4 py-4 md:px-8 md:py-5">
		<Skeleton class="mb-2 h-3 w-24" />
		<Skeleton class="h-6 w-56 max-w-full" />
		<Skeleton class="mt-2 h-4 w-80 max-w-full" />
		<div class="mt-4 flex gap-1 overflow-hidden">
			{#each Array(4) as _, i (i)}
				<Skeleton class="h-9 w-28 shrink-0" />
			{/each}
		</div>
	</div>
{:else if loadError}
	<ErrorState message={loadError} onretry={load} />
{:else if project}
	<PageHeader
		title={project.name}
		description={project.description || undefined}
		back={{ href: '/projects', label: 'All projects' }}
	>
		{#snippet actions()}
			<Badge class="text-xs">{project?.role}</Badge>
		{/snippet}
		{#snippet tabs()}
			{#each navTabs as tab (tab.href)}
				{@const active = isActive(tab.href, tab.exact)}
				<a
					href={tab.href}
					aria-current={active ? 'page' : undefined}
					class="inline-flex h-9 shrink-0 items-center gap-2 whitespace-nowrap rounded-md px-4 text-sm font-medium transition-colors {active
						? 'bg-foreground text-background'
						: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
				>
					<iconify-icon icon={tab.icon} width="18" aria-hidden="true"></iconify-icon>
					{tab.label}
				</a>
			{/each}
		{/snippet}
	</PageHeader>

	{@render children?.()}
{/if}
