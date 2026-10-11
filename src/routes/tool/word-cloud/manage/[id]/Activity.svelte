<script lang="ts">
	import { user } from '$lib/store';
	import catppuccin from '@catppuccin/palette';
	import * as Plot from '@observablehq/plot';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import type { SelectOption } from '@melt-ui/svelte';
	import { writable } from 'svelte/store';
	import type { WordCloudWord } from '../../utils';

	export let words: WordCloudWord[];

	let container: HTMLDivElement;
	let width = 0;
	$: palette = catppuccin.variants[$user.flavour];

	const GRANULARITIES: SelectOption<string>[] = [
		{ value: '10', label: '10 seconds' },
		{ value: '15', label: '15 seconds' },
		{ value: '30', label: '30 seconds' },
		{ value: '60', label: '1 minute' },
		{ value: '300', label: '5 minutes' },
		{ value: '900', label: '15 minutes' },
		{ value: '3600', label: '1 hour' }
	];
	const granularity = writable(GRANULARITIES[3]);
	$: step = Number($granularity.value) * 1000;

	const label = (time: number, step: number) =>
		new Date(time).toLocaleTimeString([], {
			hour: '2-digit',
			minute: '2-digit',
			second: step < 60_000 ? '2-digit' : undefined
		});

	/** Submissions per step, from the first to the last step, empty steps included */
	function binned(words: WordCloudWord[], step: number) {
		const times = words.map((word) => Math.floor(Date.parse(word.createdAt) / step) * step);
		const [first, last] = [Math.min(...times), Math.max(...times)];
		const counts = new Map<number, number>();
		for (const time of times) counts.set(time, (counts.get(time) ?? 0) + 1);
		const bins = [];
		for (let time = first; time <= last; time += step) {
			bins.push({ time: label(time, step), count: counts.get(time) ?? 0 });
		}
		return bins;
	}

	function render(words: WordCloudWord[], width: number, step: number) {
		if (!container || !width) return;
		const bins = binned(words, step);
		const every = Math.max(1, Math.ceil(bins.length / Math.max(1, Math.floor(width / 60))));
		container.replaceChildren(
			Plot.plot({
				width,
				height: 220,
				marginLeft: 36,
				style: { background: 'transparent', color: palette.subtext0.hex, fontSize: '11px' },
				x: {
					label: null,
					domain: bins.map((bin) => bin.time),
					ticks: bins.filter((_, i) => i % every === 0).map((bin) => bin.time),
					padding: 0.15
				},
				y: { label: `Submissions per ${$granularity.label}`, grid: true, tickFormat: 'd' },
				marks: [
					Plot.barY(bins, { x: 'time', y: 'count', fill: palette.mauve.hex, tip: true }),
					Plot.ruleY([0], { stroke: palette.surface2.hex })
				]
			})
		);
	}

	$: render(words, width, step);
</script>

<div class="flex flex-wrap items-center justify-between gap-2">
	<h2 class="text-xl font-bold text-ctp-text">Activity</h2>
	<MeltSelect name="Bars per" hideLabel options={GRANULARITIES} value={granularity} />
</div>

{#if words.length}
	<div
		bind:this={container}
		bind:clientWidth={width}
		class="w-full"
		role="img"
		aria-label="Submissions over time"
	/>
{:else}
	<p class="py-6 text-center text-ctp-subtext0">No submissions yet</p>
{/if}
