<script lang="ts">
	import Dropzone from 'svelte-file-dropzone';
	import { addToast } from '../../+layout.svelte';
	import { Download, Eraser, Repeat, RotateCcw } from 'lucide-svelte';
	import Button from '$lib/components/Button.svelte';
	import axios from 'axios';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import MeltCheckbox from '$lib/components/MeltCheckbox.svelte';
	import MeltSlider from '$lib/components/MeltSlider.svelte';
	import { type SelectOption } from '@melt-ui/svelte';
	import { writable, type Writable } from 'svelte/store';

	// Inputs
	let file: File | null = null;
	let model: Writable<SelectOption<string>> = writable({ value: 'u2net', label: 'u2net' });
	let alphaMatting: Writable<boolean> = writable(false);
	let foregroundThreshold: Writable<number[]> = writable([240]);
	let backgroundThreshold: Writable<number[]> = writable([15]);
	let erosionSize: Writable<number[]> = writable([40]);
	let onlyMask: Writable<boolean> = writable(false);
	let ppm: Writable<boolean> = writable(true);

	let processing = false;
	let processed = false;
	let result: File | null = null;

	function handleFilesSelect(e: any) {
		const { acceptedFiles, _ } = e.detail as { acceptedFiles: File[]; _: File[] };
		if (acceptedFiles.length === 0) {
			return;
		}
		file = acceptedFiles[0];
	}

	function handleFileRejected(e: any) {
		console.error(e);
		console.log(e.detail.fileRejections[e.detail.fileRejections.length - 1].errors[0].code);
		switch (e.detail.fileRejections[e.detail.fileRejections.length - 1].errors[0].code) {
			case 'file-invalid-type':
				addToast({
					data: {
						title: 'Error',
						description: 'The file type is not supported',
						color: 'bg-ctp-red'
					}
				});
				break;
			case 'file-too-large':
				addToast({
					data: {
						title: 'Error',
						description: 'The file is too large',
						color: 'bg-ctp-red'
					}
				});
				break;
			default:
				addToast({
					data: {
						title: 'Error',
						description: 'An error occurred while processing the file',
						color: 'bg-ctp-red'
					}
				});
		}
	}

	async function handleProcess() {
		if (file === null) {
			return;
		}
		processing = true;
		processed = false;

		const formData = new FormData();
		formData.append('file', file);
		formData.append('model', $model.value);
		formData.append('a', $alphaMatting.toString());
		formData.append('af', $foregroundThreshold[0].toString());
		formData.append('ab', $backgroundThreshold[0].toString());
		formData.append('ae', $erosionSize[0].toString());
		formData.append('om', $onlyMask.toString());
		formData.append('ppm', $ppm.toString());

		await axios
			.post('/api/img-removebg', formData, {
				headers: {
					'Content-Type': 'multipart/form-data'
				},
				responseType: 'blob'
			})
			.then((res) => {
				const blob = new Blob([res.data], { type: res.headers['content-type'] });
				// Update the result file
				result = new File([blob], file!.name, { type: res.headers['content-type'] });
				processing = false;
				processed = true;
			})
			.catch((error) => {
				console.error(error);
				addToast({
					data: {
						title: 'Error',
						description: 'An error occurred while processing the file',
						color: 'bg-ctp-red'
					}
				});
				processing = false;
			});
	}

	function handleReset() {
		file = null;
		processing = false;
		processed = false;
		result = null;
	}

	function handleDownload() {
		if (result === null) {
			return;
		}
		const url = URL.createObjectURL(result);
		const a = document.createElement('a');
		a.href = url;
		a.download = result.name;
		document.body.appendChild(a);
		a.click();
		URL.revokeObjectURL(url);
	}
</script>

<svelte:head>
	<title>Image background remover - Mathéo Galuba</title>
</svelte:head>

