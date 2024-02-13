<script lang="ts">
	import { onMount } from 'svelte';
	import { applyMaxPooling, loadImage } from './utils';
	import * as tf from '@tensorflow/tfjs';

	export let imgSrc: string;

	let pooledImage: HTMLCanvasElement;

	let imageTensor: tf.Tensor4D;
	$: imagePixels = imageTensor?.shape[1] * imageTensor?.shape[2];
	let pooledImageTensor: tf.Tensor4D;
	$: pooledImagePixels = pooledImageTensor?.shape[1] * pooledImageTensor?.shape[2];

	let poolSize = 2;
	let stride = 2;

	const pool = async () => {
		if (!imageTensor) return;
		pooledImageTensor = await applyMaxPooling(poolSize, stride, imageTensor);
		tf.browser.toPixels(pooledImageTensor.squeeze([0]) as tf.Tensor3D, pooledImage);
	};

	onMount(async () => {
		imageTensor = await loadImage(imgSrc);
		pool();
	});
</script>

<div class="my-4 flex gap-2 justify-between items-start">
	<div class="flex-1">
		<div class="bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md w-full overflow-hidden">
			<img class="w-full h-auto content-ignore" src={imgSrc} alt="Original" />
		</div>
		<p class="text-center text-ctp-text font-semibold">
			Number of pixels: {imageTensor?.shape[1]}x{imageTensor?.shape[2]} = {imagePixels}
		</p>
	</div>
	<div>
		<span class="block text-ctp-text font-semibold text-sm">Pooling Size</span>
		<input
			type="number"
			min="1"
			max="10"
			step="1"
			bind:value={poolSize}
			on:change={pool}
			class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-32 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
		/>
		<span class="block text-ctp-text font-semibold text-sm">Stride</span>
		<input
			type="number"
			min="1"
			max="10"
			step="1"
			bind:value={stride}
			on:change={pool}
			class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-32 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
		/>
	</div>
	<div class="flex-1">
		<div class="bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md w-full overflow-hidden">
			<canvas
				class="w-full h-auto content-ignore"
				bind:this={pooledImage}
				style="image-rendering: pixelated;"
			/>
		</div>
		<p class="text-center text-ctp-text font-semibold">
			Number of pixels: {pooledImageTensor?.shape[1]}x{pooledImageTensor?.shape[2]} = {pooledImagePixels}
			({Math.round((pooledImagePixels / imagePixels) * 100000) / 1000}% of the original)
		</p>
	</div>
</div>
