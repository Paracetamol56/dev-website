<script lang="ts">
  import { createSlider, melt } from '@melt-ui/svelte';
  import { display } from 'mathlifier';
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

  $: derivativeRelu = (x: number) => x > 0 ? 1 : 0;
</script>

<h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
  Options
</h2>

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
{@html display(`f(x)=\\left\\{\\begin{matrix}0 & \\textrm{if} \\quad x \\leq 0\\\\0 & \\textrm{if} \\quad x > 0\\end{matrix}\\right.`)}

<h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
  Result
</h2>
<Plot fx={derivativeRelu} />
