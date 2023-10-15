<script lang="ts">
	import * as tf from "@tensorflow/tfjs";
	import { onMount } from "svelte";

  export let data: tf.Tensor4D;
  export let label: tf.Tensor2D;
  export let prediction: tf.Tensor | null = null;

  $: labels = Array.from(label.argMax(1).dataSync()); console.log(labels);
  $: predictions = prediction ? Array.from(prediction.dataSync()) : null;
  $: previewCount = data.shape[0];
  $: canvas = new Array<HTMLCanvasElement>(previewCount);

  onMount(() => {
    for (let i = 0; i < previewCount; ++i) {
      const img = data.slice([i, 0], [1, data.shape[1]]);
      tf.browser.toPixels(img.reshape([28, 28]), canvas[i]);
    }
  });
</script>

<div class="grid gap-2 grid-cols-2 md:grid-cols-4 lg:grid-cols-8">
  {#each Array.from({ length: previewCount }) as _, i}
    <div class="bg-black rounded-md overflow-hidden">
      <div class="flex flex-col items-stretch">
        <div class="{predictions !== null ? predictions[i] === labels[i] ? 'bg-ctp-green' : 'bg-ctp-red' : 'bg-ctp-green'} flex flex-col items-center">
          {#if predictions === null}
            <p class="my-2 mx-3 text-center text-ctp-mantle font-semibold">
              Label: {labels[i]}
            </p>
          {:else}
            <p class="my-2 mx-3 text-center text-ctp-mantle font-semibold">
              Label: {labels[i]}/Pred: {predictions[i]}
            </p>
          {/if}
        </div>
        <div class="mx-auto square-24">
          <canvas
            bind:this={canvas[i]}
            class="w-full h-full"
            width={28}
            height={28}
          />
        </div>
      </div>
    </div>
  {/each}
</div>
