<script lang="ts">
	import { createPopover, createRadioGroup, melt, type CreateRadioGroupProps } from '@melt-ui/svelte';
	import { fade } from 'svelte/transition';
  import { variants } from '@catppuccin/palette';
	import { User, X } from 'lucide-svelte';

	const {
		elements: { trigger, content, arrow, close },
		states: { open }
	} = createPopover({
		forceVisible: true
	});
</script>

<button
	type="button"
	class="inline-flex square-8 items-center justify-center rounded-full bg-ctp-mauve p-0
        text-sm font-medium text-ctp-mantle hover:opacity-75 active:opacity-50 transition-opacity"
	use:melt={$trigger}
	aria-label="Update dimensions"
>
	<User size="20" />
	<span class="sr-only">Open profile popover</span>
</button>

{#if $open}
  <div
    use:melt={$content}
    transition:fade={{ duration: 100 }}
    class="z-10 w-60 rounded-[4px] bg-ctp-base p-5 shadow-md shadow-ctp-crust"
  >
    <div use:melt={$arrow} />
    <div class="flex flex-col gap-2.5">
      <p class="mb-2 font-semibold text-ctp-text">Profile</p>
      <p>TODO</p>
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
