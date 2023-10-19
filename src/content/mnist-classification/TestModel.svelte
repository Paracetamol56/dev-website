<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import type { Writable } from 'svelte/store';
	import * as tf from '@tensorflow/tfjs';
	import { clearCanvas, initDrawCanvas } from './drawUtils';
	import { Eraser } from 'lucide-svelte';
	import BarChart from './BarChart.svelte';

	export let model: Writable<tf.Sequential>;

	let canvas: HTMLCanvasElement;
	let previewCanvas: HTMLCanvasElement;
	let classPrediction: Map<string, number> = new Map();

	$: {
		if (canvas) {
			initDrawCanvas(canvas);
			clearCanvas(previewCanvas);
		}
	}

	const clear = () => {
		clearCanvas(canvas);
		clearCanvas(previewCanvas);
		classPrediction = new Map();
	}

	const predict = () => {
		if (!canvas) return;
		if (!model) return;

		let img = tf.browser.fromPixels(canvas, 4);
		img = tf.cast(img, 'float32').div(255);
		let resized = tf.image.resizeBilinear(img, [28, 28]);
		tf.browser.toPixels(resized, previewCanvas);
		resized = resized.slice([0, 0, 0], [28, 28, 1]);

		const input = tf.cast(resized.reshape([1, 28, 28, 1]), 'float32');
		const prediction = $model.predict(input) as tf.Tensor;
		const predictionArray = Array.from(prediction.dataSync())
		for (let i = 0; i < predictionArray.length; i++) {
			classPrediction = classPrediction.set(i.toString(), predictionArray[i]);
		}
	}
</script>

<div>
	<h3>Draw New Image</h3>
	<div class="w-full h-fit flex flex-col md:flex-row md:gap-4 items-stretch">
		<div class="flex flex-row gap-4 md:flex-col md:gap-0 justify-star items-end">
			<div class="relative w-fit h-fit">
				<canvas bind:this={canvas} on:mouseup={predict} width="140" height="140" class="border-ctp-crust border-2 square-36" />
				<button type="button" class="absolute bottom-2 right-2" on:click={clear}>
					<Eraser size="16" />
				</button>
			</div>
			<div>
				<h4>Preview</h4>
				<canvas bind:this={previewCanvas} width="28" height="28" class="border-ctp-crust border-2 square-36" />
			</div>
		</div>
		<div class="w-full ">
			<h4>Prediction Class Histogram</h4>
			<BarChart data={classPrediction} />
		</div>
	</div>
</div>
