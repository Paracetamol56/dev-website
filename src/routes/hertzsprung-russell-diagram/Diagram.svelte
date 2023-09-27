<script lang="ts">
	import * as d3 from 'd3';
	import { bv2rgb, temperature } from './utils';
	import { onMount } from 'svelte';

	type Star = {
		Amag: number;
		BV: number;
		Plx: number;
		Vmag: number;
	};

	export let data: Star[];
	export let width = 928;
	export let height = Math.round(width * 1.2);
	export let margin = { top: 40, right: 40, bottom: 40, left: 40 };

	const x = d3.scaleLinear([-0.39, 2.19], [margin.left, width - margin.right]);
	const y = d3.scaleLinear([-9, 16], [margin.top, height - margin.bottom]);
	const z = bv2rgb;

	let g1: SVGGElement;
	let g2: SVGGElement;
	let g3: SVGGElement;
	let g4: SVGGElement;

  $: d3.select(g1).call(
    d3.axisLeft(
      d3.scaleLog(
        y.domain().map((m: any) => Math.pow(10, 4.83 - m)),
        y.range()
      )
    )
  );
  $: d3.select(g2).call(d3.axisRight(y).ticks(null, '+'));
  $: d3.select(g3).call(d3.axisTop(d3.scaleLinear(x.domain().map(temperature), x.range())));
  $: d3.select(g4).call(d3.axisBottom(x).ticks(null, '+f'));
</script>

<div class="container mx-auto">
	<svg
		id="plot"
		class="mx-auto bg-ctp-crust rounded-md text-ctp-blue fill-ctp-blue text-xs max-w-full"
		{width}
		{height}
		viewBox={`0 0 ${width} ${height}`}
	>
		<g bind:this={g1} transform="translate({margin.left}, 0)" />
		<g bind:this={g2} transform="translate({width - margin.right}, 0)" />
		<g bind:this={g3} transform="translate(0, {margin.top})" />
		<g bind:this={g4} transform="translate(0, {height - margin.bottom})" />
		<text dy="12" text-anchor="middle" transform="translate({margin.left}, {(margin.top + height - margin.bottom) / 2}) rotate(-90)">
			<tspan fill-opacity="0.8">← darker&nbsp;</tspan>
			<tspan font-weight="bold">&nbsp;Luminosity L☉&nbsp;</tspan>
			<tspan fill-opacity="0.8">&nbsp;brighter →</tspan>
		</text>
		<text dy="-6" text-anchor="middle" transform="translate({width - margin.right}, {(margin.top + height - margin.bottom) / 2}) rotate(-90)">
			<tspan fill-opacity="0.8">← darker&nbsp;</tspan>
			<tspan font-weight="bold">&nbsp;Absolute magnitude M&nbsp;</tspan>
			<tspan fill-opacity="0.8">&nbsp;brighter →</tspan>
		</text>
		<text x="464" y="40" dy="12" text-anchor="middle">
			<tspan fill-opacity="0.8">← hotter&nbsp; </tspan>
			<tspan font-weight="bold">&nbsp;Temperature K&nbsp;</tspan>
			<tspan fill-opacity="0.8">&nbsp;colder → </tspan>
		</text>
		<text x="464" y="1074" dy="-6" text-anchor="middle">
			<tspan fill-opacity="0.8">← blue&nbsp;</tspan>
			<tspan font-weight="bold">&nbsp;Color B-V&nbsp;</tspan>
			<tspan fill-opacity="0.8">&nbsp;red →</tspan>
		</text>
		<g>
			{#each data as star}
				<rect
					x={x(star.BV)}
					y={y(star.Amag)}
					fill={z(star.BV)}
					width="0.75"
					height="0.75"
				></rect>
			{/each}
		</g>
	</svg>
</div>
