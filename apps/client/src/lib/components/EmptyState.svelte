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

<div class="rounded-lg border border-dashed border-border p-10 text-center">
	{#if icon}
		<div class="mb-3 flex justify-center text-muted-foreground">
			{@render icon()}
		</div>
	{/if}
	<p class="text-base font-semibold text-foreground">{title}</p>
	{#if description}
		<p class="mx-auto mt-1 max-w-sm text-sm text-muted-foreground">{description}</p>
	{/if}
	{#if children}
		<div class="mt-5 flex justify-center">
			{@render children()}
		</div>
	{/if}
</div>
