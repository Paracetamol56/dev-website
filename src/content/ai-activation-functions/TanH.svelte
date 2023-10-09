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

  $: tanh = (x: number) => (1 - Math.exp(- $value[0] * x)) / (1 + Math.exp(- $value[0] * x));
</script>

<section class="flex flex-col gap-8 lg:flex-row lg:justify-between">
  <div class="flex-1">
    <h3 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
      Formula
    </h3>

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
      {@html display(`\\phi(x)=\\frac{1-e^{-\\lambda x}}{1+e^{-\\lambda x}}`)}
    </div>
  </div>

  <div class="flex-1" bind:clientWidth={containerWidth}>
    <h3 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
      Result
    </h3>
    <Plot fx={tanh} yDomain={[-1, 1]} width={plotWidth} height={plotWidth * 2 / 3} />
  </div>
</section>
