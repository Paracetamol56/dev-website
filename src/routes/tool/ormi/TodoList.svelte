<script lang="ts">
	import { Check, GripHorizontal, X } from 'lucide-svelte';
	import TodoContent from './TodoContent.svelte';
	import type { Todo } from './+page';
	import AddForm from './AddForm.svelte';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { fade, fly } from 'svelte/transition';
	import TodoDetail from './TodoDetail.svelte';
	import api from '$lib/api';
	import { addToast } from '../../+layout.svelte';
	import { invalidateAll } from '$app/navigation';
	import { tick } from 'svelte';

	export let todos: Todo[];

	let items: Todo[] = todos;
	$: items = todos;

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

	const completeTodo = (todo: Todo) => {
		api
			.callWithAuth('PATCH', `/ormi/${todo.id}`, { state: 'DONE' })
			.then(() => invalidateAll())
			.catch((error) => {
				console.error('Failed to complete todo:', error);
				showError('Failed to complete todo.');
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
<div class="mb-16">
	<AddForm />
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
					<button
						role="checkbox"
						aria-checked="false"
						aria-label="Mark as done"
						class="group flex shrink-0 items-center justify-center rounded-md bg-ctp-crust shadow square-5 hover:opacity-75"
						on:click={() => completeTodo(todo)}
					>
						<Check class="opacity-0 group-hover:opacity-50" size="16" stroke-width="3" />
					</button>
					<button
						class="flex flex-1 items-center gap-2 hover:text-ctp-lavender hover:cursor-pointer"
						on:click={() => (focusedTodoId = todo.id)}
						use:melt={$trigger}
					>
						<TodoContent {todo} />
					</button>
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
				</div>
			</li>
		{/each}
	</ul>
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
			class="fixed right-0 top-0 z-50 h-screen w-full max-w-[350px] bg-ctp-base p-6
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
