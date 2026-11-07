<script lang="ts">
	import { PreferencesSection, type Locale, type Theme } from '@facile/muse';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { theme } from '$lib/theme.svelte';
</script>

<svelte:head>
	<title>{t('settings.head.preferences')}</title>
</svelte:head>

<PreferencesSection
	theme={theme.preference}
	onThemeChange={(next: Theme) => theme.set(next)}
	onLocaleChange={(next: Locale) => rememberLocale(next)}
/>
