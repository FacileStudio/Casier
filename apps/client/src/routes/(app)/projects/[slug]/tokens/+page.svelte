<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { Card, Field, Input, Skeleton, TokensSection, toast, type TokenDraft, type TokenItem } from '@facile/muse';
	import {
		api,
		type ApiToken,
		type ApiTokensResponse,
		type Environment,
		type Project,
	} from '$lib/backend';
	import { formatDate, formatDateTime } from '$lib/format';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { t } from '$lib/i18n/index.svelte';

	type ProjectContext = {
		readonly project: Project | null;
		readonly environments: Environment[];
		readonly canWrite: boolean;
		readonly canAdmin: boolean;
		readonly slug: string;
		reloadEnvironments: () => Promise<void>;
		reloadProject: () => Promise<void>;
	};

	const ctx = getContext<ProjectContext>('project');

	let tokens = $state<ApiToken[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	let newEnvironments = $state('*');

	const environmentHelper = $derived(
		ctx.environments.length > 0
			? t('tokens.environments.helperList', { slugs: ctx.environments.map((env) => env.slug).join(', ') })
			: t('tokens.environments.helperAll')
	);

	function isRevoked(token: ApiToken) {
		return Boolean(token.revoked_at);
	}

	function isExpired(token: ApiToken) {
		return !isRevoked(token) && new Date(token.expires_at).getTime() < Date.now();
	}

	async function loadTokens(silent = false) {
		if (!silent) {
			loading = true;
			loadError = '';
		}
		try {
			const res = await api.get<ApiTokensResponse>(`/projects/${ctx.slug}/tokens`);
			tokens = res.tokens ?? [];
		} catch (err) {
			const message = err instanceof Error ? err.message : t('tokens.error.load');
			if (silent) toast.danger(message);
			else loadError = message;
		} finally {
			loading = false;
		}
	}

	const scopes = $derived([
		{ value: 'read', label: t('tokens.permissions.read') },
		{ value: 'read,write', label: t('tokens.permissions.write') },
	]);

	const items = $derived(
		tokens.map(
			(token): TokenItem => ({
				id: String(token.id),
				name: token.environments && token.environments !== '*' ? `${token.name} (${token.environments})` : token.name,
				scope: token.permissions,
				lastUsed: formatDateTime(token.last_used_at) === '—' ? undefined : formatDateTime(token.last_used_at),
				expires: isExpired(token)
					? `${formatDate(token.expires_at)} (${t('tokens.status.expired')})`
					: formatDate(token.expires_at),
				revoked: isRevoked(token),
			}),
		),
	);

	async function handleCreate(draft: TokenDraft): Promise<string> {
		try {
			const token = await api.post<ApiToken>(`/projects/${ctx.slug}/tokens`, {
				name: draft.name,
				permissions: draft.scope ?? 'read',
				environments: newEnvironments.trim() || '*',
				expires_in_days: draft.days,
			});
			const secret = token.token ?? '';
			if (secret) toast.success(t('tokens.created', { name: token.name }));
			else toast.warning(t('tokens.createdNoValue', { name: token.name }));
			await loadTokens(true);
			return secret;
		} catch (err) {
			throw new Error(err instanceof Error ? err.message : t('tokens.error.create'));
		}
	}

	async function handleRevoke(item: TokenItem) {
		try {
			await api.del(`/projects/${ctx.slug}/tokens/${item.id}`);
			toast.success(t('tokens.revoked', { name: item.name }));
			await loadTokens(true);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('tokens.error.revoke'));
			throw err;
		}
	}

	$effect(() => {
		if (ctx.project && !ctx.canAdmin) {
			goto(`/projects/${ctx.slug}`, { replaceState: true });
		}
	});

	$effect(() => {
		ctx.slug;
		loadTokens();
	});
</script>

<svelte:head>
	<title>{t('tokens.head', { project: ctx.project?.name ?? t('project.fallbackName') })}</title>
</svelte:head>

{#if ctx.canAdmin}
	<div class="flex flex-col gap-10 px-4 py-6 md:px-8">
		{#if loading}
			<Card class="flex flex-col gap-3">
				{#each Array(3) as _, index (index)}
					<Skeleton class="h-10 w-full" />
				{/each}
			</Card>
		{:else if loadError}
			<ErrorState message={loadError} onretry={() => loadTokens()} />
		{:else}
			<TokensSection
				title={t('project.tab.tokens')}
				description={t('tokens.description')}
				tokens={items}
				{scopes}
				expiryDays={[30, 90, 365, 730]}
				onCreate={handleCreate}
				onRevoke={handleRevoke}
			>
				<Field label={t('tokens.col.environments')} helper={environmentHelper}>
					<Input placeholder="*" bind:value={newEnvironments} class="font-fc-mono" />
				</Field>
			</TokensSection>
		{/if}
	</div>
{/if}
