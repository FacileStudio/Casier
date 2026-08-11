<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Badge,
		Button,
		IconButton,
		Input,
		SettingsRow,
		SettingsSection,
		Textarea,
		icons,
		toast,
	} from '@facile/muse';
	import { auth } from '$lib/auth.svelte';
	import { api, type Environment, type Project } from '$lib/backend';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';

	type ProjectContext = {
		readonly project: Project | null;
		readonly environments: Environment[];
		readonly canWrite: boolean;
		readonly canAdmin: boolean;
		readonly slug: string;
		reloadEnvironments: () => Promise<void>;
		reloadProject: () => Promise<void>;
	};

	type AppContext = { refreshProjects: () => Promise<void> };

	const ctx = getContext<ProjectContext>('project');
	const app = getContext<AppContext | undefined>('app');

	let name = $state('');
	let description = $state('');
	let rotationDays = $state<string | number>(0);
	let saving = $state(false);

	let newEnvName = $state('');
	let addingEnv = $state(false);
	let togglingEnvSlug = $state<string | null>(null);

	let envToDelete = $state<Environment | null>(null);
	let deleteEnvOpen = $state(false);
	let deletingEnv = $state(false);

	let deleteProjectOpen = $state(false);
	let deletingProject = $state(false);

	let rotation = $derived(rotationDays === '' ? NaN : Number(rotationDays));

	let dirty = $derived(
		ctx.project !== null &&
			(name.trim() !== ctx.project.name ||
				description.trim() !== ctx.project.description ||
				rotation !== ctx.project.rotation_days)
	);

	let rotationError = $derived(
		Number.isInteger(rotation) && rotation >= 0 && rotation <= 3650
			? ''
			: 'Use a whole number of days between 0 and 3650'
	);

	let rotationDescription = $derived(
		rotationError
			? rotationError
			: rotation === 0
				? 'Rotation reminders are off. Secrets are never flagged as stale.'
				: `Secrets whose value has not changed in ${rotation} days are flagged for rotation in every environment of this project.`
	);

	let canSave = $derived(dirty && name.trim().length >= 2 && !rotationError && !saving);

	$effect(() => {
		if (ctx.project && !ctx.canAdmin) {
			goto(`/projects/${ctx.slug}`, { replaceState: true });
		}
	});

	$effect(() => {
		const project = ctx.project;
		if (!project) return;
		name = project.name;
		description = project.description;
		rotationDays = project.rotation_days;
	});

	async function handleSave(event: Event) {
		event.preventDefault();
		if (!canSave) return;
		saving = true;
		try {
			const updated = await api.put<Project>(
				`/projects/${ctx.slug}`,
				{
					name: name.trim(),
					description: description.trim(),
					rotation_days: rotation,
				},
			);
			toast.success('Project updated');
			await app?.refreshProjects();
			if (updated.slug && updated.slug !== ctx.slug) {
				await goto(`/projects/${updated.slug}/settings`, { replaceState: true });
			} else {
				await ctx.reloadProject();
			}
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to update project');
		} finally {
			saving = false;
		}
	}

	async function handleAddEnv(event: Event) {
		event.preventDefault();
		if (!newEnvName.trim()) return;
		addingEnv = true;
		try {
			const env = await api.post<Environment>(
				`/projects/${ctx.slug}/environments`,
				{ name: newEnvName.trim() },
			);
			newEnvName = '';
			await ctx.reloadEnvironments();
			toast.success(`Environment "${env.name}" created`);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to create environment');
		} finally {
			addingEnv = false;
		}
	}

	async function toggleEnvProtection(env: Environment) {
		togglingEnvSlug = env.slug;
		try {
			await api.put<Environment>(
				`/projects/${ctx.slug}/environments/${env.slug}`,
				{ name: env.name, protected: !env.protected },
			);
			await ctx.reloadEnvironments();
			toast.success(
				env.protected
					? `"${env.name}" is no longer protected`
					: `"${env.name}" now requires owner or admin to write`
			);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to change protection');
		} finally {
			togglingEnvSlug = null;
		}
	}

	function askDeleteEnv(env: Environment) {
		envToDelete = env;
		deleteEnvOpen = true;
	}

	async function handleDeleteEnv() {
		const env = envToDelete;
		if (!env) return;
		deletingEnv = true;
		try {
			await api.del(`/projects/${ctx.slug}/environments/${env.slug}`);
			deleteEnvOpen = false;
			envToDelete = null;
			await ctx.reloadEnvironments();
			toast.success(`Environment "${env.name}" deleted`);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to delete environment');
		} finally {
			deletingEnv = false;
		}
	}

	async function handleDeleteProject() {
		const project = ctx.project;
		if (!project) return;
		deletingProject = true;
		try {
			await api.del(`/projects/${ctx.slug}`);
			toast.success(`Project "${project.name}" deleted`);
			await app?.refreshProjects();
			await goto('/projects');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to delete project');
			deletingProject = false;
		}
	}
</script>

<svelte:head>
	<title>Settings — {ctx.project?.name ?? 'Project'} — Casier</title>
</svelte:head>

