<script lang="ts">
	import { getContext } from 'svelte';
	import {
		Avatar,
		Badge,
		Button,
		Card,
		EmptyState,
		Field,
		IconButton,
		Input,
		Select,
		SettingsSection,
		Skeleton,
		Spinner,
		Table,
		icons,
		toast,
	} from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		type Environment,
		type Member,
		type MembersResponse,
		type Project,
		type ProjectRole,
	} from '$lib/backend';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { displayName } from '$lib/format';
	import Drawer from '$lib/components/Drawer.svelte';
	import { t } from '$lib/i18n/index.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';

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

	const roleTones: Record<ProjectRole, 'owner' | 'admin' | 'neutral'> = {
		owner: 'owner',
		admin: 'admin',
		member: 'neutral',
		viewer: 'neutral',
	};


	let members = $state<Member[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	let addOpen = $state(false);
	let newEmail = $state('');
	let newRole = $state('member');
	let adding = $state(false);

	let removeTarget = $state<Member | null>(null);
	let removeOpen = $state(false);
	let removing = $state(false);

	async function loadMembers() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<MembersResponse>(`/projects/${ctx.slug}/members`);
			members = res.members ?? [];
		} catch (err) {
			loadError = err instanceof Error ? err.message : t('members.error.load');
		} finally {
			loading = false;
		}
	}

	function openAdd() {
		newEmail = '';
		newRole = 'member';
		addOpen = true;
	}

	async function handleAdd(event: Event) {
		event.preventDefault();
		if (!newEmail.trim()) return;
		adding = true;
		try {
			await api.post(
				`/projects/${ctx.slug}/members`,
				{ email: newEmail.trim(), role: newRole },
			);
			addOpen = false;
			toast.success(t('members.added', { email: newEmail.trim(), role: t(`role.${newRole as ProjectRole}`) }));
			await loadMembers();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('members.error.add'));
		} finally {
			adding = false;
		}
	}

	function openRemove(member: Member) {
		removeTarget = member;
		removeOpen = true;
	}

	async function handleRemove() {
		if (!removeTarget) return;
		removing = true;
		try {
			await api.del(`/projects/${ctx.slug}/members/${removeTarget.id}`);
			members = members.filter((member) => member.id !== removeTarget?.id);
			toast.success(t('members.removed', { email: removeTarget.email }));
			removeOpen = false;
			removeTarget = null;
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('members.error.remove'));
		} finally {
			removing = false;
		}
	}

	$effect(() => {
		ctx.slug;
		loadMembers();
	});
</script>

<svelte:head>
	<title>{t('members.head', { project: ctx.project?.name ?? t('project.fallbackName') })}</title>
</svelte:head>

<div class="flex flex-col gap-10 px-4 py-6 md:px-8">
	<SettingsSection
		title={t('members.title')}
		description={t('members.description')}
		bare
	>
		{#snippet actions()}
			{#if ctx.canAdmin}
				<Button icon={icons.plus} onclick={openAdd}>{t('members.add')}</Button>
			{/if}
		{/snippet}

		{#if loading}
			<Card class="flex flex-col gap-3">
				{#each Array(3) as _, index (index)}
					<Skeleton class="h-12 w-full" />
				{/each}
			</Card>
		{:else if loadError}
			<ErrorState message={loadError} onretry={loadMembers} />
		{:else if members.length === 0}
			<EmptyState
				icon={icons.usersGroup}
				title={t('members.empty.title')}
				description={t('members.empty.description')}
			>
				{#if ctx.canAdmin}
					<Button icon={icons.plus} onclick={openAdd}>{t('members.add')}</Button>
				{/if}
			</EmptyState>
		{:else}
			<Table>
				<thead>
					<tr>
						<th>{t('members.col.person')}</th>
						<th>{t('members.col.role')}</th>
						<th class="w-1 text-right" aria-label={t('common.actions')}></th>
					</tr>
				</thead>
				<tbody>
					{#each members as member (member.id)}
						<tr>
							<td>
								<div class="flex min-w-0 items-center gap-3">
									<Avatar src={member.avatar_url || undefined} name={displayName(member)} alt="" size="sm" />
									<div class="flex min-w-0 flex-col">
										<span class="truncate font-medium text-fc-fg">{displayName(member)}</span>
										<span class="truncate text-fc-xs text-fc-fg-muted">
											{member.email}{member.email === auth.email ? ` · ${t('members.you')}` : ''}
										</span>
									</div>
								</div>
							</td>
							<td>
								<div class="flex min-w-0 flex-col gap-1">
									<Badge tone={roleTones[member.role] ?? 'neutral'}>{t(`role.${member.role}`)}</Badge>
									<span class="text-fc-xs text-fc-fg-muted">
										{t(`members.role.${member.role}`)}
									</span>
								</div>
							</td>
							<td>
								<div class="flex justify-end">
									{#if ctx.canAdmin && member.email !== auth.email && member.role !== 'owner'}
										<IconButton
											variant="danger"
											aria-label={t('members.remove.aria', { email: member.email })}
											onclick={() => openRemove(member)}
										>
											<iconify-icon icon={icons.remove} width="18" aria-hidden="true"
											></iconify-icon>
										</IconButton>
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		{/if}
	</SettingsSection>
</div>

<Drawer bind:open={addOpen} title={t('members.add')}>
	<form onsubmit={handleAdd} class="flex flex-col gap-4">
		<Field label={t('common.email')} helper={t('members.emailHelper')}>
			<Input
				type="email"
				placeholder="teammate@example.com"
				bind:value={newEmail}
				required
				disabled={adding}
			/>
		</Field>
		<Field label={t('members.col.role')} helper={t('members.roleHelper')}>
			<Select bind:value={newRole} disabled={adding}>
				<option value="admin">{t('members.option.admin')}</option>
				<option value="member">{t('members.option.member')}</option>
				<option value="viewer">{t('members.option.viewer')}</option>
			</Select>
		</Field>
		<div class="flex justify-end gap-2 pt-1">
			<Button
				type="button"
				variant="ghost"
				size="lg"
				disabled={adding}
				onclick={() => (addOpen = false)}
			>
				{t('common.cancel')}
			</Button>
			<Button type="submit" size="lg" disabled={adding || !newEmail.trim()}>
				{#if adding}<Spinner label={t('common.loading')} size="sm" />{/if}
				{adding ? t('common.adding') : t('members.add')}
			</Button>
		</div>
	</form>
</Drawer>

<ConfirmDialog
	bind:open={removeOpen}
	title={t('members.remove.title')}
	message={t('members.remove.message', { email: removeTarget?.email ?? '' })}
	confirmLabel={t('members.remove.title')}
	loading={removing}
	onconfirm={handleRemove}
/>
