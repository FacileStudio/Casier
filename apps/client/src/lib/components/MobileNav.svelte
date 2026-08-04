<script lang="ts">
	import { page } from '$app/state';

	const items = [
		{ href: '/projects', label: 'Projects', icon: 'solar:widget-2-linear' },
		{ href: '/settings', label: 'Settings', icon: 'solar:settings-linear' },
	];

	function isActive(href: string) {
		return page.url.pathname === href || page.url.pathname.startsWith(href + '/');
	}
</script>

<nav
	class="fixed inset-x-0 z-50 flex justify-center px-4 md:hidden"
	style="bottom: max(0.75rem, env(safe-area-inset-bottom))"
	aria-label="Primary"
>
	<div
		class="flex items-center gap-1 rounded-full border border-border/40 bg-background/70 p-1.5 shadow-lg shadow-black/10 backdrop-blur-2xl backdrop-saturate-150"
	>
		{#each items as item (item.href)}
			{@const active = isActive(item.href)}
			<a
				href={item.href}
				aria-label={item.label}
				aria-current={active ? 'page' : undefined}
				class="flex items-center gap-2 rounded-full px-4 py-2 text-sm transition-colors {active
					? 'bg-foreground font-medium text-background'
					: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
			>
				<iconify-icon icon={item.icon} width="18" aria-hidden="true"></iconify-icon>
				{item.label}
			</a>
		{/each}
	</div>
</nav>
