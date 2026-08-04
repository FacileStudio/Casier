<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api } from '$lib/backend';
	import { toast } from 'svelte-sonner';

	const inputClass =
		'flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50';

	let token = $state(page.url.searchParams.get('token') ?? '');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let loading = $state(false);

	let passwordMismatch = $derived(confirmPassword.length > 0 && password !== confirmPassword);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';

		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}

		loading = true;
		try {
			await api.post('/auth/reset-password', { token, password });
			toast.success('Password updated. You can sign in now.');
			goto('/login');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Reset failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Set a new password — Casier</title>
</svelte:head>

<div class="mb-8">
	<h1 class="font-heading text-2xl font-bold tracking-tight text-foreground">Set a new password</h1>
	<p class="mt-1.5 text-sm text-muted-foreground">
		Enter your reset token and choose a new password.
	</p>
</div>

<form onsubmit={handleSubmit} class="space-y-4">
	<div class="space-y-1.5">
		<label for="token" class="text-sm font-medium leading-none">Reset token</label>
		<input
			id="token"
			name="token"
			type="text"
			bind:value={token}
			placeholder="Paste your reset token"
			required
			autocomplete="off"
			class={inputClass}
		/>
	</div>

	<div class="space-y-1.5">
		<label for="password" class="text-sm font-medium leading-none">New password</label>
		<input
			id="password"
			name="password"
			type="password"
			bind:value={password}
			placeholder="••••••••"
			required
			minlength={8}
			autocomplete="new-password"
			class={inputClass}
		/>
		<p class="text-xs text-muted-foreground">At least 8 characters.</p>
	</div>

	<div class="space-y-1.5">
		<label for="confirm-password" class="text-sm font-medium leading-none">Confirm password</label>
		<input
			id="confirm-password"
			name="confirm-password"
			type="password"
			bind:value={confirmPassword}
			placeholder="••••••••"
			required
			minlength={8}
			autocomplete="new-password"
			class={inputClass}
		/>
		{#if passwordMismatch}
			<p class="text-sm text-destructive">Passwords do not match</p>
		{/if}
	</div>

	{#if error}
		<p class="text-sm text-destructive">{error}</p>
	{/if}

	<button
		type="submit"
		disabled={loading || passwordMismatch}
		class="inline-flex h-10 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50"
	>
		{loading ? 'Resetting…' : 'Reset password'}
	</button>
</form>

<p class="mt-6 text-center text-sm text-muted-foreground">
	<a href="/login" class="underline underline-offset-2 transition-colors hover:text-foreground">
		Back to log in
	</a>
</p>
