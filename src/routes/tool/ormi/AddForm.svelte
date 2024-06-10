<script lang="ts">
	import { Plus } from 'lucide-svelte';
	import api from '$lib/api';
	import { createDatePicker, melt } from '@melt-ui/svelte';
	import { ChevronRight, ChevronLeft, Calendar } from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { addToast } from '../../+layout.svelte';
	import { invalidateAll } from '$app/navigation';
	import { today } from '@internationalized/date';

	let newTodoName = '';
	const addTodo = () => {
		if (newTodoName === '') return;
		console.log('Adding todo:', $value);
		api
			.callWithAuth('POST', '/ormi', {
				title: newTodoName,
				dueDate: $value ? $value.toDate('UTC') : null
			})
			.then((response) => {
				if (response.status === 201) {
					invalidateAll();
				} else {
					console.error('Failed to add todo:', response);
					addToast({
						data: {
							title: 'Error',
							description: 'Failed to add todo.',
							color: 'bg-ctp-peach'
						}
					});
				}
			})
			.catch((error) => {
				console.error('Failed to add todo:', error);
				addToast({
					data: {
						title: 'Error',
						description: 'Failed to add todo.',
						color: 'bg-ctp-peach'
					}
				});
			});
		newTodoName = '';
		$value = undefined;
	};

	const {
		elements: { calendar, cell, content, grid, heading, nextButton, prevButton, trigger },
		states: { months, headingValue, weekdays, open, value },
		helpers: { isDateDisabled, isDateUnavailable },
	} = createDatePicker({
		forceVisible: true,
		minValue: today('UTC'),
	});
</script>

<form
	class="mb-6 rounded-md bg-ctp-mantle shadow-md shadow-ctp-crust min-h-[40px] px-3 py-2 flex gap-2"
	on:submit|preventDefault={addTodo}
>
	<input
		type="text"
		class="w-full bg-transparent outline-none"
		placeholder="Add a new todo..."
		bind:value={newTodoName}
		on:keydown={(event) => event.key === 'Enter' && addTodo()}
	/>
	<button use:melt={$trigger} class="transition-opacity hover:opacity-80 active:opacity-60 text-ctp-text" type="button">
		<Calendar size={16} />
	</button>
	<button
		class="transition-opacity hover:opacity-80 active:opacity-60 text-ctp-mauve"
		type="submit"
	>
		<Plus size="18" />
	</button>
</form>

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
