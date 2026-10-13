<script lang="ts">
	import { createDropdownMenu, melt } from '@melt-ui/svelte';
	import { fly } from 'svelte/transition';
	import { Tone, toneText } from './styles';
	import { TextSize } from './typography';
	import type { MenuItem } from './types';

	export let items: MenuItem[];
	export let label: string;
	export let placement: 'bottom-start' | 'bottom-end' = 'bottom-end';
	let className = '';
	export { className as class };

	const {
		elements: { trigger, menu, item },
		states: { open }
	} = createDropdownMenu({ forceVisible: true, positioning: { placement } });
</script>

<button
	use:melt={$trigger}
	type="button"
	aria-label={label}
	class="inline-flex rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve {className}"
>
	<slot />
</button>

{#if $open}
	<div
		use:melt={$menu}
		class="z-50 flex min-w-[10rem] flex-col rounded-md bg-ctp-surface0 p-1 shadow-md shadow-ctp-crust"
		transition:fly={{ duration: 150, y: -4 }}
	>
		{#each items as entry}
			<div
				use:melt={$item}
				data-disabled={entry.disabled ? '' : undefined}
				class="flex cursor-pointer items-center gap-2 rounded-md px-3 py-1.5 {TextSize.Sm} {toneText[
					entry.tone ?? Tone.Neutral
				]}
					data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50
					data-[highlighted]:bg-ctp-mauve/25 data-[highlighted]:text-ctp-mauve"
				on:m-click={entry.onSelect}
			>
				{#if entry.icon}
					<svelte:component this={entry.icon} size="14" />
				{/if}
				{entry.label}
			</div>
		{/each}
	</div>
{/if}
