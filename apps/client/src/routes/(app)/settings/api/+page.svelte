<script lang="ts">
	import { browser } from '$app/environment';
	import { Button, SecretField, SettingsRow, SettingsSection, icons } from '@facile/muse';
	import { API_BASE } from '$lib/backend';

	const commands = [
		{
			label: 'Sign in',
			command: 'casier login',
			description: 'Discovers this instance through /auth/config and stores a session token.',
		},
		{
			label: 'Link a directory',
			command: 'casier init',
			description: 'Writes the project and environment this directory reads from.',
		},
		{
			label: 'Run with secrets',
			command: 'casier run -- npm start',
			description: 'Injects the environment for one process. Nothing is written to disk.',
		},
	];

	const baseUrl = $derived(browser ? `${window.location.origin}${API_BASE}` : API_BASE);
</script>

<svelte:head>
	<title>API — Casier</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Endpoint"
		description="Point the CLI or any HTTP client here. The URL is not a secret — the token is."
	>
		<SettingsRow
			label="Base URL"
			description="Behind a reverse proxy the API answers under /api; in development it answers at the root."
			stacked
		>
			<SecretField value={baseUrl} sensitive={false} class="w-full" />
		</SettingsRow>
	</SettingsSection>

	<SettingsSection
		title="Command line"
		description="The casier CLI pulls the same secrets into your shell, your scripts and your CI."
	>
		{#each commands as entry (entry.command)}
			<SettingsRow label={entry.label} description={entry.description} stacked>
				<SecretField value={entry.command} sensitive={false} class="w-full" />
			</SettingsRow>
		{/each}
	</SettingsSection>

	<SettingsSection
		title="Tokens"
		description="Casier scopes a machine credential to one project, so tokens are issued there rather than here."
	>
		{#snippet actions()}
			<Button variant="outline" href="/projects" icon={icons.folder}>Projects</Button>
		{/snippet}

		<SettingsRow
			label="Where tokens live"
			description="Open a project, then API tokens. A token is bound to that project, to a permission set and to the environments you name — it can never read past them."
		>
			<Button variant="outline" href="/projects" icon={icons.key}>Issue a token</Button>
		</SettingsRow>

		<SettingsRow
			label="How the CLI uses one"
			description="Non-interactive callers send it as CASIER_TOKEN. Tokens expire — 90 days by default, 730 at most — and are revoked rather than deleted so the audit log still names them."
		>
			<SecretField value="CASIER_TOKEN" sensitive={false} class="w-full sm:w-64" />
		</SettingsRow>
	</SettingsSection>
</div>
