<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import RadioGroup from '$lib/components/ui/RadioGroup.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import TextInput from '$lib/components/ui/TextInput.svelte';
	import { findPlace, type Place } from '$lib/place';
	import { user } from '$lib/store';
	import catppuccin from '@catppuccin/palette';
	import type { SelectOption } from '@melt-ui/svelte';
	import { Search } from 'lucide-svelte';
	import { onDestroy, onMount } from 'svelte';
	import { writable } from 'svelte/store';
	import RadarChart from './RadarChart.svelte';
	import Timeline from './Timeline.svelte';
	import {
		DATASETS,
		MONTH_NAMES,
		VARIABLES,
		YEAR_COLORS,
		aggregate,
		fetchDaily,
		type Daily,
		type Dataset,
		type Resolution,
		type Variable
	} from './era5';

	// Opening the page shows this place without querying Nominatim
	let place: Place = { name: 'Paris, France', lat: 48.8566, lon: 2.3522 };
	const query = writable(place.name);

	const datasetOptions = Object.entries(DATASETS).map(([value, { label }]) => ({ value, label }));
	const variableOptions = Object.entries(VARIABLES).map(([value, label]) => ({ value, label }));
	const dataset = writable<SelectOption<string>>(datasetOptions[0]);
	const variable = writable<SelectOption<string>>(variableOptions[0]);
	const resolution = writable<string>('week');
	const values = writable<string>('absolute');
	const speed = writable<string>('normal');
	/** Milliseconds to draw one year */
	const DURATIONS: Record<string, number> = { slow: 3000, normal: 1000, fast: 250 };

	let daily: Daily | null = null;
	let loading = false;
	let error: string | null = null;
	let controller: AbortController | null = null;

	async function search() {
		const text = $query.trim();
		if (!text) return;
		controller?.abort();
		const current = (controller = new AbortController());
		loading = true;
		error = null;
		try {
			place = await findPlace(text, current.signal);
		} catch (e) {
			if (current.signal.aborted) return;
			error = e instanceof Error ? e.message : 'Could not find this place.';
			loading = false;
		}
	}

	async function load(place: Place, dataset: Dataset, variable: Variable) {
		controller?.abort();
		const current = (controller = new AbortController());
		loading = true;
		error = null;
		try {
			daily = await fetchDaily(place, dataset, variable);
		} catch (e) {
			if (current.signal.aborted) return;
			daily = null;
			error = e instanceof Error ? e.message : 'Could not load ERA5 data.';
		} finally {
			if (controller === current) loading = false;
		}
	}

	let mounted = false;
	onMount(() => (mounted = true));
	onDestroy(() => controller?.abort());
	$: if (mounted) load(place, $dataset.value as Dataset, $variable.value as Variable);

	$: climate = daily ? aggregate(daily, $resolution as Resolution) : null;
	$: monthly = daily ? aggregate(daily, 'month').years : [];
	$: table = monthly.slice().reverse();
	$: first = climate?.years[0].year ?? 1940;
	$: last = climate?.years.at(-1)?.year ?? first;
	$: latest = climate?.years.findLast((entry) => entry.annual !== null)?.year ?? last;

	// New data shows every year up to the most recent complete one
	let from = 1940;
	let to = new Date().getFullYear();
	/** Playhead in years, see RadarChart; past `to` the whole range is shown */
	let progress = Infinity;
	let shownDaily: Daily | null = null;
	$: if (daily && climate && daily !== shownDaily) {
		shownDaily = daily;
		[from, to] = [first, latest];
	}

	$: shownYear = Math.min(Math.floor(progress), to);
	$: selected = climate?.years.find((entry) => entry.year === shownYear);
	$: ranked = (climate?.years ?? [])
		.filter((entry) => entry.annual !== null)
		.sort((a, b) => (b.annual as number) - (a.annual as number));
	$: rank = selected ? ranked.indexOf(selected) + 1 : 0;
	const ordinal = (n: number) =>
		n === 1
			? ''
			: `${n}${['th', 'st', 'nd', 'rd'][n % 100 > 10 && n % 100 < 14 ? 0 : n % 10] ?? 'th'} `;

	$: palette = catppuccin.variants[$user.flavour];
</script>

