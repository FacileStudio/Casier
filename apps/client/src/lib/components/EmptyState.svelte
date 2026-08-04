<script lang="ts">
	import type { Snippet } from 'svelte';

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

<div class="flex flex-col items-center justify-center gap-2 px-4 py-16 text-center">
	{#if icon}
		<div class="mb-1 text-fc-fg-muted opacity-60">
			{@render icon()}
		</div>
	{/if}
	<p class="text-fc-lg font-semibold text-fc-fg">{title}</p>
	{#if description}
		<p class="max-w-sm text-fc-md text-fc-fg-muted">{description}</p>
	{/if}
	{#if children}
		<div class="pt-3">
			{@render children()}
		</div>
	{/if}
</div>
