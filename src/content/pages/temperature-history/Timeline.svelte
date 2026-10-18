<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import type { Palette } from '$lib/colors';
	import * as d3 from 'd3';
	import { Pause, Play } from 'lucide-svelte';
	import { onDestroy } from 'svelte';
	import { MONTH_NAMES, type Year } from './era5';

	/** Monthly values, one entry per year */
	export let months: Year[];
	export let from: number;
	export let to: number;
	/** Playhead in years; past `to` the whole range is shown */
	export let progress: number;
	/** Milliseconds to play one year */
	export let duration: number;
	export let palette: Palette;
	export let unit = '°C';

	const HEIGHT = 96;
	const AXIS = 20;
	const MARGIN = 10;

	let width = 0;
	let svg: SVGSVGElement;
	let frame: number | null = null;
	let dragging: 'from' | 'to' | 'seek' | null = null;
	let hover: number | null = null;

	$: first = months[0]?.year ?? from;
	$: last = months.at(-1)?.year ?? to;
	$: x = d3
		.scaleLinear()
		.domain([first, last + 1])
		.range([MARGIN, Math.max(width - MARGIN, MARGIN + 1)]);

	// Each month against the average of the same month over every year, so the seasons cancel out
	$: normals = MONTH_NAMES.map((_, m) => d3.mean(months, ({ values }) => values[m] ?? undefined));
	$: series = months.flatMap(({ year, values }) =>
		values.map((value, m) => {
			const normal = normals[m];
			return {
				time: year + m / 12,
				label: `${MONTH_NAMES[m]} ${year}`,
				value,
				anomaly: value === null || normal === undefined ? null : value - normal
			};
		})
	);
	$: rolling = series.map((_, i) => {
		const window = series.slice(Math.max(0, i - 11), i + 1);
		if (window.length < 12 || window.some(({ anomaly }) => anomaly === null)) return null;
		return d3.mean(window, ({ anomaly }) => anomaly as number) as number;
	});

	// The rare extreme months are clipped so the ordinary ones, and the trend, stay readable
	$: magnitudes = series.flatMap(({ anomaly }) => (anomaly === null ? [] : [Math.abs(anomaly)]));
	$: spread = d3.quantile(magnitudes, 0.98) || 1;
	$: y = d3
		.scaleLinear()
		.domain([-spread, spread])
		.range([HEIGHT - 4, 4])
		.clamp(true);
	$: color = d3
		.scaleDiverging<string>()
		.domain([-spread * 0.6, 0, spread * 0.6])
		.interpolator(
			d3.piecewise(d3.interpolateLab, [palette.blue.hex, palette.overlay0.hex, palette.red.hex])
		)
		.clamp(true);
	$: trend = d3
		.line<number | null>()
		.defined((value) => value !== null)
		.x((_, i) => x(series[i].time + 1 / 24))
		.y((value) => y(value as number))(rolling);
	$: decades = d3.range(Math.ceil(first / 10) * 10, last + 1, 10);
	$: barWidth = Math.max(x(first + 1 / 12) - x(first), 1);

	$: playing = frame !== null;
	$: done = progress >= to + 1;

	function stop() {
		if (frame !== null) cancelAnimationFrame(frame);
		frame = null;
	}

	function toggle() {
		if (playing) return stop();
		if (progress >= to + 1) progress = from;
		let previous = performance.now();
		const tick = (now: number) => {
			progress = Math.min(progress + (now - previous) / duration, to + 1);
			previous = now;
			frame = progress >= to + 1 ? null : requestAnimationFrame(tick);
		};
		frame = requestAnimationFrame(tick);
	}

	// New data, or a new range, is shown in full
	$: months, stop(), (progress = Infinity);
	onDestroy(stop);

	function setRange(start: number, end: number) {
		const nextFrom = Math.max(first, Math.min(start, to));
		const nextTo = Math.min(last, Math.max(end, nextFrom));
		if (nextFrom === from && nextTo === to) return;
		[from, to] = [nextFrom, nextTo];
		stop();
		progress = Infinity;
	}

	const timeAt = (event: PointerEvent) =>
		x.invert(event.clientX - svg.getBoundingClientRect().left);

	function seek(time: number) {
		progress = time >= to + 1 ? Infinity : Math.max(time, from);
	}

	function down(event: PointerEvent) {
		const handle = (event.target as Element).closest('[data-handle]')?.getAttribute('data-handle');
		dragging = handle === 'from' || handle === 'to' ? handle : 'seek';
		svg.setPointerCapture(event.pointerId);
		if (dragging === 'seek') seek(timeAt(event));
	}

	function move(event: PointerEvent) {
		const time = timeAt(event);
		if (dragging === 'from') setRange(Math.round(time), to);
		else if (dragging === 'to') setRange(from, Math.round(time) - 1);
		else if (dragging === 'seek') seek(time);
		const index = Math.floor((time - first) * 12);
		hover = dragging || index < 0 || index >= series.length ? null : index;
	}

	function key(handle: 'from' | 'to', event: KeyboardEvent) {
		const steps: Record<string, number> = {
			ArrowLeft: -1,
			ArrowDown: -1,
			ArrowRight: 1,
			ArrowUp: 1,
			PageDown: -10,
			PageUp: 10,
			Home: -Infinity,
			End: Infinity
		};
		const step = steps[event.key];
		if (step === undefined) return;
		event.preventDefault();
		const clamp = (year: number) => Math.max(first, Math.min(last, year));
		if (handle === 'from') setRange(clamp(from + step), to);
		else setRange(from, clamp(to + step));
	}

	$: hovered = hover === null ? null : series[hover];
	const signed = (value: number) => `${value > 0 ? '+' : ''}${value.toFixed(1)} ${unit}`;
