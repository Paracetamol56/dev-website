<script lang="ts">
	import * as tf from '@tensorflow/tfjs';
	import Button from '$lib/components/Button.svelte';
	import { writable, type Writable } from 'svelte/store';
	import { BrainCog } from 'lucide-svelte';
	import { addToast } from '../../routes/+layout.svelte';
	import MeltRadioGroup from '$lib/components/MeltRadioGroup.svelte';

	export let model: Writable<tf.Sequential>;

	let modelLayers: tf.layers.Layer[] = [];

	const initModel = () => {
		let newModel = tf.sequential();

		if ($value === 'ANN') {
			newModel.add(tf.layers.flatten({ inputShape: [28, 28, 1] }));
			newModel.add(tf.layers.dense({ units: 42, activation: 'relu' }));
			newModel.add(tf.layers.dense({ units: 10, activation: 'softmax' }));
		} else if ($value === 'CNN') {
			newModel.add(
				tf.layers.conv2d({
					kernelSize: 3,
					filters: 16,
					activation: 'relu',
					inputShape: [28, 28, 1]
				})
			);
			newModel.add(tf.layers.conv2d({ kernelSize: 3, filters: 16, activation: 'relu' }));
			newModel.add(tf.layers.maxPooling2d({ poolSize: 2, strides: 2 }));
			newModel.add(tf.layers.flatten({}));
			newModel.add(tf.layers.dense({ units: 10, activation: 'softmax' }));
		} else {
			addToast({
				data: {
					title: 'Warning',
					description: 'Please select a model architecture.',
					color: 'bg-ctp-peach'
				}
			});
			return;
		}

		const myOptim = 'rmsprop';
		newModel.compile({
			loss: 'categoricalCrossentropy',
			optimizer: myOptim,
			metrics: ['accuracy']
		});

		modelLayers = newModel.layers;
		$model = newModel;
	};

	const modelOptions = [
		{ name: 'ANN', label: 'ANN' },
		{ name: 'CNN', label: 'CNN' }
	];
	const value: Writable<string> = writable(modelOptions[0].name);
</script>

<div>
	<!-- <div>
		<textarea
      class="flex h-44 w-full items-center justify-between rounded-md bg-ctp-surface0 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
      spellcheck="false"
      bind:value={modelArchitecture}
    ></textarea>
	</div> -->
	<div
		class="my-4 p-4 rounded-md bg-ctp-mantle flex flex-col md:flex-row gap-4 justify-between items-center"
	>
		<MeltRadioGroup options={modelOptions} value={value} orientation="horizontal" />
		<Button on:click={initModel}>
			<span>Initialize Model</span>
			<BrainCog size="18" stroke-width="3" />
		</Button>
	</div>
	{#if modelLayers.length > 0}
		<h4>Model Summary</h4>
		<div class="mb-8 w-full p-4 bg-ctp-mantle rounded-md overflow-x-auto">
			<table class="w-full table-auto">
				<thead>
					<tr>
						<th class="px-4 py-1">Layer Name</th>
						<th class="px-4 py-1">Output Shape</th>
						<th class="px-4 py-1"># Of Params</th>
						<th class="px-4 py-1">Trainable</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-ctp-surface0">
					{#each modelLayers as layer}
						<tr class="hover:bg-ctp-crust">
							<td class="px-4 py-1">{layer.name}</td>
							<td class="px-4 py-1">[batch, {layer.outputShape.slice(1).join(', ')}]</td>
							<td class="px-4 py-1">{layer.countParams()}</td>
							<td class="px-4 py-1">{layer.trainable}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
