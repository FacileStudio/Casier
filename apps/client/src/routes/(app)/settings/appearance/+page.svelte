<script lang="ts">
	import { OptionCards, SettingsRow, SettingsSection, StatusDot, icons } from '@facile/muse';
	import { theme, type ThemePreference } from '$lib/theme.svelte';

	const modes = [
		{ value: 'system', label: 'System', icon: icons.monitor },
		{ value: 'light', label: 'Light', icon: icons.sun },
		{ value: 'dark', label: 'Dark', icon: icons.moon },
	];

	let mode = $state<string>(theme.preference);
</script>

<svelte:head>
	<title>Appearance — Casier</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Theme"
		description="Stored in this browser. Every device you sign in from keeps its own choice."
	>
		<SettingsRow
			label="Colour scheme"
			description="System follows your operating system and keeps following it while the app is open."
			stacked
		>
			<OptionCards
				options={modes}
				bind:value={mode}
				name="theme-mode"
				label="Colour scheme"
				onSelect={(next) => theme.set(next as ThemePreference)}
			/>
		</SettingsRow>

		<SettingsRow
			label="Showing now"
			description="What the two-token palette resolves to right now — chroma-zero either way."
		>
			<StatusDot
				tone="neutral"
				label={theme.resolved === 'dark' ? 'Dark' : 'Light'}
			/>
		</SettingsRow>
	</SettingsSection>
</div>
