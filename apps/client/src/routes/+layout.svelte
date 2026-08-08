<script lang="ts">
	import '../app.css';
	import { Toaster } from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { theme } from '$lib/theme.svelte';

	let { children } = $props();

	auth.restore();
	theme.restore();
</script>

<!--
	One Toaster for the whole app, mounted beside the shell rather than inside a page, so a
	route change cannot unmount a toast mid-flight. The extra bottom padding keeps the stack
	clear of MobileNav on a phone.
-->
<Toaster class="pb-28 md:pb-6" />
{@render children?.()}
