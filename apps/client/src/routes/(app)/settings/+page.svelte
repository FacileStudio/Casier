<script lang="ts">
	import { Badge, ProfileSection, SettingsRow, SettingsSection } from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { displayName } from '$lib/format';
	import { t } from '$lib/i18n/index.svelte';

	const email = $derived(auth.email ?? '');
	const name = $derived(displayName({ name: auth.name, email }) || t('common.account'));
</script>

<svelte:head>
	<title>{t('settings.head.profile')}</title>
</svelte:head>

<ProfileSection {name} {email} avatar={auth.avatarUrl || undefined} />

<SettingsSection title={t('profile.auth.title')} description={t('profile.auth.description')}>
	<SettingsRow label={t('profile.auth.label')} description={t('profile.auth.rowDescription')}>
		<Badge tone="info">{t('profile.auth.badge')}</Badge>
	</SettingsRow>
</SettingsSection>
