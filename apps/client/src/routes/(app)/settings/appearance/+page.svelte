<script lang="ts">
	import { OptionCards, SettingsRow, SettingsSection, StatusDot, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';
	import { theme, type ThemePreference } from '$lib/theme.svelte';

	const modes = $derived([
		{ value: 'system', label: t('appearance.mode.system'), icon: icons.monitor },
		{ value: 'light', label: t('appearance.mode.light'), icon: icons.sun },
		{ value: 'dark', label: t('appearance.mode.dark'), icon: icons.moon },
	]);

	let mode = $state<string>(theme.preference);
</script>

<svelte:head>
	<title>{t('settings.head.appearance')}</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection title={t('appearance.theme.title')} description={t('appearance.theme.description')}>
		<SettingsRow
			label={t('appearance.scheme.label')}
			description={t('appearance.scheme.description')}
			stacked
		>
			<OptionCards
				options={modes}
				bind:value={mode}
				name="theme-mode"
				label={t('appearance.scheme.label')}
				onSelect={(next) => theme.set(next as ThemePreference)}
			/>
		</SettingsRow>

		<SettingsRow label={t('appearance.now.label')} description={t('appearance.now.description')}>
			<StatusDot
				tone="neutral"
				label={theme.resolved === 'dark' ? t('appearance.mode.dark') : t('appearance.mode.light')}
			/>
		</SettingsRow>
	</SettingsSection>
</div>
