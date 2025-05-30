<script lang="ts">
	import { type Writable } from 'svelte/store';
	import { type OutlineSection } from './outline';

	export let outline: Writable<OutlineSection[] | null>;
</script>

<div class="flex justify-start items-stretch gap-3 p-3 me-auto">
	{#if $outline !== null}
		{#each $outline as section}
			<div class="h-32 aspect-video bg-ctp-base rounded-md shadow-md shadow-ctp-crust p-2">
				<h3 class="w-full text-lg text-ctp-text font-semibold">{section.heading}</h3>
			</div>
		{/each}
	{/if}
</div>
