<script lang="ts">
	import { createCheckbox, melt } from "@melt-ui/svelte";
	import { CalendarClock, CalendarX2, Check, Dot } from "lucide-svelte";
	import type { Todo } from "./+page";
	import MeltTooltip from "$lib/components/MeltTooltip.svelte";

  export let todo: Todo;

  const {
    elements: { root, input },
    helpers: { isChecked, isIndeterminate },
    states: { checked },
  } = createCheckbox();

  function toRelative(date: Date) {
    const diff = new Date(date).getTime() - new Date().getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    if (days === 0) {
      return "today";
    } else if (days === 1) {
      return "tomorrow";
    } else if (days === -1) {
      return "yesterday";
    } else if (days > 0) {
      return `in ${days} days`;
    } else {
      return `${-days} days ago`;
    }
  }
</script>

<button
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
</button>
<span class="font-semibold mr-auto">{todo.title}</span>
{#if todo.dueDate !== undefined && todo.dueDate !== null}
  <MeltTooltip text={toRelative(todo.dueDate)}>
    {#if new Date(todo.dueDate) < new Date()}
    <CalendarX2 size="16" />
    {:else}
    <CalendarClock size="16" />
    {/if}
  </MeltTooltip>
{/if}

