<script lang="ts">
	import FileUploader from '$lib/components/FileUploader.svelte';
	import Button from '$lib/components/Button.svelte';
	import MeltCheckbox from '$lib/components/MeltCheckbox.svelte';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import NumberInput from '$lib/components/NumberInput.svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import { writable, type Writable } from 'svelte/store';
	import type { SelectOption } from '@melt-ui/svelte';
	import { Lock, LockOpen, Download, Eye } from 'lucide-svelte';
	import { generatePDF, type PrintingMarksOptions } from './generatePDF';

	let cropMarks: Writable<boolean> = writable(true);
	let registrationMarks: Writable<boolean> = writable(true);
	let colorBars: Writable<boolean> = writable(true);
	let pageInfo: Writable<boolean> = writable(true);
	let pageNumbering: Writable<boolean> = writable(true);
	let pageInfoText: Writable<string> = writable('');
	let unit: Writable<SelectOption<'mm' | 'in'>> = writable({ value: 'mm', label: 'mm' });
	let bleedTop: Writable<number> = writable(3);
	let bleedRight: Writable<number> = writable(3);
	let bleedBottom: Writable<number> = writable(3);
	let bleedLeft: Writable<number> = writable(3);
	let lockBleed: Writable<boolean> = writable(true);

	let uploadedFile: File | null = null;

	$: if (uploadedFile && !$pageInfoText) {
		pageInfoText.set(uploadedFile.name);
	}

	let previewUrl: string | null = null;
	let generatedPdfBlob: Blob | null = null;

	$: if (uploadedFile) {
		previewUrl = URL.createObjectURL(uploadedFile);
	}

	// Bleed lock functionality
	$: if ($lockBleed) {
		bleedRight.set($bleedTop);
		bleedBottom.set($bleedTop);
		bleedLeft.set($bleedTop);
	}

	// Unit conversion functionality
	let previousUnit: 'mm' | 'in' | null = null;
	$: if (unit && $unit.value !== previousUnit) {
		if (previousUnit) {
			// Convert values from previous unit to new unit
			if (previousUnit === 'mm' && $unit.value === 'in') {
				// Convert mm to inches (1 inch = 25.4 mm)
				const factor = 1 / 25.4;
				bleedTop.update((v) => Math.round(v * factor * 1000) / 1000);
				bleedRight.update((v) => Math.round(v * factor * 1000) / 1000);
				bleedBottom.update((v) => Math.round(v * factor * 1000) / 1000);
				bleedLeft.update((v) => Math.round(v * factor * 1000) / 1000);
			} else if (previousUnit === 'in' && $unit.value === 'mm') {
				// Convert inches to mm (1 inch = 25.4 mm)
				const factor = 25.4;
				bleedTop.update((v) => Math.round(v * factor * 1000) / 1000);
				bleedRight.update((v) => Math.round(v * factor * 1000) / 1000);
				bleedBottom.update((v) => Math.round(v * factor * 1000) / 1000);
				bleedLeft.update((v) => Math.round(v * factor * 1000) / 1000);
			}
		}
		previousUnit = $unit.value;
	}

	function getOptions(): PrintingMarksOptions {
		return {
			unit: $unit.value,
			cropMarks: $cropMarks,
			bleeds: {
				top: $bleedTop,
				right: $bleedRight,
				bottom: $bleedBottom,
				left: $bleedLeft
			},
			registrationMarks: $registrationMarks,
			colorBars: $colorBars,
			pageInfo: $pageInfo
				? {
						pageNumbering: $pageNumbering,
						pageInfoText: $pageInfoText
				  }
				: null
		};
	}

	// Generate preview PDF
	async function generatePreview() {
		if (!uploadedFile) {
			console.error('No file uploaded for preview.');
			return;
		}
		try {
			const blob = await generatePDF(uploadedFile, getOptions());
			if (blob) {
				generatedPdfBlob = blob;
				// Revoke previous object URLs to free memory
				if (previewUrl) {
					URL.revokeObjectURL(previewUrl);
				}
			}
		} catch (error) {
			console.error('Error generating preview:', error);
		}
	}

	async function downloadPDF() {
		if (!uploadedFile) {
			console.error('No file uploaded for preview.');
			return;
		}
		try {
			const blob = await generatePDF(uploadedFile, getOptions());
			if (blob) {
				const url = URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `print-ready_${uploadedFile.name}`;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				URL.revokeObjectURL(url);
			}
		} catch (error) {
			console.error('Error downloading PDF:', error);
		}
	}
</script>

