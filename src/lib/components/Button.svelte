<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let disabled: boolean = false;

  const dispatch = createEventDispatcher();

  function handleClick() {
    if (disabled) return;
    dispatch("click");
  }
</script>

<button
  class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
        font-semibold text-ctp-mantle
        shadow-md shadow-ctp-crust transition-opacity
        {disabled ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-80 active:opacity-60'}"
  on:click={handleClick}
>
  <slot></slot>
</button>
