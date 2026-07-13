<script lang="ts">
	import { page } from '$app/stores';
	import { auth } from '$lib/auth.svelte';
	import { api, type Space, type Secret, type Member } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import * as Table from '$lib/components/ui/table';
	import { toast } from 'svelte-sonner';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import TrashIcon from '@lucide/svelte/icons/trash-2';
	import UserPlusIcon from '@lucide/svelte/icons/user-plus';
	import XIcon from '@lucide/svelte/icons/x';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';

	const DEFAULT_ENVS = ['development', 'staging', 'production'];

	let space = $state<Space | null>(null);
	let secrets = $state<Secret[]>([]);
	let members = $state<Member[]>([]);
	let loading = $state(true);
	let selectedEnv = $state('development');
	let revealedKeys = $state<Set<string>>(new Set());

	let showAddSecret = $state(false);
	let newKey = $state('');
	let newValue = $state('');
	let addingSecret = $state(false);

	let editingId = $state<string | null>(null);
	let editValue = $state('');

	let showAddMember = $state(false);
	let newMemberEmail = $state('');
	let addingMember = $state(false);

	let slug = $derived($page.params.slug);
	let envSecrets = $derived(secrets.filter((s) => s.environment === selectedEnv));

	async function loadSpace() {
		loading = true;
		try {
			space = await api.get<Space>(`/spaces/${slug}`, auth.token);
			await Promise.all([loadSecrets(), loadMembers()]);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to load space');
		} finally {
			loading = false;
		}
	}

	async function loadSecrets() {
		try {
			secrets = await api.get<Secret[]>(`/spaces/${slug}/secrets`, auth.token);
		} catch {
			secrets = [];
		}
	}

	async function loadMembers() {
		try {
			members = await api.get<Member[]>(`/spaces/${slug}/members`, auth.token);
		} catch {
			members = [];
		}
	}

	function toggleReveal(id: string) {
		const next = new Set(revealedKeys);
		if (next.has(id)) {
			next.delete(id);
		} else {
			next.add(id);
		}
		revealedKeys = next;
	}

	function maskValue(value: string): string {
		if (value.length <= 4) return '*'.repeat(value.length || 8);
		return value.slice(0, 2) + '*'.repeat(Math.min(value.length - 4, 16)) + value.slice(-2);
	}

	async function handleAddSecret(e: Event) {
		e.preventDefault();
		addingSecret = true;
		try {
			const secret = await api.post<Secret>(`/spaces/${slug}/secrets`, {
				environment: selectedEnv,
				key: newKey,
				value: newValue,
			}, auth.token);
			secrets = [...secrets, secret];
			showAddSecret = false;
			newKey = '';
			newValue = '';
			toast.success('Secret added');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to add secret');
		} finally {
			addingSecret = false;
		}
	}

	function startEdit(secret: Secret) {
		editingId = secret.id;
		editValue = secret.value;
	}

	async function saveEdit(secret: Secret) {
		try {
			await api.put(`/spaces/${slug}/secrets/${secret.id}`, {
				value: editValue,
			}, auth.token);
			secrets = secrets.map((s) =>
				s.id === secret.id ? { ...s, value: editValue } : s
			);
			editingId = null;
			toast.success('Secret updated');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to update secret');
		}
	}

	async function deleteSecret(id: string) {
		try {
			await api.del(`/spaces/${slug}/secrets/${id}`, auth.token);
			secrets = secrets.filter((s) => s.id !== id);
			toast.success('Secret deleted');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to delete secret');
		}
	}

	async function handleAddMember(e: Event) {
		e.preventDefault();
		addingMember = true;
		try {
			const member = await api.post<Member>(`/spaces/${slug}/members`, {
				email: newMemberEmail,
			}, auth.token);
			members = [...members, member];
			showAddMember = false;
			newMemberEmail = '';
			toast.success('Member added');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to add member');
		} finally {
			addingMember = false;
		}
	}

	async function removeMember(id: string) {
		try {
			await api.del(`/spaces/${slug}/members/${id}`, auth.token);
			members = members.filter((m) => m.id !== id);
			toast.success('Member removed');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to remove member');
		}
	}

	$effect(() => {
		slug;
		loadSpace();
	});
</script>

