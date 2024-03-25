
<script lang="ts">
	import { createCheckbox, melt } from "@melt-ui/svelte";
	import { Check, Dot } from "lucide-svelte";
	import type { Todo } from "./+page";

  export let todo: Todo;

  const {
    elements: { root, input },
    helpers: { isChecked, isIndeterminate },
    states: { checked },
  } = createCheckbox();
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
<span class="font-semibold">{todo.title}</span>
