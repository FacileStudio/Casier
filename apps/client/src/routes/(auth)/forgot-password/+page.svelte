<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/backend';
	import { Alert, Button, Field, Input, Spinner } from '@facile/lib';

	let email = $state('');
	let error = $state('');
	let success = $state(false);
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;

		try {
			await api.post('/auth/forgot-password', { email });
			success = true;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Request failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Casier — Forgot password</title>
</svelte:head>

<div class="flex flex-col gap-5">
	<div class="flex flex-col gap-1 text-center">
		<h1 class="text-fc-xl font-semibold text-fc-fg">Reset password</h1>
		<p class="text-fc-md text-fc-fg-muted">Enter your email to receive a reset token</p>
	</div>

	{#if success}
		<Alert tone="success">
			If an account exists with that email, a reset token has been generated. Ask your
			administrator for the token, then set a new password.
		</Alert>
		<div class="flex flex-col gap-3">
			<Button class="w-full" onclick={() => goto('/reset-password')}>I have a token</Button>
			<p class="text-center text-fc-sm text-fc-fg-muted">
				<a href="/login" class="text-fc-accent hover:underline">Back to login</a>
			</p>
		</div>
	{:else}
		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			{#if error}
				<Alert tone="danger">{error}</Alert>
			{/if}

			<Field label="Email">
				<Input
					type="email"
					name="email"
					placeholder="you@example.com"
					bind:value={email}
					required
					autocomplete="email"
				/>
			</Field>

			<Button type="submit" class="w-full" disabled={loading}>
				{#if loading}<Spinner size="sm" />{/if}
				{loading ? 'Sending…' : 'Send reset token'}
			</Button>

			<p class="text-center text-fc-sm text-fc-fg-muted">
				<a href="/login" class="text-fc-accent hover:underline">Back to login</a>
			</p>
		</form>
	{/if}
</div>
