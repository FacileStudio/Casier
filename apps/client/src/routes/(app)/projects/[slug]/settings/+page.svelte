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
	import { t } from '$lib/i18n/index.svelte';

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
		Number.isInteger(rotation) && rotation >= 0 && rotation <= 3650 ? '' : t('psettings.rotation.invalid')
	);

	let rotationDescription = $derived(
		rotationError
			? rotationError
			: rotation === 0
				? t('psettings.rotation.off')
				: t('psettings.rotation.on', { days: rotation })
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
			toast.success(t('psettings.updated'));
			await app?.refreshProjects();
			if (updated.slug && updated.slug !== ctx.slug) {
				await goto(`/projects/${updated.slug}/settings`, { replaceState: true });
			} else {
				await ctx.reloadProject();
			}
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('psettings.error.update'));
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
			toast.success(t('psettings.env.created', { name: env.name }));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('psettings.error.createEnv'));
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
					? t('psettings.env.unprotected', { name: env.name })
					: t('psettings.env.nowProtected', { name: env.name })
			);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('psettings.error.protect'));
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
			toast.success(t('psettings.env.deleted', { name: env.name }));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('psettings.error.deleteEnv'));
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
			toast.success(t('psettings.delete.deleted', { name: project.name }));
			await app?.refreshProjects();
			await goto('/projects');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('psettings.error.deleteProject'));
			deletingProject = false;
		}
	}
</script>

<svelte:head>
	<title>{t('psettings.head', { project: ctx.project?.name ?? t('project.fallbackName') })}</title>
</svelte:head>

{#if ctx.canAdmin}
	<div class="mx-auto flex w-full max-w-fc-lg flex-col gap-10 px-4 py-6 md:px-8">
		<form onsubmit={handleSave}>
			<SettingsSection
				title={t('psettings.general.title')}
				description={t('psettings.general.description')}
			>
				<SettingsRow
					label={t('common.name')}
					description={t('psettings.name.description')}
					for="project-name"
					stacked
				>
					<Input id="project-name" bind:value={name} disabled={saving} required maxlength={80} />
				</SettingsRow>

				<SettingsRow
					label={t('projects.description.label')}
					description={t('psettings.description.description')}
					for="project-description"
					stacked
				>
					<Textarea
						id="project-description"
						rows={3}
						bind:value={description}
						disabled={saving}
						placeholder={t('projects.description.placeholder')}
					/>
				</SettingsRow>

				<SettingsRow
					label={t('psettings.rotation.label')}
					description={rotationDescription}
					for="project-rotation"
				>
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
						<span class="text-fc-sm text-fc-fg-muted">{t('psettings.rotation.days')}</span>
					</div>
				</SettingsRow>

				<SettingsRow
					label={t('common.save')}
					description={dirty ? t('psettings.save.dirty') : t('psettings.save.clean')}
				>
					<Button type="submit" disabled={!canSave}>
						{saving ? t('common.saving') : t('psettings.save.button')}
					</Button>
				</SettingsRow>
			</SettingsSection>
		</form>

		<SettingsSection
			title={t('psettings.envs.title')}
			description={t('psettings.envs.description')}
		>
			{#each ctx.environments as env (env.id)}
				<SettingsRow label={env.name} description={env.slug}>
					<div class="flex shrink-0 items-center gap-1">
							{#if env.protected}
								<Badge tone="info">{t('psettings.env.protected')}</Badge>
							{/if}
							<IconButton
								variant="ghost"
								aria-label={env.protected
									? t('psettings.env.unprotectAria', { name: env.name })
									: t('psettings.env.protectAria', { name: env.name })}
								title={env.protected
									? t('psettings.env.unprotectTitle')
									: t('psettings.env.protectTitle')}
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
								variant="ghost-danger"
								aria-label={t('psettings.env.deleteAria', { name: env.name })}
								title={ctx.environments.length <= 1
									? t('psettings.env.keepOne')
									: t('psettings.env.deleteTitle', { name: env.name })}
								disabled={ctx.environments.length <= 1}
								onclick={() => askDeleteEnv(env)}
							>
								<iconify-icon icon={icons.remove} width="18" aria-hidden="true"></iconify-icon>
							</IconButton>
					</div>
				</SettingsRow>
			{/each}

			<SettingsRow
				label={t('psettings.env.add.label')}
				description={t('psettings.env.add.description')}
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
						{addingEnv ? t('common.adding') : t('common.add')}
					</Button>
				</form>
			</SettingsRow>
		</SettingsSection>

		<SettingsSection
			title={t('psettings.danger.title')}
			description={t('psettings.danger.description')}
		>
			<SettingsRow
				label={t('psettings.delete.label')}
				description={t('psettings.delete.description', { name: ctx.project?.name ?? '' })}
			>
				<Button variant="danger" icon={icons.remove} onclick={() => (deleteProjectOpen = true)}>
					{t('psettings.delete.button')}
				</Button>
			</SettingsRow>
		</SettingsSection>
	</div>

	<ConfirmDialog
		bind:open={deleteEnvOpen}
		title={t('psettings.env.deleteDialog')}
		message={t('psettings.env.deleteMessage', { name: envToDelete?.name ?? '' })}
		confirmLabel={t('psettings.env.deleteDialog')}
		loading={deletingEnv}
		onconfirm={handleDeleteEnv}
	/>

	<ConfirmDialog
		bind:open={deleteProjectOpen}
		title={t('psettings.delete.button')}
		message={t('psettings.delete.message', { name: ctx.project?.name ?? '' })}
		confirmLabel={t('psettings.delete.button')}
		loading={deletingProject}
		onconfirm={handleDeleteProject}
	/>
{/if}
