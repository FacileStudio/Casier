<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import { api, COOKIE_SESSION, type Me } from '$lib/backend';
	import { nav } from '$lib/nav.svelte';
	import { Avatar, Spinner, Topbar } from '@facile/lib';
	import { DropdownMenu } from 'bits-ui';
	import Logo from '$lib/components/Logo.svelte';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import LogOutIcon from '@lucide/svelte/icons/log-out';

	let { children } = $props();

	let spaceSlug = $derived(page.params.slug ?? null);

	$effect(() => {
		if (auth.isAuthenticated) return;
		api.get<Me>('/auth/me')
			.then((me) => auth.login(COOKIE_SESSION, me.id, me.email))
			.catch(() => goto('/login'));
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
	<div class="min-h-dvh bg-fc-page">
		<Topbar class="px-4 sm:px-6">
			<nav class="flex min-w-0 items-center gap-2" aria-label="Breadcrumb">
				<a
					href="/spaces"
					class="shrink-0 rounded-fc-sm focus-visible:outline-2 focus-visible:outline-fc-accent"
				>
					<Logo size={26} wordmark />
				</a>
				{#if spaceSlug}
					<ChevronRightIcon class="size-3.5 shrink-0 text-fc-fg-muted" aria-hidden="true" />
					<span class="truncate text-fc-md font-medium text-fc-fg">
						{nav.spaceName ?? spaceSlug}
					</span>
				{/if}
			</nav>

			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="rounded-fc-pill transition-opacity hover:opacity-80 focus-visible:outline-2 focus-visible:outline-fc-accent"
					aria-label="Account menu"
				>
					<Avatar name={auth.email ?? '?'} size="sm" />
				</DropdownMenu.Trigger>
				<DropdownMenu.Portal>
					<DropdownMenu.Content
						align="end"
						sideOffset={8}
						class="z-50 min-w-52 rounded-fc-md border border-fc-fg/10 bg-fc-bg p-1.5 shadow-lg"
					>
						<div class="px-3 py-2">
							<p class="text-fc-xs uppercase tracking-wide text-fc-fg-muted">Signed in as</p>
							<p class="truncate text-fc-md text-fc-fg">{auth.email}</p>
						</div>
						<DropdownMenu.Separator class="my-1 h-px bg-fc-fg/10" />
						<DropdownMenu.Item
							class="flex cursor-pointer items-center gap-2 rounded-fc-sm px-3 py-2 text-fc-md text-fc-danger outline-none data-highlighted:bg-fc-surface"
							onSelect={handleLogout}
						>
							<LogOutIcon class="size-4" aria-hidden="true" />
							Log out
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Portal>
			</DropdownMenu.Root>
		</Topbar>

		<main class="mx-auto w-full max-w-fc-xl px-4 py-6 sm:px-6 sm:py-8">
			{@render children?.()}
		</main>
	</div>
{:else}
	<div class="flex min-h-dvh items-center justify-center bg-fc-page">
		<Spinner size="md" />
	</div>
{/if}
