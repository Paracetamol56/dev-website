<script lang="ts">
	import type { ComponentType } from 'svelte';

	export let icon: ComponentType;
	export let iconClass: string;
	export let label: string;
	export let value: string | number;
	export let detail = '';
</script>

<div class="rounded-md bg-ctp-mantle p-4 shadow-md shadow-ctp-crust">
	<span class="flex items-center gap-2 text-sm font-semibold text-ctp-subtext0">
		<svelte:component this={icon} size="16" class={iconClass} />
		{label}
	</span>
	<span class="block text-3xl font-bold">{value}</span>
	{#if detail || $$slots.default}
		<span class="block truncate text-sm text-ctp-subtext0"><slot>{detail}</slot></span>
	{/if}
</div>
