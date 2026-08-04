<script lang="ts">
	import { getContext } from 'svelte';
	import type { Environment, Project } from '$lib/backend';
	import AuditSection from '$lib/components/project/AuditSection.svelte';

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
</script>

<svelte:head>
	<title>Audit log — {ctx.project?.name ?? 'Project'} — Casier</title>
</svelte:head>

<div class="px-4 py-6 md:px-8">
	<AuditSection projectSlug={ctx.slug} />
</div>
