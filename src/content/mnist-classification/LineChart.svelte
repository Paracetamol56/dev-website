
<script lang="ts">
  import * as d3 from "d3";
  import type * as tf from "@tensorflow/tfjs";

  export let data: number[] = [];
  export let maxY: number = 10;
  export let title: string = "Loss";
  export let margin: { top: number, right: number, bottom: number, left: number } = {
    top: 30,
    right: 30,
    bottom: 30,
    left: 30
  }

  let width: number;
  let height: number;

  $: x = d3.scaleLinear([0, data.length], [margin.left, width - margin.right]);
  $: y = d3.scaleLinear([0, maxY], [height - margin.bottom, margin.top]);

  const line = d3.line<number>()
    .x((d, i) => x(i))
    .y((d) => y(d));

  let g1: SVGGElement;
	let g2: SVGGElement;

  $: d3.select(g1).call(d3.axisLeft(y));
  $: d3.select(g2).call(d3.axisBottom(x));
</script>

<div class="flex flex-col items-center w-full rounded-md bg-ctp-mantle">
  <h4 class="m-0">{title}</h4>
  <div
    class="w-full h-64"
    bind:clientWidth={width}
    bind:clientHeight={height}
  >
    <svg
      class="text-ctp-lavender fill-ctp-lavender text-xs max-w-full"
      {width}
      {height}
      viewBox={`0 0 ${width} ${height}`}
    >
      <g bind:this={g1} transform="translate({margin.left}, 0)" />
      <g bind:this={g2} transform="translate(0, {height - margin.bottom})" />
      <g transform="translate({margin.left}, 0)">
        {#each Array(10).fill(0) as _, i}
          <line
            x1={0}
            x2={width - margin.left - margin.right}
            y1={y(i)}
            y2={y(i)}
            stroke="currentColor"
            stroke-width="1"
            stroke-dasharray="2 2"
            opacity="0.5"
          />
        {/each}
      </g>
      <path
        class="stroke-ctp-blue fill-none stroke-1"
        stroke-width="1"
        d={line(data)}
      />
    </svg>
  </div>
</div>
