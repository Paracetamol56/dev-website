<script lang="ts">
	import Dropzone from 'svelte-file-dropzone';
	import { addToast } from '../../+layout.svelte';
	import { Download, Repeat, RotateCcw } from 'lucide-svelte';
	import Button from '$lib/components/Button.svelte';
	import axios from 'axios';

	let file: File | null = null;
	let processing = false;
	let processed = false;
	let result: File | null = null;

	function handleFilesSelect(e: any) {
		const { acceptedFiles, _ } = e.detail as { acceptedFiles: File[]; _: File[] };
		file = acceptedFiles[0];
	}

	function handleFileRejected(e: any) {
		console.error(e);
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
			<form action="" class="flex flex-col gap-2 md:gap-4 items-stretch justify-start">
				{#if file === null}
					<div>
					<Dropzone
						on:drop={handleFilesSelect}
						on:droprejected={handleFileRejected}
						accept={['image/png', 'image/jpeg', 'image/webp', 'image/avif']}
						containerClasses="w-full aspect-video border-2 border-dashed border-ctp-lavender rounded-lg flex flex-col items-center justify-center"
						disableDefaultStyles
						maxSize={20 * 1024 * 1024}
						required
					>
						<p class="content-ignore text-center"><strong>Drop an image here</strong><br>- or -<br><strong>Click to upload</strong></p>
					</Dropzone>
					<p class="content-ignore text-xs mt-1">
						Supported formats are <a
							href="https://en.wikipedia.org/wiki/Portable_Network_Graphics">PNG</a
						>,
						<a href="https://en.wikipedia.org/wiki/JPEG">JPEG/JPG</a>,
						<a href="https://en.wikipedia.org/wiki/WebP">WebP</a>,
						<a href="https://en.wikipedia.org/wiki/AVIF">AVIF</a>.<br />
						The maximum file size is 20MB.
					</p>
				</div>
				{:else}
					<img
						src={URL.createObjectURL(file)}
						alt="Uploaded file preview"
						class="w-full h-auto border-2 border-dashed border-ctp-lavender rounded-lg"
					/>
				{/if}
				<div class="flex flex-row items-center justify-center gap-4">
					<Button on:click={handleProcess} type="button" disabled={file === null}>
						<span>Remove background</span>
						<Repeat size="18" />
					</Button>
					<Button on:click={handleReset} type="button" disabled={file === null}>
						<span>Reset</span>
						<RotateCcw size="18" />
					</Button>
				</div>
			</form>
		</div>
		<div class="w-full bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md p-4 flex flex-col gap-2 md:gap-4 items-stretch justify-start">
			<div
				class="w-full aspect-video border-2 border-dashed border-ctp-lavender rounded-lg flex items-center justify-center"
			>
				{#if processing}
					<p>Processing...</p>
				{:else if result !== null}
					<img
						src={URL.createObjectURL(result)}
						alt="Processed file preview"
						class="w-full h-auto"
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
