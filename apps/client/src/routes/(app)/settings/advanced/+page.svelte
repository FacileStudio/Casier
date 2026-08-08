<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Button, SettingsRow, SettingsSection, StatusDot, icons } from '@facile/muse';
	import { api, type AuthConfig } from '$lib/backend';

	let ssoOnly = $state(false);
	let oidcEnabled = $state(false);
	let configLoaded = $state(false);

	onMount(async () => {
		try {
			const config = await api.get<AuthConfig>('/auth/config');
			ssoOnly = config.sso_only;
			oidcEnabled = config.oidc_enabled;
		} catch {
			ssoOnly = false;
			oidcEnabled = false;
		} finally {
			configLoaded = true;
		}
	});
</script>

<svelte:head>
	<title>Advanced — Casier</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Instance"
		description="Worth quoting when you file a bug against a self-hosted install."
	>
		<SettingsRow label="Interface" description="This app is the browser client for the Casier API.">
			<StatusDot tone="neutral" label="Casier web" />
		</SettingsRow>

		<SettingsRow
			label="Password sign-in"
			description="Turned off instance-wide by SSO_ONLY, which hides the password form entirely."
		>
			<StatusDot
				tone={configLoaded && ssoOnly ? 'neutral' : 'success'}
				label={configLoaded ? (ssoOnly ? 'Disabled' : 'Enabled') : 'Checking…'}
				pulse={!configLoaded}
			/>
		</SettingsRow>

		<SettingsRow
			label="Single sign-on"
			description="OIDC federation against the instance's identity provider. Enabled by OIDC_ISSUER on the server."
		>
			<StatusDot
				tone={configLoaded && oidcEnabled ? 'success' : 'neutral'}
				label={configLoaded ? (oidcEnabled ? 'Enabled' : 'Not configured') : 'Checking…'}
				pulse={!configLoaded}
			/>
		</SettingsRow>

		<SettingsRow
			label="Secret storage"
			description="Values are encrypted with a per-project data key and a fresh nonce per secret. The server holds a hash of every session and API token, never the token itself."
		>
			<StatusDot tone="success" label="AES-256-GCM at rest" />
		</SettingsRow>
	</SettingsSection>

	<SettingsSection
		title="About"
		description="Casier is the secrets and environment manager of the Facile Suite."
	>
		{#snippet actions()}
			<Button
				variant="outline"
				href="https://facile.studio"
				target="_blank"
				rel="noopener"
				icon={icons.globe}
			>
				facile.studio
			</Button>
		{/snippet}

		<SettingsRow
			label="What it does"
			description="Groups secrets into environments inside a project, encrypts every value at rest, versions each change and writes an audit row for it. The CLI reads the same secrets into a shell, a script or CI."
		>
			<Button variant="outline" href="/projects" icon={icons.folder}>Open projects</Button>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection
		title="Danger zone"
		description="Irreversible actions live at the bottom of this tab — never in the tab strip."
	>
		<Alert tone="info">
			Casier has nothing destructive at the account level: an account is only an email, and it
			owns no data of its own. What is destructive belongs to a project — deleting an
			environment, purging a secret, deleting the project itself — and lives in that project's
			settings, gated on owner or admin.
		</Alert>
	</SettingsSection>
</div>
