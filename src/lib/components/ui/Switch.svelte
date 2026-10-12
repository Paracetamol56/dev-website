<script lang="ts">
	import { createSwitch, melt } from '@melt-ui/svelte';
	import { writable, type Writable } from 'svelte/store';
	import { TextSize, TextWeight } from './typography';

	export let checked: Writable<boolean> = writable(false);
	export let name: string;
	export let label: string;
	export let hideLabel: boolean = false;
	export let required: boolean = false;
	export let disabled: boolean = false;

	const {
		elements: { root, input }
	} = createSwitch({ checked, name, required, disabled });
</script>

<div class="flex items-center gap-3 {disabled ? 'opacity-50' : ''}">
	<button
		use:melt={$root}
		id={name}
		class="relative h-5 w-9 shrink-0 cursor-default rounded-full bg-ctp-surface0 shadow-md shadow-ctp-crust transition-colors
			focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve focus-visible:ring-offset-2 focus-visible:ring-offset-ctp-base
			data-[state=checked]:bg-ctp-mauve"
	>
		<span
			class="block square-4 rounded-full bg-ctp-text transition-transform {$checked
				? 'translate-x-[1.125rem]'
				: 'translate-x-0.5'}"
		/>
	</button>
	<input use:melt={$input} />
	<label
		for={name}
		class="{TextSize.Sm} {TextWeight.Semibold} leading-none {hideLabel ? 'sr-only' : ''}"
	>
		{label}
	</label>
</div>
