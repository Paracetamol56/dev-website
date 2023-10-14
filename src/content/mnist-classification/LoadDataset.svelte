<script lang="ts">
	import Button from "$lib/components/Button.svelte";
	import type * as tf from "@tensorflow/tfjs";
  import MnistPreview from "./MnistPreview.svelte";
	import type { MnistData } from "./mnistData";

  export let data: MnistData;
  let loading: boolean = false;

  let preview: [tf.Tensor4D, tf.Tensor2D];

  const loadData = async (e: Event) => {
    e.preventDefault();
    loading = true;
    
    await data.load(40000, 10000);
    preview = data.getTestData(8);

    loading = false;
  };
</script>


<div>
  <h3>MNIST Dataset</h3>
  <Button on:click={loadData}>
    Load Data
  </Button>
  {#if loading}
    <p class="mt-2 text-ctp-peach font-semibold">Loading...</p>
  {:else}
    {#if preview}
      <h4>Test Data Example</h4>
      <MnistPreview data={preview[0]} label={preview[1]}/>
    {/if}
  {/if}
</div>
