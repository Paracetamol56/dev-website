<script lang="ts">
	import { writable, type Writable } from 'svelte/store';
	import EvaluateModel from './EvaluateModel.svelte';
	import InitModel from './InitModel.svelte';
	import LoadDataset from './LoadDataset.svelte';
	import SaveLoadModel from './SaveLoadModel.svelte';
	import TestModel from './TestModel.svelte';
	import TrainModel from './TrainModel.svelte';
	import { MnistData } from './mnistData';
	import * as tf from '@tensorflow/tfjs';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { fly } from 'svelte/transition';
	import { AlertTriangle } from 'lucide-svelte';

	const data: Writable<MnistData> = writable(new MnistData());
	const model: Writable<tf.Sequential> = writable(tf.sequential());

	export { data, model };

	const {
		elements: { close, content, title, description },
		states: { open }
	} = createDialog({
		forceVisible: true,
		closeOnOutsideClick: false,
		defaultOpen: true,
		preventScroll: false
	});
</script>

{#if $open}
	<div class="flex justify-center">
		<div
			class="max-h-[85vh] w-[90vw]
            max-w-[450px] rounded-md bg-ctp-mantle
            p-6 shadow-md"
			transition:fly={{ duration: 200, y: 10 }}
			use:melt={$content}
		>
			<h2 use:melt={$title} class="m-0 text-lg font-medium text-ctp-peach flex items-center gap-2">
				<AlertTriangle size="36" /> Warning
			</h2>
			<p use:melt={$description} class="mb-5 mt-2 leading-normal text-ctp-peach">
				This page is computationally intensive and may cause your browser to freeze.<br />
				Please, make sure your computer is powerful enough before continuing.<br />
				Mobile devices are strongly discouraged.
			</p>
			<div class="mt-6 flex justify-end gap-4">
				<button
					use:melt={$close}
					class="flex items-center gap-1 rounded-md bg-ctp-peach px-3 py-1
          outline-none font-semibold text-ctp-mantle
          shadow-md shadow-ctp-crust transition-opacity
          hover:opacity-80 active:opacity-60"
				>
					Continue
				</button>
			</div>
		</div>
	</div>
{/if}

{#if !$open}
	<LoadDataset {data} />
	<InitModel {model} />
	<TrainModel {data} {model} />
	<EvaluateModel {data} {model} />
	<SaveLoadModel {model} />
	<TestModel {model} />
{/if}
