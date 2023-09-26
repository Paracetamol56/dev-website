<script lang="ts">
	import { createPopover, createRadioGroup, melt, type CreateRadioGroupProps } from '@melt-ui/svelte';
	import { fade } from 'svelte/transition';
  import { variants } from '@catppuccin/palette';
	import { Palette, Settings2, X } from 'lucide-svelte';

  export let theme: Writable<string>|undefined;

	const {
		elements: { trigger, content, arrow, close },
		states: { open }
	} = createPopover({
		forceVisible: true
	});

	const {
		elements: { root, item, hiddenInput },
    states: { value },
		helpers: { isChecked }
	} = createRadioGroup({
    value: theme,
	});
</script>

<button
	type="button"
	class="inline-flex square-8 items-center justify-center rounded-full bg-transparent p-0
        text-sm font-medium text-ctp-text transition-colors hover:bg-ctp-mauve/20"
	use:melt={$trigger}
	aria-label="Update dimensions"
>
	<Palette size="16" />
	<span class="sr-only">Open theme popover</span>
</button>

{#if $open}
  <div
    use:melt={$content}
    transition:fade={{ duration: 100 }}
    class="z-10 w-60 rounded-[4px] bg-ctp-overlay0 p-5 shadow-sm"
  >
    <div use:melt={$arrow} />
    <div class="flex flex-col gap-2.5">
      <p class="mb-2 font-medium text-ctp-text">Theme</p>

      <div
        use:melt={$root}
        class="flex flex-col gap-3 data-[orientation=horizontal]:flex-row"
        aria-label="View density"
      >
        {#each Object.keys(variants) as option}
          <div class="flex items-center gap-3">
            <button
              use:melt={$item(option)}
              class="grid h-6 w-6 cursor-default place-items-center rounded-full bg-white shadow-sm
                      hover:bg-ctp-blue transition-colors"
              id={option}
              aria-labelledby="{option}-label"
            >
              {#if $isChecked(option)}
                <div class="h-3 w-3 rounded-full bg-ctp-mauve" />
              {/if}
            </button>
            <label
              class="font-medium capitalize leading-none text-magnum-900"
              for={option}
              id="{option}-label"
            >
              {option}
            </label>
          </div>
        {/each}
        <input name="line-height" use:melt={$hiddenInput} />
      </div>

      <p class="font-medium text-sm text-ctp-text">Powered by Catppuccin</p>
    </div>
    <button
      use:melt={$close}
      aria-label="close"
      class="absolute right-4 top-4 inline-flex h-6 w-6 appearance-none
                items-center justify-center rounded-full p-1 text-base
                hover:bg-ctp-mauve hover:text-ctp-base transition-colors"
    >
      <X class="square-4" />
    </button>
  </div>
{/if}
