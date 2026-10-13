<script lang="ts">
	import { writable, type Writable } from 'svelte/store';
	import Field from './Field.svelte';
	import { controlClass, invalidClass, uid } from './styles';

	export let value: Writable<string> = writable('');
	export let type: 'text' | 'email' | 'password' | 'search' | 'url' | 'tel' = 'text';
	export let label: string = '';
	export let hideLabel: boolean = false;
	export let error: string = '';
	export let hint: string = '';
	export let id: string = uid('text-input');
	let className = '';
	export { className as class };
</script>

<Field {id} {label} {hideLabel} {error} {hint}>
	<input
		{id}
		{type}
		value={$value}
		autocomplete="off"
		aria-invalid={error ? 'true' : undefined}
		aria-describedby={error || hint ? `${id}-message` : undefined}
		class="h-8 {controlClass} {error ? invalidClass : ''} {className}"
		on:input={(e) => value.set(e.currentTarget.value)}
		on:input
		on:change
		on:blur
		on:keydown
		{...$$restProps}
	/>
</Field>
