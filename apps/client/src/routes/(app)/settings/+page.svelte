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

	let ssoOnly = $state(false);
	let oidcEnabled = $state(false);
	let loggingOut = $state(false);

	const email = $derived(auth.email ?? '');
	const name = $derived(displayName({ name: auth.name, email }) || 'Account');
	const signedInWith = $derived(ssoOnly ? 'Single sign-on' : oidcEnabled ? 'Password or SSO' : 'Password');

	onMount(async () => {
		try {
			const config = await api.get<AuthConfig>('/auth/config');
			ssoOnly = config.sso_only;
			oidcEnabled = config.oidc_enabled;
		} catch {
			ssoOnly = false;
			oidcEnabled = false;
		}
	});

	async function handleLogout() {
		loggingOut = true;
		try {
			await api.post('/auth/logout', {});
		} catch {
			toast.warning('The server did not confirm the sign-out, ending the session locally anyway.');
		}
		auth.logout();
		await goto('/login');
	}
</script>

<svelte:head>
	<title>Settings — Casier</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<ProfileCard
		{name}
		{email}
		avatar={auth.avatarUrl || undefined}
		meta={[
			{ label: 'Signed in with', value: signedInWith },
			{ label: 'Interface', value: 'Casier web' },
		]}
	/>

	<SettingsSection
		title="Account"
		description="An account in Casier is an email address. With single sign-on, your name and picture come from your Facile account."
	>
		<SettingsRow
			label="Email"
			description="Your identity across every project you belong to. Members are invited by this address."
			for="account-email"
			stacked
		>
			<Input id="account-email" type="email" value={email} disabled />
		</SettingsRow>
		{#if oidcEnabled}
			<SettingsRow
				label="Name and picture"
				description="Set them in your Facile account. A change shows here on the next page load."
			>
				<Button variant="outline" href="https://sso.facile.studio/account" target="_blank">
					Edit in Facile account
				</Button>
			</SettingsRow>
		{/if}
	</SettingsSection>

	<SettingsSection
		title="Authentication"
		description="How this instance lets you in. It is set by the server, not by you."
	>
		{#if ssoOnly}
			<SettingsRow
				label="Single sign-on only"
				description="Passwords are disabled on this instance. Your identity provider owns your credentials — change them there."
			>
				<Badge tone="info">Managed by your identity provider</Badge>
			</SettingsRow>
		{:else}
			<SettingsRow
				label="Password"
				description="Casier never shows a password back to you, so a change goes through the reset email like any other."
			>
				<Button variant="outline" href="/forgot-password" icon={icons.mail}>Reset password</Button>
			</SettingsRow>

			{#if oidcEnabled}
				<SettingsRow
					label="Single sign-on"
					description="This instance also accepts SSO. Signing in that way links to the same account as long as the email matches."
				>
					<Badge tone="neutral">Available at sign-in</Badge>
				</SettingsRow>
			{/if}
		{/if}
	</SettingsSection>

	<SettingsSection
		title="Session"
		description="Sessions are per browser. Signing out here leaves your other devices alone."
	>
		<SettingsRow
			label="Log out"
			description="Ends this session and returns you to the sign-in page. API tokens keep working — revoke those from the project that owns them."
		>
			<Button
				variant="outline"
				icon={icons.logout}
				disabled={loggingOut}
				onclick={handleLogout}
			>
				{loggingOut ? 'Signing out…' : 'Log out'}
			</Button>
		</SettingsRow>
	</SettingsSection>
</div>
