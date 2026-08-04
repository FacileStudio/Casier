<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, COOKIE_SESSION, type Me } from '$lib/backend';
	import { Spinner } from '@facile/lib';
	import Logo from '$lib/components/Logo.svelte';

	$effect(() => {
		if (auth.isAuthenticated) {
			goto('/spaces');
			return;
		}

		api.get<Me>('/auth/me')
			.then((me) => {
				auth.login(COOKIE_SESSION, me.id, me.email);
				goto('/spaces', { replaceState: true });
			})
			.catch(() => {
				goto('/login', { replaceState: true });
			});
	});
</script>

<svelte:head>
	<title>Casier</title>
</svelte:head>

<div class="flex min-h-dvh flex-col items-center justify-center gap-4 bg-fc-page">
	<Logo size={40} />
	<Spinner size="md" />
</div>
