<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Badge,
		Button,
		Input,
		ProfileCard,
		SettingsRow,
		SettingsSection,
		icons,
		toast,
	} from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { api, type AuthConfig } from '$lib/backend';
	import { displayName } from '$lib/format';
	import { t } from '$lib/i18n/index.svelte';

	let oidcEnabled = $state(false);
	let loggingOut = $state(false);

	const email = $derived(auth.email ?? '');
	const name = $derived(displayName({ name: auth.name, email }) || t('common.account'));

	onMount(async () => {
		try {
			const config = await api.get<AuthConfig>('/auth/config');
			oidcEnabled = config.oidc_enabled;
		} catch {
			oidcEnabled = false;
		}
	});

	async function handleLogout() {
		loggingOut = true;
		try {
			await api.post('/auth/logout', {});
		} catch {
			toast.warning(t('profile.logout.warning'));
		}
		auth.logout();
		await goto('/login');
	}
</script>

<svelte:head>
	<title>{t('settings.head.profile')}</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<ProfileCard
		{name}
		{email}
		avatar={auth.avatarUrl || undefined}
		meta={[
			{ label: t('profile.signedInWith'), value: 'Facile' },
			{ label: t('profile.interface'), value: 'Casier web' },
		]}
	/>

	<SettingsSection title={t('profile.account.title')} description={t('profile.account.description')}>
		<SettingsRow
			label={t('common.email')}
			description={t('profile.email.description')}
			for="account-email"
			stacked
		>
			<Input id="account-email" type="email" value={email} disabled />
		</SettingsRow>
		{#if oidcEnabled}
			<SettingsRow label={t('profile.picture.label')} description={t('profile.picture.description')}>
				<Button variant="outline" href="https://sso.facile.studio/account" target="_blank">
					{t('profile.picture.edit')}
				</Button>
			</SettingsRow>
		{/if}
	</SettingsSection>

	<SettingsSection title={t('profile.auth.title')} description={t('profile.auth.description')}>
		<SettingsRow label={t('profile.auth.label')} description={t('profile.auth.rowDescription')}>
			<Badge tone="info">{t('profile.auth.badge')}</Badge>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection title={t('profile.session.title')} description={t('profile.session.description')}>
		<SettingsRow label={t('profile.logout.label')} description={t('profile.logout.description')}>
			<Button
				variant="outline"
				icon={icons.logout}
				disabled={loggingOut}
				onclick={handleLogout}
			>
				{loggingOut ? t('profile.logout.busy') : t('profile.logout.label')}
			</Button>
		</SettingsRow>
	</SettingsSection>
</div>
