<script lang="ts">
  import { onMount } from 'svelte';
  import * as d3 from 'd3';

  export let data: {text: string, occurence: number}[];
  export let margin: { top: number; right: number; bottom: number; left: number; } = {
    top: 40,
    right: 40,
    bottom: 40,
    left: 100
  };
  
  let containerWidth: number;
  let containerHeight: number;

  $: width = containerWidth;
  $: height = containerWidth;
  
  $: x = d3.scaleLinear()
    .domain([0, d3.max(data, (d: any) => d.occurence)])
    .range([margin.left, width - margin.right])
  $: y = d3.scaleBand()
    .domain(data.map(d => d.text))
    .range([margin.top, height - margin.bottom])
    .padding(0.1)

  let g1: SVGGElement;
  let g2: SVGGElement;

  $: d3.select(g1).call(d3.axisLeft(y).ticks(null));
  $: d3.select(g2).call(d3.axisTop(x).ticks(null));

</script>

<div class="mb-8 w-full p-4 bg-ctp-mantle rounded-md">
  <h2 class="text-2xl font-bold text-ctp-blue">
    Bar chart
  </h2>
</div>
<div
  class="mb-8 w-full"
  bind:clientWidth={containerWidth}
  bind:clientHeight={containerHeight}
>
  <svg
    id="chart"
    class="mx-auto fill-ctp-blue text-xs max-w-full"
    width={width}
    height={height}
    viewBox={`0 0 ${width} ${height}`}
  >
  <g bind:this={g1} transform="translate({margin.left}, 0)" />
  <g bind:this={g2} transform="translate(0, {margin.top})" />
  {#each data as d}
    <rect
      x={x(0)}
      y={y(d.text)}
      width={x(d.occurence) - x(0)}
      height={y.bandwidth()}
      class="fill-ctp-mauve"
      text-anchor="middle"
    >
    </rect>
  {/each}
  </svg>
</div>
