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
		<h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">Formula</h2>

		<div class="flex flex-row items-center gap-5">
			{@html display(`\\phi(x)=max(0, x)`)}
		</div>
	</div>

	<div class="flex-1" bind:clientWidth={containerWidth}>
		<h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">Result</h2>
		<Plot fx={relu} width={plotWidth} height={plotWidth * 2 / 3} />
	</div>
</section>

<section>
	<h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">Description</h2>

	<p><i>Rectified Linear Unit</i></p>

	<h3 class="mt-3 mb-2 font-semibold">Mathematical properties</h3>
	<ul class="mb-2 list-disc list-inside">
		<li>
			Continuous on {@html math(`\\mathbb{R}`)} and differentiable on {@html math(
				`\\mathbb{R}\\setminus\\{0\\}`
			)}
		</li>
		<li>Output range of {@html math(`[0,1]`)}</li>
	</ul>
	<h3 class="mt-3 mb-2">
		<strong>Application</strong>
	</h3>
	<p class="mb-2">
		Widely used in hidden layers of deep neural networks. It addresses the vanishing gradient
		problem to some extent by preventing gradients from becoming too small.
	</p>
	<h3 class="mt-3 mb-2 font-semibold">Pros</h3>
	<ul class="mb-2 list-disc list-inside">
		<li>Computationally efficient because of its simple linear formula</li>
		<li>Fast convergence during training for many cases</li>
	</ul>
	<h3 class="mt-3 mb-2 font-semibold">Cons</h3>
	<ul class="mb-2 list-disc list-inside">
		<li>
			Dead ReLU problem: Neurons can get stuck and never activate if their weights are not updated
			properly.
		</li>
		<li>Not zero-centered, which can lead to gradient issues in some cases.</li>
	</ul>
</section>
