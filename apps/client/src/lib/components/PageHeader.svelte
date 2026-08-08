<script lang="ts">
	/**
	 * Standard suite page header: optional back link, title row, optional actions,
	 * optional tab rail below.
	 *
	 * The rail stays a snippet of caller-owned anchors and buttons, so muse's `Tabs`
	 * (which takes an `items` array) cannot be dropped in without changing this API.
	 * It is `relative` on purpose: an absolutely positioned or `sr-only` descendant of a
	 * `position: static` `overflow-x-auto` box is measured against the document and
	 * inflates `documentElement.scrollWidth`, which is what makes a page scroll
	 * sideways for no visible reason. Labels in the rail belong on `aria-label`, never
	 * in a nested `sr-only` span.
	 */
	import type { Snippet } from 'svelte';

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

<div class="flex flex-col gap-4 border-b border-fc-border px-4 py-4 md:px-8 md:py-5">
	<div class="flex flex-col gap-2">
		{#if back}
			<a
				href={back.href}
				class="inline-flex w-fit items-center gap-1.5 text-fc-xs text-fc-fg-muted transition-colors hover:text-fc-fg"
			>
				<iconify-icon
					icon="solar:alt-arrow-left-linear"
					width="14"
					height="14"
					class="block shrink-0"
					aria-hidden="true"
				></iconify-icon>
				{back.label}
			</a>
		{/if}

		<div class="flex flex-wrap items-center justify-between gap-3">
			<div class="flex min-w-0 flex-col gap-1">
				<h1 class="truncate text-fc-lg font-semibold text-fc-fg">{title}</h1>
				{#if description}
					<p class="line-clamp-2 text-fc-sm text-fc-fg-muted">{description}</p>
				{/if}
			</div>
			{#if actions}
				<div class="flex shrink-0 items-center gap-2">
					{@render actions()}
				</div>
			{/if}
		</div>
	</div>

	{#if tabs}
		<div class="relative flex w-full items-center gap-1 overflow-x-auto">
			{@render tabs()}
		</div>
	{/if}
</div>