<article>
	<section class="container mx-auto mb-32 flex flex-col items-stretch justify-start md:flex-row">
		<hgroup>
			<h1 class="mb-8 text-6xl font-bold text-center">
				<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
					Image background remover
				</span>
			</h1>
		</hgroup>
	</section>

	<section
		class="container mx-auto flex flex-col gap-2 md:gap-4 items-start justify-stretch md:flex-row content"
	>
		<div class="w-full bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md p-4">
			<form class="flex flex-col gap-2 md:gap-4 items-stretch justify-start">
				{#if file === null}
					<div>
						<Dropzone
							on:drop={handleFilesSelect}
							on:droprejected={handleFileRejected}
							name="file"
							accept={['image/png']}
							containerClasses="w-full aspect-video border-2 border-dashed border-ctp-lavender rounded-lg flex flex-col items-center justify-center"
							disableDefaultStyles
							maxSize={20 * 1024 * 1024}
							required
						>
							<p class="content-ignore text-center">
								<strong>Drop an image here</strong><br />- or -<br /><strong>Click to upload</strong
								>
							</p>
						</Dropzone>
						<p class="content-ignore text-xs mt-1">
							Supported format is <a
								href="https://en.wikipedia.org/wiki/Portable_Network_Graphics">PNG</a
							>, The maximum file size is 20MB.
						</p>
					</div>
				{:else}
					<div class="relative w-full h-fit">
						<img
							src={URL.createObjectURL(file)}
							alt="Uploaded file preview"
							class="content-ignore w-full h-auto border-2 border-dashed border-ctp-lavender rounded-lg"
						/>
						<button type="button" class="absolute bottom-2 right-2" on:click={handleReset}>
							<Eraser size="16" />
						</button>
						</div>
				{/if}
				<details class="content-ignore">
					<summary class="font-semibold select-none">Advanced options</summary>
					<div class="flex flex-col gap-4">
						<MeltSelect
							name="Model"
							value={model}
							options={[
								'isnet-anime',
								'isnet-general-use',
								'sam',
								'silueta',
								'u2net_cloth_seg',
								'u2net_custom',
								'u2net_human_seg',
								'u2net',
								'u2netp'
							]}
						/>
						<MeltCheckbox name="alpha-mattin" label="Alpha matting" checked={alphaMatting} />
						<MeltSlider
							name="Foreground threshold"
							value={foregroundThreshold}
							defaultValue={240}
							min={0}
							max={255}
						/>
						<MeltSlider
							name="Background threshold"
							value={backgroundThreshold}
							defaultValue={15}
							min={0}
							max={255}
						/>
						<MeltSlider
							name="Erosion size"
							value={erosionSize}
							defaultValue={40}
							min={0}
							max={255}
						/>
						<MeltCheckbox name="only-mask" label="Only mask" checked={onlyMask} />
						<MeltCheckbox name="ppm" label="Prepocess mask" checked={ppm} defaultChecked={true} />
					</div>
				</details>
				<div class="flex flex-row items-center justify-center gap-4">
					<Button on:click={handleProcess} type="button" disabled={file === null}>
						<span>Remove background</span>
						<Repeat size="18" />
					</Button>
				</div>
			</form>
		</div>
		<div
			class="w-full bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md p-4 flex flex-col gap-2 md:gap-4 items-stretch justify-start"
		>
			<div
				class="w-full aspect-video border-2 border-dashed border-ctp-lavender rounded-lg flex items-center justify-center"
			>
				{#if processing}
					<p>Processing...</p>
				{:else if result !== null}
					<img
						src={URL.createObjectURL(result)}
						alt="Processed file preview"
						class="content-ignore w-full h-auto"
					/>
				{/if}
			</div>
			<div class="flex flex-col items-center justify-start gap-4">
				<Button on:click={handleDownload} type="button" disabled={result === null}>
					<span>Download</span>
					<Download size="18" />
				</Button>
			</div>
		</div>
	</section>
</article>
