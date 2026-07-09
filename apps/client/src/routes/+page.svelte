<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import { api } from '$lib/backend';

	$effect(() => {
		const tokenParam = page.url.searchParams.get('token');
		if (tokenParam) {
			auth.token = tokenParam;
			api.get<{ id: string; email: string }>('/auth/me', tokenParam)
				.then((me) => {
					auth.login(tokenParam, me.id, me.email);
					goto('/spaces', { replaceState: true });
				})
				.catch(() => {
					auth.logout();
					goto('/login', { replaceState: true });
				});
			return;
		}

		if (auth.isAuthenticated) {
			goto('/spaces');
		} else {
			goto('/login');
		}
	});
</script>
