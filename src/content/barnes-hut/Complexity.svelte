<script lang="ts">
	import { user } from '$lib/store';
	import { variants } from '@catppuccin/palette';
	import * as Plot from '@observablehq/plot';
	import * as d3 from 'd3';

	let div: HTMLDivElement;
	let OLinear = d3.ticks(0, 100, 100).map((n) => ({ n, O: n }));
  let ONSquared = d3.ticks(0, 100, 100).map((n) => ({ n, O: n ** 2 }));
  let ONlogN = d3.ticks(0, 100, 100).map((n) => ({ n, O: n * Math.log2(n) }));

  let data = d3.ticks(0, 100, 100).map((n) => ({ n, ON: n * Math.log2(n), ONSquared: n ** 2, ONlogN: n * Math.log2(n) }));

	$: {
		div?.firstChild?.remove();
		div?.append(
      Plot
      .plot({
        grid: true,
        marks: [
          Plot.lineY(OLinear, { x: "n", y: "O", stroke: variants[$user.flavour].green.rgb, strokeWidth: 2 }),
          Plot.lineY(ONSquared, { x: "n", y: "O", stroke: variants[$user.flavour].red.rgb, strokeWidth: 2 }),
          Plot.lineY(ONlogN, { x: "n", y: "O", stroke: variants[$user.flavour].blue.rgb, strokeWidth: 2 }),
        ],
      })
    );
	}
</script>

<div bind:this={div} class="flex justify-center" role="img" />
