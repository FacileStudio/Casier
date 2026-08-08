<script lang="ts">
	import { setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { MobileNav, SideBar, icons } from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		ApiError,
		COOKIE_SESSION,
		type Me,
		type Project,
		type ProjectsResponse,
	} from '$lib/backend';

	let { children } = $props();

	let projects = $state<Project[]>([]);
	let collapsed = $state(false);

	async function refreshProjects() {
		try {
			const result = await api.get<ProjectsResponse>('/projects', auth.token);
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
		api.get<Me>('/auth/me')
			.then((me) => {
				auth.login(COOKIE_SESSION, me.id, me.email);
				refreshProjects();
			})
			.catch(() => goto('/login'));
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

	const user = $derived({ name: auth.email ?? 'Account' });

	const spaces = $derived(projects.map((project) => ({ id: project.slug, name: project.name })));

	const activeSpaceId = $derived(page.params.slug ?? null);

	function selectProject(slug: string | null) {
		goto(slug ? `/projects/${slug}` : '/projects');
	}
</script>

{#if auth.isAuthenticated}
	<div class="flex h-[100dvh] w-full overflow-hidden">
		<SideBar
			class="hidden shrink-0 rounded-none border-r border-border bg-background md:flex"
			icon="solar:safe-square-bold-duotone"
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
		<main class="min-w-0 flex-1 overflow-auto overscroll-contain pb-24 md:pb-0">
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
