<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import FolderLockIcon from '@lucide/svelte/icons/folder-lock';
	import LogOutIcon from '@lucide/svelte/icons/log-out';

	let { children } = $props();

	$effect(() => {
		if (!auth.isAuthenticated) {
			goto('/login');
		}
	});

	async function handleLogout() {
		try {
			await api.post('/auth/logout', {}, auth.token);
		} catch {}
		auth.logout();
		goto('/login');
	}
</script>

{#if auth.isAuthenticated}
	<div class="flex min-h-screen">
		<aside class="flex w-56 shrink-0 flex-col border-r border-border bg-sidebar p-4">
			<div class="flex items-center gap-2 px-1 mb-4">
				<div class="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground">
					<KeyRoundIcon class="size-3.5" />
				</div>
				<span class="text-sm font-semibold text-sidebar-foreground">Clef</span>
			</div>

			<nav class="flex flex-col gap-1">
				<Button variant="ghost" href="/spaces" class="justify-start gap-2">
					<FolderLockIcon class="size-4" />
					Spaces
				</Button>
			</nav>

			<div class="mt-auto flex flex-col gap-2">
				<Separator />
				<div class="flex items-center justify-between px-1">
					<span class="truncate text-xs text-muted-foreground">{auth.email}</span>
					<Button variant="ghost" size="icon" onclick={handleLogout}>
						<LogOutIcon class="size-4" />
					</Button>
				</div>
			</div>
		</aside>

		<main class="flex-1 overflow-auto">
			{@render children?.()}
		</main>
	</div>
{/if}
