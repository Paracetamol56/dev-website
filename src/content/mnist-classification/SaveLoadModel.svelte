<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import * as tf from '@tensorflow/tfjs';
	import type { Writable } from 'svelte/store';
	import { addToast } from '../../routes/+layout.svelte';
	import { Binary, Braces, Download, Upload } from 'lucide-svelte';

	export let model: Writable<tf.Sequential>;

	let jsonFileList: FileList;
	let weightsFileList: FileList;

	const save = (e: Event) => {
		e.preventDefault();
		$model.save(`downloads://my-model-${Date.now()}`);
		addToast({
			data: {
				title: 'Model Saved',
				description: 'Your model has been saved to your downloads folder.',
				color: 'bg-ctp-green'
			}
		});
	};

	const load = async (e: Event) => {
		e.preventDefault();
		const jsonFile: File = jsonFileList[0];
		const weightsFile: File = weightsFileList[0];
		try {
			$model = (await tf.loadLayersModel(
				tf.io.browserFiles([jsonFile, weightsFile])
			)) as tf.Sequential;
			addToast({
				data: {
					title: 'Model Loaded',
					description: 'Your model has been loaded.',
					color: 'bg-ctp-green'
				}
			});
		} catch (err) {
			console.error(err);
			addToast({
				data: {
					title: 'Error',
					description: 'There was an error loading your model.',
					color: 'bg-ctp-red'
				}
			});
		}
	};
</script>

<div>
	<div class="w-full">
		<h4>Save Model</h4>
		<Button on:click={save} disabled={$model.layers.length === 0}>
			<span>Save Model</span>
			<Download size="18" stroke-width="3" />
		</Button>
		<h4>
			<div id="saved" style="display:none;">Model Saved</div>
		</h4>
	</div>
	<div class="w-full">
		<h4>Load Model</h4>
		<div class="flex flex-col items-start gap-2">
			<div class="flex bg-ctp-mantle rounded-md items-center gap-2 shadow-md shadow-ctp-crust">
				<label
					class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1 font-semibold text-ctp-mantle
                      shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60"
					for="json-upload"
				>
					<span class="whitespace-nowrap">Browse JSON</span>
					<Braces size="18" stroke-width="3" />
				</label>
				<input
					type="file"
					accept=".json"
					id="json-upload"
					name="json-upload"
					class="hidden"
					bind:files={jsonFileList}
				/>
				{#if jsonFileList && jsonFileList.length > 0}
					<span class="text-xs me-2 mb-0">{jsonFileList[0].name}</span>
				{:else}
					<span class="text-xs me-2 mb-0">No file selected</span>
				{/if}
			</div>
			<div class="flex bg-ctp-mantle rounded-md items-center gap-2 shadow-md shadow-ctp-crust">
				<label
					class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1 font-semibold text-ctp-mantle
                      shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60"
					for="weights-upload"
				>
					<span class="whitespace-nowrap">Browse Weights</span>
					<Binary size="18" stroke-width="3" />
				</label>
				<input
					type="file"
					accept=".bin"
					id="weights-upload"
					name="weights-upload"
					class="hidden"
					bind:files={weightsFileList}
				/>
				{#if weightsFileList && weightsFileList.length > 0}
					<span class="text-xs me-2 mb-0">{weightsFileList[0].name}</span>
				{:else}
					<span class="text-xs me-2 mb-0">No file selected</span>
				{/if}
			</div>
			<Button on:click={load} disabled={!jsonFileList || !weightsFileList}>
				<span>Load Model</span>
				<Upload size="18" stroke-width="3" />
			</Button>
		</div>
	</div>
</div>
