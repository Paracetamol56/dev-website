<script lang="ts">
  import { createCheckbox, melt } from '@melt-ui/svelte';
  import { Check, Minus } from 'lucide-svelte';
	import type { Writable } from 'svelte/store';

  export let defaultChecked: boolean | 'indeterminate' = false;
  export let checked: Writable<boolean>;
  export let name: string;
  export let required: boolean = false;
  export let disabled: boolean = false;
  export let label: string;

  const {
    elements: { root, input },
    helpers: { isChecked, isIndeterminate },
  } = createCheckbox({
    defaultChecked,
    checked,
    name,
    required,
    disabled,
  });
</script>

  <div class="flex items-center gap-2">
    <button
      use:melt={$root}
      class="grid square-4 cursor-default place-items-center rounded-md bg-ctp-surface0 shadow-sm shadow-ctp-crust
                  hover:bg-ctp-surface1"
      id={name}
    >
      {#if $isIndeterminate}
        <Minus class="square-3 stroke-[4] text-ctp-mauve" />
      {:else if $isChecked}
        <Check class="square-3 stroke-[4] text-ctp-mauve" />
      {/if}
      <input use:melt={$input} />
    </button>
    <label class="font-semibold leading-none" for={name}>
      {label}
    </label>
  </div>