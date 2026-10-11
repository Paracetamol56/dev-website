<script lang="ts">
	import {
		ArrowDown,
		ArrowUp,
		ArrowUpDown,
		Ban,
		Check,
		FilterX,
		GripHorizontal,
		X
	} from 'lucide-svelte';
	import TodoContent from './TodoContent.svelte';
	import type { Todo } from './+page';
	import AddForm from './AddForm.svelte';
	import { createDialog, createSelect, melt } from '@melt-ui/svelte';
	import { getLocalTimeZone } from '@internationalized/date';
	import MeltTags from '$lib/components/MeltTags.svelte';
	import MeltDatePicker from '$lib/components/MeltDatePicker.svelte';
	import { suggestLabels } from './labels';
	import { from, labels, setFilters, sort, sortOptions, states, to } from './filters';
	import { fade, fly } from 'svelte/transition';
	import TodoDetail from './TodoDetail.svelte';
	import api from '$lib/api';
	import { addToast } from '../../+layout.svelte';
	import { invalidateAll } from '$app/navigation';
	import { tick } from 'svelte';
	import { closedStates, openStates, todoStates, type TodoState } from './states';

	export let refresh: unknown;

	const PAGE_SIZE = 50;
	const allStates = Object.keys(todoStates) as TodoState[];

	let items: Todo[] = [];
	let total = 0;
	$: hasMore = items.length < total;
	let lastRequest = 0;

	$: localStorage.setItem('ormi-sort', $sort.value);
	let descending = false;
	$: descending = ['created', 'updated'].includes($sort.value);

	const {
		elements: { trigger: sortTrigger, menu: sortMenu, option: sortOption },
		states: { open: sortOpen }
	} = createSelect({
		selected: sort,
		forceVisible: true,
		positioning: { placement: 'bottom-end' }
	});

	const load = (reset: boolean) => {
		if ($states.length === 0) {
			items = [];
			total = 0;
			return;
		}
		const params = new URLSearchParams({
			sort: $sort.value,
			order: descending ? 'desc' : 'asc',
			limit: String(PAGE_SIZE),
			offset: String(reset ? 0 : items.length)
		});
		$states.forEach((state) => params.append('state', state));
		$labels.forEach((label) => params.append('label', label.value));
		if ($from) params.set('from', $from.toDate(getLocalTimeZone()).toISOString());
		if ($to) params.set('to', $to.add({ days: 1 }).toDate(getLocalTimeZone()).toISOString());

		const request = ++lastRequest;
		api
			.callWithAuth('GET', `/ormi?${params}`)
			.then((response) => {
				if (request !== lastRequest) return;
				items = reset ? response.data.items : [...items, ...response.data.items];
				total = response.data.total;
			})
			.catch((error) => {
				console.error('Failed to fetch todos:', error);
				showError('Failed to fetch todos.');
			});
	};

	$: refresh, $states, $labels, $from, $to, $sort, descending, load(true);

	$: defaultFilters =
		$states.length === openStates.length &&
		openStates.every((state) => $states.includes(state)) &&
		$labels.length === 0 &&
		!$from &&
		!$to;
	$: canReorder = defaultFilters && $sort.value === 'position' && !descending && !hasMore;

	const toggleState = (state: TodoState) => {
		$states = $states.includes(state)
			? $states.filter((other) => other !== state)
			: [...$states, state];
	};

	let focusedTodoId: string | null = null;

	let list: HTMLUListElement;
	let listoffset: number;
	let moving: number | null = null;
	let moved = false;
	let startY: number;
	let pointerY: number;

	const showError = (description: string) => {
		addToast({
			data: {
				title: 'Error',
				description,
				color: 'bg-ctp-red'
			}
		});
	};

	const flipListItems = (a: number, b: number) => {
		const tmp = items[a];
		items[a] = items[b];
		items[b] = tmp;
	};

	const saveOrder = () => {
		api.callWithAuth('PUT', '/ormi/order', { ids: items.map((todo) => todo.id) }).catch((error) => {
			console.error('Failed to save order:', error);
			showError('Failed to save the new order.');
			invalidateAll();
		});
	};

	const moveTo = (target: number) => {
		if (moving === null) return;
		flipListItems(moving, target);
		moving = target;
		moved = true;
		const row = list.querySelector(`[data-index="${moving}"]`) as HTMLElement;
		if (row) {
			startY = row.getBoundingClientRect().top - listoffset + 20;
		}
	};

	const updatePointerY = (event: PointerEvent) => {
		if (moving === null) return;
		event.preventDefault();
		pointerY = event.clientY - listoffset;
		const deltaY = pointerY - startY;
		if (deltaY < -30 && moving > 0) {
			moveTo(moving - 1);
		} else if (deltaY > 30 && moving < items.length - 1) {
			moveTo(moving + 1);
		}
	};

	const startMoving = (event: PointerEvent, index: number) => {
		event.preventDefault();
		listoffset = list?.getBoundingClientRect().top;
		pointerY = event.clientY - listoffset;
		moving = index;
		moved = false;
		const target = event.currentTarget as HTMLElement;
		startY = target.getBoundingClientRect().top - listoffset + 20;
	};

	const endMoving = () => {
		if (moving === null) return;
		moving = null;
		if (moved) saveOrder();
	};

	const moveWithKeyboard = async (event: KeyboardEvent, index: number) => {
		const target = event.key === 'ArrowUp' ? index - 1 : event.key === 'ArrowDown' ? index + 1 : -1;
		if (target < 0 || target >= items.length) return;
		event.preventDefault();
		flipListItems(index, target);
		saveOrder();
		await tick();
		(list.querySelector(`[data-index="${target}"] [data-handle]`) as HTMLElement)?.focus();
	};

	const setState = (todo: Todo, state: TodoState) => {
		api
			.callWithAuth('PATCH', `/ormi/${todo.id}`, { state })
			.then(() => invalidateAll())
			.catch((error) => {
				console.error('Failed to update todo:', error);
				showError('Failed to update todo.');
			});
	};

	const {
		elements: { trigger, overlay, content, title, description, close, portalled },
		states: { open }
	} = createDialog({
		forceVisible: true
	});
