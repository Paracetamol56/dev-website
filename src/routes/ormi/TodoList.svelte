<script lang="ts">
	import { createCheckbox, melt } from "@melt-ui/svelte";
	import type { Todo } from "./todo";
	import { Check, Dot, GripHorizontal } from "lucide-svelte";
	import { flip } from "svelte/animate";

  export let todos: Todo[];

  // const {
  //   elements: { root, input },
  //   helpers: { isChecked, isIndeterminate },
  //   states: { checked },
  // } = createCheckbox();

  // switch (todo.state) {
  //   case TodoState.Backlog:
  //     $checked = false;
  //   case TodoState.Pending:
  //     $checked = false;
  //   case TodoState.InProgress:
  //     $checked = 'indeterminate';
  //   case TodoState.Done:
  //     $checked = true;
  //   default:
  //     $checked = false;
  // }

  let list: HTMLUListElement;
  $: listoffset = list?.getBoundingClientRect().top;
  let moving: number | null = null;
  let startY: number;
  let mouseY: number;

  const updateMouseY = (event: MouseEvent) => {
    mouseY = event.clientY - listoffset;
    if (moving === null) return;
    const deltaY = mouseY - startY;
    if (deltaY < -30 && moving > 0) {
      const todo = todos[moving];
      todos[moving] = todos[moving - 1];
      todos[moving - 1] = todo;
      moving--;
    } else if (deltaY > 30 && moving < todos.length - 1) {
      const todo = todos[moving];
      todos[moving] = todos[moving + 1];
      todos[moving + 1] = todo;
      moving++;
    }
    // Update startY
    const target = document.querySelector(`[data-index="${moving}"]`) as HTMLElement;
    if (target) {
      startY = document.querySelector(`[data-index="${moving}"]`)!.getBoundingClientRect().top - listoffset + 20;
    }
  }

  const startMoving = (event: MouseEvent, index: number) => {
    moving = index;
    const target = event.currentTarget as HTMLElement;
    startY = target.getBoundingClientRect().top - listoffset + 20;
  }

  const endMoving = (event: MouseEvent) => {
    if (moving === null) return;
    moving = null;
  }
</script>

<svelte:window on:mousemove={updateMouseY} on:mouseup={endMoving} />

<ul class="relative flex flex-col gap-4" bind:this={list}>
  {#each todos as todo, index (todo._id)}
    <li
      class="rounded-md bg-ctp-mantle/50 min-h-[40px]"
      animate:flip={{ duration: moving === index ? 0 : 200 }}
      data-index={index}
    >
      <div
        class="bg-ctp-mantle rounded-md px-3 py-2 flex items-center gap-2 {moving !== null ? 'pointer-events-none' : ''}
              {moving === index ? 'absolute left-0 right-0 transform scale-[99%] z-10' : ''}"
        style={moving === index ? `top: calc(${mouseY}px - 20px)` : ''}
      >
        <!-- <button
          use:melt={$root}
          class="flex appearance-none items-center justify-center rounded-md
                bg-ctp-crust shadow square-5 hover:opacity-75"
          id="checkbox"
        >
          {#if $isIndeterminate}
            <Dot size="16" stroke-width="3" />
          {:else if $isChecked}
            <Check size="16" stroke-width="3" />
          {/if}
          <input use:melt={$input} />
        </button> -->
        <span class="font-semibold">{todo.name}</span>
        <div
          class="ml-auto"
          role="button"
          tabindex="0"
          on:mousedown={event => startMoving(event, index)}
        >
          <GripHorizontal class="cursor-grab active:cursor-grabbing opacity-50 active:opacity-100 transition-opacity"
            size="18" stroke-width="3" />
        </div>
      </div>
    </li>
  {/each}
</ul>
