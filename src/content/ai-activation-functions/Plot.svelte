<script lang="ts">
	import { user } from '$lib/store';
	import { variants } from '@catppuccin/palette';
	import * as Plot from '@observablehq/plot';
	import * as d3 from 'd3';

	export let fx: (x: number) => number;
	export let xDomain: [number, number] = [-1, 1];
	export let yDomain: [number, number] = [0, 1];
	export let width: number = 500;
	export let height: number = 300;

	$: fxPts = d3.ticks(xDomain[0], xDomain[1], 200).map((x) => ({ x, ϕ: fx(x) }));
	$: derivPts = d3.ticks(xDomain[0], xDomain[1], 200).map((n) => ({ n, O: (fx(n + 0.01) - fx(n - 0.01)) / 0.02 }));

	let div: HTMLDivElement;
	$: {
		div?.firstChild?.remove();
		div?.append(
      Plot
      .plot({
				width,
				height,
				x: { domain: xDomain },
				y: { domain: yDomain },
        grid: true,
        marks: [
          Plot.lineY(fxPts, { x: "x", y: "ϕ", stroke: variants[$user.flavour].blue.rgb, strokeWidth: 2 }),
					Plot.lineY(derivPts, { x: "n", y: "O", stroke: variants[$user.flavour].red.rgb, strokeWidth: 1, strokeDasharray: "4 4" }),
          Plot.crosshairX(fxPts, { x: "x", y: "ϕ", ruleStrokeOpacity: 1, textStroke: "none", textFill: variants[$user.flavour].blue.rgb }),
        ],
      })
    );
	}
</script>

<div bind:this={div} class="flex justify-start" role="img" />
