<script lang="ts">
	import type { Snippet } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import Logo from '$lib/components/Logo.svelte';

	let { children }: { children?: Snippet } = $props();

	$effect(() => {
		if (auth.isAuthenticated) {
			goto('/projects', { replaceState: true });
		}
	});
</script>

<div class="flex min-h-screen">
	<div class="hidden flex-col bg-black px-12 py-10 lg:flex lg:w-1/2">
		<a href="/" class="mb-auto flex items-center gap-3 text-xl text-white">
			<Logo size={28} wordmark />
		</a>

		<div class="mb-auto">
			<h2 class="font-heading text-4xl font-bold leading-tight tracking-tight text-white">
				Your secrets.<br />Your server.
			</h2>
			<p class="mt-4 max-w-xs text-sm leading-relaxed text-white/50">
				Self-hosted secrets and environment management for teams.
			</p>
		</div>

		<p class="text-xs text-white/30">
			&copy; {new Date().getFullYear()} Casier by Facile.
		</p>
	</div>

	<div class="flex w-full flex-col items-center justify-center bg-background px-8 py-12 lg:w-1/2">
		<div class="w-full max-w-sm">
			<a href="/" class="mb-8 flex items-center gap-3 text-xl lg:hidden">
				<Logo size={28} wordmark />
			</a>
			{@render children?.()}
		</div>
	</div>
</div>
