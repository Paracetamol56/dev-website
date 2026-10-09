<script lang="ts">
	import * as d3 from 'd3';
	import { variants } from '@catppuccin/palette';
	import { user } from '$lib/store';
	import type { DayCount } from './+page';

	export let stats: DayCount[];

	type Day = { date: Date; value: number };
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

	const toKey = (date: Date) =>
		[date.getFullYear(), date.getMonth() + 1, date.getDate()]
			.map((part) => String(part).padStart(2, '0'))
			.join('-');

	const today = new Date();
	const monday = new Date(
		today.getFullYear(),
		today.getMonth(),
		today.getDate() - ((today.getDay() + 6) % 7)
	);

	$: counts = new Map(stats.map((day) => [day.date, day.count]));
	$: weeks = Array.from({ length: 52 }, (_, i) => {
		const days: Day[] = [];
		for (let j = 0; j < 7; j++) {
			const date = new Date(
				monday.getFullYear(),
				monday.getMonth(),
				monday.getDate() - (51 - i) * 7 + j
			);
			if (date > today) break;
			days.push({ date, value: counts.get(toKey(date)) ?? 0 });
		}
		return days;
	});

	$: max = Math.max(4, d3.max(stats, (day) => day.count) ?? 0);
	$: colorScale = d3
		.scaleSequential(
			d3.interpolateRgb(variants[$user.flavour].crust.rgb, variants[$user.flavour].mauve.rgb)
		)
		.domain([0, max]);

	let hovering: Day | null = null;
</script>

<div class="w-full max-w-4xl px-4 py-8 mx-auto bg-ctp-mantle rounded-md">
	<div class="w-fit max-w-full mx-auto">
		<div class="max-w-full overflow-y-hidden overflow-x-auto">
			<table class="mb-2">
				<caption class="sr-only">Activity Calendar</caption>
				<tbody>
					<tr>
						<td />
						{#each weeks as week}
							{#if week[0].date.getDate() <= 7}
								<td class="relative h-4">
									<p class="absolute top-0 left-0 text-xs text-left">
										{months[week[0].date.getMonth()]}
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
								<td class="w-8">
									<p class="text-xs text-left">{day}</p>
								</td>
							{:else}
								<td />
							{/if}
							{#each weeks as week}
								<td>
									{#if week[i]}
										<!-- svelte-ignore a11y-no-static-element-interactions -->
										<div
											class="square-3 rounded-sm"
											style="background-color: {colorScale(week[i].value)};"
											on:mouseenter={() => (hovering = week[i])}
											on:mouseleave={() => (hovering = null)}
										/>
									{/if}
								</td>
							{/each}
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<div class="flex justify-between gap-4">
			<p class="text-sm">
				{#if hovering}
					{hovering.value} completed on {hovering.date.toDateString()}
				{/if}
			</p>
			<div class="flex items-center gap-1 text-sm">
				<span>Less</span>
				{#each [0, 1, 2, 3, 4] as i}
					<div class="square-3 rounded-sm" style="background-color: {colorScale((i * max) / 4)};" />
				{/each}
				<span>More</span>
			</div>
		</div>
	</div>
</div>
