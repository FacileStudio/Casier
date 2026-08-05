<script lang="ts">
	import { setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		ApiError,
		COOKIE_SESSION,
		type Me,
		type Project,
		type ProjectsResponse,
	} from '$lib/backend';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import MobileNav from '$lib/components/MobileNav.svelte';

	let { children } = $props();

	let projects = $state<Project[]>([]);

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

	async function handleLogout() {
		try {
			await api.post('/auth/logout', {}, auth.token);
		} catch {}
		auth.logout();
		goto('/login');
	}
</script>

{#if auth.isAuthenticated}
	<div class="flex h-[100dvh] w-full overflow-hidden">
		<Sidebar email={auth.email} {projects} onlogout={handleLogout} />
		<main class="min-w-0 flex-1 overflow-auto pb-24 md:pb-0">
			{@render children?.()}
		</main>
		<MobileNav />
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
