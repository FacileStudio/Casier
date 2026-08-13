<script lang="ts">
	import { setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { MobileNav, SideBar, SpaceSwitcher, Topbar, icons } from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		ApiError,
		type Me,
		type Project,
		type ProjectsResponse,
	} from '$lib/backend';

	let { children } = $props();

	let projects = $state<Project[]>([]);
	let collapsed = $state(false);

	async function refreshProjects() {
		try {
			const result = await api.get<ProjectsResponse>('/projects');
			projects = result.projects ?? [];
		} catch (err) {
			projects = [];
			if (err instanceof ApiError && err.status === 401) {
				auth.logout();
				goto('/login');
			}
		}
	}

	setContext('app', {
		get projects() {
			return projects;
		},
		refreshProjects,
	});

	$effect(() => {
		if (auth.isAuthenticated) {
			refreshProjects();
			return;
		}
		auth.check().then(() => {
			if (auth.isAuthenticated) refreshProjects();
			else goto('/login');
		});
	});

	const navPages = $derived([
		{
			label: 'Projects',
			href: '/projects',
			icon: icons.folder,
			active:
				page.url.pathname === '/projects' || page.url.pathname.startsWith('/projects/'),
		},
	]);

	const onSettings = $derived(
		page.url.pathname === '/settings' || page.url.pathname.startsWith('/settings/'),
	);

	/**
	 * The API has no display name — an account is only ever an email. The user card
	 * wants a name, not an address, so the local part stands in for one: a rail is
	 * 220px wide and a full address either truncates to nonsense or blows the layout.
	 */
	const user = $derived({ name: auth.email?.split('@')[0] || 'Account' });

	const spaces = $derived(projects.map((project) => ({ id: project.slug, name: project.name })));

	const activeSpaceId = $derived(page.params.slug ?? null);

	function selectProject(slug: string | null) {
		goto(slug ? `/projects/${slug}` : '/projects');
	}
</script>

{#if auth.isAuthenticated}
	<div class="flex h-[100dvh] w-full overflow-hidden bg-fc-page">
		<div class="hidden h-full shrink-0 p-3 md:block">
			<SideBar
				class="h-full"
				icon="solar:key-minimalistic-square-bold-duotone"
				title="Casier"
				bind:collapsed
				pages={navPages}
				{user}
				userHref="/settings"
				userActive={onSettings}
				{spaces}
				{activeSpaceId}
				onSpaceSelect={selectProject}
				personalSpaceLabel="All projects"
			/>
		</div>
		<main class="min-w-0 flex-1 overflow-auto overscroll-contain pb-24 md:pb-0">
			<Topbar class="md:hidden">
				<span class="text-fc-md font-semibold text-fc-fg">Casier</span>
				<div class="min-w-0 max-w-56 flex-1">
					<SpaceSwitcher
						{spaces}
						activeId={activeSpaceId}
						onSelect={selectProject}
						personalLabel="All projects"
					/>
				</div>
			</Topbar>
			{@render children?.()}
		</main>
		<MobileNav
			items={navPages.map(({ href, label, icon, active }) => ({ href, label, icon, active }))}
			{user}
			profileHref="/settings"
			profileActive={onSettings}
		/>
	</div>
{:else}
	<div class="flex h-[100dvh] items-center justify-center">
		<span
			class="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-foreground motion-reduce:animate-none"
			role="status"
			aria-label="Loading"
		></span>
	</div>
{/if}
