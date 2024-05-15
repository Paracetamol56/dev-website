<script lang="ts">
	import api from '$lib/api';
	import * as Plot from '@observablehq/plot';
	import { onMount } from 'svelte';
	import { user } from '$lib/store';
	import { variants } from '@catppuccin/palette';
	import * as d3 from 'd3';
	import MeltRadioGroup from '$lib/components/MeltRadioGroup.svelte';
	import { get, writable, type Writable } from 'svelte/store';
	import MeltCheckbox from '$lib/components/MeltCheckbox.svelte';
	import { valueAndGrad } from '@tensorflow/tfjs';

	type Microprocessors = {
		id: string;
		name: string;
		type: 'CPU' | 'GPU';
		release: Date;
		gateSize: number;
		tdp: number;
		transistors: number;
		frequency: number;
		vendor: string;
	};

	let div: HTMLDivElement;
	let width: number;

	let chipsData: Microprocessors[] = [];
	const mooreData = d3
		.ticks(2000, 2024, 24)
		.map((year) => ({ release: new Date(year, 0), transistors: 30 * 2 ** ((year - 2000) / 2.5) }));

	const features = [
		{ name: 'transistors', label: 'Transistors (millions)' },
		{ name: 'gateSize', label: 'Gate size (nm)' },
		{ name: 'frequency', label: 'Frequency (MHz)' },
		{ name: 'tdp', label: 'TDP (W)' }
	];
	let selectedFeature: Writable<string> = writable(features[0].name);
	const types = [
		{ name: 'CPU', label: 'CPU' },
		{ name: 'GPU', label: 'GPU' }
	];
	const cpuSelected: Writable<boolean> = writable(true);
	const gpuSelected: Writable<boolean> = writable(false);
	const scales = [
		{ name: 'linear', label: 'Linear' },
		{ name: 'log', label: 'Logarithmic' }
	];
	let selectedScale: Writable<string> = writable(scales[0].name);

	onMount(() => {
		api.call('GET', '/microprocessors').then((res) => {
			chipsData = Array.from(res.data, (chip: any) => ({
				id: chip.id,
				name: chip.name,
				type: chip.type,
				release: new Date(chip.release),
				gateSize: chip.gateSize,
				tdp: chip.tdp,
				transistors: chip.transistors,
				frequency: chip.frequency,
				vendor: chip.vendor
			}));
		});
	});

	$: {
		div?.firstChild?.remove();

		let data: Microprocessors[] = [];
		if ($cpuSelected && $gpuSelected) {
			data = chipsData;
		} else if ($cpuSelected) {
			data = chipsData.filter((chip) => chip.type === 'CPU');
		} else if ($gpuSelected) {
			data = chipsData.filter((chip) => chip.type === 'GPU');
		}

		div?.append(
			Plot.plot({
				width: Math.max(width, 800),
				height: 600,
				grid: true,
				x: {
					label: 'Release date',
					type: 'time'
				},
				y: {
					label: features.find((f) => f.name === $selectedFeature)?.label,
					type: $selectedScale as Plot.ScaleType
				},
				symbol: {
					legend: true,
					range: ['circle', 'square'],
					domain: ['CPU', 'GPU']
				},
				color: {
					legend: true,
					range: [
						variants[$user.flavour].peach.rgb,
						variants[$user.flavour].maroon.rgb,
						variants[$user.flavour].red.rgb,
						variants[$user.flavour].lavender.rgb,
						variants[$user.flavour].green.rgb,
						variants[$user.flavour].mauve.rgb,
						variants[$user.flavour].blue.rgb
					],
					domain: ['3dfx', 'AMD', 'ATI', 'Intel', 'NVIDIA', 'Sony', 'VIA']
				},
				marks: [
					Plot.dot(data, {
						x: 'release',
						y: $selectedFeature,
						stroke: 'vendor',
						strokeWidth: 1,
						symbol: 'type',
						r: 4
					}),
					Plot.line($selectedFeature === features[0].name ? mooreData : [], {
						x: 'release',
						y: 'transistors',
						stroke: variants[$user.flavour].blue.rgb,
						strokeWidth: 2
					})
				]
			})
		);
	}
</script>

<form class="mb-2 p-4 bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md flex justify-center gap-4">
	<div class="flex flex-col">
		<h4 class="content-ignore mb-2 text-lg font-bold">Features</h4>
		<MeltRadioGroup options={features} value={selectedFeature} orientation="vertical" />
	</div>
	<div class="flex flex-col">
		<h4 class="content-ignore mb-2 text-lg font-bold">Types</h4>
		<div class="flex flex-col gap-3">
			<MeltCheckbox name={types[0].name} label={types[0].label} checked={cpuSelected} />
			<MeltCheckbox name={types[1].name} label={types[1].label} checked={gpuSelected} />
		</div>
	</div>
	<div class=" flex flex-col">
		<h4 class="content-ignore mb-2 text-lg font-bold">Scale</h4>
		<MeltRadioGroup options={scales} value={selectedScale} orientation="vertical" />
	</div>
</form>

<div bind:clientWidth={width} bind:this={div} class="flex justify-center" role="img" />
