<script lang="ts">
	import { PreferencesSection, type Locale, type Theme } from '@facile/muse';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { theme } from '$lib/theme.svelte';
</script>

<svelte:head>
	<title>{t('settings.head.preferences')}</title>
</svelte:head>

<PreferencesSection
	subscription={{ plan: 'Gratuit', href: 'https://sso.facile.studio/spaces' }}
	theme={theme.mode}
	onThemeChange={(next: Theme) => theme.set(next)}
	onLocaleChange={(next: Locale) => rememberLocale(next)}
/>
