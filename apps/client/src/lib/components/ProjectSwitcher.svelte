<script lang="ts">
	import { page } from '$app/state';
	import type { Project } from '$lib/backend';
	import { DropdownMenu } from 'bits-ui';

	let { projects = [] }: { projects?: Project[] } = $props();

	let currentSlug = $derived(page.params.slug ?? null);
	let current = $derived(projects.find((project) => project.slug === currentSlug) ?? null);
</script>

<div class="px-3 pb-3">
	<DropdownMenu.Root>
		<DropdownMenu.Trigger
			class="flex w-full items-center gap-2 rounded-md border border-border px-3 py-2 text-left text-sm transition-colors hover:bg-muted"
		>
			<iconify-icon
				icon="solar:widget-2-linear"
				width="16"
				class="shrink-0 text-muted-foreground"
				aria-hidden="true"
			></iconify-icon>
			<span class="min-w-0 flex-1 truncate">{current?.name ?? 'All projects'}</span>
			<iconify-icon
				icon="solar:alt-arrow-down-linear"
				width="14"
				class="shrink-0 text-muted-foreground"
				aria-hidden="true"
			></iconify-icon>
		</DropdownMenu.Trigger>

		<DropdownMenu.Portal>
			<DropdownMenu.Content
				align="start"
				sideOffset={6}
				class="z-50 max-h-80 w-56 overflow-y-auto rounded-lg border border-border bg-popover p-1 text-popover-foreground shadow-lg"
			>
				<DropdownMenu.Item>
					{#snippet child({ props })}
						<a
							{...props}
							href="/projects"
							class="flex cursor-pointer items-center gap-2 rounded-md px-3 py-2 text-sm outline-none data-highlighted:bg-muted"
						>
							<iconify-icon icon="solar:widget-2-linear" width="16" aria-hidden="true"
							></iconify-icon>
							All projects
						</a>
					{/snippet}
				</DropdownMenu.Item>

				{#if projects.length > 0}
					<DropdownMenu.Separator class="my-1 h-px bg-border" />
					{#each projects as project (project.id)}
						<DropdownMenu.Item>
							{#snippet child({ props })}
								<a
									{...props}
									href="/projects/{project.slug}"
									class="flex cursor-pointer items-center justify-between gap-2 rounded-md px-3 py-2 text-sm outline-none data-highlighted:bg-muted"
								>
									<span class="min-w-0 truncate">{project.name}</span>
									{#if project.slug === currentSlug}
										<iconify-icon
											icon="solar:check-read-linear"
											width="14"
											class="shrink-0 text-muted-foreground"
											aria-hidden="true"
										></iconify-icon>
									{/if}
								</a>
							{/snippet}
						</DropdownMenu.Item>
					{/each}
				{/if}
			</DropdownMenu.Content>
		</DropdownMenu.Portal>
	</DropdownMenu.Root>
</div>
