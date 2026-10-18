<script lang="ts">
	import type { Palette } from '$lib/colors';
	import * as d3 from 'd3';
	import { MONTH_NAMES, YEAR_COLORS, type Climate } from './era5';

	export let climate: Climate;
	export let from: number;
	export let to: number;
	/**
	 * Playhead in years: 1975.5 has drawn every year from `from` to 1974 and half of 1975.
	 * At `to + 1` or later, the whole range is drawn.
	 */
	export let progress: number;
	/** Show each value as its difference from the average of the range */
	export let relative: boolean;
	export let palette: Palette;
	export let unit = '°C';

	let width = 0;
	let svg: SVGSVGElement;
	let hover: { index: number; bin: number; x: number; y: number } | null = null;

	const MONTH_START = [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365];
	const TAU = Math.PI * 2;

	$: size = Math.min(width, 720);
	$: outer = size / 2 - 36;
	$: inner = outer * 0.16;

	$: range = climate.years.filter(({ year }) => year >= from && year <= to);
	$: average = climate.bins.map((_, i) => d3.mean(range, ({ values }) => values[i] ?? undefined));
	$: shift = (values: (number | null)[]) =>
		values.map((value, i) => {
			if (value === null) return null;
			if (!relative) return value;
			const reference = average[i];
			return reference === undefined ? null : value - reference;
		});
	$: years = range.map(({ year, values }) => ({ year, values: shift(values) }));
	// Years before the range stay in the background for comparison
	$: earlier = climate.years
		.filter(({ year }) => year < from)
		.map(({ year, values }) => ({ year, values: shift(values) }));

	// The scale covers every year drawn, so it stays still while the timeline plays
	$: extent = d3.extent(
		[...earlier, ...years].flatMap((year) => year.values).concat(relative ? [0] : []),
		(value) => value ?? undefined
	) as [number, number];
	$: radius = d3
		.scaleLinear()
		.domain(extent[0] === undefined ? [0, 1] : extent)
		.nice()
		.range([inner, outer]);
	$: rings = radius.ticks(5);

	// Oldest to newest along blue, teal, green and yellow, a viridis-like scale in the theme's colours
	$: ramp = d3.piecewise(
		d3.interpolateLab,
		YEAR_COLORS.map((name) => palette[name].hex)
	);
	$: colorOf = (index: number) => ramp(years.length > 1 ? index / (years.length - 1) : 1);

	// Each year runs on into the next one's first spoke, so the years form one continuous spiral
	$: line = d3
		.lineRadial<number | null>()
		.defined((value) => value !== null)
		.angle(
			(_, i) =>
				(i === climate.bins.length ? 1 + climate.bins[0].angle : climate.bins[i].angle) * TAU
		)
		.radius((value) => radius(value as number))
		.curve(d3.curveCatmullRom);
	$: paths = years.map(({ values }, i) => line([...values, years[i + 1]?.values[0] ?? null]));
	$: earlierPaths = earlier.map(({ values }, i) =>
		line([...values, (earlier[i + 1] ?? years[0])?.values[0] ?? null])
	);

	$: done = progress >= to + 1;
	$: currentIndex = done ? years.length - 1 : Math.floor(progress) - from;
	$: current = years[currentIndex];
	$: fraction = done ? 1 : progress - Math.floor(progress);

	// The year being drawn runs up to the playhead, ending on a point between two spokes
	$: head = (() => {
		if (done || !current) return null;
		const bins = climate.bins;
		const points: [number, number][] = [];
		for (let i = 0; i < bins.length && bins[i].angle <= fraction; i++) {
			const value = current.values[i];
			if (value === null) break;
			points.push([bins[i].angle, value]);
			const next = bins[i + 1];
			const nextValue = current.values[i + 1];
			if (next && next.angle > fraction && nextValue != null) {
				const t = (fraction - bins[i].angle) / (next.angle - bins[i].angle);
				points.push([fraction, value + (nextValue - value) * t]);
			}
		}
		return { points, tip: points.at(-1) ?? null };
	})();
	$: headPath = head
		? d3
				.lineRadial<[number, number]>()
				.angle(([angle]) => angle * TAU)
				.radius(([, value]) => radius(value))
				.curve(d3.curveCatmullRom)(head.points)
		: null;

	const point = (angle: number, r: number): [number, number] => [
		Math.sin(angle * TAU) * r,
		-Math.cos(angle * TAU) * r
	];

	$: hovered = hover ? years[hover.index] : null;
	$: hoveredValue = hover && hovered ? hovered.values[hover.bin] : null;
	const signed = (value: number) => `${value > 0 ? '+' : ''}${value.toFixed(1)} ${unit}`;
	$: format = relative ? signed : (value: number) => `${value.toFixed(1)} ${unit}`;

	function move(event: PointerEvent) {
		const rect = svg.getBoundingClientRect();
		const x = event.clientX - rect.left - rect.width / 2;
		const y = event.clientY - rect.top - rect.height / 2;
		const r = Math.hypot(x, y);
		if (r > outer + 12) return (hover = null);
		const angle = (((Math.atan2(x, -y) / TAU) % 1) + 1) % 1;
		const distance = (a: number) => Math.min(Math.abs(a - angle), 1 - Math.abs(a - angle));
		let bin = 0;
		climate.bins.forEach(({ angle }, i) => {
			if (distance(angle) < distance(climate.bins[bin].angle)) bin = i;
		});
		let index = -1;
		let best = Infinity;
		years.slice(0, done ? years.length : currentIndex).forEach(({ values }, i) => {
			const value = values[bin];
			if (value === null) return;
			const gap = Math.abs(radius(value) - r);
			if (gap < best) {
				best = gap;
				index = i;
			}
		});
		hover = index < 0 ? null : { index, bin, x: x + rect.width / 2, y: y + rect.height / 2 };
	}