</script>

<svelte:window
	on:pointermove={updatePointerY}
	on:pointerup={endMoving}
	on:pointercancel={endMoving}
/>
<div class="mb-16 scroll-mt-24" id="todos">
	<AddForm />
	<div class="mb-6 flex flex-col gap-4 rounded-md bg-ctp-mantle p-4">
		<div class="flex flex-wrap items-center gap-x-4 gap-y-2">
			{#each allStates as state}
				<button
					type="button"
					role="checkbox"
					aria-checked={$states.includes(state)}
					class="flex items-center gap-1.5 rounded-md bg-ctp-base px-3 py-1 text-sm font-semibold transition-colors
						hover:bg-ctp-surface0 aria-checked:ring-2 aria-checked:ring-ctp-mauve"
					on:click={() => toggleState(state)}
				>
					<svelte:component
						this={todoStates[state].icon}
						size="14"
						class={todoStates[state].color}
					/>
					{todoStates[state].label}
				</button>
			{/each}
			<button
				type="button"
				class="ml-auto flex items-center gap-2 rounded-md px-3 py-1 text-sm font-semibold transition-colors hover:bg-ctp-surface0"
				aria-label="Sort todos"
				use:melt={$sortTrigger}
			>
				<ArrowUpDown size="14" />
				{$sort.label}
			</button>
			<button
				type="button"
				class="flex items-center gap-2 rounded-md px-3 py-1 text-sm font-semibold transition-colors hover:bg-ctp-surface0 disabled:opacity-40 disabled:hover:bg-transparent"
				disabled={$sort.value === 'position'}
				on:click={() => (descending = !descending)}
			>
				{#if descending}
					<ArrowDown size="14" />
					Descending
				{:else}
					<ArrowUp size="14" />
					Ascending
				{/if}
			</button>
		</div>
		<div class="grid gap-4 md:grid-cols-3">
			<div>
				<span class="mb-2 text-sm font-semibold">Labels</span>
				<MeltTags tags={labels} placeholder="Search labels..." suggest={suggestLabels} />
			</div>
			<div>
				<MeltDatePicker value={from} label="Changed from" />
			</div>
			<div>
				<MeltDatePicker value={to} label="Changed until" />
			</div>
		</div>
		{#if !defaultFilters}
			<button
				type="button"
				class="flex w-fit items-center gap-2 rounded-md px-3 py-1 text-sm font-semibold transition-colors hover:bg-ctp-surface0"
				on:click={() => setFilters()}
			>
				<FilterX size="14" />
				Reset filters
			</button>
		{/if}
	</div>
	{#if $sortOpen}
		<div
			class="z-10 flex flex-col gap-1 rounded-md bg-ctp-surface0 p-1 shadow-md shadow-ctp-crust focus:!ring-0"
			use:melt={$sortMenu}
			transition:fade={{ duration: 150 }}
		>
			{#each sortOptions as option}
				<div
					class="cursor-pointer rounded-md px-3 py-1 text-sm data-[highlighted]:bg-ctp-mauve/25
						data-[selected]:bg-ctp-mauve/25 data-[highlighted]:text-ctp-mauve data-[selected]:text-ctp-mauve"
					use:melt={$sortOption(option)}
				>
					{option.label}
				</div>
			{/each}
		</div>
	{/if}
	<ul class="relative mb-4 h-full flex flex-col gap-4" bind:this={list}>
		{#each items as todo, index (todo.id)}
			<li class="rounded-md bg-ctp-mantle/50 min-h-[40px]" data-index={index}>
				<div
					class="h-min bg-ctp-mantle rounded-md overflow-hidden px-3 py-2 flex items-center gap-2
						{focusedTodoId === todo.id && $open ? 'ring-2 ring-ctp-mauve' : ''}
						{moving !== null ? 'pointer-events-none' : ''}
						{moving === index ? 'absolute left-0 right-0 transform scale-[99%] z-10' : 'transition-all'}"
					style={moving === index ? `top: calc(${pointerY}px - 20px)` : ''}
				>
					{#if !closedStates.includes(todo.state)}
						<button
							role="checkbox"
							aria-checked="false"
							aria-label="Mark as done"
							class="group flex shrink-0 items-center justify-center rounded-md bg-ctp-crust shadow square-5 hover:opacity-75"
							on:click={() => setState(todo, 'DONE')}
						>
							<Check class="opacity-0 group-hover:opacity-50" size="16" stroke-width="3" />
						</button>
					{:else}
						<button
							role="checkbox"
							aria-checked="true"
							aria-label="Reopen"
							class="flex shrink-0 items-center justify-center rounded-md bg-ctp-crust shadow square-5 hover:opacity-75"
							on:click={() => setState(todo, 'TODO')}
						>
							{#if todo.state === 'CANCELLED'}
								<Ban size="14" stroke-width="3" />
							{:else}
								<Check size="16" stroke-width="3" />
							{/if}
						</button>
					{/if}
					<button
						class="flex flex-1 items-center gap-2 hover:text-ctp-lavender hover:cursor-pointer"
						on:click={() => (focusedTodoId = todo.id)}
						use:melt={$trigger}
					>
						<TodoContent {todo} />
					</button>
					{#if canReorder}
						<button
							type="button"
							data-handle
							aria-label="Reorder, use the arrow keys to move up or down"
							class="touch-none cursor-grab active:cursor-grabbing opacity-50 hover:opacity-100 focus:opacity-100 transition-opacity"
							on:pointerdown={(event) => startMoving(event, index)}
							on:keydown={(event) => moveWithKeyboard(event, index)}
						>
							<GripHorizontal size="18" stroke-width="3" />
						</button>
					{/if}
				</div>
			</li>
		{/each}
	</ul>
	{#if items.length === 0}
		<p class="text-center text-sm text-ctp-subtext0">Nothing here yet.</p>
	{/if}
	{#if hasMore}
		<button
			type="button"
			class="mx-auto block rounded-md px-3 py-1 text-sm font-semibold hover:bg-ctp-surface0"
			on:click={() => load(false)}
		>
			Load more ({items.length} of {total})
		</button>
	{/if}
</div>

{#if $open}
	<div class="" use:melt={$portalled}>
		<div
			use:melt={$overlay}
			class="fixed inset-0 z-50 bg-black/50"
			transition:fade={{ duration: 150 }}
		/>
		<div
			use:melt={$content}
			class="fixed right-0 top-0 z-50 h-screen w-full max-w-[350px] overflow-y-auto bg-ctp-base p-6
            shadow-md focus:outline-none"
			transition:fly={{
				x: 350,
				duration: 300,
				opacity: 1
			}}
		>
			<button
				use:melt={$close}
				aria-label="Close"
				class="absolute right-4 top-4 inline-flex h-6 w-6 appearance-none
                items-center justify-center rounded-full p-1 text-base
                hover:bg-ctp-mauve hover:text-ctp-base transition-colors"
			>
				<X class="size-4" />
			</button>
			{#if focusedTodoId}
				<TodoDetail todoId={focusedTodoId} {title} {description} />
			{/if}
		</div>
	</div>
{/if}
