<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import { ArrowRight, Download, Repeat, Trash } from 'lucide-svelte';
	import Dropzone from 'svelte-file-dropzone';
	import { addToast } from '../../+layout.svelte';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import { type Writable, writable } from 'svelte/store';
	import type { SelectOption } from '@melt-ui/svelte';
	import axios from 'axios';
	import Loader from '$lib/components/Loader.svelte';
	import MeltSlider from '$lib/components/MeltSlider.svelte';
	import JSZip from 'jszip';

	type ImageFile = {
		file: File;
		processing: boolean;
		converted: boolean;
		convertedFile?: File;
	};

	let files: ImageFile[] = [];
	const format: Writable<SelectOption<string>> = writable({ value: 'PNG', label: 'PNG' });
	const quality: Writable<number[]> = writable([90]);

	function handleFilesSelect(e: any) {
		const { acceptedFiles, _ } = e.detail as { acceptedFiles: File[]; _: File[] };
		files = [
			...files,
			...acceptedFiles.map((file) => ({ file, processing: false, converted: false }))
		];
	}

	function handleFileRejected(e: any) {
		console.log(e);
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

	function handleDelete(index: number) {
		files = files.filter((_, i) => i !== index);
	}

	async function handleConvert(index: number) {
		files[index].processing = true;
		// Post the form data to the server
		const formData = new FormData();
		formData.append('format', $format.value);
		formData.append('quality', $quality.toString());
		formData.append('image', files[index].file);
		await axios
			.post('/api/img-convert', formData, {
				headers: {
					'Content-Type': 'multipart/form-data'
				},
				responseType: 'blob'
			})
			.then((res) => {
				// Create a blob from the response data
				const blob = new Blob([res.data], { type: res.headers['content-type'] });
				// Update the file object with the converted file
				files[index].convertedFile = new File(
					[blob],
					`${files[index].file.name.split('.')[0]}.${$format.value.toLowerCase()}`,
					{ type: res.headers['content-type'] }
				);
				files[index].processing = false;
				files[index].converted = true;
			})
			.catch((err) => {
				addToast({
					data: {
						title: 'Error',
						description: 'An error occurred while converting the image',
						color: 'bg-ctp-red'
					}
				});
				files[index].processing = false;
			});
	}

	function handleDownload(index: number) {
		if (!files[index].convertedFile) return;
		const url = URL.createObjectURL(files[index].convertedFile);
		const a = document.createElement('a');
		a.href = url;
		a.download = files[index].convertedFile.name;
		document.body.appendChild(a);
		a.click();
		URL.revokeObjectURL(url);
		document.body.removeChild(a);
	}

	async function handleConvertAll() {
		for (let i = 0; i < files.length; i++) {
			if (!files[i].converted) {
				await handleConvert(i);
			}
		}
	}

	async function handleDownloadAll() {
		if (files.length === 1) {
			handleDownload(0);
			return;
		}
		const zip = new JSZip();
		const fileNamesCounter: Record<string, number> = {};
		files.forEach((file) => {
			if (file.convertedFile) {
				let fileName = file.convertedFile.name;
				if (fileNamesCounter[fileName]) {
					fileNamesCounter[fileName]++;
					fileName = `${fileName.split('.')[0]}_${fileNamesCounter[fileName]}.${fileName.split('.')[1]}`;
				} else {
					fileNamesCounter[fileName] = 1;
				}
				zip.file(fileName, file.convertedFile);
			}
		});
		zip.generateAsync({ type: 'blob' }).then((content) => {
			const url = URL.createObjectURL(content);
			const timestamp = new Date().getTime();
			const a = document.createElement('a');
			a.href = url;
			a.download = `converted_images_${timestamp}.zip`;
			document.body.appendChild(a);
			a.click();
			URL.revokeObjectURL(url);
			document.body.removeChild(a);
		});
	}
</script>

<svelte:head>
	<title>Image converter - Mathéo Galuba</title>
</svelte:head>

<article>
	<section class="container mx-auto mb-32">
		<hgroup>
			<h1 class="mb-8 text-6xl font-bold text-center">
				<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
					Image converter
				</span>
			</h1>
		</hgroup>
	</section>

	<form action="">
		<section class="container max-w-5xl mx-auto content">
			<p class="mb-8">
				Supported formats are <a href="https://en.wikipedia.org/wiki/Portable_Network_Graphics"
					>PNG</a
				>,
				<a href="https://en.wikipedia.org/wiki/JPEG">JPEG/JPG</a>,
				<a href="https://en.wikipedia.org/wiki/WebP">WebP</a>,
				<a href="https://en.wikipedia.org/wiki/Graphics_Interchange_Format">GIF</a>,
				<a href="https://en.wikipedia.org/wiki/BMP_file_format">BMP</a>,
				<a href="https://en.wikipedia.org/wiki/Tagged_Image_File_Format">TIFF</a>, and
				<a href="https://en.wikipedia.org/wiki/AVIF">AVIF</a>.<br />
				The maximum file size is 20MB.
			</p>
			<Dropzone
				on:drop={handleFilesSelect}
				on:droprejected={handleFileRejected}
				noClick="{files.length > 0}"
				accept="{['image/png', 'image/jpeg', 'image/webp', 'image/gif', 'image/bmp', 'image/tiff', 'image/avif']}"
				multiple
				containerClasses="w-full min-h-60 mb-8 border-2 border-dashed border-ctp-lavender rounded-lg flex items-center justify-center"
				disableDefaultStyles
				maxSize={20 * 1024 * 1024}
				required
			>
				{#if files.length === 0}
					<p>Drop an image here</p>
				{:else}
					<ul class="content-ignore flex flex-row justify-center flex-wrap gap-2 p-2 lg:gap-4 lg:p-4">
						{#each files as item, i}
							<li
								class="cursor-default aspect-square w-48 relative bg-cover bg-no-repeat bg-center rounded-md shadow-md shadow-ctp-crust svelte-file-dropzone overflow-hidden"
								style="background-image: url('{URL.createObjectURL(item.file)}')"
							>
								<div
									class="absolute top-0 left-0 bg-gradient-to-b from-ctp-crust/0 to-ctp-crust/75 hover:to-ctp-base/75 transition-colors w-full h-full p-2 flex flex-row items-end justify-end gap-2"
								>
									<span
										class="flex-1 uppercase font-semibold text-sm text-ctp-mauve ms-1 overflow-hidden flex items-center gap-1"
									>
										{item.file.type.split('/')[1]}
										{#if item.converted}
											<ArrowRight size="14" />
											{item.convertedFile?.type.split('/')[1]}
										{/if}
									</span>
									<button
										class="flex items-center gap-1 rounded-md bg-ctp-crust p-1
												font-ctp-red text-ctp-mantle
												shadow-md shadow-ctp-crust transition-opacity
												hover:opacity-80 active:opacity-60"
										on:click={() => handleDelete(i)}
										type="button"
									>
										<Trash class="text-ctp-red" size="18" />
									</button>
									{#if !item.converted}
										<button
											class="flex items-center gap-1 rounded-md bg-ctp-crust p-1
												font-ctp-blue text-ctp-mantle
												shadow-md shadow-ctp-crust transition-opacity
												hover:opacity-80 active:opacity-60"
											on:click={() => handleConvert(i)}
											type="button"
										>
											<Repeat class="text-ctp-blue" size="18" />
										</button>
									{:else}
										<button
											class="flex items-center gap-1 rounded-md bg-ctp-crust p-1
												font-ctp-green text-ctp-mantle
												shadow-md shadow-ctp-crust transition-opacity
												hover:opacity-80 active:opacity-60"
											on:click={() => handleDownload(i)}
											type="button"
										>
											<Download class="text-ctp-green" size="18" />
										</button>
									{/if}
								</div>
								{#if item.processing}
									<div
										class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2"
									>
										<Loader />
									</div>
								{/if}
							</li>
						{/each}
					</ul>
				{/if}
			</Dropzone>
			<div class="flex flex-col items-center justify-start gap-4">
				<MeltSelect
					name="Format"
					value={format}
					options={['PNG', 'JPEG', 'WebP', 'GIF', 'BMP', 'TIFF', 'AVIF']}
				/>
				{#if $format.value === 'JPEG' || $format.value === 'WebP' || $format.value === 'AVIF'}
					<MeltSlider
						name="Quality"
						min={0}
						max={100}
						step={5}
						value={quality}
					/>
				{/if}
				{#if files.reduce((acc, curr) => acc || curr.converted, false)}
					<Button on:click={handleDownloadAll} type="button">
						<span>Download all</span>
						<Download size="18" />
					</Button>
				{:else}
					<Button on:click={handleConvertAll} type="button">
						<span>Convert all</span>
						<Repeat size="18" />
					</Button>
				{/if}
			</div>
		</section>
	</form>
</article>
