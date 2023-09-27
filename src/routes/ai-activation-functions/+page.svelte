<script lang="ts">
	import { createTabs, melt } from "@melt-ui/svelte";
	import { crossfade } from "svelte/transition";
	import { cubicInOut } from "svelte/easing";
	import Sigmoid from "./Sigmoid.svelte";
	import ReLu from "./ReLu.svelte";
	import LeakyReLu from "./LeakyReLu.svelte";
	import DerivativeReLu from "./DerivativeReLu.svelte";
	import TanH from "./TanH.svelte";

  const {
    elements: { root, list, content, trigger },
    states: { value },
  } = createTabs({
    defaultValue: 'tab-1',
  });
 
  let className = '';
  export { className as class };
 
  const triggers = [
    { id: 'tab-1', title: 'Sigmoid' },
    { id: 'tab-2', title: 'ReLU' },
    { id: 'tab-3', title: 'LeakyReLU' },
    { id: 'tab-4', title: 'DerivativeReLU' },
    { id: 'tab-5', title: 'TanH' },
  ];
 
  const [send, receive] = crossfade({
    duration: 250,
    easing: cubicInOut,
  });

</script>

<div class="container mx-auto">
  <h1 class="mt-16 mb-16 text-6xl font-semibold text-center">
    AI activation functions
  </h1>

  <div
    use:melt={$root}
    class={'flex w-full flex-col overflow-hidden rounded-md shadow-md shadow-ctp-crust data-[orientation=vertical]:flex-row' + className}
  >
    <div
      use:melt={$list}
      class="flex shrink-0 overflow-x-auto bg-ctp-crust
      data-[orientation=vertical]:flex-col data-[orientation=vertical]:border-r"
      aria-label="Tabs"
    >
      {#each triggers as triggerItem}
        <button
          use:melt={$trigger(triggerItem.id)}
          class="relative flex items-center justify-center cursor-default select-none
                rounded-none font-medium leading-1
                flex-1 h-12 px-2 focus:relative { $value === triggerItem.id ? 'bg-ctp-mantle' : 'bg-ctp-crust'}"
        >
          {triggerItem.title}
          {#if $value === triggerItem.id}
            <div
              in:send={{ key: 'trigger' }}
              out:receive={{ key: 'trigger' }}
              class="absolute bottom-1 left-1/2 h-1 w-6 -translate-x-1/2 rounded-full bg-ctp-mauve"
            />
          {/if}
        </button>
      {/each}
    </div>
    <div use:melt={$content('tab-1')} class="grow bg-ctp-mantle p-5 xl:px-32 2xl:px-64">
      <Sigmoid />
    </div>
    <div use:melt={$content('tab-2')} class="grow bg-ctp-mantle p-5 xl:px-32 2xl:px-64">
      <ReLu />
    </div>
    <div use:melt={$content('tab-3')} class="grow bg-ctp-mantle p-5 xl:px-32 2xl:px-64">
      <LeakyReLu />
    </div>
    <div use:melt={$content('tab-4')} class="grow bg-ctp-mantle p-5 xl:px-32 2xl:px-64">
      <DerivativeReLu />
    </div>
    <div use:melt={$content('tab-5')} class="grow bg-ctp-mantle p-5 xl:px-32 2xl:px-64">
      <TanH />
    </div>
  </div>
</div>