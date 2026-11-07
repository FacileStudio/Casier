<script lang="ts">
	import { page } from '$app/state';
	import { SettingsLayout, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let { children } = $props();

	const sections = $derived([
		{ id: 'profile', label: t('settings.tab.profile'), icon: icons.userCircle, href: '/settings' },
		{ id: 'preferences', label: t('settings.tab.preferences'), icon: icons.palette, href: '/settings/preferences' },
		{ id: 'tokens', label: t('settings.tab.tokens'), icon: icons.key, href: '/settings/tokens' },
		{ id: 'advanced', label: t('settings.tab.advanced'), icon: icons.shield, href: '/settings/advanced' },
	]);

	const active = $derived(sections.find((section) => section.href === page.url.pathname)?.id);
</script>

<div class="mx-auto w-full max-w-fc-lg px-4 py-6 md:px-8 md:py-10">
	<SettingsLayout {sections} {active} backHref="/settings" description={t('settings.subtitle')}>
		{@render children?.()}
	</SettingsLayout>
</div>
