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
	import { Button, EmptyState, Tabs, icons } from '@facile/muse';
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

	let selectedEnvSlug = $state('');

	let selectedEnv = $derived(ctx.environments.find((env) => env.slug === selectedEnvSlug) ?? null);

	let rotation = $state<Record<string, EnvironmentRotation>>({});

	function staleCount(envSlug: string): number {
		const entry = rotation[envSlug];
		if (!entry) return 0;
		return entry.overdue + entry.due_soon;
	}

	let envTabs = $derived(
		ctx.environments.map((env) => {
			const stale = staleCount(env.slug);
			return {
				id: env.slug,
				label: env.name,
				icon: env.protected ? icons.shield : undefined,
				badge: stale > 0 ? stale : undefined,
			};
		})
	);

	async function loadRotation() {
		try {
			const res = await api.get<RotationSummary>(
				`/projects/${ctx.slug}/secrets/rotation`,
			);
			rotation = Object.fromEntries((res.environments ?? []).map((env) => [env.slug, env]));
		} catch {
			rotation = {};
		}
	}

	$effect(() => {
		if (!ctx.environments.some((env) => env.slug === selectedEnvSlug)) {
			selectedEnvSlug = ctx.environments[0]?.slug ?? '';
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

<div class="flex flex-col gap-6 px-4 py-6 md:px-8">
	{#if ctx.environments.length === 0}
		<EmptyState
			icon={icons.server}
			title="No environments"
			description="A project needs at least one environment before it can hold secrets. Environments keep production, staging and local values apart."
		>
			{#if ctx.canAdmin}
				<Button href="/projects/{ctx.slug}/settings" icon={icons.plus}>
					Create an environment
				</Button>
			{/if}
		</EmptyState>
	{:else}
		<div class="flex flex-wrap items-center gap-3">
			<div class="min-w-0 flex-1">
				<Tabs items={envTabs} bind:value={selectedEnvSlug} label="Environments" />
			</div>
			{#if ctx.canAdmin}
				<Button href="/projects/{ctx.slug}/settings" variant="outline" icon={icons.settings}>
					Manage environments
				</Button>
			{/if}
		</div>

		{#if selectedEnv}
			<SecretsSection
				projectSlug={ctx.slug}
				environment={selectedEnv}
				canWrite={ctx.canWrite}
			/>
		{/if}
	{/if}
</div>
