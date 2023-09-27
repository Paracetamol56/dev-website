<script lang="ts">
  import { createSlider, melt } from '@melt-ui/svelte';
  import { display, math } from 'mathlifier';
	import Plot from "./Plot.svelte";

  const {
    elements: { root, range, thumb },
    states: { value },
  } = createSlider({
    defaultValue: [10],
    min: 0,
    max: 20,
    step: 0.1,
  });

  let containerWidth: number;
  $: containerWidth;
  $: plotWidth = Math.min(500, containerWidth);

  $: derivativeRelu = (x: number) => x > 0 ? 1 : 0;
</script>

<section class="flex flex-col gap-8 lg:flex-row lg:justify-between">
  <div class="flex-1">
    <h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
      Formula
    </h2>
    <div class="flex flex-row items-center gap-5">
      {@html display(`f(x)=\\left\\{\\begin{matrix}0 & \\textrm{if} \\quad x \\leq 0\\\\0 & \\textrm{if} \\quad x > 0\\end{matrix}\\right.`)}
    </div>
  </div>

  <div class="flex-1" bind:clientWidth={containerWidth}>
    <h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
      Result
    </h2>
    <Plot fx={derivativeRelu} width={plotWidth} height={plotWidth * 2 / 3} />
  </div>
</section>

<section>
  <h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
    Description
  </h2>

  <h3 class="mt-3 mb-2 font-semibold">
    Mathematical properties
  </h3>
  <ul class="mb-2 list-disc list-inside">
    <li>Continuous on {@html math(`\\mathbb{R}`)} and differentiable on {@html math(`\\mathbb{R}\\setminus\\{0\\}`)}</li>
    <li>Output range of {@html math(`[0,1]`)}</li>
  </ul>
  <h3 class="mt-3 mb-2">
    <strong>Application</strong>
  </h3>
  <p class="mb-2">
    
  </p>
</section>
