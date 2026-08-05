<script lang="ts">
	/**
	 * Side drawer, in the shape shadcn-svelte gives one, built on the bits-ui
	 * Dialog primitive already in the bundle. It replaces muse's `Modal`, whose
	 * native `<dialog>` loses its centering to Tailwind's preflight margin reset
	 * and lands in the top-left corner. A drawer is anchored to an edge, so
	 * there is no centering left to get wrong.
	 */
	import type { Snippet } from 'svelte';
	import { Dialog } from 'bits-ui';

	let {
		open = $bindable(false),
		title,
		description,
		class: className = '',
		children,
	}: {
		open?: boolean;
		title: string;
		description?: string;
		class?: string;
		children: Snippet;
	} = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay class="drawer-overlay fixed inset-0 z-50 bg-black/50" />
		<Dialog.Content
			class="drawer-panel fixed inset-y-0 right-0 z-50 flex w-full flex-col border-l border-border bg-background outline-none sm:max-w-lg {className}"
		>
			<div class="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
				<div class="min-w-0">
					<Dialog.Title class="truncate text-base font-semibold">{title}</Dialog.Title>
					<Dialog.Description class="mt-1 text-sm text-muted-foreground {description ? '' : 'sr-only'}">
						{description ?? title}
					</Dialog.Description>
				</div>
				<Dialog.Close
					class="-mr-1.5 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					aria-label="Close"
				>
					<iconify-icon icon="solar:close-circle-linear" width="18" aria-hidden="true"></iconify-icon>
				</Dialog.Close>
			</div>
			<div class="min-h-0 flex-1 overflow-y-auto px-5 py-5">
				{@render children()}
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
