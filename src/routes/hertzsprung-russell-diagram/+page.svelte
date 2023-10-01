<script lang="ts">
	import axios from "axios";
  import type { PageData } from "./$types";
	import Diagram from "./Diagram.svelte";
	import type { StarHR } from "./utils";
	import { AlertTriangle, Search } from "lucide-svelte";
  import { page } from '$app/stores';
	import { createDialog, melt } from "@melt-ui/svelte";
	import { fly } from "svelte/transition";
	import { goto } from "$app/navigation";
	import { onMount } from "svelte";

  export let data: PageData;
  let search: string = $page.url.searchParams.get("q") ?? "";

  let searchError: string = "";
  let selectedStar: StarHR | null = null;
  const updateSelectedStar = () => {
    axios.get(`/hertzsprung-russell-diagram/${search}`).then((response) => {
      if (response.data) {
        searchError = "";
        selectedStar = response.data;
      }
    }).catch((error) => {
      console.error(error);
      searchError = "No star found with this HIP number";
      selectedStar = null;
    });
  };
  const handleSearch = (event: Event) => {
    event.preventDefault();

    if (search !== "" && search !== null) {
      updateSelectedStar();
      $page.url.searchParams.set("q", search);
      goto($page.url.toString());
    } else {
      selectedStar = null;
      $page.url.searchParams.delete("q");
      goto($page.url.toString());
    }
  };
  onMount(() => {
    if (search) {
      updateSelectedStar();
    }
  });

  const {
    elements: {
      close,
      content,
      title,
      description,
    },
    states: { open },
  } = createDialog({
    forceVisible: true,
    closeOnOutsideClick: false,
    defaultOpen: true,
  });
</script>

<div class="container mx-auto">
  <h1 class="my-16 text-6xl font-semibold text-center">
    Herzsprung-Russell diagram
  </h1>

  {#if $open}
    <div class="flex justify-center">
      <div
        class="relative max-h-[85vh] w-[90vw]
              max-w-[450px] rounded-md bg-ctp-mantle
              p-6 shadow-md"
        transition:fly={{ duration: 200, y: 10 }}
        use:melt={$content}
      >
        <h2 use:melt={$title} class="m-0 text-lg font-medium text-ctp-peach flex gap-2">
          <AlertTriangle /> Warning
        </h2>
        <p use:melt={$description} class="mb-5 mt-2 leading-normal text-ctp-peach">
          This page is computationally intensive and may cause your browser to freeze.<br>
          Please, make sure your computer is powerful enough before continuing.<br>
          Mobile devices are strongly discouraged.
        </p>
        <div class="mt-6 flex justify-end gap-4">
          <button
            use:melt={$close}
            class="inline-flex h-8 items-center justify-center rounded-md
                    bg-ctp-peach px-4 font-medium leading-none text-ctp-mantle
                    outline-none hover:opacity-75 active:opacity-50 transition-opacity"
          >
            Continue
          </button>
        </div>
      </div>
    </div>
  {/if}

  {#if !$open}
  <div class="my-8 mx-auto w-fit">
    <label for="hip" class="mb-1 text-sm font-medium">
      Search by HIP number
    </label>
    <div>
      <div class="flex gap-2">
        <input
          id="hip"
          type="number"
          min="0"
          class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none
                flex h-8 items-center justify-between rounded-md bg-ctp-text
                px-3 pr-12 text-ctp-base focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
          bind:value={search}
        />
        <button
          class="inline-flex square-8 items-center justify-center rounded-md
                  bg-ctp-mauve font-medium leading-none text-ctp-mantle
                  outline-none hover:opacity-75 active:opacity-50 transition-opacity"
          on:click={handleSearch}
        >
          <Search size="18" />
        </button>
      </div>
      <p class="mb-4 text-left text-sm text-ctp-red">{searchError}</p>
    </div>
  </div>
  
  <Diagram data={data.stars} selected={selectedStar} />
  {/if}
</div>
