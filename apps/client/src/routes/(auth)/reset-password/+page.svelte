<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/backend';
	import { toast } from 'svelte-sonner';
	import { Alert, Button, Field, Input, Spinner } from '@facile/lib';

	let token = $state('');
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
	<title>Casier — Reset password</title>
</svelte:head>

<div class="flex flex-col gap-5">
	<div class="flex flex-col gap-1 text-center">
		<h1 class="text-fc-xl font-semibold text-fc-fg">Set new password</h1>
		<p class="text-fc-md text-fc-fg-muted">Enter your reset token and a new password</p>
	</div>

	<form onsubmit={handleSubmit} class="flex flex-col gap-4">
		{#if error}
			<Alert tone="danger">{error}</Alert>
		{/if}

		<Field label="Reset token">
			<Input
				type="text"
				name="token"
				placeholder="Paste your reset token"
				bind:value={token}
				required
				autocomplete="off"
			/>
		</Field>

		<Field label="New password" helper="At least 8 characters">
			<Input
				type="password"
				name="password"
				placeholder="Password"
				bind:value={password}
				required
				minlength={8}
				autocomplete="new-password"
			/>
		</Field>

		<Field label="Confirm password" error={passwordMismatch ? 'Passwords do not match' : undefined}>
			<Input
				type="password"
				name="confirm"
				placeholder="Confirm password"
				bind:value={confirmPassword}
				required
				minlength={8}
				autocomplete="new-password"
			/>
		</Field>

		<Button type="submit" class="w-full" disabled={loading || passwordMismatch}>
			{#if loading}<Spinner size="sm" />{/if}
			{loading ? 'Resetting…' : 'Reset password'}
		</Button>

		<p class="text-center text-fc-sm text-fc-fg-muted">
			<a href="/login" class="text-fc-accent hover:underline">Back to login</a>
		</p>
	</form>
</div>
