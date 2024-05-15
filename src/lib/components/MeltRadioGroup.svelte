<script lang="ts">
  import { createRadioGroup, melt } from '@melt-ui/svelte';
	import type { Writable } from 'svelte/store';

  export let options: { name: string; label: string }[];
  export let required: boolean = false;
  export let disabled: boolean = false;
  export let orientation: 'horizontal' | 'vertical' = 'horizontal';
  export let value: Writable<string>;

  const {
    elements: { root, item, hiddenInput },
    helpers: { isChecked },
  } = createRadioGroup({
    required,
    disabled,
    orientation,
    value,
  });
</script>

<div
  use:melt={$root}
  class="flex flex-col gap-3 data-[orientation=horizontal]:flex-row"
  aria-label="Radio group"
>
  {#each options as option}
    <div class="flex items-center gap-2">
      <button
        use:melt={$item(option.name)}
        class="grid square-3.5 cursor-default place-items-center rounded-full bg-ctp-surface0 shadow-sm shadow-ctp-crust
                  hover:bg-ctp-surface1"
        id={option.name}
        aria-labelledby="{option.name}-label"
      >
        {#if $isChecked(option.name)}
          <div class="square-2 rounded-full bg-ctp-mauve" />
        {/if}
      </button>
      <label
        class="leading-none font-semibold"
        for={option.name}
        id="{option}-label"
      >
        {option.label}
      </label>
    </div>
  {/each}
  <input name="line-height" use:melt={$hiddenInput} />
</div>