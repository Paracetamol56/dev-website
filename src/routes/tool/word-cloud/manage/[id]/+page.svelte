<script lang="ts">
	import WordCloud from '../../WordCloud.svelte';
	import BarChart from './BarChart.svelte';
	import Table from './Table.svelte';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { QrCode, X } from 'lucide-svelte';
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import { fade, fly } from 'svelte/transition';
	import QRCode from 'qrcode';
	import type { PageData } from './$types';
	import { addToast } from '../../../../+layout.svelte';
	import { onDestroy, onMount } from 'svelte';
	import { user } from '$lib/store';
	import { followSession } from '../../socket';
	import type { WordCloudWord } from '../../utils';
	import api from '$lib/api';

	export let data: PageData;

	function addWord(word: WordCloudWord) {
		data.session.words = [...data.session.words, word];
		const found = data.distribution.find((entry) => entry.text === word.text);
		if (found) found.occurence++;
		else data.distribution.push({ text: word.text, occurence: 1 });
		data.distribution = data.distribution.sort((a, b) => b.occurence - a.occurence);
	}

	let stopFollowing = () => {};
	onMount(() => {
		if (!data.session.open) return;
		stopFollowing = followSession(
			data.session.id,
			(event) => {
				if (event.type === 'word') addWord(event.word);
				else if (event.type === 'session' && !event.open) {
					data.session.open = false;
					data.session.closedAt ??= new Date();
				}
			},
			$user.accessToken
		);
	});
	onDestroy(() => stopFollowing());

	let qrWidthAvailable: number;
	let canvas: HTMLCanvasElement;

	$: {
		if (canvas !== undefined) {
			const url = `${window.location.origin}/tool/word-cloud?code=${data.session.code}`;
			QRCode.toCanvas(
				canvas,
				url,
				{
					errorCorrectionLevel: 'H',
					scale: 3 + qrWidthAvailable / 175
				},
				(error) => {
					if (error) console.error(error);
				}
			);
		}
	}

	const humanReadableDate = (date: Date | null) => {
		if (date === null) return '';
		return date.toLocaleString();
	};

	const closeSession = (e: Event) => {
		e.preventDefault();

		api
			.callWithAuth('DELETE', `/word-cloud/${data.session.id}`)
			.then(async (res) => {
				if (res.status === 204) {
					addToast({
						data: {
							title: 'Success',
							description: 'The session has been closed',
							color: 'bg-ctp-green'
						}
					});
					data.session.open = false;
					data.session.closedAt = new Date();
				}
			})
			.catch((err) => {
				console.error(err);
				addToast({
					data: {
						title: 'Error',
						description: 'An error occured while closing the session',
						color: 'bg-ctp-red'
					}
				});
			});
	};

	const {
		elements: { trigger, overlay, content, title, close, portalled },
		states: { open }
	} = createDialog({
		forceVisible: true
	});
</script>

<svelte:head>
	<title>{data.session.name} - Word cloud - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-4xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				{data.session.name}
			</span>
		</h1>
		{#if data.session.description}
			<p class="text-center text-ctp-subtext0 mb-8">{data.session.description}</p>
		{/if}
	</hgroup>

	<div class="flex flex-col gap-8">
		<div
			class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-wrap items-center gap-x-8 gap-y-4"
		>
			<div class="flex items-center gap-2">
				{#if data.session.open}
					<span class="relative square-2 rounded-full bg-ctp-green">
						<span class="absolute inset-0 animate-ping rounded-full bg-ctp-green" />
					</span>
					<span class="font-semibold">Open</span>
				{:else}
					<span class="square-2 rounded-full bg-ctp-red" />
					<span class="font-semibold">Closed</span>
				{/if}
			</div>
			<dl class="flex flex-wrap gap-x-8 gap-y-2 text-sm">
				<div>
					<dt class="text-ctp-subtext0">Code</dt>
					<dd class="font-mono text-lg font-bold text-ctp-mauve">{data.session.code}</dd>
				</div>
				<div>
					<dt class="text-ctp-subtext0">Submissions</dt>
					<dd class="text-lg font-semibold">{data.session.words.length}</dd>
				</div>
				<div>
					<dt class="text-ctp-subtext0">Unique words</dt>
					<dd class="text-lg font-semibold">{data.distribution.length}</dd>
				</div>
				<div>
					<dt class="text-ctp-subtext0">Created</dt>
					<dd class="text-lg font-semibold">
						{humanReadableDate(new Date(data.session.createdAt))}
					</dd>
				</div>
				{#if !data.session.open && data.session.closedAt}
					<div>
						<dt class="text-ctp-subtext0">Closed</dt>
						<dd class="text-lg font-semibold">
							{humanReadableDate(new Date(data.session.closedAt))}
						</dd>
					</div>
				{/if}
			</dl>
			{#if data.session.open}
				<div class="ml-auto flex flex-wrap gap-2">
					<MeltTooltip text="Show the code and QR code to the audience">
						<button
							class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1 font-semibold text-ctp-mantle
								shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60"
							use:melt={$trigger}
						>
							<QrCode size="16" />
							Join info
						</button>
					</MeltTooltip>
					<button
						class="flex items-center gap-1 rounded-md bg-ctp-red px-3 py-1 font-semibold text-ctp-mantle
							shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60"
						on:click={closeSession}
					>
						<X size="16" />
						Close session
					</button>
				</div>
			{/if}
		</div>

		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col gap-4">
			<h2 class="text-2xl font-bold text-ctp-text">Word cloud</h2>
			<WordCloud data={data.distribution} filename="word-cloud-{data.session.code}" />
		</div>

		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col gap-4">
			<h2 class="text-2xl font-bold text-ctp-text">Most frequent words</h2>
			<BarChart data={data.distribution} />
		</div>

		<Table data={data.session.words} id={data.session.id} />
	</div>
</section>

<div use:melt={$portalled}>
	{#if $open}
		<div
			use:melt={$overlay}
			class="fixed inset-0 z-30 bg-black/50"
			transition:fade={{ duration: 200 }}
		/>
		<div
			class="fixed left-[50%] top-[50%] z-50 h-[90vh] w-[90vw]
            translate-x-[-50%] translate-y-[-50%] rounded-md bg-ctp-base
            p-6 shadow-md overflow-auto"
			transition:fly={{ duration: 200, y: 10 }}
			use:melt={$content}
		>
			<h2 use:melt={$title} class="mt-4 mb-8 text-center text-4xl font-bold">Join the session !</h2>
			<div class="w-full flex flex-col items-center" bind:clientWidth={qrWidthAvailable}>
				<canvas bind:this={canvas} />
				<a
					class="my-8 text-3xl font-medium"
					href="{window.location.origin}/tool/word-cloud?code={data.session.code}"
					target="_blank"
				>
					{window.location.origin}/tool/word-cloud
				</a>
				<p class="my-8 text-4xl font-bold">
					Code: <span class="text-ctp-mauve">{data.session.code}</span>
				</p>
			</div>

			<button
				use:melt={$close}
				aria-label="close"
				class="absolute right-4 top-4 inline-flex h-6 w-6 appearance-none
                items-center justify-center rounded-full p-1 text-base
                hover:bg-ctp-mauve hover:text-ctp-base transition-colors"
			>
				<X class="square-4" />
			</button>
		</div>
	{/if}
</div>
