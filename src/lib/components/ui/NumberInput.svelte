<script lang="ts">
	import { writable, type Writable } from 'svelte/store';
	import Field from './Field.svelte';
	import { controlClass, invalidClass, uid } from './styles';

	export let value: Writable<number> = writable(0);
	export let min: number = 0;
	export let max: number = 100;
	export let step: number = 1;
	export let label: string = '';
	export let hideLabel: boolean = false;
	export let error: string = '';
	export let hint: string = '';
	export let id: string = uid('number-input');
	let className = '';
	export { className as class };
</script>

<Field {id} {label} {hideLabel} {error} {hint}>
	<input
		{id}
		type="number"
		{min}
		{max}
		{step}
		bind:value={$value}
		aria-invalid={error ? 'true' : undefined}
		aria-describedby={error || hint ? `${id}-message` : undefined}
		class="h-8 [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none
			{controlClass} {error ? invalidClass : ''} {className}"
		on:input
		on:change
		on:blur
		{...$$restProps}
	/>
</Field>
