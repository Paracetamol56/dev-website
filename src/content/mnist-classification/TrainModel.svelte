<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import type { Writable } from 'svelte/store';
	import type { MnistData } from './mnistData';
	import type * as tf from '@tensorflow/tfjs';
	import LineChart from './LineChart.svelte';

	export let data: Writable<MnistData>;
	export let model: Writable<tf.Sequential>;

	let maxEpoch: number = 10;
	let maxEpochError: string = '';

	let batchSize: number = 100;
	let batchSizeError: string = '';

	let loading: boolean = false;
	let lossLogs: number[] = [];
	let accLogs: number[] = [];

	const trainModel = async () => {
		loading = true;
		lossLogs = [];
		accLogs = [];

		const [trainData, trainLabels] = $data.getTrainData();
		let counter = 0;
		const history = await $model.fit(trainData, trainLabels, {
			batchSize,
			epochs: maxEpoch,
			shuffle: true,
			callbacks: {
				onBatchEnd: async (batch, logs) => {
					counter++;
					if (logs) {
						lossLogs = [...lossLogs, logs.loss];
						accLogs = [...accLogs, logs.acc];
					}
				}
			}
		});

		loading = false;
	};
</script>

<div>
	<div class="flex flex-col lg:flex-row gap-4">
		<div class="lg:w-64">
			<h3>Train Model</h3>
			<div class="mb-2">
				<fieldset class="w-full">
					<label for="epoch" class="mb-2 text-sm font-semibold">Max Epoch</label>
					<input
						id="epoch"
						type="number"
						bind:value={maxEpoch}
						min="1"
						class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none
									flex h-8 items-center justify-between rounded-md bg-ctp-surface0
									px-3 pr-12 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
					/>
					<p class="text-left text-sm font-semibold text-ctp-red">
						{maxEpochError}
					</p>
				</fieldset>

				<fieldset class="w-full">
					<label for="batch" class="mb-2 text-sm font-semibold">Batch Size</label>
					<input
						id="batch"
						type="number"
						bind:value={batchSize}
						min="1"
						class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none
									flex h-8 items-center justify-between rounded-md bg-ctp-surface0
									px-3 pr-12 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
					/>
					<p class="text-left text-sm font-semibold text-ctp-red">
						{batchSizeError}
					</p>
				</fieldset>
			</div>
			<Button on:click={trainModel}>Train Model</Button>
			{#if loading}
    		<p class="mt-2 text-ctp-peach font-semibold">Training...</p>
			{/if}
		</div>
		<div class="w-full">
			<h3>Training Progress</h3>
			<div class="w-full flex flex-col gap-4 md:flex-row justify-stretch">
				<LineChart data={lossLogs} maxY={10} title="Loss" />
				<LineChart data={accLogs} maxY={1} title="Accuracy" />
      </div>
		</div>
	</div>
</div>
