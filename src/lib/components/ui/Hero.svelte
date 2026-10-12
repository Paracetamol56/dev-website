<script lang="ts">
	import Heading from './Heading.svelte';
	import Text from './Text.svelte';
	import { TextColor, TextSize } from './typography';

	export let highlight: string;
	export let title: string = '';
	export let description: string = '';
	let className = 'mb-32';
	export { className as class };
</script>

<section class="relative isolate container mx-auto pt-8 text-center {className}">
	<div
		class="pointer-events-none absolute left-1/2 top-0 -z-10 h-72 w-[min(48rem,100%)] -translate-x-1/2
			rounded-full bg-gradient-to-r from-ctp-mauve/25 to-ctp-blue/20 blur-3xl"
	/>
	<Heading level={1} size={TextSize.Xl5} class="mb-6 sm:text-6xl lg:text-7xl">
		<span class="bg-gradient-to-r from-ctp-mauve to-ctp-lavender bg-clip-text text-transparent">
			{highlight}
		</span>
		{title}
	</Heading>
	{#if description}
		<Text size={TextSize.Lg} color={TextColor.Muted} class="mx-auto mb-8 max-w-2xl">
			{description}
		</Text>
	{/if}
	<slot />
</section>
