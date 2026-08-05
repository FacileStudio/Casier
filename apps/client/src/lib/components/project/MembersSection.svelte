<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { api, type Member, type MembersResponse } from '$lib/backend';
	import { Avatar, Badge, Button, Field, Input, Select, Skeleton, Spinner } from '@facile/lib';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { toast } from 'svelte-sonner';

	let { projectSlug, canAdmin }: { projectSlug: string; canAdmin: boolean } = $props();

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
		admin: 'neutral',
		member: 'neutral',
		viewer: 'neutral',
	} as const;

	async function loadMembers() {
		loading = true;
		loadError = '';
		try {
			const res = await api.get<MembersResponse>(`/projects/${projectSlug}/members`, auth.token);
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
				`/projects/${projectSlug}/members`,
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
			await api.del(`/projects/${projectSlug}/members/${removeTarget.id}`, auth.token);
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
		projectSlug;
		loadMembers();
	});
</script>

<section class="flex flex-col gap-4" aria-label="Members">
	{#if canAdmin}
		<div class="flex justify-end">
			<Button class="h-9" onclick={openAdd}>
				<iconify-icon icon="solar:add-circle-linear" width="16" aria-hidden="true"></iconify-icon>
				Add member
			</Button>
		</div>
	{/if}

	{#if loading}
		<div class="flex flex-col gap-2 rounded-lg border border-border bg-card p-4">
			{#each Array(3) as _, i (i)}
				<Skeleton class="h-12 w-full" />
			{/each}
		</div>
	{:else if loadError}
		<ErrorState message={loadError} onretry={loadMembers} />
	{:else if members.length === 0}
		<EmptyState title="No members" description="Invite teammates to share this project's secrets.">
			{#snippet icon()}
				<iconify-icon icon="solar:users-group-rounded-linear" width="32" aria-hidden="true"></iconify-icon>
			{/snippet}
		</EmptyState>
	{:else}
		<ul class="flex flex-col divide-y divide-border rounded-lg border border-border bg-card px-4 py-2">
			{#each members as member (member.id)}
				<li class="flex items-center justify-between gap-3 py-3">
					<div class="flex min-w-0 items-center gap-3">
						<Avatar name={member.email} size="sm" />
						<div class="min-w-0">
							<p class="truncate text-sm text-foreground">
								{member.email}
								{#if member.email === auth.email}
									<span class="text-xs text-muted-foreground">(you)</span>
								{/if}
							</p>
						</div>
					</div>
					<div class="flex shrink-0 items-center gap-2">
						<Badge tone={roleTones[member.role] ?? 'neutral'} class="text-xs">{member.role}</Badge>
						{#if canAdmin && member.email !== auth.email && member.role !== 'owner'}
							<button
								type="button"
								class="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
								aria-label={`Remove ${member.email}`}
								onclick={() => openRemove(member)}
							>
								<iconify-icon icon="solar:trash-bin-2-linear" width="16" aria-hidden="true"></iconify-icon>
							</button>
						{/if}
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</section>

<Drawer bind:open={addOpen} title="Add member">
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
			<Button type="button" variant="ghost" class="h-9" disabled={adding} onclick={() => (addOpen = false)}>
				Cancel
			</Button>
			<Button type="submit" class="h-9" disabled={adding || !newEmail.trim()}>
				{#if adding}<Spinner size="sm" />{/if}
				{adding ? 'Adding…' : 'Add member'}
			</Button>
		</div>
	</form>
</Drawer>

<ConfirmDialog
	bind:open={removeOpen}
	title="Remove member"
	message={`This removes ${removeTarget?.email} from the project. They immediately lose access to every secret in every environment, and any session they have open stops working.`}
	confirmLabel="Remove member"
	loading={removing}
	onconfirm={handleRemove}
/>
