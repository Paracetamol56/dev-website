<script lang="ts">
	import * as tf from "@tensorflow/tfjs";
	import { onMount } from "svelte";

  export let data: Tensor4D;
  export let label: Tensor2D;
  export let prediction: Tensor2D | null = null;

  const labels = Array.from(label.argMax(1).dataSync());
  const previewCount = data.shape[0];
  const canvas: HTMLCanvasElement[] = new Array(previewCount);

  onMount(() => {
    for (let i = 0; i < previewCount; ++i) {
      const img = data.slice([i, 0], [1, data.shape[1]]);
      tf.browser.toPixels(img.reshape([28, 28]), canvas[i]);
    }
  });
</script>

<div class="grid grid-cols-8 gap-2">
  {#each Array.from({ length: previewCount }) as _, i}
    <div class="p-2 bg-ctp-mantle rounded-md">
      <div class="flex flex-col items-center gap-2">
        <p class="text-ctp-peach">
          Label: {labels[i]}
        </p>
        <div class="square-24">
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
