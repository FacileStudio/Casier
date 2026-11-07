<script lang="ts">
	/**
	 * The failure twin of `EmptyState` — same muse `Card`, same CHARTE §4 geometry, so a
	 * list that failed to load and a list with nothing in it occupy the same shape.
	 */
	import { Button, Card } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let {
		message,
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
	<p class="mx-auto max-w-sm text-fc-sm text-fc-fg-muted">{message ?? t('common.genericError')}</p>
	{#if onretry}
		<Button variant="secondary" onclick={onretry}>{t('common.tryAgain')}</Button>
	{/if}
</Card>
