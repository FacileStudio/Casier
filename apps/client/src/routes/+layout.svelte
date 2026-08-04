<script lang="ts">
	import '../app.css';
	import { auth } from '$lib/auth.svelte';
	import { theme } from '$lib/theme.svelte';
	import { Toaster } from 'svelte-sonner';

	let { children } = $props();

	auth.restore();
	theme.restore();
</script>

<Toaster position="bottom-right" theme={theme.resolved} />
{@render children?.()}