{#if ctx.canAdmin}
	<div class="mx-auto flex w-full max-w-fc-lg flex-col gap-10 px-4 py-6 md:px-8">
		<form onsubmit={handleSave}>
			<SettingsSection
				title="General"
				description="How this project is named and how long a secret may sit unchanged before Casier flags it."
			>
				<SettingsRow
					label="Name"
					description="Renaming a project rewrites its slug, and every URL that names it moves with it."
					for="project-name"
					stacked
				>
					<Input id="project-name" bind:value={name} disabled={saving} required maxlength={80} />
				</SettingsRow>

				<SettingsRow
					label="Description"
					description="What lives in this project. Shown on the project list."
					for="project-description"
					stacked
				>
					<Textarea
						id="project-description"
						rows={3}
						bind:value={description}
						disabled={saving}
						placeholder="What lives in this project?"
					/>
				</SettingsRow>

				<SettingsRow label="Rotation reminder" description={rotationDescription} for="project-rotation">
					<div class="flex items-center gap-2">
						<Input
							id="project-rotation"
							type="number"
							min="0"
							max="3650"
							step="1"
							bind:value={rotationDays}
							disabled={saving}
							aria-invalid={rotationError ? 'true' : undefined}
							class="w-28"
						/>
						<span class="text-fc-sm text-fc-fg-muted">days</span>
					</div>
				</SettingsRow>

				<SettingsRow
					label="Save"
					description={dirty ? 'You have unsaved changes.' : 'Nothing has changed yet.'}
				>
					<Button type="submit" disabled={!canSave}>
						{saving ? 'Saving…' : 'Save changes'}
					</Button>
				</SettingsRow>
			</SettingsSection>
		</form>

		<SettingsSection
			title="Environments"
			description="Each environment holds its own copy of every secret. A protected one takes owner or admin to write to."
		>
			{#each ctx.environments as env (env.id)}
				<SettingsRow label={env.name} description={env.slug}>
					<div class="flex shrink-0 items-center gap-1">
							{#if env.protected}
								<Badge tone="info">Protected</Badge>
							{/if}
							<IconButton
								variant="ghost"
								aria-label={env.protected
									? `Remove protection from ${env.name}`
									: `Protect ${env.name}`}
								title={env.protected
									? 'Allow any member to write here'
									: 'Require owner or admin to write here'}
								disabled={togglingEnvSlug === env.slug}
								onclick={() => toggleEnvProtection(env)}
							>
								<iconify-icon
									icon={env.protected
										? 'solar:lock-keyhole-unlocked-linear'
										: 'solar:lock-keyhole-linear'}
									width="18"
									aria-hidden="true"
								></iconify-icon>
							</IconButton>
							<IconButton
								variant="danger"
								aria-label={`Delete environment ${env.name}`}
								title={ctx.environments.length <= 1
									? 'A project keeps at least one environment'
									: `Delete ${env.name}`}
								disabled={ctx.environments.length <= 1}
								onclick={() => askDeleteEnv(env)}
							>
								<iconify-icon icon={icons.remove} width="18" aria-hidden="true"></iconify-icon>
							</IconButton>
					</div>
				</SettingsRow>
			{/each}

			<SettingsRow
				label="Add an environment"
				description="The slug is derived from the name and is what the CLI and every token scope refer to."
				for="new-environment"
				stacked
			>
				<form class="flex flex-col gap-2 sm:flex-row" onsubmit={handleAddEnv}>
					<Input
						id="new-environment"
						bind:value={newEnvName}
						disabled={addingEnv}
						placeholder="staging"
						class="flex-1"
					/>
					<Button
						type="submit"
						size="lg"
						icon={icons.plus}
						disabled={addingEnv || !newEnvName.trim()}
						class="sm:shrink-0"
					>
						{addingEnv ? 'Adding…' : 'Add'}
					</Button>
				</form>
			</SettingsRow>
		</SettingsSection>

		<SettingsSection
			title="Danger zone"
			description="Irreversible, and nobody can undo it for you."
		>
			<SettingsRow
				label="Delete this project"
				description={`Erases every environment, every secret value and version, every member and every API token ${ctx.project?.name ?? 'this project'} holds. Anything reading from it breaks immediately. There is no backup and no undo.`}
			>
				<Button variant="danger" icon={icons.remove} onclick={() => (deleteProjectOpen = true)}>
					Delete project
				</Button>
			</SettingsRow>
		</SettingsSection>
	</div>

	<ConfirmDialog
		bind:open={deleteEnvOpen}
		title="Delete environment"
		message={`Deleting "${envToDelete?.name ?? ''}" destroys every secret stored in it, including all past versions, and every API token scope that named it. Other environments in this project are untouched. This cannot be undone.`}
		confirmLabel="Delete environment"
		loading={deletingEnv}
		onconfirm={handleDeleteEnv}
	/>

	<ConfirmDialog
		bind:open={deleteProjectOpen}
		title="Delete project"
		message={`Deleting "${ctx.project?.name ?? ''}" permanently destroys all of its environments, secrets, members and API tokens. Anything reading these secrets will stop working. This cannot be undone.`}
		confirmLabel="Delete project"
		loading={deletingProject}
		onconfirm={handleDeleteProject}
	/>
{/if}
