<script lang="ts">
	let {
		size = 24,
		wordmark = false,
		class: className = '',
	}: { size?: number; wordmark?: boolean; class?: string } = $props();
</script>

<span class="inline-flex items-center gap-2 {className}">
	<svg
		width={size}
		height={size}
		viewBox="0 0 24 24"
		role="img"
		aria-label="Casier"
		class="shrink-0"
	>
		<rect width="24" height="24" rx="6" fill="var(--color-fc-accent)" />
		<rect
			x="6.5"
			y="4.5"
			width="11"
			height="15"
			rx="2"
			fill="none"
			stroke="var(--color-fc-accent-fg)"
			stroke-width="1.8"
		/>
		<line
			x1="9.5"
			y1="8"
			x2="14.5"
			y2="8"
			stroke="var(--color-fc-accent-fg)"
			stroke-width="1.6"
			stroke-linecap="round"
		/>
		<line
			x1="9.5"
			y1="11"
			x2="14.5"
			y2="11"
			stroke="var(--color-fc-accent-fg)"
			stroke-width="1.6"
			stroke-linecap="round"
		/>
		<circle cx="12" cy="15.5" r="1.4" fill="var(--color-fc-accent-fg)" />
	</svg>
	{#if wordmark}
		<span class="font-fc-title text-fc-lg font-semibold tracking-tight text-fc-fg">Casier</span>
	{/if}
</span>
