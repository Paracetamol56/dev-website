<script lang="ts">
	import { onMount } from 'svelte';
	import Heading from './Heading.svelte';
	import Text from './Text.svelte';
	import { TextColor, TextSize } from './typography';

	export let highlight: string;
	export let title: string = '';
	export let description: string = '';
	let className = 'mb-32';
	export { className as class };

	let backdrop: HTMLDivElement;
	let active = false;

	onMount(() => {
		if (!matchMedia('(hover: hover) and (prefers-reduced-motion: no-preference)').matches) return;

		const target = { x: 0, y: 0 };
		const current = { x: 0, y: 0 };
		let frame = 0;

		const follow = () => {
			current.x += (target.x - current.x) * 0.12;
			current.y += (target.y - current.y) * 0.12;
			backdrop.style.setProperty('--x', `${current.x}px`);
			backdrop.style.setProperty('--y', `${current.y}px`);
			const settled = Math.abs(target.x - current.x) + Math.abs(target.y - current.y) < 0.5;
			frame = settled ? 0 : requestAnimationFrame(follow);
		};
		const move = (event: MouseEvent) => {
			const rect = backdrop.getBoundingClientRect();
			target.x = event.clientX - rect.left;
			target.y = event.clientY - rect.top;
			if (!active) {
				current.x = target.x;
				current.y = target.y;
				active = true;
			}
			if (!frame) frame = requestAnimationFrame(follow);
		};
		const leave = () => (active = false);

		addEventListener('mousemove', move);
		document.documentElement.addEventListener('mouseleave', leave);
		return () => {
			removeEventListener('mousemove', move);
			document.documentElement.removeEventListener('mouseleave', leave);
			cancelAnimationFrame(frame);
		};
	});
</script>

<div
	bind:this={backdrop}
	class="backdrop pointer-events-none absolute inset-x-0 top-0 -z-10 h-[64rem] max-h-full"
>
	<div class="dots absolute inset-0 text-ctp-surface2 opacity-60" />
	<div
		class="dots glow absolute inset-0 text-ctp-mauve transition-opacity duration-500 {active
			? 'opacity-100'
			: 'opacity-0'}"
	/>
</div>
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

<style>
	.backdrop {
		mask-image: radial-gradient(ellipse 70% 60% at 50% 40%, black, transparent);
	}

	.dots {
		background-image: radial-gradient(currentColor 1px, transparent 1px);
		background-size: 28px 28px;
	}

	.glow {
		background-image: radial-gradient(currentColor 1.5px, transparent 1.5px);
		mask-image: radial-gradient(circle 14rem at var(--x) var(--y), black, transparent);
	}
</style>
