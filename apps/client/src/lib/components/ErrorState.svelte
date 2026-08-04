<script lang="ts">
	import { Button } from '@facile/lib';

	let {
		message = 'Something went wrong.',
		onretry,
	}: { message?: string; onretry?: () => void } = $props();
</script>

<div class="flex flex-col items-center justify-center gap-2 px-4 py-16 text-center">
	<iconify-icon
		icon="solar:danger-triangle-linear"
		width="32"
		class="mb-1 text-muted-foreground"
		aria-hidden="true"
	></iconify-icon>
	<p class="max-w-sm text-sm text-muted-foreground">{message}</p>
	{#if onretry}
		<div class="pt-3">
			<Button variant="outline" class="h-9" onclick={onretry}>Try again</Button>
		</div>
	{/if}
</div>
