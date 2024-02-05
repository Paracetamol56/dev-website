<script lang="ts">
	import * as d3 from 'd3';
	import { bv2rgb, temperature, type Star, type StarHR } from './utils';

	export let data: StarHR[];
	export let selected: StarHR | null = null;
	export let margin = { top: 40, right: 40, bottom: 40, left: 40 };

	let containerWidth: number;
	let containerHeight: number;
	$: height = Math.min(containerHeight, containerWidth * 1.2);
	$: width = Math.min(containerWidth, height / 1.2);

	$: x = d3.scaleLinear([-0.39, 2.19], [margin.left, width - margin.right]);
	$: y = d3.scaleLinear([-9, 16], [margin.top, height - margin.bottom]);
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

<div
	class="relative container h-screen mx-auto"
	bind:clientWidth={containerWidth}
	bind:clientHeight={containerHeight}
>
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
		<text
			dy="12"
			text-anchor="middle"
			transform="translate({margin.left}, {(margin.top + height - margin.bottom) / 2}) rotate(-90)"
		>
			<tspan fill-opacity="0.8">← darker&nbsp;</tspan>
			<tspan font-weight="bold">&nbsp;Luminosity L☉&nbsp;</tspan>
			<tspan fill-opacity="0.8">&nbsp;brighter →</tspan>
		</text>
		<text
			dy="-6"
			text-anchor="middle"
			transform="translate({width - margin.right}, {(margin.top + height - margin.bottom) /
				2}) rotate(-90)"
		>
			<tspan fill-opacity="0.8">← darker&nbsp;</tspan>
			<tspan font-weight="bold">&nbsp;Absolute magnitude M&nbsp;</tspan>
			<tspan fill-opacity="0.8">&nbsp;brighter →</tspan>
		</text>
		<text
			dy="12"
			text-anchor="middle"
			transform="translate({(margin.left + width - margin.right) / 2}, {margin.top})"
		>
			<tspan fill-opacity="0.8">← hotter&nbsp; </tspan>
			<tspan font-weight="bold">&nbsp;Temperature K&nbsp;</tspan>
			<tspan fill-opacity="0.8">&nbsp;colder → </tspan>
		</text>
		<text
			dy="-6"
			text-anchor="middle"
			transform="translate({(margin.left + width - margin.right) / 2}, {height - margin.bottom})"
		>
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
					opacity="0.8"
					width="0.75"
					height="0.75"
				/>
			{/each}
		</g>
		{#if selected}
			<line
				x1={margin.left}
				y1={y(selected.Amag)}
				x2={width - margin.right}
				y2={y(selected.Amag)}
				stroke="currentColor"
				stroke-width="1"
			/>
			<line
				x1={x(selected.BV)}
				y1={margin.top}
				x2={x(selected.BV)}
				y2={height - margin.bottom}
				stroke="currentColor"
				stroke-width="1"
			/>
		{/if}
	</svg>
	<p class="text-center text-xs text-ctp-blue">
		{data.length} stars loaded
	</p>
</div>
