<script lang="ts">
	/**
	 * The failure twin of `EmptyState` — same muse `Card`, same CHARTE §4 geometry, so a
	 * list that failed to load and a list with nothing in it occupy the same shape.
	 */
	import { Button, Card } from '@facile/muse';

	let {
		message = 'Something went wrong.',
		onretry,
	}: { message?: string; onretry?: () => void } = $props();
</script>

<Card class="flex flex-col items-center gap-4 py-12 text-center">
	<iconify-icon
		icon="solar:danger-triangle-linear"
		width="24"
		height="24"
		class="block text-fc-fg-muted"
		aria-hidden="true"
	></iconify-icon>
	<p class="mx-auto max-w-sm text-fc-sm text-fc-fg-muted">{message}</p>
	{#if onretry}
		<Button variant="outline" onclick={onretry}>Try again</Button>
	{/if}
</Card>
