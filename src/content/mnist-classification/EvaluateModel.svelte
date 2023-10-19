<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import type { Writable } from 'svelte/store';
	import type { MnistData } from './mnistData';
	import * as tf from '@tensorflow/tfjs';
	import BarChart from './BarChart.svelte';
	import ConfusionMatrix from './ConfusionMatrix.svelte';
	import MnistPreview from './MnistPreview.svelte';
	import { FlaskConical, View } from 'lucide-svelte';
	import { addToast } from '../../routes/+layout.svelte';

	export let data: Writable<MnistData>;
	export let model: Writable<tf.Sequential>;

	let classAccuracyMap: Map<string, number> = new Map();
	let confusionMap: Map<string, Map<string, number>> = new Map();
	const classNames = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'];

	// Prediction example
	let exampleTestData: tf.Tensor4D;
	let exampleTestLabels: tf.Tensor2D;
	let examplePredictions: tf.Tensor;

	const evaluateModel = () => {
		let testData;
		let testLabels;
		try {
			[testData, testLabels] = $data.getTestData();
		} catch (err) {
			console.error(err);
			addToast({
				data: {
					title: 'Error',
					description: 'There was an error loading your data.',
					color: 'bg-ctp-red'
				}
			});
			return;
		}

		const predictions = ($model.predict(testData) as tf.Tensor).argMax(1);
		const labels = testLabels.argMax(1);
		
		// Compute confusion matrix
		const confusionMatrix = tf.math.confusionMatrix(labels as tf.Tensor1D, predictions as tf.Tensor1D, 10);
		classAccuracyMap = new Map();
		confusionMap = new Map();
		for (let i = 0; i < 10; i++) {
			const row = confusionMatrix.slice([i, 0], [1, 10]).dataSync();
			const rowMap = new Map();
			for (let j = 0; j < 10; j++) {
				rowMap.set(classNames[j], row[j]);
			}
			classAccuracyMap.set(classNames[i], row[i] / Array.from(row).reduce((a, b) => a + b));
			confusionMap.set(classNames[i], rowMap);
		}
		classAccuracyMap = new Map([...classAccuracyMap.entries()].sort((a, b) => a[0].localeCompare(b[0])));	
		confusionMap = new Map([...confusionMap.entries()].sort((a, b) => a[0].localeCompare(b[0])));	

		// Load prediction example
		[exampleTestData, exampleTestLabels] = $data.getTestData(32);
		examplePredictions = ($model.predict(exampleTestData) as tf.Tensor).argMax(1);
	};
</script>

<div>
	<h3>Evaluate Model</h3>
  <Button on:click={evaluateModel} disabled={$data.isDownloaded === false || $model.layers.length === 0}>
		<span>Evaluate Model</span>
		<FlaskConical size="18" stroke-width="3" />
	</Button>
	
	<div class="w-full flex flex-col md:flex-row gap-4">
		<div class="w-full h-72">
			<h4>Class Accuracy</h4>
			<BarChart data={classAccuracyMap} />
		</div>
		<div class="w-full">
			<h4>Confusion Matrix</h4>
			<ConfusionMatrix data={confusionMap} />
		</div>
	</div>
	
	<h4>Prediction Example</h4>
	{#if examplePredictions !== undefined}
		<MnistPreview data={exampleTestData} label={exampleTestLabels} prediction={examplePredictions} />
	{/if}
</div>
