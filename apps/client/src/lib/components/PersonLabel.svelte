<script lang="ts">
	/**
	 * A person as every audit, history and token row names them: their picture
	 * from the identity provider, initials when they have none or it fails to
	 * load, and their name beside it. `fallback` is shown when the person no
	 * longer resolves to an account, such as the email an audit row recorded.
	 */
	import { Avatar } from '@facile/muse';
	import type { Person } from '$lib/backend';
	import { displayName } from '$lib/format';

	let {
		person,
		fallback = 'unknown',
	}: { person: Person | null | undefined; fallback?: string } = $props();

	const label = $derived(displayName(person) || fallback);
</script>

<span class="inline-flex min-w-0 items-center gap-2" title={person?.email || fallback}>
	<Avatar src={person?.avatar_url || undefined} name={label} alt="" size="sm" class="h-6 w-6 text-fc-xs" />
	<span class="truncate">{label}</span>
</span>
