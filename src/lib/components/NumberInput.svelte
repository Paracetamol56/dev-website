<script lang="ts">
	import { writable, type Writable } from 'svelte/store';

	export let value: Writable<number> = writable(0);
	export let label: string = '';
	export let min: number = 0;
	export let max: number = 100;
	export let step: number = 1;
	export let disabled: boolean = false;
</script>

<div class="flex flex-col gap-1">
	{#if label}
		<label class="block text-ctp-text font-semibold text-sm">{label}</label>
	{/if}
	<input
		type="number"
		{min}
		{max}
		{step}
		bind:value={$value}
		{disabled}
		class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 min-w-8 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve shadow-md shadow-ctp-crust"
	/>
</div>

