
<script lang="ts">
  import * as d3 from "d3";

  export let data: Map<string, number> = new Map();
  export let margin: { top: number, right: number, bottom: number, left: number } = {
    top: 30,
    right: 30,
    bottom: 30,
    left: 30
  }

  let width: number;
  let height: number;

  $: x = d3.scaleBand<string>()
		.domain(data.keys())
		.range([margin.left, width - margin.right])
		.padding(0.1);
	
	$: y = d3.scaleLinear<number>()
		.domain([0, d3.max(data.values()) ?? 1])
		.range([height - margin.bottom, margin.top]);

  let g1: SVGGElement;
	let g2: SVGGElement;

  $: d3.select(g1).call(d3.axisLeft(y));
  $: d3.select(g2).call(d3.axisBottom(x));
</script>

<div
  class="w-full h-64 rounded-md bg-ctp-mantle"
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
      {#each y.ticks() as t}
        <line
          x1={0}
          x2={width - margin.left - margin.right}
          y1={y(t)}
          y2={y(t)}
          stroke="currentColor"
          stroke-width="1"
          stroke-dasharray="2 2"
          opacity="0.5"
        />
      {/each}
    </g>
    {#each data.entries() as d, i}
      <rect
        class="fill-ctp-blue"
        x={x(d[0])}
        y={y(d[1])}
        width={x.bandwidth()}
        height={height - margin.bottom - y(d[1])}
      />
      <text
        class="text-ctp-lavender"
        x={(x(d[0]) ?? 0) + x.bandwidth() / 2}
        y={(y(d[1]) ?? 0) - 3}
        text-anchor="middle"
        dominant-baseline="baseline"
      >
        {d[1].toFixed(2)}
      </text>
    {/each}
  </svg>
</div>
