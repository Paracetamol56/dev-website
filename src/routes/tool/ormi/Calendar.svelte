<script lang="ts">
	import { onMount } from 'svelte';
	import * as d3 from 'd3';
	import { variants } from '@catppuccin/palette';
	import { user } from '$lib/store';

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

	// Create an array of 52 weeks before the current date
	const today = new Date();
	const weeks: Day[][] = Array.from({ length: 52 }, (_, i) => {
		const days: Day[] = [];
		for (let j = 0; j < 7; j++) {
			const date = new Date();
			date.setDate(date.getDate() - date.getDay() - i * 7 + j + 1);
			if (date > today) break;
			days.push({
				date,
				value: Math.floor(Math.random() * 10)
			});
		}

		return days;
	}).reverse();

	const colorScale = d3
		.scaleSequential()
		.domain([0, d3.max(weeks.flat(), (d) => d.value)!])
		.range([variants[$user.flavour].crust.rgb, variants[$user.flavour].mauve.rgb]);

	let hovering: Day | null = null;
</script>

<div class="w-full max-w-4xl px-4 py-8 mx-auto bg-ctp-mantle rounded-md">
  <div class="w-fit mx-auto">
    <div class="w-fit max-w-full overflow-y-hidden overflow-x-auto">
      <table class="mb-2">
        <caption class="sr-only">Activity Calendar</caption>
        <tbody>
          <tr>
            <td />
            {#each weeks as week, i}
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
              {#each weeks as week, j}
                <td>
                  {#if week[i]}
                    <!-- svelte-ignore a11y-no-static-element-interactions -->
                    <div
                      class="relative square-3 bg-ctp-crust rounded-sm"
                      style="background-color: {colorScale(week[i].value)};"
                      on:mouseenter={(e) => (hovering = week[i])}
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

    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="flex justify-between">
      <p class="text-sm">
        {#if hovering}
          {hovering.value} activity on {hovering.date.toDateString()}
        {/if}
      </p>
      <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
      <table class="flex items-center gap-[]" on:click={() => (hovering = null)}>
        <tbody>
          <tr>
            <td class="text-sm">
              <span>Less</span>
            </td>
            {#each [1, 2, 3, 4] as i}
              <td>
                <div class="square-3 rounded-sm" style="background-color: {colorScale(i * 2.5)};" />
              </td>
            {/each}
            <td class="text-sm">
              <span>More</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</div>
