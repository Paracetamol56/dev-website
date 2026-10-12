<script lang="ts">
	import { labelClass } from './styles';
	import { TextColor, TextSize, TextWeight } from './typography';

	export let id: string;
	export let label: string = '';
	export let hideLabel: boolean = false;
	export let error: string = '';
	export let hint: string = '';
</script>

<div class="flex flex-col gap-1">
	{#if label}
		<label for={id} class="{labelClass} {hideLabel ? 'sr-only' : ''}">{label}</label>
	{/if}
	<slot />
	{#if error}
		<p id="{id}-message" class="{TextSize.Sm} {TextWeight.Semibold} {TextColor.Danger}">{error}</p>
	{:else if hint}
		<p id="{id}-message" class="{TextSize.Xs} {TextColor.Muted}">{hint}</p>
	{/if}
</div>
