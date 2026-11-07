<script lang="ts">
	import { setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { MobileNav, SideBar, SpaceSwitcher, Spinner, Topbar, icons, toast, type Locale, type Theme } from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { displayName } from '$lib/format';
	import { locale, rememberLocale, t } from '$lib/i18n/index.svelte';
	import { theme } from '$lib/theme.svelte';
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
			label: t('nav.projects'),
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
	 * The user card shows the identity provider's display name and picture. An
	 * account without a display name falls back to the email's local part, and
	 * muse's Avatar draws its initial.
	 */
	const user = $derived({
		name: displayName({ name: auth.name, email: auth.email ?? '' }) || t('common.account'),
		email: auth.email ?? undefined,
		avatar: auth.avatarUrl || undefined,
	});

	async function signOut() {
		try {
			await api.post('/auth/logout', {});
		} catch {
			toast.warning(t('profile.logout.warning'));
		}
		auth.logout();
		await goto('/login');
	}

	const userMenu = $derived({
		helpHref: 'https://github.com/FacileStudio/Casier/tree/main/docs',
		onSignOut: signOut,
		locale: locale(),
		onLocaleChange: (next: Locale) => rememberLocale(next),
		theme: theme.preference,
		onThemeChange: (next: Theme) => theme.set(next),
	});

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
				{userMenu}
				{spaces}
				{activeSpaceId}
				onSpaceSelect={selectProject}
				personalSpaceLabel={t('nav.allProjects')}
			/>
		</div>
		<main class="min-w-0 flex-1 overflow-auto overscroll-contain pb-24 md:pb-0">
			<Topbar class="md:hidden">
				<span class="text-fc-md font-semibold text-fc-fg">Casier</span>
				{#if spaces.length > 0}
					<div class="min-w-0 max-w-56 flex-1">
						<SpaceSwitcher
							{spaces}
							activeId={activeSpaceId}
							onSelect={selectProject}
							personalLabel={t('nav.allProjects')}
						/>
					</div>
				{/if}
			</Topbar>
			{@render children?.()}
		</main>
		<MobileNav
			items={navPages.map(({ href, label, icon, active }) => ({ href, label, icon, active }))}
			{user}
			profileHref="/settings"
			profileActive={onSettings}
			{userMenu}
		/>
	</div>
{:else}
	<div class="flex h-[100dvh] items-center justify-center">
		<Spinner label={t('common.loading')} />
	</div>
{/if}
