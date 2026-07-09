<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api } from '$lib/backend';

	$effect(() => {
		if (auth.isAuthenticated) {
			goto('/spaces');
			return;
		}

		api.get<{ id: string; email: string }>('/auth/me')
			.then((me) => {
				auth.login('cookie', me.id, me.email);
				goto('/spaces', { replaceState: true });
			})
			.catch(() => {
				goto('/login', { replaceState: true });
			});
	});
</script>
