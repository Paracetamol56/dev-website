<script lang="ts">
	import { TextAlign, TextColor, TextSize, TextWeight } from './typography';

	const sizes: Record<number, TextSize> = {
		1: TextSize.Xl6,
		2: TextSize.Xl4,
		3: TextSize.Xl2,
		4: TextSize.Xl,
		5: TextSize.Lg,
		6: TextSize.Base
	};

	export let level: 1 | 2 | 3 | 4 | 5 | 6 = 2;
	export let size: TextSize | undefined = undefined;
	export let weight: TextWeight = TextWeight.Bold;
	export let color: TextColor = TextColor.Default;
	export let align: TextAlign | '' = '';
	export let gradient: boolean = false;
	let className = '';
	export { className as class };
</script>

<svelte:element
	this={`h${level}`}
	class="{size ?? sizes[level]} {weight} {color} {align} {className}"
	{...$$restProps}
>
	{#if gradient}
		<span class="bg-gradient-to-r from-ctp-mauve to-ctp-lavender bg-clip-text text-transparent">
			<slot />
		</span>
	{:else}
		<slot />
	{/if}
</svelte:element>
