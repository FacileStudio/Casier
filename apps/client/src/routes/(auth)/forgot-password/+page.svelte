<script lang="ts">
	import { api } from '$lib/backend';

	const inputClass =
		'flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50';
	const labelClass = 'text-sm font-medium leading-none';
	const primaryButtonClass =
		'inline-flex h-10 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50';
	const outlineButtonClass =
		'inline-flex h-10 w-full items-center justify-center rounded-md border border-border bg-background px-4 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50';

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
	<title>Reset password — Casier</title>
</svelte:head>

<div class="mb-8">
	<h1 class="font-heading text-2xl font-bold tracking-tight text-foreground">Reset password</h1>
	<p class="mt-1.5 text-sm text-muted-foreground">
		We'll generate a reset token for your account.
	</p>
</div>

{#if success}
	<p role="status" class="rounded-md bg-muted px-3 py-2 text-sm text-muted-foreground">
		If an account exists with that email, a reset token has been generated. Ask your administrator
		for the token, then set a new password.
	</p>

	<a href="/reset-password" class="mt-4 {outlineButtonClass}">I have a token</a>
{:else}
	<form onsubmit={handleSubmit} class="space-y-4">
		<div class="space-y-1.5">
			<label for="email" class={labelClass}>Email</label>
			<input
				id="email"
				name="email"
				type="email"
				bind:value={email}
				placeholder="you@example.com"
				required
				autocomplete="email"
				disabled={loading}
				class={inputClass}
			/>
		</div>

		{#if error}
			<p role="alert" class="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
				{error}
			</p>
		{/if}

		<button type="submit" disabled={loading} aria-busy={loading} class={primaryButtonClass}>
			{loading ? 'Sending…' : 'Send reset token'}
		</button>
	</form>
{/if}

<p class="mt-6 text-center text-sm text-muted-foreground">
	<a href="/login" class="underline underline-offset-2 transition-colors hover:text-foreground">
		Back to log in
	</a>
</p>
