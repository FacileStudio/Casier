<script lang="ts">
	import type { Snippet } from 'svelte';

	/** Standard suite page header: title row, optional actions, optional tab strip below. */
	let {
		title,
		description,
		back,
		actions,
		tabs,
	}: {
		title: string;
		description?: string;
		back?: { href: string; label: string };
		actions?: Snippet;
		tabs?: Snippet;
	} = $props();
</script>

<div class="border-b border-border px-4 py-4 md:px-8 md:py-5">
	{#if back}
		<a
			href={back.href}
			class="mb-2 inline-flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
		>
			<iconify-icon icon="solar:alt-arrow-left-linear" width="14" aria-hidden="true"></iconify-icon>
			{back.label}
		</a>
	{/if}

	<div class="flex flex-wrap items-center justify-between gap-3">
		<div class="min-w-0">
			<h1 class="truncate text-lg font-semibold">{title}</h1>
			{#if description}
				<p class="mt-0.5 line-clamp-2 text-sm text-muted-foreground">{description}</p>
			{/if}
		</div>
		{#if actions}
			<div class="flex shrink-0 items-center gap-2">
				{@render actions()}
			</div>
		{/if}
	</div>

	{#if tabs}
		<div class="mt-4 flex gap-1 overflow-x-auto">
			{@render tabs()}
		</div>
	{/if}
</div>
