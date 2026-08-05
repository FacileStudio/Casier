<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api, type Environment, type Project } from '$lib/backend';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { toast } from 'svelte-sonner';

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
	let rotationDays = $state(0);
	let saving = $state(false);

	let newEnvName = $state('');
	let addingEnv = $state(false);
	let togglingEnvSlug = $state<string | null>(null);

	let envToDelete = $state<Environment | null>(null);
	let deleteEnvOpen = $state(false);
	let deletingEnv = $state(false);

	let deleteProjectOpen = $state(false);
	let deletingProject = $state(false);

	let dirty = $derived(
		ctx.project !== null &&
			(name.trim() !== ctx.project.name ||
				description.trim() !== ctx.project.description ||
				rotationDays !== ctx.project.rotation_days)
	);

	let rotationError = $derived(
		Number.isInteger(rotationDays) && rotationDays >= 0 && rotationDays <= 3650
			? ''
			: 'Use a whole number of days between 0 and 3650'
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
					rotation_days: rotationDays,
				},
				auth.token
			);
			toast.success('Project updated');
			await app?.refreshProjects();
			if (updated.slug && updated.slug !== ctx.slug) {
				await goto(`/projects/${updated.slug}/settings`, { replaceState: true });
			} else {
				await ctx.reloadProject();
			}
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to update project');
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
				auth.token
			);
			newEnvName = '';
			await ctx.reloadEnvironments();
			toast.success(`Environment "${env.name}" created`);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to create environment');
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
				auth.token
			);
			await ctx.reloadEnvironments();
			toast.success(
				env.protected
					? `"${env.name}" is no longer protected`
					: `"${env.name}" now requires owner or admin to write`
			);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to change protection');
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
			await api.del(`/projects/${ctx.slug}/environments/${env.slug}`, auth.token);
			deleteEnvOpen = false;
			envToDelete = null;
			await ctx.reloadEnvironments();
			toast.success(`Environment "${env.name}" deleted`);
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to delete environment');
		} finally {
			deletingEnv = false;
		}
	}

	async function handleDeleteProject() {
		const project = ctx.project;
		if (!project) return;
		deletingProject = true;
		try {
			await api.del(`/projects/${ctx.slug}`, auth.token);
			toast.success(`Project "${project.name}" deleted`);
			await app?.refreshProjects();
			await goto('/projects');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to delete project');
			deletingProject = false;
		}
	}
</script>

<svelte:head>
	<title>Settings — {ctx.project?.name ?? 'Project'} — Casier</title>
</svelte:head>

