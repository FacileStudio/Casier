<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import type { Environment, Project } from '$lib/backend';
	import TokensSection from '$lib/components/project/TokensSection.svelte';

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

	$effect(() => {
		if (ctx.project && !ctx.canAdmin) {
			goto(`/projects/${ctx.slug}`, { replaceState: true });
		}
	});
</script>

<svelte:head>
	<title>API tokens — {ctx.project?.name ?? 'Project'} — Casier</title>
</svelte:head>

{#if ctx.canAdmin}
	<div class="px-4 py-6 md:px-8">
		<TokensSection projectSlug={ctx.slug} />
	</div>
{/if}
