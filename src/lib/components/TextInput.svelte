<script lang="ts" context="module">
	let count = 0;
</script>

<script lang="ts">
	import { writable, type Writable } from 'svelte/store';

	export let value: Writable<string> = writable('');
	export let label: string = '';
	export let placeholder: string = '';
	export let disabled: boolean = false;
	let className = '';
	const id = `text-input-${++count}`;
	export { className as class };
</script>

<div class="flex flex-col gap-1">
	{#if label}
		<label for={id} class="block text-ctp-text font-semibold text-sm">{label}</label>
	{/if}
	<input
		{id}
		type="text"
		bind:value={$value}
		{placeholder}
		{disabled}
		class="h-8 min-w-8 w-full rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve shadow-md shadow-ctp-crust {className}"
	/>
</div>

