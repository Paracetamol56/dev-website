<script lang="ts">
	import api from '$lib/api';
	import * as Plot from '@observablehq/plot';
	import { onMount } from 'svelte';
	import { user } from '$lib/store';
	import { variants } from '@catppuccin/palette';
	import * as d3 from 'd3';

	type Chip = {
		id: string;
		name: string;
		type: string;
		release: Date;
		vendor: string;
		transistors?: number;
		gateSize?: number;
		dieSize?: number;
		density?: number;
		frequency?: number;
		tdp?: number;
		sourceUrl?: string;
	};
	type Metric = 'transistors' | 'gateSize' | 'dieSize' | 'density' | 'frequency' | 'tdp';

	const compact = (value: number) => d3.format('.3~s')(value).replace('G', 'B');
	const plain = (value: number) => d3.format(',.4~r')(value);
	const metrics: {
		name: Metric;
		label: string;
		axis: string;
		format: (value: number) => string;
	}[] = [
		{ name: 'transistors', label: 'Transistors', axis: 'Transistors', format: compact },
		{ name: 'gateSize', label: 'Process node', axis: 'Process node (nm)', format: plain },
		{ name: 'dieSize', label: 'Die area', axis: 'Die area (mm²)', format: plain },
		{ name: 'density', label: 'Density', axis: 'Transistors per mm²', format: compact },
		{ name: 'frequency', label: 'Frequency', axis: 'Frequency (MHz)', format: plain },
		{ name: 'tdp', label: 'TDP', axis: 'TDP (W)', format: plain }
	];
	const types = [
		{ name: 'CPU', symbol: 'circle' },
		{ name: 'GPU', symbol: 'square' }
	];
	const scales = ['log', 'linear'];
	const VENDORS = 5;
	const OTHER = 'Other';

	let chips: Chip[] = [];
	let loaded = false;
	let metric: Metric = 'transistors';
	let selectedTypes = ['CPU'];
	let scale = 'log';

	let container: HTMLDivElement;
	let width = 0;
	let legend: { name: string; color: string }[] = [];
	let shown = 0;
	let showTrend = true;
	let trend: { slope: number; doubling: number; r2: number } | null = null;

	// Least squares on log2(value) over the release year: an exponential trend, straight on the log scale
	const fitTrend = (points: { year: number; value: number }[]) => {
		if (points.length < 3) return null;
		const x = d3.mean(points, (point) => point.year) as number;
		const y = d3.mean(points, (point) => Math.log2(point.value)) as number;
		const sxx = d3.sum(points, (point) => (point.year - x) ** 2);
		const syy = d3.sum(points, (point) => (Math.log2(point.value) - y) ** 2);
		const sxy = d3.sum(points, (point) => (point.year - x) * (Math.log2(point.value) - y));
		if (sxx === 0 || syy === 0) return null;
		const slope = sxy / sxx;
		return { slope, intercept: y - slope * x, r2: sxy ** 2 / (sxx * syy) };
	};

	$: sourceUrl = chips.find((chip) => chip.sourceUrl)?.sourceUrl;
	$: availableMetrics = metrics.filter(({ name }) => chips.some((chip) => (chip[name] ?? 0) > 0));
	$: availableTypes = types.filter(({ name }) => chips.some((chip) => chip.type === name));
	$: current = metrics.find(({ name }) => name === metric) ?? metrics[0];

	onMount(() => {
		api
			.call('GET', '/chips')
			.then((res) => {
				chips = (res.data as any[])
					.map((chip) => ({
						...chip,
						release: new Date(chip.release),
						// The API counts transistors in millions
						transistors: chip.transistors ? chip.transistors * 1e6 : undefined
					}))
					.filter((chip) => !isNaN(chip.release.getTime()));
			})
			.catch((error) => console.error('Failed to fetch chips:', error))
			.finally(() => (loaded = true));
	});

	const toggleType = (name: string) => {
		selectedTypes = selectedTypes.includes(name)
			? selectedTypes.filter((type) => type !== name)
			: [...selectedTypes, name];
	};

	$: if (container && width > 0) {
		const palette = variants[$user.flavour];
		const colors = [palette.blue, palette.peach, palette.mauve, palette.green, palette.red];

		const data = chips.filter(
			(chip) => selectedTypes.includes(chip.type) && (chip[metric] ?? 0) > 0
		);
		const vendors = d3
			.rollups(
				data,
				(group) => group.length,
				(chip) => chip.vendor
			)
			.sort((a, b) => b[1] - a[1])
			.slice(0, VENDORS)
			.map(([vendor]) => vendor);
		const group = (chip: Chip) => (vendors.includes(chip.vendor) ? chip.vendor : OTHER);

		const domain =
			vendors.length < new Set(data.map((chip) => chip.vendor)).size
				? [...vendors, OTHER]
				: vendors;
		const range = domain.map((name, i) => (name === OTHER ? palette.overlay1.hex : colors[i].hex));
		legend = domain.map((name, i) => ({ name, color: range[i] }));
		shown = data.length;

		const extent = d3.extent(data, (chip) => chip.release) as [Date, Date];
		const [min, max] = d3.extent(data, (chip) => chip[metric] ?? 0) as [number, number];
		const powersOfTen = d3
			.range(Math.ceil(Math.log10(min || 1)), Math.floor(Math.log10(max || 1)) + 1)
			.map((exponent) => 10 ** exponent);
		const fit = showTrend
			? fitTrend(
					data.map((chip) => ({ year: chip.release.getFullYear(), value: chip[metric] ?? 0 }))
			  )
			: null;
		trend = fit && { slope: fit.slope, doubling: 1 / fit.slope, r2: fit.r2 };
		const trendLine = fit
			? d3.range(extent[0].getFullYear(), extent[1].getFullYear() + 1).map((year) => ({
					release: new Date(year, 0),
					value: 2 ** (fit.intercept + fit.slope * year)
			  }))
			: [];

		container.replaceChildren(
			Plot.plot({
				width,
				height: Math.max(320, Math.min(560, width * 0.6)),
				marginLeft: 56,
				marginRight: 16,
				style: { background: 'transparent', fontSize: '12px', overflow: 'visible' },
				x: { type: 'time', label: null, grid: true },
				y: {
					type: scale as Plot.ScaleType,
					label: current.axis,
					grid: true,
					ticks: scale === 'log' ? powersOfTen : undefined,
					tickFormat: current.format
				},
				color: { domain, range },
				symbol: {
					domain: types.map(({ name }) => name),
					range: types.map(({ symbol }) => symbol)
				},
				marks: [
					Plot.line(trendLine, {
						x: 'release',
						y: 'value',
						stroke: palette.text.hex,
						strokeWidth: 2
					}),
					Plot.dot(data, {
						x: 'release',
						y: metric,
						fill: group,
						fillOpacity: 0.75,
						stroke: palette.mantle.hex,
						strokeWidth: 0.75,
						symbol: 'type',
						r: 4
					}),
					Plot.tip(
						data,
						Plot.pointer({
							x: 'release',
							y: metric,
							title: (chip: Chip) =>
								`${chip.name}\n${chip.vendor || 'Unknown'} · ${
									chip.type
								} · ${chip.release.getFullYear()}\n${current.axis}: ${current.format(
									chip[metric] ?? 0
								)}`,
							fill: palette.crust.hex,
							stroke: palette.surface1.hex,
							lineWidth: 24
						})
					)
				]
			})
		);
	}

	const pill =
		'rounded-md px-3 py-1 text-sm font-semibold transition-colors bg-ctp-surface0 hover:bg-ctp-surface1 aria-pressed:bg-ctp-mauve aria-pressed:text-ctp-base';