<svelte:head>
	<title>PDF Printing Marks - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-4xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				PDF Printing Marks
			</span>
		</h1>
		<p class="text-center text-ctp-subtext0 mb-8">
			Add crop marks, bleed marks, registration marks, color bars, and page information to your PDF
		</p>
	</hgroup>

	<div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
		<!-- Upload and Preview Section -->
		<div class="flex flex-col gap-6">
			<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust">
				<h2 class="text-2xl font-bold mb-4 text-ctp-text">Upload PDF</h2>

				<FileUploader onFileSelect={(file) => (uploadedFile = file)} accept=".pdf" />
			</div>

			<!-- Preview Section -->
			<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust">
				<h2 class="text-2xl font-bold mb-4 text-ctp-text">Preview</h2>
				{#if generatedPdfBlob}
					<div class="rounded-md overflow-hidden bg-ctp-crust">
						<iframe
							src={URL.createObjectURL(generatedPdfBlob) + '#toolbar=0&navpanes=0&scrollbar=0'}
							class="w-full h-96"
							type="application/pdf"
						/>
					</div>
				{:else if previewUrl}
					<div class="rounded-md overflow-hidden bg-ctp-crust">
						<!-- Original PDF preview -->
						<div class="bg-ctp-crust aspect-video flex flex-col items-center justify-center">
							<p class="text-ctp-text mb-4">Original PDF Preview</p>
						</div>
					</div>
				{:else}
					<div
						class="rounded-md overflow-hidden bg-ctp-crust aspect-video flex items-center justify-center"
					>
						<p class="text-ctp-subtext0">Upload a PDF to see preview</p>
					</div>
				{/if}
				<div class="mt-4 flex justify-end gap-2">
					<Button type="button" on:click={generatePreview} disabled={!uploadedFile}>
						<Eye size="16" />
						Preview PDF
					</Button>
					<Button type="button" on:click={downloadPDF} disabled={!uploadedFile}>
						<Download size="16" />
						Download PDF
					</Button>
				</div>
			</div>
		</div>

		<!-- Settings Panel -->
		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust">
			<h2 class="text-2xl font-bold mb-4 text-ctp-text">Settings</h2>

			<div class="space-y-6">
				<div class="space-y-2">
					<MeltCheckbox bind:checked={cropMarks} label="Crop Marks" name="cropMarks" />

					{#if $cropMarks}
						<div class="ml-6">
							<h3 class="text-sm font-semibold text-ctp-text mb-2">Bleed settings</h3>
							<!-- Bleed Settings -->
							<div class="mt-2">
								<div class="mb-2 w-fit">
									<MeltSelect name="Unit" options={['mm', 'in']} bind:value={unit} />
								</div>

								<!-- Bleed grid with lock button -->
								<div class="grid grid-cols-3 gap-2 w-fit">
									<div />
									<div>
										<NumberInput label="Top" bind:value={bleedTop} min={0} step={0.1} />
									</div>
									<div />

									<div>
										<NumberInput label="Left" bind:value={bleedLeft} min={0} step={0.1} />
									</div>

									<div class="flex items-center justify-center">
										<button
											class="square-20 bg-ctp-crust rounded-md p-2 shadow-md shadow-ctp-crust hover:opacity-90 flex items-center justify-center transition-opacity"
											on:click={() => lockBleed.update((v) => !v)}
										>
											{#if $lockBleed}
												<Lock class="text-ctp-text" size="16" />
											{:else}
												<LockOpen class="text-ctp-text" size="16" />
											{/if}
										</button>
									</div>

									<div>
										<NumberInput label="Right" bind:value={bleedRight} min={0} step={0.1} />
									</div>

									<div />
									<div>
										<NumberInput label="Bottom" bind:value={bleedBottom} min={0} step={0.1} />
									</div>
									<div />
								</div>
							</div>
						</div>
					{/if}

					<MeltCheckbox
						bind:checked={registrationMarks}
						label="Registration Marks"
						name="registrationMarks"
					/>

					<MeltCheckbox bind:checked={colorBars} label="Color Bars" name="colorBars" />

					<MeltCheckbox bind:checked={pageInfo} label="Page Information" name="pageInfo" />

					{#if $pageInfo}
						<div class="ml-6">
							<div class="mt-2 space-y-2">
								<MeltCheckbox
									bind:checked={pageNumbering}
									label="Page Numbering"
									name="pageNumbering"
								/>
								<TextInput
									label="Custom Text"
									bind:value={pageInfoText}
									placeholder="Enter page information text"
								/>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
</section>
