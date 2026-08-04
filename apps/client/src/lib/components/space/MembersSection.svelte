<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { api, type Member, type MembersResponse } from '$lib/backend';
	import { Avatar, Badge, Button, Field, Input, Modal, Select, Skeleton, Spinner } from '@facile/lib';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { toast } from 'svelte-sonner';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import UserPlusIcon from '@lucide/svelte/icons/user-plus';
	import UsersIcon from '@lucide/svelte/icons/users';

	let { spaceSlug, canAdmin }: { spaceSlug: string; canAdmin: boolean } = $props();

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

	const roleTones = {
		owner: 'accent',
		admin: 'success',
		member: 'neutral',
		viewer: 'neutral',
	} as const;

	async function loadMembers() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<MembersResponse>(`/spaces/${spaceSlug}/members`, auth.token);
			members = res.members ?? [];
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load members';
		} finally {
			loading = false;
		}
	}

	function openAdd() {
		newEmail = '';
		newRole = 'member';
		addOpen = true;
	}

	async function handleAdd(e: Event) {
		e.preventDefault();
		if (!newEmail.trim()) return;
		adding = true;
		try {
			await api.post(
				`/spaces/${spaceSlug}/members`,
				{ email: newEmail.trim(), role: newRole },
				auth.token
			);
			addOpen = false;
			toast.success(`${newEmail.trim()} added as ${newRole}`);
			await loadMembers();
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to add member');
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
			await api.del(`/spaces/${spaceSlug}/members/${removeTarget.id}`, auth.token);
			members = members.filter((m) => m.id !== removeTarget!.id);
			toast.success(`${removeTarget.email} removed`);
			removeOpen = false;
			removeTarget = null;
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to remove member');
		} finally {
			removing = false;
		}
	}

	$effect(() => {
		spaceSlug;
		loadMembers();
	});
</script>

<section class="flex flex-col gap-4" aria-label="Members">
	{#if canAdmin}
		<div class="flex justify-end">
			<Button onclick={openAdd}>
				<UserPlusIcon class="size-4" aria-hidden="true" />
				Add member
			</Button>
		</div>
	{/if}

	{#if loading}
		<div class="flex flex-col gap-2 rounded-fc-md bg-fc-component px-6 py-5">
			{#each Array(3) as _, i (i)}
				<Skeleton class="h-12 w-full" />
			{/each}
		</div>
	{:else if loadError}
		<ErrorState message={loadError} onretry={loadMembers} />
	{:else if members.length === 0}
		<EmptyState title="No members" description="Invite teammates to share this space's secrets.">
			{#snippet icon()}
				<UsersIcon class="size-10" aria-hidden="true" />
			{/snippet}
		</EmptyState>
	{:else}
		<ul class="flex flex-col divide-y divide-fc-fg/7 rounded-fc-md bg-fc-component px-4 py-2">
			{#each members as member (member.id)}
				<li class="flex items-center justify-between gap-3 py-3">
					<div class="flex min-w-0 items-center gap-3">
						<Avatar name={member.email} size="sm" />
						<div class="min-w-0">
							<p class="truncate text-fc-md text-fc-fg">
								{member.email}
								{#if member.email === auth.email}
									<span class="text-fc-xs text-fc-fg-muted">(you)</span>
								{/if}
							</p>
						</div>
					</div>
					<div class="flex shrink-0 items-center gap-2">
						<Badge tone={roleTones[member.role] ?? 'neutral'}>{member.role}</Badge>
						{#if canAdmin && member.email !== auth.email && member.role !== 'owner'}
							<button
								type="button"
								class="inline-flex h-8 w-8 items-center justify-center rounded-fc-sm text-fc-fg-muted transition-colors hover:bg-fc-surface hover:text-fc-danger focus-visible:outline-2 focus-visible:outline-fc-accent"
								aria-label={`Remove ${member.email}`}
								onclick={() => openRemove(member)}
							>
								<Trash2Icon class="size-4" aria-hidden="true" />
							</button>
						{/if}
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</section>

<Modal bind:open={addOpen} title="Add member">
	<form onsubmit={handleAdd} class="flex flex-col gap-4">
		<Field label="Email" helper="Must be an existing Casier account">
			<Input
				type="email"
				placeholder="teammate@example.com"
				bind:value={newEmail}
				required
				disabled={adding}
			/>
		</Field>
		<Field label="Role">
			<Select bind:value={newRole} disabled={adding}>
				<option value="admin">Admin — manage members, tokens and environments</option>
				<option value="member">Member — read and write secrets</option>
				<option value="viewer">Viewer — read-only</option>
			</Select>
		</Field>
		<div class="flex justify-end gap-2 pt-1">
			<Button type="button" variant="ghost" disabled={adding} onclick={() => (addOpen = false)}>
				Cancel
			</Button>
			<Button type="submit" disabled={adding || !newEmail.trim()}>
				{#if adding}<Spinner size="sm" />{/if}
				{adding ? 'Adding…' : 'Add member'}
			</Button>
		</div>
	</form>
</Modal>

<ConfirmDialog
	bind:open={removeOpen}
	title="Remove member"
	message={`This removes ${removeTarget?.email} from the space. They will immediately lose access to all of its secrets.`}
	confirmLabel="Remove member"
	loading={removing}
	onconfirm={handleRemove}
/>
