<script lang="ts">
	import { TextColor, TextSize, TextWeight } from './typography';

	export let href: string;
	export let external: boolean = false;
	export let size: TextSize | '' = '';
	export let weight: TextWeight | '' = '';
	export let color: TextColor = TextColor.Link;
	let className = '';
	export { className as class };
</script>

<a
	{href}
	target={external ? '_blank' : undefined}
	rel={external ? 'noopener noreferrer' : undefined}
	class="{size} {weight} {color} rounded-sm hover:underline focus:outline-none focus-visible:underline {className}"
	{...$$restProps}
>
	<slot />
</a>
