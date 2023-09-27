<script lang="ts">
  import { createSlider, melt } from '@melt-ui/svelte';
  import { display, math } from 'mathlifier';
	import Plot from "./Plot.svelte";

  const {
    elements: { root, range, thumb },
    states: { value },
  } = createSlider({
    defaultValue: [0.1],
    min: 0,
    max: 1,
    step: 0.01,
  });

  let containerWidth: number;
  $: containerWidth;
  $: plotWidth = Math.min(500, containerWidth);

  $: leakyRelu = (x: number) => Math.max($value[0] * x, x);
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
      {@html display(`\\phi(x)=max(\\lambda x, x)`)}
    </div>
  </div>

  <div class="flex-1" bind:clientWidth={containerWidth}>
    <h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
      Result
    </h2>
    <Plot fx={leakyRelu} yDomain={[-1, 1]} width={plotWidth} height={plotWidth * 2 / 3} />
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
    <li>Similar to ReLU for {@html math(`x>0`)}, but with a small gradient for {@html math(`x<0`)}</li>
    <li>Output range of {@html math(`[-\\lambda,1]`)}</li>
  </ul>
  <h3 class="mt-3 mb-2">
    <strong>Application</strong>
  </h3>
  <p class="mb-2">
    Leaky ReLU addresses the dead ReLU problem by allowing a small gradient for negative inputs.
  </p>
  <h3 class="mt-3 mb-2 font-semibold">
    Pros
  </h3>
  <ul class="mb-2 list-disc list-inside">
    <li>Addresses the dead ReLU problem</li>
    <li>Efficient and easy to implement</li>
  </ul>
  <h3 class="mt-3 mb-2 font-semibold">
    Cons
  </h3>
  <ul class="mb-2 list-disc list-inside">
    <li>The choice of {@html math(`\\lambda`)} is important and can affect performance</li>
    <li>Not zero-centered.</li>
  </ul>
</section>
