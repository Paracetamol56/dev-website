<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import type { Writable } from 'svelte/store';
	import type { MnistData } from './mnistData';
	import type * as tf from '@tensorflow/tfjs';
	import LineChart from './LineChart.svelte';
	import { BrainCircuit } from 'lucide-svelte';
	import { addToast } from '../../routes/+layout.svelte';

	export let data: Writable<MnistData>;
	export let model: Writable<tf.Sequential>;

	let maxEpoch: number = 1;
	let maxEpochError: string = '';

	let batchSize: number = 100;
	let batchSizeError: string = '';

	let training: boolean = false;
	let currentEpoch: number = 0;
	let currentBatch: number = 0;
	let lossLogs: number[] = [];
	let accLogs: number[] = [];

	const trainModel = async () => {
		if (maxEpoch < 1) {
			maxEpochError = 'Max Epoch must be greater than 0';
			return;
		}
		if (maxEpoch > 100) {
			maxEpochError = 'Max Epoch must be less than 100';
			return;
		}
		if (batchSize < 1) {
			batchSizeError = 'Batch Size must be greater than 0';
			return;
		}
		if (batchSize > 10000) {
			batchSizeError = 'Batch Size must be less than 10000';
			return;
		}
		maxEpochError = '';
		batchSizeError = '';

		training = true;
		lossLogs = [];
		accLogs = [];

		const [trainData, trainLabels] = $data.getTrainData();
		let counter = 0;
		try {
			await $model.fit(trainData, trainLabels, {
				batchSize,
				epochs: maxEpoch,
				shuffle: true,
				callbacks: {
					onBatchEnd: async (batch, logs) => {
						counter++;
						currentBatch = batch;
						if (logs) {
							lossLogs = [...lossLogs, logs.loss];
							accLogs = [...accLogs, logs.acc];
						}
					},
					onEpochEnd: async (epoch, logs) => {
						currentEpoch = epoch;
					},
				}
			});
		} catch (e) {
			console.error(e);
			addToast({
				data: {
					title: 'Error',
					description: 'There was an error training your model.',
					color: 'bg-ctp-red'
				}
			});
		}

		training = false;
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
			<Button on:click={trainModel} disabled={$data.isDownloaded === false || $model.layers.length === 0}>
				<span>Train Model</span>
				<BrainCircuit size="18" stroke-width="3" />
			</Button>
			{#if training}
    		<p class="mt-2 text-ctp-peach font-semibold">Training...</p>
				<p class="text-sm">
					<strong>Epoch:</strong> {currentEpoch + 1} / {maxEpoch}<br />
					<strong>Batch:</strong> {currentBatch + 1} / {Math.ceil($data.trainSize / batchSize)}
				</p>
			{/if}
		</div>
		<div class="w-full">
			<h3>Training Progress</h3>
			<div class="w-full flex flex-col gap-4 md:flex-row justify-stretch">
				<div class="w-full">
					<h4>Loss{lossLogs.length > 0 ? ` - ${lossLogs[lossLogs.length - 1].toFixed(4)}` : ''}</h4>
					<LineChart data={lossLogs} />
				</div>
				<div class="w-full">
					<h4>Accuracy{accLogs.length > 0 ? ` - ${accLogs[accLogs.length - 1].toFixed(4)}` : ''}</h4>
					<LineChart data={accLogs} />
				</div>
      </div>
		</div>
	</div>
</div>
