<script lang="ts">
	import axios from 'axios';
	import Diagram from './Diagram.svelte';
	import type { StarHR } from './utils';
	import { AlertTriangle, Search } from 'lucide-svelte';
	import { page } from '$app/stores';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { fly } from 'svelte/transition';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { addToast } from '../../routes/+layout.svelte';
	import { browser } from '$app/environment';

	let data: StarHR[] = [];
	let search: string = $page.url.searchParams.get('q') ?? '';

	let searchError: string = '';
	let selectedStar: StarHR | null = null;

	const getData = () => {
		axios
			.get('/api/hipparcos')
			.then((response) => {
				data = response.data;
			})
			.catch((error) => {
				console.error(error);
				if (browser) {
					addToast({
						data: {
							title: 'Error',
							description: 'An error occurred while fetching the data',
							color: 'bg-ctp-red'
						}
					});
				}
			});
	};

	const updateSelectedStar = () => {
		axios
			.get(`/api/hipparcos/${search}`)
			.then((response) => {
				if (response.data) {
					searchError = '';
					selectedStar = response.data;
				}
			})
			.catch((error) => {
				console.error(error);
				searchError = 'No star found with this HIP number';
				selectedStar = null;
			});
	};

	const handleSearch = (event: Event) => {
		event.preventDefault();

		if (search !== '' && search !== null) {
			updateSelectedStar();
			$page.url.searchParams.set('q', search);
			goto($page.url.toString());
		} else {
			selectedStar = null;
			$page.url.searchParams.delete('q');
			goto($page.url.toString());
		}
	};

	getData();
	onMount(() => {
		if (search) {
			updateSelectedStar();
		}
	});

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
	<div class="my-8 mx-auto w-fit">
		<label for="hip" class="mb-2 text-sm font-semibold"> Search by HIP number </label>
		<div>
			<div class="flex gap-2">
				<input
					id="hip"
					type="number"
					min="0"
					class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none
              flex h-8 items-center justify-between rounded-md bg-ctp-surface0
              px-3 pr-12 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
					bind:value={search}
				/>
				<button
					class="inline-flex square-8 items-center justify-center rounded-md
                bg-ctp-mauve font-medium leading-none text-ctp-mantle
                outline-none hover:opacity-75 active:opacity-50 transition-opacity"
					on:click={handleSearch}
				>
					<Search size="18" />
				</button>
			</div>
			<p class="mb-4 text-left text-sm font-semibold text-ctp-red">{searchError}</p>
		</div>
	</div>

	<Diagram {data} selected={selectedStar} />
{/if}
