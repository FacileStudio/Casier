<script lang="ts">
	/**
	 * CHARTE §4's empty state, in muse's geometry, with this app's API.
	 *
	 * muse's own `EmptyState` takes `icon` as an icon *name*; every call site here passes
	 * a snippet, so the layout is composed on muse's `Card` instead — same card, same
	 * `py-12`, same `gap-4` with title and description bound together in a `gap-1` block.
	 */
	import type { Snippet } from 'svelte';
	import { Card } from '@facile/muse';

	let {
		title,
		description,
		icon,
		children,
	}: {
		title: string;
		description?: string;
		icon?: Snippet;
		children?: Snippet;
	} = $props();
</script>

<Card class="flex flex-col items-center gap-4 py-12 text-center">
	{#if icon}
		<div class="flex justify-center text-fc-fg-muted">
			{@render icon()}
		</div>
	{/if}
	<div class="flex flex-col gap-1">
		<p class="text-fc-sm font-medium text-fc-fg">{title}</p>
		{#if description}
			<p class="mx-auto max-w-sm text-fc-sm text-fc-fg-muted">{description}</p>
		{/if}
	</div>
	{#if children}
		{@render children()}
	{/if}
</Card>
