<script lang="ts">
  import { createSelect, melt, type SelectOption } from '@melt-ui/svelte';
  import { Check, ChevronDown } from 'lucide-svelte';
	import type { Writable } from 'svelte/store';
  import { fade } from 'svelte/transition';
 
  export let name: string;
  export let options: string[];
  export let value: Writable<SelectOption<string>>;
  export let disabled: boolean = false;
 
  const {
    elements: { trigger, menu, option, label },
    states: { selectedLabel, open },
    helpers: { isSelected },
  } = createSelect({
    selected: value,
    forceVisible: true,
    disabled,
    positioning: {
      placement: 'bottom',
      fitViewport: true,
      sameWidth: true,
    },
  });
;
</script>
 
<div class="flex flex-col gap-1">
  <!-- svelte-ignore a11y-label-has-associated-control - $label contains the 'for' attribute -->
  <label class="block text-ctp-text font-semibold text-sm" use:melt={$label}>{name}</label>
  <button
    class="flex min-w-[220px] items-center justify-between rounded-md bg-ctp-surface0 px-3 py-1
        shadow-md shadow-ctp-crust transition-opacity hover:opacity-90"
    use:melt={$trigger}
    aria-label="Food"
  >
    {$selectedLabel || 'Select an option'}
    <ChevronDown class="square-5" />
  </button>
  {#if $open}
    <div
      class="z-10 flex max-h-[300px] flex-col
        overflow-y-auto rounded-md bg-ctp-surface0 p-1
        shadow-md shadow-ctp-crust focus:!ring-0"
      use:melt={$menu}
      transition:fade={{ duration: 150 }}
    >
      <div>
        {#each options as item}
          <div
            class="relative cursor-pointer rounded-md py-1 pl-8 pr-4
              focus:z-10
              data-[highlighted]:bg-ctp-mauve/25 data-[selected]:bg-ctp-mauve/25
              data-[highlighted]:text-ctp-mauve data-[selected]:text-ctp-mauve"
            use:melt={$option({ value: item, label: item })}
          >
            <div class="absolute left-[0.5rem] top-1/2 transform -translate-y-1/2 z-20 text-ctp-mauve {$isSelected(item) ? 'block' : 'hidden'}">
              <Check class="square-4" />
            </div>

            {item}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>
