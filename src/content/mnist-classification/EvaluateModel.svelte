<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import type { Writable } from 'svelte/store';
	import type { MnistData } from './mnistData';
	import type * as tf from '@tensorflow/tfjs';
	import BarChart from './BarChart.svelte';
	import ConfusionMatrix from './ConfusionMatrix.svelte';
	import MnistPreview from './MnistPreview.svelte';

	export let data: Writable<MnistData>;
	export let model: Writable<tf.Sequential>;

	let classAccuracy: Map<string, {correct: number, total: number}> = new Map();
	let confusionMatrix: Map<string, Map<string, number>> = new Map();
	const classNames = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'];

	const evaluateModel = () => {
		// Reset class accuracy and confusion matrix
		classAccuracy = new Map();
		confusionMatrix = new Map();
		for (const className of classNames) {
			classAccuracy.set(className, {correct: 0, total: 0});
			confusionMatrix.set(className, new Map());
			for (const className2 of classNames) {
				confusionMatrix.get(className)!.set(className2, 0);
			}
		}

		const [testData, testLabels] = $data.getTestData();

		const predictions = ($model.predict(testData) as tf.Tensor).argMax(1);
		const labels = testLabels.argMax(1);

		for (let i = 0; i < predictions.shape[0]; i++) {
			const prediction = predictions.dataSync()[i];
			const label = labels.dataSync()[i];

			const predictionClass = classNames[prediction];
			const labelClass = classNames[label];

			classAccuracy.get(predictionClass)!.total++;
			if (prediction === label) {
				classAccuracy.get(predictionClass)!.correct++;
			}

			if (!confusionMatrix.get(predictionClass)!.has(labelClass)) {
				confusionMatrix.get(predictionClass)!.set(labelClass, 0);
			}
			confusionMatrix.get(predictionClass)!.set(labelClass, confusionMatrix.get(predictionClass)!.get(labelClass)! + 1);
		}

		classAccuracy = new Map(classAccuracy);
		confusionMatrix = new Map(confusionMatrix);
	};

	let testData: tf.Tensor4D;
	let testLabels: tf.Tensor2D;
	let predictions: tf.Tensor;

	const showPredicions = () => {
		[testData, testLabels] = $data.getTestData(32);
		
		predictions = ($model.predict(testData) as tf.Tensor).argMax(1);
	}
</script>

<div>
	<h3>Evaluate Model</h3>
  <Button on:click={evaluateModel}>Evaluate Model</Button>

	<div class="w-full flex flex-col md:flex-row gap-4">
		<div class="w-full">
			<h4>Class Accuracy</h4>
			<BarChart data={classAccuracy} />
		</div>
		<div class="w-full">
			<h4>Confusion Matrix</h4>
			<ConfusionMatrix data={confusionMatrix} />
		</div>
	</div>

	<h3>Prediction Example</h3>
	<Button on:click={showPredicions}>Show Prediction</Button>
	<h4 />
	{#if predictions !== undefined}
		<MnistPreview data={testData} label={testLabels} prediction={predictions} />
	{/if}
</div>
