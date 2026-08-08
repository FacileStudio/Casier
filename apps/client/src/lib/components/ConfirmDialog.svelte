<script lang="ts">
	/**
	 * muse's `ConfirmModal` under this app's existing prop names.
	 *
	 * The one real adaptation is ownership of the close. `ConfirmModal` closes itself
	 * when `onConfirm` resolves, while every call site here drives `open` and `loading`
	 * itself and deliberately stays open on failure. So `onConfirm` returns a promise
	 * that settles only when the caller drops `loading`, and it settles by rejecting —
	 * a rejection leaves the dialog open and hands the decision back to the caller,
	 * which is the semantics these props already promised. Until then `ConfirmModal` is
	 * pending: spinner on Confirm, both buttons disabled, Escape and backdrop inert.
	 */
	import { ConfirmModal } from '@facile/muse';

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

	let release: (() => void) | null = $state(null);

	$effect(() => {
		if (loading || !release) return;
		const settle = release;
		release = null;
		settle();
	});

	function confirm() {
		onconfirm();
		return new Promise<void>((_resolve, reject) => {
			release = () => reject(new Error('handled by caller'));
		});
	}
</script>

<ConfirmModal
	bind:open
	{title}
	description={message}
	{confirmLabel}
	tone="danger"
	onConfirm={confirm}
/>
