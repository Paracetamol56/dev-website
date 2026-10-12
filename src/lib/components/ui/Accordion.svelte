<script lang="ts">
	import { createAccordion, melt } from '@melt-ui/svelte';
	import { ChevronDown } from 'lucide-svelte';
	import { slide } from 'svelte/transition';
	import { TextColor, TextSize, TextWeight } from './typography';
	import type { Section } from './types';

	export let items: Section[];
	export let multiple: boolean = true;
	export let expanded: string[] = [];

	const {
		elements: { root, item, trigger, content },
		helpers: { isSelected }
	} = createAccordion({ multiple, defaultValue: multiple ? expanded : expanded[0] });
</script>

<div use:melt={$root} class="overflow-hidden rounded-xl shadow-md shadow-ctp-crust">
	{#each items as { id, title }}
		<div use:melt={$item(id)}>
			<h3 class="flex">
				<button
					use:melt={$trigger(id)}
					class="flex flex-1 cursor-pointer items-center justify-between gap-2 bg-ctp-mantle px-5 py-5 leading-none
						{TextSize.Lg} {TextWeight.Semibold} {TextColor.Default}
						transition-colors hover:bg-ctp-crust focus:outline-none focus-visible:bg-ctp-crust"
				>
					<span class="flex items-center gap-2">
						{title}
						<slot name="title" {id} />
					</span>
					<ChevronDown
						class="square-5 shrink-0 transition-transform {$isSelected(id) ? 'rotate-180' : ''}"
					/>
				</button>
			</h3>
			{#if $isSelected(id)}
				<div
					use:melt={$content(id)}
					class="overflow-hidden bg-ctp-base px-5 py-4 {TextColor.Default}"
					transition:slide={{ duration: 200 }}
				>
					<slot {id} />
				</div>
			{/if}
		</div>
	{/each}
</div>
