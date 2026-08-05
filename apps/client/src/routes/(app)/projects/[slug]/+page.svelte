<script lang="ts">
	import { getContext } from 'svelte';
	import { auth } from '$lib/auth.svelte';
	import {
		api,
		type Environment,
		type EnvironmentRotation,
		type Project,
		type RotationSummary,
	} from '$lib/backend';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import SecretsSection from '$lib/components/project/SecretsSection.svelte';

	type ProjectContext = {
		readonly project: Project | null;
		readonly environments: Environment[];
		readonly canWrite: boolean;
		readonly canAdmin: boolean;
		readonly slug: string;
		reloadEnvironments: () => Promise<void>;
		reloadProject: () => Promise<void>;
	};

	const ctx = getContext<ProjectContext>('project');

	let selectedEnvSlug = $state<string | null>(null);

	let selectedEnv = $derived(
		ctx.environments.find((env) => env.slug === selectedEnvSlug) ?? null
	);

	let rotation = $state<Record<string, EnvironmentRotation>>({});

	function staleCount(envSlug: string): number {
		const entry = rotation[envSlug];
		if (!entry) return 0;
		return entry.overdue + entry.due_soon;
	}

	async function loadRotation() {
		try {
			const res = await api.get<RotationSummary>(
				`/projects/${ctx.slug}/secrets/rotation`,
				auth.token
			);
			rotation = Object.fromEntries((res.environments ?? []).map((env) => [env.slug, env]));
		} catch {
			rotation = {};
		}
	}

	$effect(() => {
		if (!ctx.environments.some((env) => env.slug === selectedEnvSlug)) {
			selectedEnvSlug = ctx.environments[0]?.slug ?? null;
		}
	});

	$effect(() => {
		ctx.slug;
		loadRotation();
	});
</script>

<svelte:head>
	<title>{ctx.project?.name ?? 'Project'} — Casier</title>
</svelte:head>

<div class="px-4 py-6 md:px-8">
	{#if ctx.environments.length === 0}
		<EmptyState
			title="No environments"
			description="A project needs at least one environment before it can hold secrets. Environments keep production, staging and local values apart."
		>
			{#snippet icon()}
				<iconify-icon icon="solar:layers-linear" width="40" aria-hidden="true"></iconify-icon>
			{/snippet}
			{#if ctx.canAdmin}
				<a
					href="/projects/{ctx.slug}/settings"
					class="inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
				>
					<iconify-icon icon="solar:layers-linear" width="16" aria-hidden="true"></iconify-icon>
					Create an environment
				</a>
			{/if}
		</EmptyState>
	{:else}
		<div class="flex flex-wrap items-center gap-2">
			<div
				class="inline-flex max-w-full items-center gap-1 overflow-x-auto rounded-lg border border-border bg-muted p-1"
				role="group"
				aria-label="Environments"
			>
				{#each ctx.environments as env (env.id)}
					{@const stale = staleCount(env.slug)}
					<button
						type="button"
						aria-pressed={selectedEnvSlug === env.slug}
						class="inline-flex shrink-0 items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium whitespace-nowrap transition-colors {selectedEnvSlug ===
						env.slug
							? 'bg-background text-foreground shadow-sm'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (selectedEnvSlug = env.slug)}
					>
						{env.name}
						{#if stale > 0}
							<span
								class="inline-flex h-5 min-w-5 items-center justify-center rounded-full border border-border px-1.5 text-xs tabular-nums"
								title="{stale} {stale === 1 ? 'secret needs' : 'secrets need'} rotating"
							>
								{stale}
							</span>
						{/if}
					</button>
				{/each}
			</div>
			{#if ctx.canAdmin}
				<a
					href="/projects/{ctx.slug}/settings"
					class="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-border bg-background px-4 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
				>
					<iconify-icon icon="solar:layers-linear" width="16" aria-hidden="true"></iconify-icon>
					Manage environments
				</a>
			{/if}
		</div>

		{#if selectedEnv}
			<div class="mt-6">
				<SecretsSection
					projectSlug={ctx.slug}
					environment={selectedEnv}
					canWrite={ctx.canWrite}
				/>
			</div>
		{/if}
	{/if}
</div>
