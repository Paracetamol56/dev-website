<!-- src/SigmoidPlot.svelte -->
<script lang="ts">
  import * as d3 from "d3";

  export let fx: (x: number) => number;
  export let xDomain: [number, number] = [-1, 1];
  export let yDomain: [number, number] = [0, 1];
  export let width: number = 500;
  export let height: number = 300;
  let margin = { top: 20, right: 20, bottom: 30, left: 40 };

  let gx: SVGGElement;
  let gy: SVGGElement;

  $: x = d3.scaleLinear(xDomain, [margin.left, width - margin.right]);
  $: y = d3.scaleLinear(yDomain, [height - margin.bottom, margin.top]);
  $: line = d3
    .line()
    .x((d: any) => x(d.x))
    .y((d: any) => y(fx(d.x)));
  $: data = d3.range(-1, 1, 0.01).map((x: any) => ({ x }));
  $: d3.select(gy).call(d3.axisLeft(y));
  $: d3.select(gx).call(d3.axisBottom(x));
</script>

<svg
  id="plot"
  width={width}
  height={height}
>
  <g bind:this={gx} transform="translate(0, {height - margin.bottom})" />
  <g bind:this={gy} transform="translate({margin.left}, 0)" />
  <path
    class="fill-none stroke-ctp-mauve stroke-2"
    d={line(data)}
  />
</svg>