{#if ctx.canAdmin}
	<div class="px-4 py-6 md:px-8">
		<div class="max-w-xl space-y-8">
			<section class="space-y-4">
				<h2 class="text-sm font-semibold tracking-wider text-muted-foreground uppercase">
					General
				</h2>
				<form class="space-y-4" onsubmit={handleSave}>
					<div class="space-y-1.5">
						<label for="project-name" class="block text-sm font-medium">Name</label>
						<input
							id="project-name"
							type="text"
							bind:value={name}
							disabled={saving}
							required
							class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none disabled:opacity-50"
						/>
					</div>
					<div class="space-y-1.5">
						<label for="project-description" class="block text-sm font-medium">Description</label>
						<textarea
							id="project-description"
							rows="3"
							bind:value={description}
							disabled={saving}
							placeholder="What lives in this project?"
							class="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none disabled:opacity-50"
						></textarea>
					</div>
					<div class="space-y-1.5">
						<label for="project-rotation" class="block text-sm font-medium">
							Rotation reminder
						</label>
						<div class="flex items-center gap-2">
							<input
								id="project-rotation"
								type="number"
								min="0"
								max="3650"
								step="1"
								bind:value={rotationDays}
								disabled={saving}
								class="flex h-10 w-28 rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none disabled:opacity-50"
							/>
							<span class="text-sm text-muted-foreground">days</span>
						</div>
						<p class="text-xs text-muted-foreground">
							{#if rotationError}
								<span class="text-destructive">{rotationError}</span>
							{:else if rotationDays === 0}
								Rotation reminders are off. Secrets are never flagged as stale.
							{:else}
								Secrets whose value has not changed in {rotationDays} days are flagged for rotation
								in every environment of this project.
							{/if}
						</p>
					</div>
					<button
						type="submit"
						disabled={!canSave}
						class="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50"
					>
						{saving ? 'Saving…' : 'Save changes'}
					</button>
				</form>
			</section>

			<section class="space-y-4">
				<h2 class="text-sm font-semibold tracking-wider text-muted-foreground uppercase">
					Environments
				</h2>
				<div class="space-y-2">
					{#each ctx.environments as env (env.id)}
						<div class="flex items-center justify-between gap-3 rounded-lg border border-border p-3">
							<div class="flex min-w-0 flex-wrap items-center gap-2">
								<span class="truncate text-sm font-medium">{env.name}</span>
								<code class="rounded bg-muted px-2 py-1 font-mono text-xs">{env.slug}</code>
								{#if env.protected}
									<span
										class="inline-flex items-center gap-1 rounded-md border border-border px-2 py-1 text-xs text-muted-foreground"
									>
										<iconify-icon icon="solar:lock-keyhole-linear" width="12" aria-hidden="true"
										></iconify-icon>
										Protected
									</span>
								{/if}
							</div>
							<div class="flex shrink-0 items-center gap-1">
								<button
									type="button"
									aria-label={env.protected
										? `Remove protection from ${env.name}`
										: `Protect ${env.name}`}
									title={env.protected
										? 'Allow any member to write here'
										: 'Require owner or admin to write here'}
									disabled={togglingEnvSlug === env.slug}
									onclick={() => toggleEnvProtection(env)}
									class="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
								>
									<iconify-icon
										icon={env.protected
											? 'solar:lock-keyhole-unlocked-linear'
											: 'solar:lock-keyhole-linear'}
										width="16"
										aria-hidden="true"
									></iconify-icon>
								</button>
								<button
									type="button"
									aria-label="Delete environment {env.name}"
									disabled={ctx.environments.length <= 1}
									onclick={() => askDeleteEnv(env)}
									class="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
								>
									<iconify-icon icon="solar:trash-bin-2-linear" width="16" aria-hidden="true"
									></iconify-icon>
								</button>
							</div>
						</div>
					{/each}
				</div>
				<form class="flex flex-wrap items-center gap-2" onsubmit={handleAddEnv}>
					<input
						type="text"
						bind:value={newEnvName}
						disabled={addingEnv}
						placeholder="Environment name"
						aria-label="New environment name"
						class="flex h-10 min-w-0 flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none disabled:opacity-50"
					/>
					<button
						type="submit"
						disabled={addingEnv || !newEnvName.trim()}
						class="inline-flex h-10 shrink-0 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50"
					>
						<iconify-icon icon="solar:add-circle-linear" width="16" aria-hidden="true"
						></iconify-icon>
						{addingEnv ? 'Adding…' : 'Add'}
					</button>
				</form>
			</section>

			<section class="space-y-4">
				<h2 class="text-sm font-semibold tracking-wider text-muted-foreground uppercase">
					Danger zone
				</h2>
				<div class="rounded-lg border border-border p-4">
					<p class="text-base font-semibold">Delete this project</p>
					<p class="mt-1 text-sm text-muted-foreground">
						Deleting {ctx.project?.name} erases every environment, every secret value, every
						member and every API token it holds. Applications reading from it will break
						immediately. This is permanent — there is no backup and no undo.
					</p>
					<button
						type="button"
						onclick={() => (deleteProjectOpen = true)}
						class="mt-4 inline-flex h-9 items-center justify-center gap-2 rounded-md border border-border bg-background px-4 text-sm font-medium text-destructive transition-colors hover:bg-accent hover:text-accent-foreground"
					>
						<iconify-icon icon="solar:danger-triangle-linear" width="16" aria-hidden="true"
						></iconify-icon>
						Delete project
					</button>
				</div>
			</section>
		</div>
	</div>

	<ConfirmDialog
		bind:open={deleteEnvOpen}
		title="Delete environment"
		message={`Deleting "${envToDelete?.name ?? ''}" destroys every secret stored in it, including all past versions. Other environments in this project are untouched. This cannot be undone.`}
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
