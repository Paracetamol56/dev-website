<script lang="ts">
	import api from '$lib/api';
	import { addToast } from '../../+layout.svelte';
	import type { MeltElement } from '@melt-ui/svelte/internal/helpers';
	import type { Todo } from './proxy+page';
	import Button from '$lib/components/Button.svelte';
  import { createDatePicker, melt } from '@melt-ui/svelte';
  import { ChevronRight, ChevronLeft, Calendar } from 'lucide-svelte';
  import { fade } from 'svelte/transition';
  import { CalendarDateTime } from '@internationalized/date';

	export let todoId: string;
	export let title: MeltElement<any, any, any, any>;
	export let description: MeltElement<any, any, any, any>;

	let todo: Todo | null = null;
	let newTitle = '';
	let titleError = '';
	let newDescription = '';
	let descriptionError = '';
	let newDueDate = '';
	let dueDateError = '';

	$: {
		api.callWithAuth('GET', `/ormi/${todoId}`).then((response) => {
			if (response.status === 200) {
				todo = response.data;
				newTitle = response.data.title;
				newDescription = response.data.description;
				newDueDate = response.data.dueDate;
			} else {
				console.error('Failed to fetch todo:', response);
				addToast({
					data: {
						title: 'Error',
						description: 'Failed to fetch todo.',
						color: 'bg-ctp-peach'
					}
				});
			}
		});
	}

	function validateDescription(description: string) {
		if (description.length > 1000) {
			descriptionError = 'Description must be less than 1000 characters long';
			return false;
		}
		descriptionError = '';
		return true;
	}

	const {
		elements: {
			calendar,
			cell,
			content,
			field,
			grid,
			heading,
			label,
			nextButton,
			prevButton,
			segment,
			trigger
		},
		states: { months, headingValue, weekdays, segmentContents, open },
		helpers: { isDateDisabled, isDateUnavailable }
	} = createDatePicker({
		forceVisible: true,
		defaultPlaceholder: new CalendarDateTime(2021, 2, 1)
	});
</script>

{#if todo !== null}
	<h2 use:melt={$title} class="mb-0 text-lg font-semibold text-ctp-text">
		{todo.title}
	</h2>
	<form use:melt={$description} class="mb-5 mt-2 flex flex-col gap-6">
		<fieldset>
			<label for="description" class="mb-2 text-sm font-semibold">Description</label>
			<textarea
				id="description"
				name="description"
				class="flex h-32 w-full items-center justify-between rounded-md bg-ctp-surface0
                shadow-md shadow-ctp-crust px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
				bind:value={newDescription}
				on:blur={() => validateDescription(newDescription)}
			/>
			<p class="text-left text-sm font-semibold text-ctp-red">{descriptionError}</p>
		</fieldset>
		<fieldset>
			<label for="dueDate" class="mb-2 text-sm font-semibold" use:melt={$label}>Due Date</label>
			<div>
				<div use:melt={$field} class="flex w-full min-w-[200px] items-center rounded-md bg-ctp-surface0 pl-3 p-2
        focus:outline-none focus:ring-2 focus:ring-ctp-mauve text-ctp-text">
					{#each $segmentContents as seg}
						<div use:melt={$segment(seg.part)}>
							{seg.value}
						</div>
					{/each}
					<div class="ml-4 flex w-full items-center justify-end">
						<button use:melt={$trigger} class="rounded-md bg-ctp-mauve p-1 text-ctp-surface0 transition-all hover:bg-ctp-mauve/80">
							<Calendar size={16} />
						</button>
					</div>
				</div>
			</div>
			<p class="text-left text-sm font-semibold text-ctp-red">{dueDateError}</p>
		</fieldset>

		<div class="flex flex-row-reverse gap-2">
			<Button>
				<span>Save</span>
			</Button>
		</div>
	</form>
{:else}
	<h2 use:melt={$title} class="mb-0 text-lg font-semibold text-ctp-text">Loading...</h2>
{/if}

{#if $open}
	<div transition:fade={{ duration: 100 }} use:melt={$content} class="z-10 min-w-[320px]">
		<div use:melt={$calendar} class="w-full rounded-md bg-ctp-mantle p-3 text-ctp-text shadow-md shadow-ctp-crust">
			<header class="flex items-center justify-between pb-2">
				<button use:melt={$prevButton} class="rounded-md p-1 transition-all hover:bg-ctp-mauve/20 data-[disabled]:pointer-events-none data-[disabled]:opacity-40">
					<ChevronLeft size={24} />
				</button>
				<div use:melt={$heading} class="flex items-center gap-6 font-semibold">
					{$headingValue}
				</div>
				<button use:melt={$nextButton} class="rounded-md p-1 transition-all hover:bg-ctp-mauve/20 data-[disabled]:pointer-events-none data-[disabled]:opacity-40">
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
											<div use:melt={$cell(date, month.value)} class="flex h-6 w-6 cursor-pointer select-none items-center
												justify-center rounded-md p-4 hover:bg-ctp-mauve/20 focus:ring-2 focus:ring-ctp-mauve
												data-[disabled]:pointer-events-none data-[disabled]:opacity-40 data-[unavailable]:text-ctp-red/20
												data-[unavailable]:line-through data-[selected]:bg-ctp-mauve/40 data-[selected]:text-ctp-mauve
												data-[outside-visible-months]:pointer-events-none data-[outside-visible-months]:cursor-default
												data-[outside-visible-months]:opacity-40 data-[outside-visible-months]:hover:bg-transparent
												data-[outside-month]:pointer-events-none data-[outside-month]:cursor-default
												data-[outside-month]:opacity-0 data-[outside-month]:hover:bg-transparent">
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
  [data-melt-popover-content] {
    @apply z-10 min-w-[320px] rounded-lg bg-neutral-900 shadow-sm;
  }

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

  [data-melt-datefield-validation] {
    @apply self-start text-ctp-red;
  }
</style>