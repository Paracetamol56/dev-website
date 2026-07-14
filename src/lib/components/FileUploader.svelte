<script lang="ts">
	import { FilePen } from 'lucide-svelte';
	import { addToast } from '../../routes/+layout.svelte';

	export let onFileSelect: (file: File | null) => void;
	export let accept: string = '.pdf';
	export let label: string | null = null;

	let overDropzone = false;
	let uploadedFile: File | null = null;

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		overDropzone = false;

		if (event.dataTransfer?.files && event.dataTransfer.files.length > 0) {
			const file = event.dataTransfer.files[0];
			// Check if file type is accepted
			if (accept && !file.name.endsWith(accept.replace('*', ''))) {
				addToast({
					data: {
						title: 'Invalid File Type',
						description: `Please select a ${accept} file.`,
						color: 'bg-ctp-red'
					}
				});
				return;
			}

			uploadedFile = file;
			onFileSelect(uploadedFile);
		}
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		overDropzone = true;
	}

	function handleDragLeave() {
		overDropzone = false;
	}

	function handleFileInput(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files && input.files.length > 0) {
			const file = input.files[0];
			uploadedFile = file;
			onFileSelect(uploadedFile);
		}
	}

	function clearFile() {
		uploadedFile = null;
		const fileInput = document.getElementById('file-input') as HTMLInputElement;
		if (fileInput) {
			fileInput.value = '';
		}
		onFileSelect(null);
	}

	function triggerFileInput() {
		const fileInput = document.getElementById('file-input') as HTMLInputElement;
		fileInput?.click();
	}
</script>

<div class="w-full">
	{#if label}
		<h4>{label}</h4>
	{/if}
	<div
		class="relative border-2 border-dashed rounded-md p-6 text-center cursor-pointer transition-colors bg-ctp-mantle
		       {overDropzone ? 'border-ctp-mauve bg-ctp-mantle/50' : 'border-ctp-surface0'}"
		on:drop={handleDrop}
		on:dragover={handleDragOver}
		on:dragleave={handleDragLeave}
		on:click={triggerFileInput}
	>
		<input id="file-input" type="file" {accept} class="hidden" on:change={handleFileInput} />
		{#if uploadedFile}
			<div class="flex justify-center items-center">
				<div>
					<p class="text-ctp-text">{uploadedFile.name}</p>
					<p class="text-xs text-ctp-subtext0 mt-2">
						{(uploadedFile.size / 1024 / 1024).toFixed(2)} MB
					</p>
				</div>
				<button
					on:click|preventDefault={clearFile}
					class="absolute text-ctp-text hover:text-ctp-mauve transition-colors top-2 right-2"
				>
					<FilePen size="20" />
				</button>
			</div>
		{:else}
			<p class="text-ctp-subtext0">Drag & drop your file here, or click to select</p>
		{/if}
	</div>
</div>
