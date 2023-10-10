<script lang="ts">
	import { display, math } from 'mathlifier';
	import Plot from './Plot.svelte';

	let containerWidth: number;
  $: containerWidth;
  $: plotWidth = Math.min(500, containerWidth);

	$: relu = (x: number) => Math.max(0, x);
</script>

<section class="flex flex-col gap-8 lg:flex-row lg:justify-between">
	<div class="flex-1">
		<h3 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">Formula</h3>

		<div class="flex flex-row items-center gap-5">
			{@html display(`\\phi(x)=max(0, x)`)}
		</div>
	</div>

	<div class="flex-1" bind:clientWidth={containerWidth}>
		<h3 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">Result</h3>
		<Plot fx={relu} width={plotWidth} height={plotWidth * 2 / 3} />
	</div>
</section>
