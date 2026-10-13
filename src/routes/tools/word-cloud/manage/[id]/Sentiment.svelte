<script lang="ts">
	import { user } from '$lib/store';
	import catppuccin from '@catppuccin/palette';
	import * as Plot from '@observablehq/plot';
	import type { WordCloudWord } from '../../utils';

	export let words: WordCloudWord[];
	export let compact = false;

	const MAX = 5;
	let container: HTMLDivElement;
	let width = 0;
	$: palette = catppuccin.variants[$user.flavour];

	$: rated = [...words]
		.filter((word) => word.sentiment !== undefined)
		.sort((a, b) => a.createdAt.localeCompare(b.createdAt));
	$: average = rated.length
		? rated.reduce((sum, word) => sum + word.sentiment!, 0) / rated.length
		: 0;
	$: evolution = rated.map((word, i) => ({
		date: new Date(word.createdAt),
		average: rated.slice(0, i + 1).reduce((sum, w) => sum + w.sentiment!, 0) / (i + 1)
	}));
	$: [label, tone] =
		average <= -2
			? ['Very negative', 'text-ctp-red']
			: average < -0.5
			? ['Negative', 'text-ctp-maroon']
			: average <= 0.5
			? ['Neutral', 'text-ctp-yellow']
			: average < 2
			? ['Positive', 'text-ctp-teal']
			: ['Very positive', 'text-ctp-green'];

	function render(points: typeof evolution, width: number) {
		if (!container || !width) return;
		container.replaceChildren(
			Plot.plot({
				width,
				height: 160,
				marginLeft: 30,
				style: { background: 'transparent', color: palette.subtext0.hex, fontSize: '11px' },
				x: { type: 'time', label: null },
				y: { domain: [-MAX, MAX], label: 'Mood', grid: true, ticks: 5 },
				marks: [
					Plot.ruleY([0], { stroke: palette.surface2.hex }),
					Plot.areaY(points, {
						x: 'date',
						y: 'average',
						fill: palette.teal.hex,
						fillOpacity: 0.15,
						curve: 'step-after'
					}),
					Plot.line(points, {
						x: 'date',
						y: 'average',
						stroke: palette.teal.hex,
						strokeWidth: 2,
						curve: 'step-after',
						tip: true
					})
				]
			})
		);
	}
	$: render(evolution, width);
</script>

{#if rated.length}
	<div class="flex flex-col gap-4">
		<div class="flex items-baseline justify-between gap-4">
			<p class="text-2xl font-bold {tone}">{label}</p>
			<p class="font-mono text-sm text-ctp-subtext0">
				{average > 0 ? '+' : ''}{average.toFixed(1)} / {MAX}
			</p>
		</div>
		<div
			class="relative h-3 rounded-full"
			style="background: linear-gradient(to right, {palette.red.hex}, {palette.yellow.hex}, {palette
				.green.hex})"
		>
			<div
				class="absolute top-1/2 h-5 w-1.5 -translate-x-1/2 -translate-y-1/2 rounded-full bg-ctp-text shadow-md shadow-ctp-crust transition-[left] duration-500"
				style="left: {((average + MAX) / (2 * MAX)) * 100}%"
				aria-hidden="true"
			/>
		</div>
		{#if evolution.length > 1 && !compact}
			<div
				bind:this={container}
				bind:clientWidth={width}
				class="w-full"
				role="img"
				aria-label="Mood over time"
			/>
		{/if}
		<p class="text-xs text-ctp-subtext0">
			Based on {rated.length} of {words.length} submissions with a known sentiment, rated with
			<a class="underline" href="https://github.com/fnielsen/afinn" target="_blank" rel="noreferrer"
				>AFINN</a
			>
			by Finn Årup Nielsen (English, Danish, Finnish, French, Polish, Swedish, Turkish; Apache 2.0).
		</p>
	</div>
{:else}
	<p class="py-6 text-center text-ctp-subtext0">
		{words.length ? 'None of the words has a known sentiment yet.' : 'No submissions yet'}
	</p>
{/if}
