<script lang="ts">
	import { writable, type Writable } from 'svelte/store';

	export let value: Writable<string> = writable('');
	export let label: string = '';
	export let placeholder: string = '';
	export let disabled: boolean = false;
</script>

<div class="flex flex-col gap-1">
	{#if label}
		<label class="block text-ctp-text font-semibold text-sm">{label}</label>
	{/if}
	<input
		type="text"
		bind:value={$value}
		{placeholder}
		{disabled}
		class="h-8 min-w-8 w-full rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve shadow-md shadow-ctp-crust"
	/>
</div>

