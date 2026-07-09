<script lang="ts">
	import { api } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';

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

<div class="flex min-h-screen items-center justify-center px-4">
	<Card class="w-full max-w-sm">
		<CardHeader class="text-center">
			<div class="mx-auto mb-2 flex h-10 w-10 items-center justify-center rounded-lg bg-primary text-primary-foreground">
				<KeyRoundIcon class="size-5" />
			</div>
			<CardTitle class="text-xl">Reset password</CardTitle>
			<CardDescription>Enter your email to receive a reset link</CardDescription>
		</CardHeader>
		<CardContent>
			{#if success}
				<div class="rounded-lg bg-green-500/10 px-3 py-2 text-sm text-green-500">
					If an account exists with that email, a reset token has been generated. Check server logs for the token.
				</div>
				<p class="mt-4 text-center text-sm text-muted-foreground">
					<a href="/login" class="text-primary underline underline-offset-4 hover:text-primary/80">Back to login</a>
				</p>
			{:else}
				<form onsubmit={handleSubmit} class="flex flex-col gap-4">
					{#if error}
						<div class="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</div>
					{/if}

					<div class="flex flex-col gap-1.5">
						<Label for="email">Email</Label>
						<Input id="email" type="email" placeholder="you@example.com" bind:value={email} required autocomplete="email" />
					</div>

					<Button type="submit" class="w-full" disabled={loading}>
						{loading ? 'Sending...' : 'Send reset token'}
					</Button>

					<p class="text-center text-sm text-muted-foreground">
						<a href="/login" class="text-primary underline underline-offset-4 hover:text-primary/80">Back to login</a>
					</p>
				</form>
			{/if}
		</CardContent>
	</Card>
</div>