</script>

<figure class="my-8 rounded-md bg-ctp-mantle p-4 shadow-md shadow-ctp-crust">
	<div class="mb-4 flex flex-wrap gap-x-8 gap-y-3">
		<div>
			<span class="mb-1 block text-sm font-semibold">Metric</span>
			<div class="flex flex-wrap gap-1">
				{#each availableMetrics as { name, label } (name)}
					<button
						type="button"
						class={pill}
						aria-pressed={metric === name}
						on:click={() => (metric = name)}
					>
						{label}
					</button>
				{/each}
			</div>
		</div>
		<div>
			<span class="mb-1 block text-sm font-semibold">Type</span>
			<div class="flex flex-wrap gap-1">
				{#each availableTypes as { name, symbol } (name)}
					<button
						type="button"
						class="{pill} flex items-center gap-2"
						aria-pressed={selectedTypes.includes(name)}
						on:click={() => toggleType(name)}
					>
						<span
							class="inline-block size-2.5 bg-current {symbol === 'circle' ? 'rounded-full' : ''}"
						/>
						{name}
					</button>
				{/each}
			</div>
		</div>
		<div>
			<span class="mb-1 block text-sm font-semibold">Scale</span>
			<div class="flex flex-wrap gap-1">
				{#each scales as name (name)}
					<button
						type="button"
						class="{pill} capitalize"
						aria-pressed={scale === name}
						on:click={() => (scale = name)}
					>
						{name}
					</button>
				{/each}
			</div>
		</div>
		<div class="flex flex-wrap items-center gap-x-3 gap-y-1 self-end">
			<button
				type="button"
				class={pill}
				aria-pressed={showTrend}
				on:click={() => (showTrend = !showTrend)}
			>
				Linear regression
			</button>
			{#if trend}
				<span class="text-sm text-ctp-subtext0">
					{d3.format('+.3f')(trend.slope)} log₂/yr ({d3.format('+.0%')(2 ** trend.slope - 1)}/yr)
				</span>
			{/if}
		</div>
	</div>

	<div
		bind:clientWidth={width}
		bind:this={container}
		class="text-ctp-text"
		role="img"
		aria-label="{current.axis} of {selectedTypes.join(' and ')} chips by release year"
	/>

	{#if loaded && shown === 0}
		<p class="py-8 text-center text-sm text-ctp-subtext0">
			{chips.length === 0 ? 'No data available.' : 'Nothing to show, select at least one type.'}
		</p>
	{/if}

	<figcaption
		class="mt-3 flex flex-wrap items-center justify-between gap-x-6 gap-y-2 text-sm text-ctp-subtext0"
	>
		<ul class="flex flex-wrap gap-x-4 gap-y-1">
			{#each legend as { name, color } (name)}
				<li class="flex items-center gap-1.5">
					<span class="inline-block size-2.5 rounded-full" style="background-color: {color};" />
					{name}
				</li>
			{/each}
			{#if trend}
				<li class="flex items-center gap-1.5">
					<span class="inline-block h-0.5 w-4 bg-ctp-text" />
					Trend: {trend.doubling > 0 ? '×2' : '÷2'} every {d3.format('.2~f')(
						Math.abs(trend.doubling)
					)} years (R² {d3.format('.2f')(trend.r2)})
				</li>
			{/if}
		</ul>
		<span>
			{shown} chips
			{#if sourceUrl}
				· <a class="text-ctp-blue" href={sourceUrl} target="_blank" rel="noreferrer">Wikipedia</a>,
				<a
					class="text-ctp-blue"
					href="https://creativecommons.org/licenses/by-sa/4.0/"
					target="_blank"
					rel="noreferrer">CC BY-SA</a
				>
			{/if}
		</span>
	</figcaption>
</figure>
