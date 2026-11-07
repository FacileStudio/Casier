<script lang="ts">
	import { browser } from '$app/environment';
	import { Button, SecretField, SettingsRow, SettingsSection, icons } from '@facile/muse';
	import { API_BASE } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';

	const commands = $derived([
		{
			label: t('api.cli.login.label'),
			command: 'casier login',
			description: t('api.cli.login.description'),
		},
		{
			label: t('api.cli.init.label'),
			command: 'casier init',
			description: t('api.cli.init.description'),
		},
		{
			label: t('api.cli.run.label'),
			command: 'casier run -- npm start',
			description: t('api.cli.run.description'),
		},
	]);

	const baseUrl = $derived(browser ? `${window.location.origin}${API_BASE}` : API_BASE);
</script>

<svelte:head>
	<title>{t('settings.head.tokens')}</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection title={t('api.endpoint.title')} description={t('api.endpoint.description')}>
		<SettingsRow label={t('api.baseUrl.label')} description={t('api.baseUrl.description')} stacked>
			<SecretField value={baseUrl} sensitive={false} class="w-full" />
		</SettingsRow>
	</SettingsSection>

	<SettingsSection title={t('api.cli.title')} description={t('api.cli.description')}>
		{#each commands as entry (entry.command)}
			<SettingsRow label={entry.label} description={entry.description} stacked>
				<SecretField value={entry.command} sensitive={false} class="w-full" />
			</SettingsRow>
		{/each}
	</SettingsSection>

	<SettingsSection title={t('api.tokens.title')} description={t('api.tokens.description')}>
		{#snippet actions()}
			<Button variant="secondary" href="/projects" icon={icons.folder}>{t('api.tokens.projects')}</Button>
		{/snippet}

		<SettingsRow label={t('api.tokens.where.label')} description={t('api.tokens.where.description')}>
			<Button variant="secondary" href="/projects" icon={icons.key}>{t('api.tokens.issue')}</Button>
		</SettingsRow>

		<SettingsRow label={t('api.tokens.cli.label')} description={t('api.tokens.cli.description')}>
			<SecretField value="CASIER_TOKEN" sensitive={false} class="w-full sm:w-64" />
		</SettingsRow>
	</SettingsSection>
</div>
