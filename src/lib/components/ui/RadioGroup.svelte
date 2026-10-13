<script lang="ts">
	import { createRadioGroup, melt } from '@melt-ui/svelte';
	import type { Writable } from 'svelte/store';
	import { TextSize, TextWeight } from './typography';

	export let value: Writable<string>;
	export let options: string[];
	export let name: string = 'radio-group';
	export let orientation: 'horizontal' | 'vertical' = 'horizontal';
	export let required: boolean = false;
	export let disabled: boolean = false;

	const {
		elements: { root, item, hiddenInput },
		helpers: { isChecked }
	} = createRadioGroup({ required, disabled, orientation, value });
</script>

<div
	use:melt={$root}
	aria-label={name}
	class="flex flex-col gap-3 data-[orientation=horizontal]:flex-row {disabled ? 'opacity-50' : ''}"
>
	{#each options as option}
		<div class="flex items-center gap-2">
			<button
				use:melt={$item(option)}
				id="{name}-{option}"
				aria-labelledby="{name}-{option}-label"
				class="grid square-5 shrink-0 cursor-default place-items-center rounded-full bg-ctp-surface0 shadow-sm shadow-ctp-crust
					hover:bg-ctp-surface1 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve"
			>
				{#if $isChecked(option)}
					<span class="square-3 rounded-full bg-ctp-mauve" />
				{/if}
			</button>
			<label
				for="{name}-{option}"
				id="{name}-{option}-label"
				class="{TextSize.Base} {TextWeight.Semibold} capitalize leading-none"
			>
				{option}
			</label>
		</div>
	{/each}
	<input {name} use:melt={$hiddenInput} />
</div>