<div class="p-6">
	{#if loading}
		<div class="flex items-center justify-center py-12 text-muted-foreground">
			Loading...
		</div>
	{:else if space}
		<div class="mb-6">
			<Button variant="ghost" size="sm" href="/spaces" class="mb-3 gap-1">
				<ArrowLeftIcon class="size-3.5" />
				Back
			</Button>
			<h1 class="text-xl font-semibold">{space.name}</h1>
			{#if space.description}
				<p class="text-sm text-muted-foreground">{space.description}</p>
			{/if}
		</div>

		<div class="mb-4 flex gap-1">
			{#each DEFAULT_ENVS as env (env)}
				<Button
					variant={selectedEnv === env ? 'default' : 'outline'}
					size="sm"
					onclick={() => (selectedEnv = env)}
				>
					{env}
				</Button>
			{/each}
		</div>

		<Card class="mb-6">
			<CardHeader>
				<div class="flex items-center justify-between">
					<CardTitle class="flex items-center gap-2">
						Secrets
						<Badge variant="secondary">{envSecrets.length}</Badge>
					</CardTitle>
					<Button size="sm" onclick={() => (showAddSecret = true)}>
						<PlusIcon class="size-3.5" />
						Add secret
					</Button>
				</div>
			</CardHeader>
			<CardContent>
				{#if showAddSecret}
					<form onsubmit={handleAddSecret} class="mb-4 flex items-end gap-2 rounded-lg border border-border p-3">
						<div class="flex flex-1 flex-col gap-1.5">
							<Label for="new-key">Key</Label>
							<Input id="new-key" placeholder="SECRET_KEY" bind:value={newKey} required />
						</div>
						<div class="flex flex-1 flex-col gap-1.5">
							<Label for="new-value">Value</Label>
							<Input id="new-value" placeholder="secret-value" bind:value={newValue} required />
						</div>
						<Button type="submit" size="sm" disabled={addingSecret}>
							{addingSecret ? 'Adding...' : 'Add'}
						</Button>
						<Button variant="ghost" size="icon" onclick={() => (showAddSecret = false)}>
							<XIcon class="size-4" />
						</Button>
					</form>
				{/if}

				{#if envSecrets.length === 0 && !showAddSecret}
					<p class="py-6 text-center text-sm text-muted-foreground">
						No secrets in {selectedEnv}
					</p>
				{:else if envSecrets.length > 0}
					<Table.Table>
						<Table.TableHeader>
							<Table.TableRow>
								<Table.TableHead>Key</Table.TableHead>
								<Table.TableHead>Value</Table.TableHead>
								<Table.TableHead class="w-[120px] text-right">Actions</Table.TableHead>
							</Table.TableRow>
						</Table.TableHeader>
						<Table.TableBody>
							{#each envSecrets as secret (secret.id)}
								<Table.TableRow>
									<Table.TableCell class="font-mono text-xs">{secret.key}</Table.TableCell>
									<Table.TableCell class="font-mono text-xs">
										{#if editingId === secret.id}
											<div class="flex items-center gap-1">
												<Input
													class="h-7 text-xs font-mono"
													bind:value={editValue}
												/>
												<Button variant="ghost" size="icon" onclick={() => saveEdit(secret)}>
													<CheckIcon class="size-3.5" />
												</Button>
												<Button variant="ghost" size="icon" onclick={() => (editingId = null)}>
													<XIcon class="size-3.5" />
												</Button>
											</div>
										{:else if revealedKeys.has(secret.id)}
											{secret.value}
										{:else}
											{maskValue(secret.value)}
										{/if}
									</Table.TableCell>
									<Table.TableCell class="text-right">
										{#if editingId !== secret.id}
											<div class="flex items-center justify-end gap-0.5">
												<Button variant="ghost" size="icon" onclick={() => toggleReveal(secret.id)}>
													{#if revealedKeys.has(secret.id)}
														<EyeOffIcon class="size-3.5" />
													{:else}
														<EyeIcon class="size-3.5" />
													{/if}
												</Button>
												<Button variant="ghost" size="icon" onclick={() => startEdit(secret)}>
													<PencilIcon class="size-3.5" />
												</Button>
												<Button variant="ghost" size="icon" onclick={() => deleteSecret(secret.id)}>
													<TrashIcon class="size-3.5 text-destructive" />
												</Button>
											</div>
										{/if}
									</Table.TableCell>
								</Table.TableRow>
							{/each}
						</Table.TableBody>
					</Table.Table>
				{/if}
			</CardContent>
		</Card>

		<Separator class="my-6" />

		<Card>
			<CardHeader>
				<div class="flex items-center justify-between">
					<CardTitle class="flex items-center gap-2">
						Members
						<Badge variant="secondary">{members.length}</Badge>
					</CardTitle>
					<Button size="sm" onclick={() => (showAddMember = true)}>
						<UserPlusIcon class="size-3.5" />
						Add member
					</Button>
				</div>
			</CardHeader>
			<CardContent>
				{#if showAddMember}
					<form onsubmit={handleAddMember} class="mb-4 flex items-end gap-2 rounded-lg border border-border p-3">
						<div class="flex flex-1 flex-col gap-1.5">
							<Label for="member-email">Email</Label>
							<Input id="member-email" type="email" placeholder="teammate@example.com" bind:value={newMemberEmail} required />
						</div>
						<Button type="submit" size="sm" disabled={addingMember}>
							{addingMember ? 'Adding...' : 'Add'}
						</Button>
						<Button variant="ghost" size="icon" onclick={() => (showAddMember = false)}>
							<XIcon class="size-4" />
						</Button>
					</form>
				{/if}

				{#if members.length === 0}
					<p class="py-6 text-center text-sm text-muted-foreground">No members</p>
				{:else}
					<div class="flex flex-col gap-2">
						{#each members as member (member.id)}
							<div class="flex items-center justify-between rounded-lg px-3 py-2 hover:bg-muted/50">
								<div class="flex items-center gap-3">
									<div class="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-xs font-medium uppercase">
										{member.email.slice(0, 2)}
									</div>
									<div>
										<p class="text-sm">{member.email}</p>
										<p class="text-xs text-muted-foreground">{member.role}</p>
									</div>
								</div>
								{#if member.email !== auth.email}
									<Button variant="ghost" size="icon" onclick={() => removeMember(member.id)}>
										<TrashIcon class="size-3.5 text-destructive" />
									</Button>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</CardContent>
		</Card>
	{/if}
</div>