</script>

<div class="flex flex-col gap-2">
	<div class="flex flex-wrap items-center gap-x-4 gap-y-2">
		<Button on:click={toggle} class="gap-2">
			{#if playing}
				<Pause size="16" /> Pause
			{:else}
				<Play size="16" /> Play
			{/if}
		</Button>
		<p class="content-ignore text-sm">
			<span class="font-semibold text-ctp-text">Timeline</span>
			<span class="ml-1 text-ctp-mauve tabular-nums">{from} – {to}</span>
		</p>
		<div class="ml-auto flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-ctp-subtext0">
			<span class="flex items-center gap-2">
				<span
					class="block h-2 w-16 rounded-full"
					style="background: linear-gradient(to right, {palette.blue.hex}, {palette.overlay0
						.hex}, {palette.red.hex})"
				/>
				Each month vs its {first}–{last} average
			</span>
			<span class="flex items-center gap-2">
				<span class="block h-[2px] w-6 bg-ctp-text" />
				12-month average
			</span>
		</div>
	</div>

	<div class="relative" bind:clientWidth={width}>
		{#if width > 0}
			<svg
				bind:this={svg}
				{width}
				height={HEIGHT + AXIS}
				class="block touch-none select-none {dragging === 'seek'
					? 'cursor-grabbing'
					: 'cursor-pointer'}"
				role="presentation"
				on:pointerdown={down}
				on:pointermove={move}
				on:pointerup={() => (dragging = null)}
				on:pointercancel={() => (dragging = null)}
				on:pointerleave={() => (hover = null)}
			>
				<rect
					x={MARGIN}
					width={width - MARGIN * 2}
					height={HEIGHT}
					rx="4"
					fill={palette.base.hex}
				/>

				<g shape-rendering="crispEdges">
					{#each series as { time, anomaly }}
						{#if anomaly !== null}
							<rect
								x={x(time)}
								width={barWidth}
								y={Math.min(y(anomaly), y(0))}
								height={Math.abs(y(anomaly) - y(0))}
								fill={color(anomaly)}
							/>
						{/if}
					{/each}
				</g>
				<line x1={MARGIN} x2={width - MARGIN} y1={y(0)} y2={y(0)} stroke={palette.overlay0.hex} />
				<path d={trend} fill="none" stroke={palette.text.hex} stroke-width="1.5" />

				<!-- Years outside the range are dimmed -->
				<g fill={palette.mantle.hex} opacity="0.75">
					<rect x={MARGIN} width={Math.max(x(from) - MARGIN, 0)} height={HEIGHT} />
					<rect x={x(to + 1)} width={Math.max(width - MARGIN - x(to + 1), 0)} height={HEIGHT} />
				</g>

				{#each decades as decade}
					<text
						x={x(decade)}
						y={HEIGHT + 14}
						text-anchor="middle"
						class="text-[10px]"
						fill={palette.subtext0.hex}>{decade}</text
					>
				{/each}

				{#if hovered}
					<line
						x1={x(hovered.time + 1 / 24)}
						x2={x(hovered.time + 1 / 24)}
						y2={HEIGHT}
						stroke={palette.overlay2.hex}
						stroke-dasharray="2 3"
					/>
				{/if}

				{#if !done}
					<line
						x1={x(progress)}
						x2={x(progress)}
						y1="-2"
						y2={HEIGHT + 2}
						stroke={palette.text.hex}
						stroke-width="2"
					/>
				{/if}

				{#each [{ handle: 'from', year: from, at: x(from) }, { handle: 'to', year: to, at: x(to + 1) }] as { handle, year, at } (handle)}
					<g
						data-handle={handle}
						role="slider"
						tabindex="0"
						aria-label="Timeline {handle === 'from' ? 'start' : 'end'}"
						aria-valuemin={first}
						aria-valuemax={last}
						aria-valuenow={year}
						class="cursor-ew-resize outline-none focus-visible:[&>circle]:stroke-ctp-mauve"
						on:keydown={(event) => key(handle === 'from' ? 'from' : 'to', event)}
					>
						<rect x={at - 8} width="16" height={HEIGHT} fill="transparent" />
						<line x1={at} x2={at} y2={HEIGHT} stroke={palette.mauve.hex} stroke-width="2" />
						<circle
							cx={at}
							cy={HEIGHT / 2}
							r="6"
							fill={palette.mauve.hex}
							stroke={palette.mantle.hex}
							stroke-width="2"
						/>
					</g>
				{/each}
			</svg>

			{#if hovered && hovered.value !== null && hovered.anomaly !== null}
				<div
					class="pointer-events-none absolute bottom-full z-10 mb-2 whitespace-nowrap rounded-md bg-ctp-crust px-3 py-2 text-sm shadow-md shadow-ctp-crust"
					style="left: {x(hovered.time)}px; transform: translateX({x(hovered.time) > width / 2
						? 'calc(-100% + 4px)'
						: '-4px'})"
				>
					<p class="content-ignore font-semibold text-ctp-text">{hovered.label}</p>
					<p class="content-ignore text-ctp-text">{hovered.value.toFixed(1)} {unit}</p>
					<p class="content-ignore text-xs text-ctp-subtext0">
						{signed(hovered.anomaly)} vs average {hovered.label.slice(0, 3)}
					</p>
				</div>
			{/if}
		{/if}
	</div>
</div>
