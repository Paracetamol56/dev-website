<script lang="ts">
	import { createCombobox, melt } from "@melt-ui/svelte";
	import { Check, ChevronDown, ChevronUp } from "lucide-svelte";
	import { fly } from "svelte/transition";
	import Sigmoid from "./Sigmoid.svelte";
	import LeakyReLu from "./LeakyReLu.svelte";
	import DerivativeReLu from "./DerivativeReLu.svelte";
	import TanH from "./TanH.svelte";
	import ReLu from "./ReLu.svelte";

  const {
    elements: { menu, input, option },
    states: { open },
    helpers: { isSelected },
  } = createCombobox({
    forceVisible: true,
  });

  $: functions = [
    "Sigmoid",
    "ReLU",
    "LeakyReLU",
    "DerivativeReLU",
    "TanH",
  ];

</script>

<div class="container mx-auto">
  <h1 class="mt-16 mb-16 text-6xl font-semibold text-center">
    AI activation functions
  </h1>

  <div class="relative">
    <input
      use:melt={$input}
      class="flex h-10 items-center justify-between rounded-md bg-ctp-mantle w-full
          px-3 pr-12 focus:outline-none focus:ring-2 focus:ring-ctp-mauve focus:border-transparent"
      placeholder="Activation function"
    />
    <div class="absolute right-2 top-1/2 z-10 -translate-y-1/2 text-ctp-text">
      {#if $open}
        <ChevronUp class="square-4" />
      {:else}
        <ChevronDown class="square-4" />
      {/if}
    </div>
  </div>

  {#if $isSelected("Sigmoid")}
    <Sigmoid />
  {/if}
  {#if $isSelected("ReLU")}
    <ReLu />
  {/if}
  {#if $isSelected("LeakyReLU")}
    <LeakyReLu />
  {/if}
  {#if $isSelected("DerivativeReLU")}
    <DerivativeReLu />
  {/if}
  {#if $isSelected("TanH")}
    <TanH />
  {/if}
</div>

{#if $open}
  <ul
    class="z-10 flex max-h-[300px] flex-col overflow-hidden rounded-md"
    use:melt={$menu}
    transition:fly={{ duration: 150, y: -5 }}
  >
    <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
    <div
      class="flex max-h-full flex-col gap-0 overflow-y-auto bg-ctp-mantle px-2 py-2 text-ctp-text"
      tabindex="0"
    >
      {#each functions as fn, index (index)}
        <li
          use:melt={$option({
            value: fn,
            label: fn,
          })}
          class="relative cursor-pointer scroll-my-2 rounded-md py-2 pl-4 pr-4
                data-[highlighted]:bg-ctp-mauve/25 data-[highlighted]:text-ctp-mauve
                data-[disabled]:opacity-50"
        >
          {#if $isSelected(fn)}
            <div class="check absolute left-2 top-1/2 z-10 text-ctp-mauve transform -translate-y-1/2">
              <Check class="square-4" />
            </div>
          {/if}
          <div class="pl-4">
            <span class="font-medium">{fn}</span>
          </div>
        </li>
      {:else}
        <li
          class="relative cursor-pointer rounded-md py-1 pl-8 pr-4
                data-[highlighted]:bg-ctp-mauve data-[highlighted]:text-ctp-mauve"
        >
          No results found
        </li>
      {/each}
    </div>
  </ul>
{/if}
