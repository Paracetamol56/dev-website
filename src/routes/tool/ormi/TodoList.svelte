<script lang="ts">
	import { GripHorizontal, X } from 'lucide-svelte';
	import TodoContent from './TodoContent.svelte';
	import type { Todo } from './+page';
	import { writable, type Writable } from 'svelte/store';
	import AddForm from './AddForm.svelte';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { fade, fly } from 'svelte/transition';
	import TodoDetail from './TodoDetail.svelte';

	export let todos: Todo[];
	export let focusedTodoId: Writable<string | null> = writable(null);

	let list: HTMLUListElement;
	let listoffset: number;
	let moving: number | null = null;
	let startY: number;
	let mouseY: number;

	const flipListItems = (a: number, b: number) => {
		const tmp = todos[a];
		todos[a] = todos[b];
		todos[b] = tmp;
	};

	const updateMouseY = (event: MouseEvent) => {
		event.preventDefault();
		if (moving === null) return;
		mouseY = event.clientY - listoffset;
		const deltaY = mouseY - startY;
		if (deltaY < -30 && moving > 0) {
			flipListItems(moving, moving - 1);
			moving--;
			const target = list.querySelector(`[data-index="${moving}"]`) as HTMLElement;
			if (target) {
				startY = target.getBoundingClientRect().top - listoffset + 20;
			}
		} else if (deltaY > 30 && moving < todos.length - 1) {
			flipListItems(moving, moving + 1);
			moving++;
			const target = list.querySelector(`[data-index="${moving}"]`) as HTMLElement;
			if (target) {
				startY = target.getBoundingClientRect().top - listoffset + 20;
			}
		}
	};

	const startMoving = (event: MouseEvent, index: number) => {
		event.preventDefault();
		listoffset = list?.getBoundingClientRect().top;
		mouseY = event.clientY - listoffset;
		moving = index;
		const target = event.currentTarget as HTMLElement;
		startY = target.getBoundingClientRect().top - listoffset + 20;
	};

	const endMoving = (event: MouseEvent) => {
		event.preventDefault();
		if (moving === null) return;
		moving = null;
	};

	const {
    elements: {
      trigger,
      overlay,
      content,
      title,
      description,
      close,
      portalled,
    },
    states: { open },
  } = createDialog({
    forceVisible: true,
  });
</script>

<svelte:window on:mousemove={updateMouseY} on:mouseup={endMoving} />
<div class="mb-16">
	<AddForm />
	<ul class="relative mb-4 h-full flex flex-col gap-4" bind:this={list}>
		{#each todos as todo, index (todo.id)}
			<li class="rounded-md bg-ctp-mantle/50 min-h-[40px]" data-index={index}>
				<div
					class="h-min bg-ctp-mantle rounded-md overflow-hidden {$focusedTodoId ===
					todo.id
						? 'ring-2 ring-ctp-mauve'
						: ''} {moving !== null ? 'pointer-events-none' : ''}
        {moving === index ? 'absolute left-0 right-0 transform scale-[99%] z-10' : 'transition-all'}"
					style={moving === index ? `top: calc(${mouseY}px - 20px)` : ''}
				>
					<button
						class="w-full px-3 py-2 hover:text-ctp-lavender hover:cursor-pointer flex items-center gap-2"
						on:click={() =>
							$focusedTodoId === todo.id ? ($focusedTodoId = null) : ($focusedTodoId = todo.id)}
						use:melt={$trigger}
					>
						<TodoContent {todo} />
						<div
							class="ml-auto"
							role="button"
							tabindex="0"
							on:mousedown={(event) => startMoving(event, index)}
						>
							<GripHorizontal
								class="cursor-grab active:cursor-grabbing opacity-50 active:opacity-100 transition-opacity"
								size="18"
								stroke-width="3"
							/>
						</div>
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
        opacity: 1,
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
      <TodoDetail todoId={$focusedTodoId ?? ""} {title} {description} />
    </div>
  </div>
{/if}