</script>

<div class="relative mx-auto w-full" style="max-width: 720px" bind:clientWidth={width}>
	{#if size > 0}
		<svg
			bind:this={svg}
			width={size}
			height={size}
			viewBox="{-size / 2} {-size / 2} {size} {size}"
			role="img"
			aria-label="Radar chart of temperature through the year, one line per year from {from} to {current?.year ??
				to}"
			class="block touch-none select-none"
			on:pointermove={move}
			on:pointerleave={() => (hover = null)}
		>
			<g fill="none" stroke={palette.surface1.hex}>
				{#each rings as ring}
					<circle
						r={radius(ring)}
						stroke={relative && ring === 0 ? palette.overlay1.hex : undefined}
						stroke-width={relative && ring === 0 ? 1.5 : 1}
					/>
				{/each}
				{#each MONTH_START.slice(0, 12) as start}
					<line
						x1={point(start / 365, inner)[0]}
						y1={point(start / 365, inner)[1]}
						x2={point(start / 365, outer)[0]}
						y2={point(start / 365, outer)[1]}
					/>
				{/each}
			</g>

			{#each MONTH_NAMES as name, m}
				{@const [x, y] = point((MONTH_START[m] + MONTH_START[m + 1]) / 2 / 365, outer + 18)}
				<text
					{x}
					{y}
					text-anchor="middle"
					dominant-baseline="middle"
					class="text-xs"
					fill={palette.subtext0.hex}>{name}</text
				>
			{/each}

			<g fill="none" stroke-linejoin="round" stroke-linecap="round">
				<g stroke={palette.overlay0.hex} stroke-width="1" opacity={hover ? 0.05 : 0.15}>
					{#each earlierPaths as d}
						<path {d} />
					{/each}
				</g>
				{#each paths.slice(0, done ? years.length - 1 : Math.max(currentIndex, 0)) as d, i (years[i].year)}
					<path
						{d}
						stroke={colorOf(i)}
						stroke-width={hover?.index === i ? 2.5 : 1}
						opacity={hover && hover.index !== i ? 0.2 : 0.8}
					/>
				{/each}
				{#if done && current}
					<path d={paths[currentIndex]} stroke={palette.mantle.hex} stroke-width="6" />
					<path
						d={paths[currentIndex]}
						stroke={palette.text.hex}
						stroke-width="3"
						opacity={hover && hover.index !== currentIndex ? 0.4 : 1}
					/>
				{:else if headPath}
					<path d={headPath} stroke={palette.mantle.hex} stroke-width="6" />
					<path d={headPath} stroke={palette.text.hex} stroke-width="3" />
				{/if}
			</g>

			{#if head?.tip}
				{@const [x, y] = point(head.tip[0], radius(head.tip[1]))}
				<circle
					cx={x}
					cy={y}
					r="5"
					fill={palette.text.hex}
					stroke={palette.mantle.hex}
					stroke-width="2"
				/>
			{/if}

			{#each rings as ring}
				<text
					x={4}
					y={-radius(ring) - 3}
					class="text-[10px]"
					fill={palette.subtext0.hex}
					paint-order="stroke"
					stroke={palette.mantle.hex}
					stroke-width="3">{format(ring)}</text
				>
			{/each}

			<text
				text-anchor="middle"
				dominant-baseline="central"
				class="text-xl font-bold tabular-nums"
				fill={palette.text.hex}>{current?.year ?? to}</text
			>

			{#if hover && hoveredValue !== null}
				{@const angle = climate.bins[hover.bin].angle}
				{@const [x, y] = point(angle, radius(hoveredValue))}
				<line
					x1={point(angle, inner)[0]}
					y1={point(angle, inner)[1]}
					x2={point(angle, outer)[0]}
					y2={point(angle, outer)[1]}
					stroke={palette.overlay1.hex}
					stroke-dasharray="2 3"
				/>
				<circle
					cx={x}
					cy={y}
					r="5"
					fill={hover.index === currentIndex ? palette.text.hex : colorOf(hover.index)}
					stroke={palette.mantle.hex}
					stroke-width="2"
				/>
			{/if}
		</svg>

		{#if hover && hovered && hoveredValue !== null}
			<div
				class="pointer-events-none absolute z-10 rounded-md bg-ctp-crust px-3 py-2 text-sm shadow-md shadow-ctp-crust"
				style="left: {hover.x}px; top: {hover.y}px; transform: translate({hover.x > size / 2
					? 'calc(-100% - 12px)'
					: '12px'}, -50%)"
			>
				<p class="content-ignore font-semibold text-ctp-text">{hovered.year}</p>
				<p class="content-ignore text-ctp-subtext0">{climate.bins[hover.bin].label}</p>
				<p class="content-ignore font-semibold text-ctp-text">{format(hoveredValue)}</p>
				{#if relative}
					<p class="content-ignore text-xs text-ctp-subtext0">vs {from}–{to} average</p>
				{/if}
			</div>
		{/if}
	{/if}
</div>
