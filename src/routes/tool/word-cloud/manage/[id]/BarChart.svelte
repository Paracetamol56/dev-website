<script lang="ts">
	import * as d3 from 'd3';

	export let data: { text: string; occurence: number }[];

	const TOP = 30;
	const BAR_HEIGHT = 22;
	const margin = { top: 30, right: 24, bottom: 10, left: 120 };

	let showAll = false;
	let width = 0;

	$: shown = showAll ? data : data.slice(0, TOP);
	$: height = shown.length * BAR_HEIGHT + margin.top + margin.bottom;
	$: x = d3
		.scaleLinear()
		.domain([0, d3.max(shown, (d) => d.occurence) ?? 1])
		.nice()
		.range([margin.left, Math.max(margin.left + 1, width - margin.right)]);
	$: y = d3
		.scaleBand()
		.domain(shown.map((d) => d.text))
		.range([margin.top, height - margin.bottom])
		.padding(0.15);

	let left: SVGGElement;
	let top: SVGGElement;
	$: if (left) d3.select(left).call(d3.axisLeft(y).tickSizeOuter(0));
	$: if (top)
		d3.select(top).call(
			d3
				.axisTop(x)
				.tickValues(x.ticks(Math.max(2, Math.floor(width / 80))).filter(Number.isInteger))
				.tickFormat(d3.format('d'))
		);
</script>

<div class="w-full" bind:clientWidth={width}>
	{#if data.length === 0}
		<p class="py-4 text-center text-ctp-subtext0">No words yet</p>
	{:else}
		<svg class="block max-w-full text-xs text-ctp-subtext0" {width} {height}>
			<g bind:this={left} transform="translate({margin.left}, 0)" />
			<g bind:this={top} transform="translate(0, {margin.top})" />
			{#each shown as d (d.text)}
				<rect
					x={x(0)}
					y={y(d.text)}
					width={Math.max(0, x(d.occurence) - x(0))}
					height={y.bandwidth()}
					rx="3"
					class="fill-ctp-mauve"
				>
					<title>{d.text}: {d.occurence}</title>
				</rect>
			{/each}
		</svg>
		{#if data.length > TOP}
			<button
				type="button"
				class="mt-2 text-sm text-ctp-mauve hover:underline"
				on:click={() => (showAll = !showAll)}
			>
				{showAll ? `Show the top ${TOP}` : `Show all ${data.length} words`}
			</button>
		{/if}
	{/if}
</div>
