<script lang="ts">
	import { page } from '$app/state';
	import type { Project } from '$lib/backend';
	import Logo from './Logo.svelte';
	import ProjectSwitcher from './ProjectSwitcher.svelte';

	let {
		email,
		projects = [],
		onlogout,
	}: { email: string | null; projects?: Project[]; onlogout: () => void } = $props();

	const navLinks = [
		{ href: '/projects', label: 'Projects', icon: 'solar:widget-2-linear' },
		{ href: '/settings', label: 'Settings', icon: 'solar:settings-linear' },
	];

	function initials(value: string | null) {
		const name = (value ?? '').trim();
		if (!name) return '?';
		const [local] = name.split('@');
		return local.slice(0, 2).toUpperCase();
	}
</script>

<aside
	class="sticky top-0 hidden h-[100dvh] w-60 shrink-0 flex-col border-r border-border bg-background md:flex"
>
	<a href="/projects" class="flex items-center px-5 pt-8 pb-4 text-2xl">
		<Logo size={28} wordmark />
	</a>

	<ProjectSwitcher {projects} />

	<nav class="flex flex-1 flex-col gap-1 px-3">
		{#each navLinks as link (link.href)}
			{@const active =
				page.url.pathname === link.href || page.url.pathname.startsWith(link.href + '/')}
			<a
				href={link.href}
				aria-current={active ? 'page' : undefined}
				class="flex items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors {active
					? 'bg-foreground font-medium text-background'
					: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
			>
				<iconify-icon icon={link.icon} width="16" aria-hidden="true"></iconify-icon>
				{link.label}
			</a>
		{/each}
	</nav>

	<div class="h-px bg-border"></div>

	<div class="flex flex-col gap-2 p-4">
		<a
			href="/settings"
			class="flex items-center gap-3 rounded-xl border border-border/70 bg-muted/40 p-2.5 transition-colors hover:bg-muted"
		>
			<span
				class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-border bg-foreground text-xs font-semibold text-background"
			>
				{initials(email)}
			</span>
			<div class="min-w-0 flex-1">
				<p class="truncate text-sm font-medium">Account</p>
				<p class="truncate text-xs text-muted-foreground">{email ?? ''}</p>
			</div>
		</a>
		<button
			onclick={onlogout}
			class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
		>
			<iconify-icon icon="solar:logout-2-linear" width="16" aria-hidden="true"></iconify-icon>
			Log out
		</button>
	</div>
</aside>
