<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import * as tf from '@tensorflow/tfjs';
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import { applyConvolution, loadImage } from './utils';
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';

	export let imgSrc: string;
	export const prerender = false;

	let convolvedImage: HTMLCanvasElement;

	let imageTensor: tf.Tensor4D;
  let convolvedImageTensor: tf.Tensor4D;

	let matrixMultiplier: number = 1;
	let matrix: number[][] = [
		[0, 0, 0],
		[0, 1, 0],
		[0, 0, 0]
	];
  
  const convolve = async (): Promise<void> => {
    if (!imageTensor) return;
    convolvedImageTensor = await applyConvolution(
      matrix.map((row) => row.map((value) => value * matrixMultiplier)),
      imageTensor
    );
    tf.browser.toPixels(convolvedImageTensor.squeeze([0]) as tf.Tensor3D, convolvedImage);
  };

	onMount(async () => {
		imageTensor = await loadImage(imgSrc);
    convolve();
	});
</script>

<div
	class="bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md p-4 flex flex-col lg:flex-row gap-2 md:gap-4 justify-center items-start my-8"
>
	<div>
		<span class="block text-ctp-text font-semibold text-sm">Matrix Multiplier</span>
		<input
			type="number"
			min="-100"
			max="100"
			step="0.1"
			bind:value={matrixMultiplier}
      on:change={convolve}
			class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-32 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
		/>
	</div>
	<div class="shrink-0">
		<span class="block text-ctp-text font-semibold text-sm">Kernel</span>
		<div class="grid grid-cols-3 gap-2">
			<input
				type="number"
				min="-100"
				max="100"
				step="1"
				bind:value={matrix[0][0]}
        on:change={convolve}
				class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-16 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
			<input
				type="number"
				min="-100"
				max="100"
				step="1"
				bind:value={matrix[0][1]}
        on:change={convolve}
				class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-16 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
			<input
				type="number"
				min="-100"
				max="100"
				step="1"
				bind:value={matrix[0][2]}
        on:change={convolve}
				class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-16 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
			<input
				type="number"
				min="-100"
				max="100"
				step="1"
				bind:value={matrix[1][0]}
        on:change={convolve}
				class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-16 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
			<input
				type="number"
				min="-100"
				max="100"
				step="1"
				bind:value={matrix[1][1]}
        on:change={convolve}
				class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-16 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
			<input
				type="number"
				min="-100"
				max="100"
				step="1"
				bind:value={matrix[1][2]}
        on:change={convolve}
				class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-16 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
			<input
				type="number"
				min="-100"
				max="100"
				step="1"
				bind:value={matrix[2][0]}
        on:change={convolve}
				class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-16 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
			<input
				type="number"
				min="-100"
				max="100"
				step="1"
				bind:value={matrix[2][1]}
        on:change={convolve}
				class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-16 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
			<input
				type="number"
				min="-100"
				max="100"
				step="1"
				bind:value={matrix[2][2]}
        on:change={convolve}
				class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none flex h-8 w-16 items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
		</div>
	</div>
	<div class="shrink">
		<span class="block text-ctp-text font-semibold text-sm">Predefined Kernels</span>
		<div class="flex gap-2 flex-wrap">
			<MeltTooltip math={`\\begin{bmatrix}0 & 0 & 0\\\\ 0 & 1 & 0\\\\ 0 & 0 & 0\\end{bmatrix}`}>
				<Button
					on:click={() => {
						matrixMultiplier = 1;
						matrix = [
							[0, 0, 0],
							[0, 1, 0],
							[0, 0, 0]
						];
            convolve();
					}}>Identity</Button
				>
			</MeltTooltip>
			<MeltTooltip math={`\\begin{bmatrix}0 & -1 & 0\\\\ -1 & 4 & -1\\\\ 0 & -1 & 0\\end{bmatrix}`}>
				<Button
					on:click={() => {
						matrixMultiplier = 1;
						matrix = [
							[0, -1, 0],
							[-1, 4, -1],
							[0, -1, 0]
						];
            convolve();
					}}>Edge Detection 1</Button
				>
			</MeltTooltip>
			<MeltTooltip
				math={`\\begin{bmatrix}-1 & -1 & -1\\\\ -1 & 8 & -1\\\\ -1 & -1 & -1\\end{bmatrix}`}
			>
				<Button
					on:click={() => {
						matrixMultiplier = 1;
						matrix = [
							[-1, -1, -1],
							[-1, 8, -1],
							[-1, -1, -1]
						];
            convolve();
					}}>Edge Detection 2</Button
				>
			</MeltTooltip>
			<MeltTooltip math={`\\begin{bmatrix}0 & -1 & 0\\\\ -1 & 5 & -1\\\\ 0 & -1 & 0\\end{bmatrix}`}>
				<Button
					on:click={() => {
						matrixMultiplier = 1;
						matrix = [
							[0, -1, 0],
							[-1, 5, -1],
							[0, -1, 0]
						];
            convolve();
					}}>Sharpen</Button
				>
			</MeltTooltip>
			<MeltTooltip
				math={`\\frac{1}{9}\\begin{bmatrix}1 & 1 & 1\\\\ 1 & 1 & 1\\\\ 1 & 1 & 1\\end{bmatrix}`}
			>
				<Button
					on:click={() => {
						matrixMultiplier = 1 / 9;
						matrix = [
							[1, 1, 1],
							[1, 1, 1],
							[1, 1, 1]
						];
            convolve();
					}}>Box Blur</Button
				>
			</MeltTooltip>
			<MeltTooltip
				math={`\\frac{1}{16}\\begin{bmatrix}1 & 2 & 1\\\\ 2 & 4 & 2\\\\ 1 & 2 & 1\\end{bmatrix}`}
			>
				<Button
					on:click={() => {
						matrixMultiplier = 1 / 16;
						matrix = [
							[1, 2, 1],
							[2, 4, 2],
							[1, 2, 1]
						];
            convolve();
					}}>Gaussian Blur</Button
				>
			</MeltTooltip>
		</div>
	</div>
</div>

<div class="my-4 flex gap-2 justify-between items-start">
	<div class="bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md w-full overflow-hidden">
		<canvas class="w-full h-auto content-ignore" bind:this={convolvedImage} />
	</div>
</div>
