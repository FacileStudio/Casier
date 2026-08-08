<script lang="ts">
	/**
	 * CHARTE §14: one settings page, top pill tabs, then a rule.
	 *
	 * The section lives in the URL rather than in a local `$state`, so `/settings/api`
	 * deep-links, a reload stays put and browser-back walks the sections. That is the
	 * whole reason `Tabs` items carry `href`.
	 *
	 * Only the sections Casier has real content for are here. Notifications, Pool and
	 * Members are canonical but Casier has no account-level mail, no Nook client, and
	 * membership that belongs to a project rather than to an account — an empty tab is
	 * worse than a missing one.
	 */
	import { page } from '$app/state';
	import { Divider, PageTransition, Tabs, icons } from '@facile/muse';

	let { children } = $props();

	const sections = [
		{ id: 'profile', label: 'Profile', icon: icons.userCircle, href: '/settings' },
		{ id: 'appearance', label: 'Appearance', icon: icons.palette, href: '/settings/appearance' },
		{ id: 'api', label: 'API', icon: icons.key, href: '/settings/api' },
		{ id: 'advanced', label: 'Advanced', icon: icons.shield, href: '/settings/advanced' },
	];

	const active = $derived(
		sections.find((section) => section.href === page.url.pathname)?.id ?? 'profile'
	);
</script>

<div class="mx-auto flex w-full max-w-fc-lg flex-col gap-8 px-4 py-6 md:px-8 md:py-10">
	<div class="flex flex-col gap-1">
		<h1 class="text-fc-2xl font-semibold text-fc-fg">Settings</h1>
		<p class="text-fc-sm text-fc-fg-muted">
			Your account, this browser, and how machines reach this instance.
		</p>
	</div>

	<div class="flex flex-col gap-4">
		<Tabs items={sections} value={active} label="Settings sections" />
		<Divider class="my-0" />
	</div>

	<PageTransition key={active} distance={8} duration={0.25}>
		{@render children?.()}
	</PageTransition>
</div>
