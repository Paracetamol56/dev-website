<script lang="ts">
	import { TextAlign, TextColor, TextFont, TextSize, TextWeight } from './typography';

	export let as: 'p' | 'span' | 'div' | 'small' | 'strong' | 'em' | 'li' = 'p';
	export let size: TextSize = TextSize.Base;
	export let weight: TextWeight = TextWeight.Normal;
	export let color: TextColor = TextColor.Default;
	export let align: TextAlign | '' = '';
	export let font: TextFont = TextFont.Sans;
	let className = '';
	export { className as class };
</script>

<svelte:element
	this={as}
	class="{size} {weight} {color} {align} {font} {className}"
	{...$$restProps}
>
	<slot />
</svelte:element>
