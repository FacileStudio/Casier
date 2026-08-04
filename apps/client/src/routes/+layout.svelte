<script lang="ts">
	import '../app.css';
	import { auth } from '$lib/auth.svelte';
	import { Toaster } from 'svelte-sonner';

	let { children } = $props();

	auth.restore();
</script>

<Toaster position="bottom-right" theme="system" richColors />
{@render children?.()}
