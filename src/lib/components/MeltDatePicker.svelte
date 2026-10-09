<script lang="ts">
	import { createDatePicker, melt } from '@melt-ui/svelte';
	import { Calendar, ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { writable, type Writable } from 'svelte/store';
	import type { DateValue } from '@internationalized/date';

	export let value: Writable<DateValue | undefined> = writable(undefined);
	export let minValue: DateValue | undefined = undefined;
	export let label: string | undefined = undefined;
	export let compact: boolean = false;

	const {
		elements: {
			calendar,
			cell,
			content,
			field,
			grid,
			heading,
			label: labelElement,
			nextButton,
			prevButton,
			segment,
			trigger
		},
		states: { months, headingValue, weekdays, segmentContents, open },
		helpers: { isDateDisabled, isDateUnavailable }
	} = createDatePicker({
		forceVisible: true,
		value,
		minValue
	});
</script>

{#if compact}
	<button
		use:melt={$trigger}
		class="transition-opacity hover:opacity-80 active:opacity-60 {$value
			? 'text-ctp-mauve'
			: 'text-ctp-text'}"
		type="button"
	>
		<Calendar size={16} />
	</button>
{:else}
	{#if label}
		<span class="mb-2 text-sm font-semibold" use:melt={$labelElement}>{label}</span>
	{/if}
	<div
		use:melt={$field}
		class="flex w-full min-w-[200px] items-center rounded-md bg-ctp-surface0 pl-3 p-2
			focus:outline-none focus:ring-2 focus:ring-ctp-mauve text-ctp-text"
	>
		{#each $segmentContents as seg}
			<div use:melt={$segment(seg.part)}>
				{seg.value}
			</div>
		{/each}
		<div class="ml-4 flex w-full items-center justify-end">
			<button
				use:melt={$trigger}
				class="rounded-md bg-ctp-mauve p-1 text-ctp-surface0 transition-all hover:bg-ctp-mauve/80"
				type="button"
			>
				<Calendar size={16} />
			</button>
		</div>
	</div>
{/if}

{#if $open}
	<div transition:fade={{ duration: 100 }} use:melt={$content} class="z-50 min-w-[320px]">
		<div
			use:melt={$calendar}
			class="w-full rounded-md bg-ctp-mantle p-3 text-ctp-text shadow-md shadow-ctp-crust"
		>
			<header class="flex items-center justify-between pb-2">
				<button
					use:melt={$prevButton}
					class="rounded-md p-1 transition-all hover:bg-ctp-mauve/20 data-[disabled]:pointer-events-none data-[disabled]:opacity-40"
				>
					<ChevronLeft size={24} />
				</button>
				<div use:melt={$heading} class="flex items-center gap-6 font-semibold">
					{$headingValue}
				</div>
				<button
					use:melt={$nextButton}
					class="rounded-md p-1 transition-all hover:bg-ctp-mauve/20 data-[disabled]:pointer-events-none data-[disabled]:opacity-40"
				>
					<ChevronRight size={24} />
				</button>
			</header>
			<div>
				{#each $months as month}
					<table use:melt={$grid} class="w-full">
						<thead aria-hidden="true">
							<tr>
								{#each $weekdays as day}
									<th class="text-sm font-semibold">
										<div class="flex h-6 w-6 items-center justify-center p-4">
											{day}
										</div>
									</th>
								{/each}
							</tr>
						</thead>
						<tbody>
							{#each month.weeks as weekDates}
								<tr>
									{#each weekDates as date}
										<td
											role="gridcell"
											aria-disabled={$isDateDisabled(date) || $isDateUnavailable(date)}
											class="text-sm font-semibold"
										>
											<div
												use:melt={$cell(date, month.value)}
												class="flex h-6 w-6 cursor-pointer select-none items-center
													justify-center rounded-md p-4 hover:bg-ctp-mauve/20 focus:ring-2 focus:ring-ctp-mauve
													data-[disabled]:pointer-events-none data-[disabled]:opacity-40 data-[unavailable]:text-ctp-red/20
													data-[unavailable]:line-through data-[selected]:bg-ctp-mauve/40 data-[selected]:text-ctp-mauve
													data-[outside-visible-months]:pointer-events-none data-[outside-visible-months]:cursor-default
													data-[outside-visible-months]:opacity-40 data-[outside-visible-months]:hover:bg-transparent
													data-[outside-month]:pointer-events-none data-[outside-month]:cursor-default
													data-[outside-month]:opacity-0 data-[outside-month]:hover:bg-transparent"
											>
												{date.day}
											</div>
										</td>
									{/each}
								</tr>
							{/each}
						</tbody>
					</table>
				{/each}
			</div>
		</div>
	</div>
{/if}

<style lang="postcss">
	[data-melt-datefield-label][data-invalid] {
		@apply text-ctp-red;
	}

	[data-melt-datefield-field][data-invalid] {
		@apply ring-ctp-red ring-2;
	}

	[data-melt-datefield-segment][data-invalid] {
		@apply text-ctp-red;
	}

	[data-melt-datefield-segment]:not([data-segment='literal']) {
		@apply px-0.5;
	}
</style>
