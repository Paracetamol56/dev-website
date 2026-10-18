<script lang="ts">
	import { createSlider, melt } from '@melt-ui/svelte';
	import { writable, type Writable } from 'svelte/store';
	import { labelClass } from './styles';
	import { TextColor, TextSize, TextWeight } from './typography';

	export let label: string = 'Value';
	export let defaultValue: number = 50;
	export let value: Writable<number[]> = writable([defaultValue]);
	export let min: number = 0;
	export let max: number = 100;
	export let step: number = 1;
	export let disabled: boolean = false;
	/** Text shown (and announced) for a value, e.g. to map step indexes to labels */
	export let format: (value: number) => string = String;
	/** Optional labels spread evenly under the track, one per step */
	export let ticks: string[] = [];

	const {
		elements: { root, range, thumbs }
	} = createSlider({ defaultValue: [defaultValue], min, max, step, disabled, value });
</script>

<div>
	<span class={labelClass}>
		{label}
		{#if ticks.length}
			<span class="ml-1 {TextWeight.Normal} {TextColor.Accent}">{format($value[0])}</span>
		{/if}
	</span>
	<span
		use:melt={$root}
		class="relative -mb-1 flex h-[20px] min-w-[220px] items-center {disabled ? 'opacity-50' : ''}"
	>
		<span class="block h-1 w-full rounded bg-ctp-mauve/20">
			<span use:melt={$range} class="h-1 rounded bg-ctp-mauve" />
		</span>
		<!-- A value with two numbers is a range, with a start and an end thumb -->
		{#each $thumbs as thumb, i}
			<span
				use:melt={thumb}
				aria-label={$thumbs.length > 1 ? `${label} ${i === 0 ? 'start' : 'end'}` : label}
				aria-valuetext={format($value[i])}
				class="block square-3 rounded-full bg-ctp-mauve outline-none {disabled
					? ''
					: 'focus:ring-4 focus:ring-ctp-mauve/20'}"
			/>
		{/each}
	</span>
	{#if ticks.length}
		<span class="mt-1 flex justify-between text-[10px] {TextColor.Muted}" aria-hidden="true">
			{#each ticks as tick}
				<span>{tick}</span>
			{/each}
		</span>
	{:else}
		<span class={TextSize.Sm}>{$value.map(format).join(' – ')}</span>
	{/if}
</div>
