<script lang="ts">
	import api from '$lib/api';
  import * as Plot from '@observablehq/plot';
	import { onMount } from 'svelte';
	import { user } from '$lib/store';
	import { variants } from '@catppuccin/palette';
	import * as d3 from 'd3';

	type Microprocessors = {
		id: string;
		name: string;
		type: "CPU" | "GPU";
		release: Date;
		gateSize: number;
		tdp: number;
		transistors: number;
		frequency: number;
		vendor: string;
	};

  let div: HTMLDivElement;
  let chipsData: Microprocessors[] = [];

	const features = [
		{ name: 'transistors', label: 'Transistors (millions)' },
		{ name: 'gateSize', label: 'Gate size (nm)' },
		{ name: 'frequency', label: 'Frequency (MHz)' },
		{ name: 'tdp', label: 'TDP (W)' },
	];
	let selectedFeature = 0;
	const types = [
		{ name: 'CPU', label: 'CPU' },
		{ name: 'GPU', label: 'GPU' },
	];
	let selectedType = [true, true];

  onMount(() => {
    api.call('GET', '/microprocessors?type=CPU').then((res) => {
			chipsData = Array.from(res.data, (chip) => ({
				id: chip.id,
				name: chip.name,
				type: chip.type,
				release: new Date(chip.release),
				gateSize: chip.gateSize,
				tdp: chip.tdp,
				transistors: chip.transistors,
				frequency: chip.frequency,
				vendor: chip.vendor,
			}));
    });
  });

  $: {
		console.log(selectedFeature);
		div?.firstChild?.remove();
		div?.append(
			Plot.plot({
				width: 1000,
				height: 750,
				grid: true,
				x: {
					label: 'Release date',
					type: 'time',
				},
				y: {
					label: features[selectedFeature].label,
					type: 'log',
				},
				symbol: {legend: true},
				color: {legend: true},
				marks: [
					Plot.dot(chipsData, {
						x: 'release',
						y: features[selectedFeature].name,
						stroke: 'vendor',
						strokeWidth: 1,
						symbol: 'type',
						r: 4,
					}),
				]
			})
		);
	}
</script>

<form class="flex justify-center gap-4">
	<div class="p-4 bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md flex flex-col">
		<h4 class="content-ignore text-lg font-bold">Features</h4>
		{#each features as feature, i}
			<label>
				<input type="radio" name="feature" value={feature.name} checked={selectedFeature === i} on:input={() => selectedFeature = i} />
				{feature.label}
			</label>
		{/each}
	</div>
	<div class="p-4 bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md flex flex-col">
		<h4 class="content-ignore text-lg font-bold">Types</h4>
		{#each types as type}
			<label>
				<input type="checkbox" name="type" bind:checked={selectedType[types.indexOf(type)]} />
				{type.label}
			</label>
		{/each}
	</div>
</form>

<div bind:this={div} class="flex justify-center" role="img" />