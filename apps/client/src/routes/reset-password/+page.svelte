<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';

	let token = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let loading = $state(false);

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
			goto('/login');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Reset failed';
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center px-4">
	<Card class="w-full max-w-sm">
		<CardHeader class="text-center">
			<div class="mx-auto mb-2 flex h-10 w-10 items-center justify-center rounded-lg bg-primary text-primary-foreground">
				<KeyRoundIcon class="size-5" />
			</div>
			<CardTitle class="text-xl">Set new password</CardTitle>
			<CardDescription>Enter your reset token and new password</CardDescription>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit} class="flex flex-col gap-4">
				{#if error}
					<div class="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</div>
				{/if}

				<div class="flex flex-col gap-1.5">
					<Label for="token">Reset token</Label>
					<Input id="token" type="text" placeholder="Paste your reset token" bind:value={token} required autocomplete="off" />
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="password">New password</Label>
					<Input id="password" type="password" placeholder="Password" bind:value={password} required minlength={8} autocomplete="new-password" />
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="confirm">Confirm password</Label>
					<Input id="confirm" type="password" placeholder="Confirm password" bind:value={confirmPassword} required minlength={8} autocomplete="new-password" />
				</div>

				<Button type="submit" class="w-full" disabled={loading}>
					{loading ? 'Resetting...' : 'Reset password'}
				</Button>

				<p class="text-center text-sm text-muted-foreground">
					<a href="/login" class="text-primary underline underline-offset-4 hover:text-primary/80">Back to login</a>
				</p>
			</form>
		</CardContent>
	</Card>
</div>
