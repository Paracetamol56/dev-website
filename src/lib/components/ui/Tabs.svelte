<script lang="ts">
	import { createTabs, melt } from '@melt-ui/svelte';
	import { writable, type Writable } from 'svelte/store';
	import { TextColor, TextSize, TextWeight } from './typography';
	import type { Section } from './types';

	export let tabs: Section[];
	export let value: Writable<string> = writable(tabs[0]?.id ?? '');
	export let label: string = 'Tabs';

	const {
		elements: { root, list, content, trigger }
	} = createTabs({ value });
</script>

<div use:melt={$root} class="flex flex-col overflow-hidden rounded-md shadow-md shadow-ctp-crust">
	<div use:melt={$list} aria-label={label} class="flex shrink-0 overflow-x-auto bg-ctp-crust">
		{#each tabs as tab}
			<button
				use:melt={$trigger(tab.id)}
				class="relative flex h-12 flex-1 cursor-default select-none items-center justify-center px-4
					{TextSize.Base} {TextWeight.Medium} focus:outline-none focus-visible:underline
					{$value === tab.id ? `bg-ctp-mantle ${TextColor.Default}` : `bg-ctp-crust ${TextColor.Muted}`}"
			>
				{tab.title}
				{#if $value === tab.id}
					<span
						class="absolute bottom-1 left-1/2 h-1 w-6 -translate-x-1/2 rounded-full bg-ctp-mauve"
					/>
				{/if}
			</button>
		{/each}
	</div>
	{#each tabs as tab}
		<div use:melt={$content(tab.id)} class="grow bg-ctp-mantle p-5 {TextColor.Default}">
			{#if $value === tab.id}
				<slot id={tab.id} />
			{/if}
		</div>
	{/each}
</div>
