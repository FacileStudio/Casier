<script lang="ts">
	/**
	 * Casier brand mark. `size` drives the icon; the wordmark inherits its font
	 * size from the parent, so callers set it with a text-* class.
	 */
	let {
		size = 28,
		wordmark = false,
		class: className = '',
	}: { size?: number; wordmark?: boolean; class?: string } = $props();
</script>

<span class="inline-flex items-center gap-3 {className}">
	<iconify-icon
		icon="solar:safe-square-bold-duotone"
		width={size}
		height={size}
		class="shrink-0"
		aria-hidden="true"
	></iconify-icon>
	{#if wordmark}
		<span class="font-heading font-bold tracking-tight">Casier</span>
	{/if}
</span>
