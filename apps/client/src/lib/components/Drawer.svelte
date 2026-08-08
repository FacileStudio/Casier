<script lang="ts">
	/**
	 * The app's overlay for forms and detail panels, built on muse's `Modal`.
	 *
	 * muse's own `Drawer` is a bottom sheet — the mobile counterpart to `Modal` per
	 * CHARTE §11 — and these panels carry desktop forms and scrollable lists, so
	 * `Modal` at `size="lg"` is the right side of that pair. The name stays `Drawer`
	 * because every call site imports it under that name with these props.
	 *
	 * `title` and `description` go through the `header` snippet rather than muse's
	 * `title`, which means the accessible name has to be wired by hand: `Modal` only
	 * points `aria-labelledby` at its own heading, so the id rides through `...rest`.
	 */
	import type { Snippet } from 'svelte';
	import { Modal } from '@facile/muse';

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

	const titleId = $props.id();
</script>

<Modal bind:open size="lg" showClose aria-labelledby={titleId} class={className}>
	{#snippet header()}
		<div class="flex min-w-0 flex-col gap-1">
			<h2 id={titleId} class="truncate text-fc-lg font-semibold">{title}</h2>
			{#if description}
				<p class="text-fc-sm text-fc-fg-muted">{description}</p>
			{/if}
		</div>
	{/snippet}
	{@render children()}
</Modal>
