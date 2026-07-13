<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { api, type Space } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from 'svelte-sonner';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import FolderLockIcon from '@lucide/svelte/icons/folder-lock';
	import UsersIcon from '@lucide/svelte/icons/users';
	import XIcon from '@lucide/svelte/icons/x';

	let spaces = $state<Space[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);
	let createName = $state('');
	let createDescription = $state('');
	let creating = $state(false);

	async function loadSpaces() {
		loading = true;
		try {
			spaces = await api.get<Space[]>('/spaces', auth.token);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to load spaces');
		} finally {
			loading = false;
		}
	}

	async function handleCreate(e: Event) {
		e.preventDefault();
		creating = true;
		try {
			const space = await api.post<Space>('/spaces', {
				name: createName,
				description: createDescription,
			}, auth.token);
			spaces = [...spaces, space];
			showCreate = false;
			createName = '';
			createDescription = '';
			toast.success('Space created');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to create space');
		} finally {
			creating = false;
		}
	}

	$effect(() => {
		loadSpaces();
	});
</script>

<div class="p-6">
	<div class="mb-6 flex items-center justify-between">
		<div>
			<h1 class="text-xl font-semibold">Spaces</h1>
			<p class="text-sm text-muted-foreground">Manage your secret spaces</p>
		</div>
		<Button onclick={() => (showCreate = true)}>
			<PlusIcon class="size-4" />
			New space
		</Button>
	</div>

	{#if showCreate}
		<Card class="mb-6">
			<CardHeader>
				<div class="flex items-center justify-between">
					<CardTitle>Create space</CardTitle>
					<Button variant="ghost" size="icon" onclick={() => (showCreate = false)}>
						<XIcon class="size-4" />
					</Button>
				</div>
			</CardHeader>
			<CardContent>
				<form onsubmit={handleCreate} class="flex flex-col gap-4">
					<div class="flex flex-col gap-1.5">
						<Label for="space-name">Name</Label>
						<Input id="space-name" placeholder="my-project" bind:value={createName} required />
					</div>
					<div class="flex flex-col gap-1.5">
						<Label for="space-desc">Description</Label>
						<Input id="space-desc" placeholder="Optional description" bind:value={createDescription} />
					</div>
					<div class="flex justify-end">
						<Button type="submit" disabled={creating}>
							{creating ? 'Creating...' : 'Create'}
						</Button>
					</div>
				</form>
			</CardContent>
		</Card>
	{/if}

	{#if loading}
		<div class="flex items-center justify-center py-12 text-muted-foreground">
			Loading spaces...
		</div>
	{:else if spaces.length === 0}
		<div class="flex flex-col items-center justify-center py-12 text-muted-foreground">
			<FolderLockIcon class="mb-3 size-10 opacity-40" />
			<p class="text-sm">No spaces yet</p>
			<p class="text-xs">Create a space to start managing secrets</p>
		</div>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each spaces as space (space.id)}
				<a href="/spaces/{space.slug}" class="block">
					<Card class="transition-colors hover:bg-muted/30">
						<CardHeader>
							<div class="flex items-center justify-between">
								<CardTitle>{space.name}</CardTitle>
								<Badge variant="secondary" class="gap-1">
									<UsersIcon class="size-3" />
									{space.member_count}
								</Badge>
							</div>
							{#if space.description}
								<CardDescription>{space.description}</CardDescription>
							{/if}
						</CardHeader>
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</div>
