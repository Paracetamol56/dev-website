<script lang="ts">
	import type { Readable } from 'svelte/store';
	import { type OutlineSection } from './outline';

	export let outline: Readable<OutlineSection[] | null>;
	export let current: Readable<{ h: number; v: number; f?: number }>;
</script>

<div class="flex justify-start items-stretch gap-3 p-3 me-auto">
	{#if $outline !== null}
		{#each $outline as section, index}
			<a
				href={index === 0 ? '#' : `#/${index}`}
				class={`h-32 aspect-video bg-ctp-base rounded-md shadow-md shadow-ctp-crust p-2 ${
					index === $current.h ? 'ring-2 ring-ctp-mauve' : ''
				}`}
			>
				<h3 class="w-full text-lg text-ctp-text font-semibold">{section.heading}</h3>
			</a>
		{/each}
	{/if}
</div>
