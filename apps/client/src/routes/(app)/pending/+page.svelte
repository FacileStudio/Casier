<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Alert, Button, EmptyState, Spinner, icons, toast } from '@facile/muse';
	import { API_BASE, ApiError, api, type PendingAction, type PendingList, type PendingResult } from '$lib/backend';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PendingCard from '$lib/components/PendingCard.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let actions = $state<PendingAction[]>([]);
	let loading = $state(true);
	let busy = $state('');
	let signInAgain = $state(false);

	const focused = $derived(page.url.searchParams.get('action') ?? '');
	const ordered = $derived([
		...actions.filter((action) => action.id === focused),
		...actions.filter((action) => action.id !== focused)
	]);

	onMount(load);

	/** Reports whether the API refused this session for confirming, and shows the sign-in prompt if so. */
	function refusedSession(err: unknown): boolean {
		signInAgain = err instanceof ApiError && err.status === 403;
		return signInAgain;
	}

	/** Loads the open actions. */
	async function load() {
		loading = true;
		try {
			actions = (await api.get<PendingList>('/mcp/pending')).actions ?? [];
		} catch (err) {
			if (!refusedSession(err)) toast.danger(t('pending.loadFailed'));
		}
		loading = false;
	}

	/** Runs one action once, then drops it from the list. */
	async function confirm(action: PendingAction) {
		busy = action.id;
		try {
			const result = await api.post<PendingResult>(`/mcp/pending/${encodeURIComponent(action.id)}/confirm`, {});
			toast.success(t(result.status === 'approved' ? 'pending.approved' : 'pending.trashed'));
			actions = actions.filter((item) => item.id !== action.id);
		} catch (err) {
			if (!refusedSession(err)) {
				toast.danger(err instanceof Error ? err.message : t('pending.failed'));
				await load();
			}
		}
		busy = '';
	}

	/** Closes one action without running it. */
	async function dismiss(action: PendingAction) {
		busy = action.id;
		try {
			await api.post<PendingResult>(`/mcp/pending/${encodeURIComponent(action.id)}/dismiss`, {});
			toast.success(t('pending.dismissed'));
			actions = actions.filter((item) => item.id !== action.id);
		} catch (err) {
			if (!refusedSession(err)) toast.danger(err instanceof Error ? err.message : t('pending.failed'));
		}
		busy = '';
	}
</script>

<svelte:head>
	<title>{t('pending.title')}</title>
</svelte:head>

<PageHeader title={t('pending.heading')} description={t('pending.intro')} />

<div class="mx-auto flex w-full max-w-2xl flex-col gap-4 px-4 pb-8 md:px-8">
	{#if signInAgain}
		<Alert tone="warning" title={t('pending.signInAgain.title')}>
			<p>{t('pending.signInAgain.body')}</p>
			<Button href="{API_BASE}/auth/oidc" variant="primary" class="mt-3 w-full sm:w-auto [@media(pointer:coarse)]:h-11">
				{t('pending.signInAgain.action')}
			</Button>
		</Alert>
	{:else if loading}
		<div class="flex h-48 items-center justify-center"><Spinner /></div>
	{:else if actions.length === 0}
		<EmptyState icon={icons.check} title={t('pending.none')} description={t('pending.noneBody')} />
	{:else}
		{#each ordered as action (action.id)}
			<PendingCard {action} busy={busy === action.id} onconfirm={confirm} ondismiss={dismiss} />
		{/each}
	{/if}
</div>
