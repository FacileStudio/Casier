<script lang="ts">
	import { Button } from '@facile/lib';
	import CircleAlertIcon from '@lucide/svelte/icons/circle-alert';

	let {
		message = 'Something went wrong.',
		onretry,
	}: { message?: string; onretry?: () => void } = $props();
</script>

<div class="flex flex-col items-center justify-center gap-2 px-4 py-16 text-center">
	<CircleAlertIcon class="mb-1 size-8 text-fc-danger" aria-hidden="true" />
	<p class="max-w-sm text-fc-md text-fc-fg-muted">{message}</p>
	{#if onretry}
		<div class="pt-3">
			<Button variant="outline" onclick={onretry}>Try again</Button>
		</div>
	{/if}
</div>
