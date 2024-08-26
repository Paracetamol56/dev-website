<script lang="ts">
	import { createSlider, melt } from '@melt-ui/svelte';
	import { math } from 'mathlifier';
	import { writable, type Writable } from 'svelte/store';

	export let name: string = 'Value';
	export let defaultValue: number = 50;
	export let value: Writable<number[]> = writable([defaultValue]);
	export let min: number = 0;
	export let max: number = 100;
	export let step: number = 1;
	export let disabled: boolean = false;

	const {
		elements: { root, range, thumb }
	} = createSlider({
		defaultValue: [defaultValue],
		min,
		max,
		step,
		disabled,
		value
	});
</script>

<div>
	<span class="block text-ctp-text font-semibold text-sm">{name}</span>
	<span
		use:melt={$root}
		class="relative flex h-[20px] min-w-[220px] items-center -mb-1 {disabled && 'opacity-50'}"
	>
		<span class="block h-1 w-full rounded bg-ctp-mauve/20">
			<span use:melt={$range} class="h-1 rounded bg-ctp-mauve" />
		</span>
		<span
			use:melt={$thumb()}
			class="block square-3 rounded-full outline-none bg-ctp-mauve {!disabled &&
				'focus:ring-4 focus:ring-ctp-mauve/20'}"
		/>
	</span>
	<span class="text-sm">{$value[0]}</span>
</div>
