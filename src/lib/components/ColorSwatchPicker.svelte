<script lang="ts">
	import type { Writable } from 'svelte/store';
	import type { Variant } from '@catppuccin/palette';

	export let palette: Variant;
	export let selected: Writable<string>;
</script>

<div class="flex flex-wrap gap-2">
	{#each Object.entries(palette) as [key, colorObj]}
		{#if typeof colorObj === 'object' && colorObj.hex}
			<button
				class="w-8 h-8 rounded cursor-pointer border-2 transition-all hover:scale-110
					{$selected === key ? 'border-ctp-mauve ring-2 ring-ctp-mauve/50' : 'border-ctp-surface0'}"
				style="background-color: {colorObj.hex}"
				on:click={() => selected.set(key)}
				aria-label={key}
			/>
		{/if}
	{/each}
</div>
