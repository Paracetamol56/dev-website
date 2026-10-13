<script lang="ts">
	import { createCheckbox, melt } from '@melt-ui/svelte';
	import { Check, Minus } from 'lucide-svelte';
	import { writable, type Writable } from 'svelte/store';
	import { TextColor, TextSize, TextWeight } from './typography';

	export let defaultChecked: boolean | 'indeterminate' = false;
	export let checked: Writable<boolean> = writable(
		defaultChecked === 'indeterminate' ? false : defaultChecked
	);
	export let name: string;
	export let label: string;
	export let hideLabel: boolean = false;
	export let required: boolean = false;
	export let disabled: boolean = false;

	const {
		elements: { root, input },
		helpers: { isChecked, isIndeterminate }
	} = createCheckbox({ defaultChecked, checked, name, required, disabled });
</script>

<div class="flex items-center gap-2 {disabled ? 'opacity-50' : ''}">
	<button
		use:melt={$root}
		id={name}
		class="grid square-4 shrink-0 cursor-default place-items-center rounded-md bg-ctp-surface0 shadow-sm shadow-ctp-crust
			hover:bg-ctp-surface1 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve"
	>
		{#if $isIndeterminate}
			<Minus class="square-3 stroke-[4] {TextColor.Accent}" />
		{:else if $isChecked}
			<Check class="square-3 stroke-[4] {TextColor.Accent}" />
		{/if}
		<input use:melt={$input} />
	</button>
	<label
		for={name}
		class="{TextSize.Base} {TextWeight.Semibold} leading-none {hideLabel ? 'sr-only' : ''}"
	>
		{label}
	</label>
</div>
