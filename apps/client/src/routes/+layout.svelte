<script lang="ts">
	import '../app.css';
	import { auth } from '$lib/auth.svelte';
	import { Toaster } from '$lib/components/ui/sonner';

	let { children } = $props();

	auth.restore();
</script>

<Toaster />
{@render children?.()}
