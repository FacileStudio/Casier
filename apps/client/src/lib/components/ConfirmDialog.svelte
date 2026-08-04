<script lang="ts">
	import { Modal, Button } from '@facile/lib';

	let {
		open = $bindable(false),
		title,
		message,
		confirmLabel = 'Delete',
		loading = false,
		onconfirm,
	}: {
		open?: boolean;
		title: string;
		message: string;
		confirmLabel?: string;
		loading?: boolean;
		onconfirm: () => void;
	} = $props();
</script>

<Modal bind:open {title}>
	<p class="text-sm text-muted-foreground">{message}</p>
	<div class="flex justify-end gap-2 pt-4">
		<Button variant="ghost" class="h-9" disabled={loading} onclick={() => (open = false)}>Cancel</Button>
		<Button class="h-9" disabled={loading} onclick={onconfirm}>
			{loading ? 'Working…' : confirmLabel}
		</Button>
	</div>
</Modal>