<div class="my-8">
	<div class="flex flex-col gap-8">
		<form
			class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-wrap items-end gap-4"
			on:submit|preventDefault={search}
		>
			<div class="relative flex-1 min-w-[16rem]">
				<Search class="absolute left-3 bottom-2 text-ctp-subtext0" size="16" />
				<TextInput
					label="City, place or coordinates"
					value={query}
					placeholder="Paris, or 48.8566, 2.3522"
					class="pl-9"
				/>
			</div>
			<Button type="submit" {loading}>Show</Button>
		</form>

		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col gap-6">
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div aria-live="polite" class="text-sm">
					{#if error}
						<p class="content-ignore text-ctp-red">{error}</p>
					{:else}
						<p class="content-ignore font-semibold text-ctp-text">{place.name}</p>
						{#if daily}
							<p class="content-ignore text-ctp-subtext0">
								Grid cell at {daily.lat.toFixed(2)}, {daily.lon.toFixed(2)}, {Math.round(
									daily.elevation
								)} m
							</p>
						{:else if loading}
							<p class="content-ignore text-ctp-subtext0">Loading ERA5 data...</p>
						{/if}
					{/if}
				</div>
				{#if selected && !error}
					<div class="text-right">
						{#if selected.annual !== null}
							<p class="content-ignore text-3xl font-bold text-ctp-text">
								{selected.annual.toFixed(1)} °C
							</p>
							<p class="content-ignore text-xs text-ctp-subtext0">
								{selected.year} average, {ordinal(rank)}warmest of {ranked.length} years
							</p>
						{:else}
							<p class="content-ignore text-3xl font-bold text-ctp-text">{selected.year}</p>
							<p class="content-ignore text-xs text-ctp-subtext0">Year not complete yet</p>
						{/if}
					</div>
				{/if}
			</div>

			<div class="flex flex-wrap items-end gap-x-8 gap-y-4">
				<Select label="Dataset" options={datasetOptions} value={dataset} />
				<Select label="Temperature" options={variableOptions} value={variable} />
			</div>
			<div class="flex flex-wrap items-end gap-x-8 gap-y-4">
				<fieldset>
					<legend class="text-sm font-semibold text-ctp-text mb-2">Spokes</legend>
					<RadioGroup name="resolution" options={['month', 'week', 'day']} value={resolution} />
				</fieldset>
				<fieldset>
					<legend class="text-sm font-semibold text-ctp-text mb-2">Values</legend>
					<RadioGroup name="values" options={['absolute', 'relative']} value={values} />
				</fieldset>
				<fieldset>
					<legend class="text-sm font-semibold text-ctp-text mb-2">Animation speed</legend>
					<RadioGroup name="speed" options={['slow', 'normal', 'fast']} value={speed} />
				</fieldset>
			</div>

			{#if climate}
				<Timeline
					months={monthly}
					bind:from
					bind:to
					bind:progress
					duration={DURATIONS[$speed]}
					{palette}
				/>

				<div class="flex flex-wrap items-center gap-x-6 gap-y-2 text-xs text-ctp-subtext0">
					<span class="flex items-center gap-2">
						{from}
						<span
							class="block h-2 w-24 rounded-full"
							style="background: linear-gradient(to right, {YEAR_COLORS.map(
								(name) => palette[name].hex
							).join()})"
						/>
						{to}
					</span>
					<span class="flex items-center gap-2">
						<span class="block h-[3px] w-6 rounded-full bg-ctp-text" />
						{shownYear}
					</span>
					{#if $values === 'relative'}
						<span class="flex items-center gap-2">
							<span class="block h-[2px] w-6 bg-ctp-overlay1" />
							0 = {from}–{to} average
						</span>
					{/if}
				</div>

				<div class="transition-opacity {loading ? 'opacity-50' : ''}">
					<RadarChart
						{climate}
						{palette}
						{from}
						{to}
						{progress}
						relative={$values === 'relative'}
					/>
				</div>

				<details class="content-ignore text-sm">
					<summary class="content-ignore cursor-pointer font-semibold text-ctp-text"
						>Monthly values table</summary
					>
					<div class="mt-4 overflow-x-auto">
						<table class="content-ignore w-full text-right tabular-nums">
							<thead class="text-ctp-subtext0">
								<tr>
									<th class="py-1 pr-3 text-left font-semibold">Year</th>
									{#each MONTH_NAMES as month}
										<th class="px-2 py-1 font-semibold">{month}</th>
									{/each}
									<th class="pl-3 py-1 font-semibold">Average</th>
								</tr>
							</thead>
							<tbody>
								{#each table as period (period.year)}
									<tr class="border-t border-ctp-surface0">
										<th class="py-1 pr-3 text-left font-semibold">{period.year}</th>
										{#each period.values as value}
											<td class="px-2 py-1">{value === null ? '–' : value.toFixed(1)}</td>
										{/each}
										<td class="pl-3 py-1 font-semibold">
											{period.annual === null ? '–' : period.annual.toFixed(1)}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</details>
			{/if}

			<p class="content-ignore text-xs text-ctp-subtext0">
				Temperatures 2 m above ground from the Copernicus <a
					class="content-ignore underline"
					href="https://cds.climate.copernicus.eu/datasets/reanalysis-era5-single-levels"
					target="_blank"
					rel="noreferrer">ERA5</a
				>
				reanalysis (Hersbach et al., 2023), served by
				<a
					class="content-ignore underline"
					href="https://open-meteo.com/"
					target="_blank"
					rel="noreferrer">Open-Meteo</a
				>. Values are averages over a grid cell, so they can differ from a weather station,
				especially in mountains or on coasts. Daily spokes are smoothed over 15 days. Place search:
				Nominatim, ©
				<a
					class="content-ignore underline"
					href="https://www.openstreetmap.org/copyright"
					target="_blank"
					rel="noreferrer">OpenStreetMap contributors</a
				>.
			</p>
		</div>
	</div>
</div>
