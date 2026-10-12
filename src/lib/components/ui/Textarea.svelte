<script lang="ts">
	import { writable, type Writable } from 'svelte/store';
	import Field from './Field.svelte';
	import { controlClass, invalidClass, uid } from './styles';

	export let value: Writable<string> = writable('');
	export let rows: number = 5;
	export let label: string = '';
	export let hideLabel: boolean = false;
	export let error: string = '';
	export let hint: string = '';
	export let id: string = uid('textarea');
	let className = '';
	export { className as class };
</script>

<Field {id} {label} {hideLabel} {error} {hint}>
	<textarea
		{id}
		{rows}
		bind:value={$value}
		aria-invalid={error ? 'true' : undefined}
		aria-describedby={error || hint ? `${id}-message` : undefined}
		class="py-2 {controlClass} {error ? invalidClass : ''} {className}"
		on:input
		on:change
		on:blur
		{...$$restProps}
	/>
</Field>
