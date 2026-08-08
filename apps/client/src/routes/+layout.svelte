<script lang="ts">
	import '../app.css';
	import { Toaster } from '@facile/muse';
	import { Toaster as SonnerToaster } from 'svelte-sonner';
	import { auth } from '$lib/auth.svelte';
	import { theme } from '$lib/theme.svelte';

	let { children } = $props();

	auth.restore();
	theme.restore();
</script>

<Toaster position="top" />
<SonnerToaster position="bottom-right" theme={theme.resolved} />
{@render children?.()}
