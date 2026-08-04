<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import Logo from '$lib/components/Logo.svelte';

	let { children } = $props();

	$effect(() => {
		if (auth.isAuthenticated) {
			goto('/spaces', { replaceState: true });
		}
	});
</script>

<div class="flex min-h-dvh flex-col items-center justify-center gap-6 bg-fc-page px-4 py-10">
	<a href="/login" class="rounded-fc-sm focus-visible:outline-2 focus-visible:outline-fc-accent">
		<Logo size={32} wordmark />
	</a>
	<div class="w-full max-w-sm rounded-fc-md border border-fc-fg/7 bg-fc-component p-6 sm:p-8">
		{@render children?.()}
	</div>
	<p class="text-fc-sm text-fc-fg-muted">Self-hosted secrets manager by Facile Studio</p>
</div>
