<script lang="ts">
	import * as d3 from 'd3';
	import { variants } from '@catppuccin/palette';
	import { user } from '$lib/store';
	import type { DayCount } from './+page';
	import { addDays, startOfWeek, toKey } from './dates';
	import { showTodos } from './filters';
	import { CalendarDate } from '@internationalized/date';
	import api from '$lib/api';

	export let stats: DayCount[];
	export let firstYear: number;

	type Day = { date: Date; value: number };
	type Week = { start: Date; days: (Day | null)[] };
	const weekDays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
	const months = [
		'Jan',
		'Feb',
		'Mar',
		'Apr',
		'May',
		'Jun',
		'Jul',
		'Aug',
		'Sep',
		'Oct',
		'Nov',
		'Dec'
	];

	const today = new Date();
	const years = Array.from(
		{ length: Math.max(1, today.getFullYear() - firstYear + 1) },
		(_, i) => today.getFullYear() - i
	);

	let year: number | null = null;
	let yearStats: DayCount[] = [];

	const loadYear = (selected: number) => {
		const timezone = encodeURIComponent(Intl.DateTimeFormat().resolvedOptions().timeZone);
		api
			.callWithAuth('GET', `/ormi/stats/${selected}?tz=${timezone}`)
			.then((response) => {
				if (selected === year) yearStats = response.data;
			})
			.catch((error) => {
				console.error('Failed to fetch the year activity:', error);
				year = null;
			});
	};
	$: if (year !== null) {
		stats;
		loadYear(year);
	}

	$: shown = year === null ? stats : yearStats;
	$: counts = new Map(shown.map((day) => [day.date, day.count]));
	$: first = year === null ? addDays(startOfWeek(today), -51 * 7) : new Date(year, 0, 1);
	$: last = year === null || year === today.getFullYear() ? today : new Date(year, 11, 31);
	$: weeks = buildWeeks(first, last, counts);

	const buildWeeks = (first: Date, last: Date, counts: Map<string, number>) => {
		const weeks: Week[] = [];
		for (let start = startOfWeek(first); start <= last; start = addDays(start, 7)) {
			const days = Array.from({ length: 7 }, (_, j) => {
				const date = addDays(start, j);
				return date < first || date > last ? null : { date, value: counts.get(toKey(date)) ?? 0 };
			});
			weeks.push({ start, days });
		}
		return weeks;
	};

	$: max = Math.max(4, d3.max(shown, (day) => day.count) ?? 0);
	$: colorScale = d3
		.scaleSequential(
			d3.interpolateRgb(variants[$user.flavour].crust.rgb, variants[$user.flavour].mauve.rgb)
		)
		.domain([0, max]);

	let hovering: Day | null = null;

	const showDay = (date: Date) => {
		const day = new CalendarDate(date.getFullYear(), date.getMonth() + 1, date.getDate());
		showTodos({ states: ['DONE'], from: day, to: day, sort: 'updated' });
	};
</script>

<div class="w-full p-4 bg-ctp-mantle rounded-md">
	<div class="w-fit max-w-full mx-auto">
		<div class="max-w-full overflow-y-hidden overflow-x-auto">
			<table class="mb-2">
				<caption class="sr-only">Activity Calendar</caption>
				<tbody>
					<tr>
						<td />
						{#each weeks as week, i}
							{#if week.start.getDate() <= 7 && i < weeks.length - 1}
								<td class="relative h-4">
									<p class="absolute top-0 left-0 text-xs text-left">
										{months[week.start.getMonth()]}
									</p>
								</td>
							{:else}
								<td />
							{/if}
						{/each}
					</tr>
					{#each weekDays as day, i}
						<tr>
							{#if i % 2 == 0}
								<td class="w-10">
									<p class="text-xs text-left">{day}</p>
								</td>
							{:else}
								<td />
							{/if}
							{#each weeks as week}
								{@const cell = week.days[i]}
								<td>
									{#if cell}
										<button
											type="button"
											class="block square-4 rounded-sm hover:ring-2 hover:ring-ctp-text focus-visible:ring-2 focus-visible:ring-ctp-text"
											style="background-color: {colorScale(cell.value)};"
											aria-label="{cell.value} completed on {cell.date.toDateString()}"
											on:mouseenter={() => (hovering = cell)}
											on:mouseleave={() => (hovering = null)}
											on:focus={() => (hovering = cell)}
											on:blur={() => (hovering = null)}
											on:click={() => showDay(cell.date)}
										/>
									{/if}
								</td>
							{/each}
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
			<div class="flex flex-wrap items-center gap-1">
				{#each [null, ...years] as option}
					<button
						type="button"
						aria-pressed={year === option}
						class="rounded-md px-2 py-0.5 text-sm font-semibold transition-colors hover:bg-ctp-surface0
							aria-pressed:bg-ctp-surface0 aria-pressed:text-ctp-mauve"
						on:click={() => (year = option)}
					>
						{option ?? 'Last 12 months'}
					</button>
				{/each}
			</div>
			<p class="text-sm">
				{#if hovering}
					{hovering.value} completed on {hovering.date.toDateString()}
				{/if}
			</p>
			<div class="flex items-center gap-1 text-sm">
				<span>Less</span>
				{#each [0, 1, 2, 3, 4] as i}
					<div class="square-4 rounded-sm" style="background-color: {colorScale((i * max) / 4)};" />
				{/each}
				<span>More</span>
			</div>
		</div>
	</div>
</div>
