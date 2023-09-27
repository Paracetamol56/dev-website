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

  $: sigmoid = (x: number) => 1 / (1 + Math.exp(- $value[0] * x));
</script>

<section class="flex flex-col gap-8 lg:flex-row lg:justify-between">
  <div class="flex-1">
    <h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
      Formula
    </h2>

    <div class="flex flex-row items-center gap-5">
      <div>
        <p>Lambda</p>
        <span use:melt={$root} class="relative flex h-[20px] w-[200px] items-center">
          <span class="block h-1 w-full rounded bg-ctp-mauve/20">
            <span use:melt={$range} class="h-1 rounded bg-ctp-mauve" />
          </span>
          <span
            use:melt={$thumb()}
            class="block square-3 rounded-full outline-none bg-ctp-mauve focus:ring-4 focus:ring-ctp-mauve/20"
          />
        </span>
        <p>{@html math(`\\lambda=${$value[0]}`)}</p>
      </div>
      {@html display(`\\phi(x)=\\frac{1}{1+e^{\\lambda*x}}`)}
    </div>
  </div>

  <div class="flex-1" bind:clientWidth={containerWidth}>
    <h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
      Result
    </h2>
    <Plot fx={sigmoid} width={plotWidth} height={plotWidth * 2 / 3} />
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
    <li>Continuous and differentiable on {@html math(`\\mathbb{R}`)}</li>
    <li>Output range of {@html math(`[0,1]`)}</li>
  </ul>
  <h3 class="mt-3 mb-2">
    <strong>Application</strong>
  </h3>
  <p class="mb-2">
    Commonly used in the output layer of a binary classifier.
    It is also sensitive to changes in the input, which makes it useful for backpropagation.
  </p>
  <h3 class="mt-3 mb-2 font-semibold">
    Pros
  </h3>
  <ul class="mb-2 list-disc list-inside">
    <li>Smooth gradient, which helps in stable training</li>
    <li>Outputs are in a bounded range, making it suitable for certain types of problems</li>
  </ul>
  <h3 class="mt-3 mb-2 font-semibold">
    Cons
  </h3>
  <ul class="mb-2 list-disc list-inside">
    <li>Vanishing gradient problem: Gradients become very small for extreme inputs, which can slow down training</li>
    <li>Outputs are not centered around zero, which can lead to slower convergence in some cases</li>
  </ul>
</section>
