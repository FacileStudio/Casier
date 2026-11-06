<script lang="ts">
	import '../app.css';
	import { Toaster, provideLocale } from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { bindLocale, localeSource } from '$lib/i18n/index.svelte';
	import { theme } from '$lib/theme.svelte';

	let { children } = $props();

	auth.check();
	theme.restore();

	const i18n = provideLocale(localeSource);
	bindLocale(() => i18n.locale);
</script>

<!--
	One Toaster for the whole app, mounted beside the shell rather than inside a page, so a
	route change cannot unmount a toast mid-flight. The extra bottom padding keeps the stack
	clear of MobileNav on a phone.
-->
<Toaster class="pb-28 md:pb-6" />
{@render children?.()}
