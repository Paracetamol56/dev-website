<script lang="ts">
	import * as d3 from 'd3';

	export let data: Map<string, Map<string, number>> = new Map();
	export let margin: { top: number; right: number; bottom: number; left: number } = {
		top: 30,
		right: 30,
		bottom: 30,
		left: 30
	};

	let width: number;
	let height: number;

	$: x = d3
		.scaleBand<string>()
		.domain(data.keys())
		.range([margin.left, height - margin.right])
		.padding(0.1);

	$: y = d3
		.scaleBand<string>()
		.domain(data.keys())
		.range([margin.top, height - margin.bottom])
		.padding(0.1);

	// Get the max value in the matrix
	$: TPMax = Math.max(...Array.from(data.values()).map((d) => Math.max(...Array.from(d.values()))));
	// Get the max value in the matrix ignoring the diagonal
	$: LossMax = Math.max(
		...Array.from(data.values()).map((d, i) =>
			Math.max(...Array.from(d.values()).filter((e, j) => i !== j))
		)
	);

	$: colorLoss = d3.scaleSequentialSqrt(d3.interpolateReds).domain([0, LossMax]);
	$: scaleLoss = Object.assign(
		colorLoss.copy().interpolator(d3.interpolateRound(height - margin.bottom, margin.top)),
		{
			range() {
				return [height - margin.bottom, margin.top];
			}
		}
	);
	$: colorTP = d3.scaleSequentialSqrt(d3.interpolateBlues).domain([0, TPMax]);
	$: scaleTP = Object.assign(
		colorTP.copy().interpolator(d3.interpolateRound(height - margin.bottom, margin.top)),
		{
			range() {
				return [height - margin.bottom, margin.top];
			}
		}
	);

	// Axis
	let g1: SVGGElement;
	let g2: SVGGElement;
	// Legend
	let g3: SVGGElement;
	let g4: SVGGElement;

	$: d3.select(g1).call(d3.axisLeft(y));
	$: d3.select(g2).call(d3.axisTop(x));
	$: d3.select(g3).call(d3.axisRight(scaleLoss));
	$: d3.select(g4).call(d3.axisRight(scaleTP));
</script>

<div
	class="w-full h-64 rounded-md bg-ctp-mantle"
	bind:clientWidth={width}
	bind:clientHeight={height}
>
	<!-- Matrix -->
	<svg
		class="text-ctp-lavender fill-ctp-lavender text-xs max-w-full"
		{width}
		{height}
		viewBox={`0 0 ${width} ${height}`}
	>
		<g bind:this={g1} transform="translate({margin.left}, 0)" />
		<text
			x={height / 2}
			y={margin.top / 3}
			text-anchor="middle"
			dominant-baseline="middle"
			font-size="0.75rem"
		>
			Predicted
		</text>
		<g bind:this={g2} transform="translate(0, {margin.top})" />
		<text
			x={margin.left / 3}
			y={height / 2}
			transform={`rotate(-90, ${margin.left / 3}, ${height / 2})`}
			text-anchor="middle"
			dominant-baseline="middle"
			font-size="0.75rem"
		>
			Actual
		</text>
		{#each data.entries() as d, i}
			<g>
				{#each d[1].entries() as e, j}
					<rect
						x={x(d[0])}
						y={y(e[0])}
						width={x.bandwidth()}
						height={y.bandwidth()}
						fill={e[0] === d[0] ? colorTP(e[1]) : colorLoss(e[1])}
					/>
				{/each}
			</g>
		{/each}

		<!-- Legend -->
		<g transform={`translate(${height}, 0)`}>
			<defs>
				<linearGradient id="gradient-loss" x1="0%" y1="0%" x2="0%" y2="100%">
					<stop offset="0%" stop-color={colorLoss(LossMax)} />
					<stop offset="100%" stop-color={colorLoss(0)} />
				</linearGradient>
				<linearGradient id="gradient-tp" x1="0%" y1="0%" x2="0%" y2="100%">
					<stop offset="0%" stop-color={colorTP(TPMax)} />
					<stop offset="100%" stop-color={colorTP(0)} />
				</linearGradient>
			</defs>
			<g>
				<text
					x={0}
					y={(2 * margin.top) / 3}
					text-anchor="left"
					dominant-baseline="middle"
					font-size="0.75rem"
				>
					Loss
				</text>
				<rect
					x={0}
					y={margin.top}
					width={15}
					height={height - margin.bottom - margin.top}
					fill="url(#gradient-loss)"
				/>
				<g bind:this={g3} transform={`translate(15, 0)`} />
			</g>
			<g transform={`translate(40, 0)`}>
				<text
					x={0}
					y={(2 * margin.top) / 3}
					text-anchor="left"
					dominant-baseline="middle"
					font-size="0.75rem"
				>
					True Positive
				</text>
				<rect
					x={0}
					y={margin.top}
					width={15}
					height={height - margin.bottom - margin.top}
					fill="url(#gradient-tp)"
				/>
				<g bind:this={g4} transform={`translate(15, 0)`} />
			</g>
		</g>
	</svg>
</div>
