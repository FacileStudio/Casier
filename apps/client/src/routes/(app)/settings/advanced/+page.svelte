<script lang="ts">
	import { onMount } from 'svelte';
	import { Button, SettingsRow, SettingsSection, StatusDot, icons } from '@facile/muse';
	import { api, type AuthConfig } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';

	let oidcEnabled = $state(false);
	let configLoaded = $state(false);

	onMount(async () => {
		try {
			const config = await api.get<AuthConfig>('/auth/config');
			oidcEnabled = config.oidc_enabled;
		} catch {
			oidcEnabled = false;
		} finally {
			configLoaded = true;
		}
	});
</script>

<svelte:head>
	<title>{t('settings.head.advanced')}</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection title={t('advanced.instance.title')} description={t('advanced.instance.description')}>
		<SettingsRow label={t('profile.interface')} description={t('advanced.interface.description')}>
			<StatusDot tone="neutral" label="Casier web" />
		</SettingsRow>

		<SettingsRow label={t('advanced.sso.label')} description={t('advanced.sso.description')}>
			<StatusDot
				tone={configLoaded && oidcEnabled ? 'success' : 'neutral'}
				label={configLoaded
					? oidcEnabled
						? t('advanced.enabled')
						: t('advanced.notConfigured')
					: t('advanced.checking')}
				pulse={!configLoaded}
			/>
		</SettingsRow>

		<SettingsRow label={t('advanced.storage.label')} description={t('advanced.storage.description')}>
			<StatusDot tone="success" label={t('advanced.storage.status')} />
		</SettingsRow>
	</SettingsSection>

	<SettingsSection title={t('advanced.about.title')} description={t('advanced.about.description')}>
		{#snippet actions()}
			<Button
				variant="secondary"
				href="https://facile.studio"
				target="_blank"
				rel="noopener"
				icon={icons.globe}
			>
				facile.studio
			</Button>
		{/snippet}

		<SettingsRow
			label={t('advanced.about.what.label')}
			description={t('advanced.about.what.description')}
		>
			<Button variant="secondary" href="/projects" icon={icons.folder}>{t('advanced.about.open')}</Button>
		</SettingsRow>
	</SettingsSection>
</div>
