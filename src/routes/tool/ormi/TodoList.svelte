<script lang="ts">
	import { GripHorizontal, Plus } from "lucide-svelte";
	import { flip } from "svelte/animate";
	import TodoContent from "./TodoContent.svelte";
	import type { Todo } from "./+page";
	import { writable, type Writable } from "svelte/store";
	import AddDialog from "./AddDialog.svelte";

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
  }
  
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
  }

  const startMoving = (event: MouseEvent, index: number) => {
    event.preventDefault();
    listoffset = list?.getBoundingClientRect().top;
    mouseY = event.clientY - listoffset;
    moving = index;
    const target = event.currentTarget as HTMLElement;
    startY = target.getBoundingClientRect().top - listoffset + 20;
  }

  const endMoving = (event: MouseEvent) => {
    event.preventDefault();
    if (moving === null) return;
    moving = null;
  }
</script>

<svelte:window on:mousemove={updateMouseY} on:mouseup={endMoving} />

<div class="flex gap-2">
  <AddDialog />
</div>

<ul class="relative h-full flex flex-col gap-4" bind:this={list}>
  {#each todos as todo, index (todo.id)}
    <li
      class="rounded-md bg-ctp-mantle/50 min-h-[40px]"
      animate:flip={{ duration: moving === index ? 0 : 150 }}
      data-index={index}
    >
      <div
        class="bg-ctp-mantle rounded-md overflow-hidden {$focusedTodoId === todo.id ? 'ring-2 ring-ctp-mauve' : ''} {moving !== null ? 'pointer-events-none' : ''}
        {moving === index ? 'absolute left-0 right-0 transform scale-[99%] z-10' : ''}"
        style={moving === index ? `top: calc(${mouseY}px - 20px)` : ''}
      >
        <button
          class="w-full px-3 py-2 hover:text-ctp-lavender hover:cursor-pointer flex items-center gap-2"
          
          on:click={() => focusedTodoId.set(todo.id)}
        >
          <TodoContent {todo} />
          <div
            class="ml-auto"
            role="button"
            tabindex="0"
            on:mousedown={event => startMoving(event, index)}
          >
            <GripHorizontal class="cursor-grab active:cursor-grabbing opacity-50 active:opacity-100 transition-opacity"
              size="18" stroke-width="3" />
          </div>
        </button>
        {#if $focusedTodoId === todo.id}
          <div class="bg-ctp-crust p-2">
            <h4 class="text-xl font-semibold">{todo.title}</h4>
            <p>{todo.description}</p>
          </div>
        {/if}
      </div>
    </li>
  {/each}
</ul>
