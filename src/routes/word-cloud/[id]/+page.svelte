<script lang="ts">
	import WordCloud from './WordCloud.svelte';
	import BarChart from './BarChart.svelte';
	import Table from './Table.svelte';
	import axios from 'axios';
	import type { PageLoad } from './$types';
	import { browser } from '$app/environment';
	import type { WordCloudSession } from '../utils';
	import { addToast } from '../../+layout.svelte';
	import { goto } from '$app/navigation';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { QrCode, X } from 'lucide-svelte';
	import { fade, fly } from 'svelte/transition';
	import { onMount } from 'svelte';
	import QRCode from 'qrcode';
	import { create } from 'd3';
	import Button from '$lib/components/Button.svelte';

	export let data: PageLoad;
	const distribution: { text: string; occurence: number }[] = [];
	let session: WordCloudSession;

	if (browser) {
		axios
			.get(`/api/word-cloud/${data.id}`)
			.then((res) => {
				session = res.data;
				if (session.words === undefined) {
					addToast({
						data: {
							title: 'Unauthorized',
							description: 'You are not the owner of this session',
							color: 'bg-ctp-red'
						}
					});
					goto('/word-cloud');
				}
				for (const word of session.words) {
					const wDistribition = distribution.find((w) => w.text.toLowerCase() === word.text);
					if (wDistribition) {
						wDistribition.occurence++;
					} else {
						distribution.push({ text: word.text.toLowerCase(), occurence: 1 });
					}
					distribution.sort((a, b) => b.occurence - a.occurence);
				}
			})
			.catch((err) => {
				console.error(err);
			});
	}

	let qrWidthAvailable: number;
	let canvas: HTMLCanvasElement;

	$: {
		if (canvas !== undefined) {
			const url = `${window.location.origin}/word-cloud?code=${session.code}`;
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

		axios
			.delete(`/api/word-cloud/${session.id}`)
			.then((res) => {
				if (res.status === 204) {
					addToast({
						data: {
							title: 'Success',
							description: 'The session has been closed',
							color: 'bg-ctp-green'
						}
					});
					session.open = false;
					session.closedAt = new Date();
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
	<title>Word cloud - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender"
				>Word cloud</span
			>
			{session?.name || ''}
		</h1>
	</hgroup>
</section>

<section class="container mx-auto">
	{#if session}
		<WordCloud data={distribution} />
		<div class="mb-4 p-4 w-full bg-ctp-mantle rounded-md">
			<p><strong>Submitions:</strong> {session.words.length}</p>
			<p><strong>Unique words:</strong> {distribution.length}</p>
			<p><strong>Created at:</strong> {humanReadableDate(new Date(session.createdAt))}</p>
			{#if !session.open}
				<p>
					<strong>Closed at:</strong>
					{humanReadableDate(session.closedAt ? new Date(session.closedAt) : null)}
				</p>
			{/if}

			<div class="mt-2 flex justify-start gap-2">
				{#if session.open}
					<button
						class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
                  font-semibold text-ctp-mantle
                  shadow-md shadow-ctp-crust transition-opacity
                  hover:opacity-80 active:opacity-60"
						use:melt={$trigger}
					>
						Session info <QrCode size="20" stroke-width="3" />
					</button>
					<button
						class="flex items-center gap-1 rounded-md bg-ctp-red px-3 py-1
                  font-semibold text-ctp-mantle
                  shadow-md shadow-ctp-crust transition-opacity
                  hover:opacity-80 active:opacity-60"
						on:click={closeSession}
					>
						Close session <X size="20" stroke-width="3" />
					</button>
				{/if}
			</div>
		</div>
		<BarChart data={distribution} />
		<Table data={session.words} id={session.id} />
	{/if}
</section>

<div use:melt={$portalled}>
	{#if $open}
		<div
			use:melt={$overlay}
			class="fixed inset-0 z-10 bg-black/50"
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
					href="{window.location.origin}/word-cloud?code={session.code}"
					target="_blank"
				>
					{window.location.origin}/word-cloud
				</a>
				<p class="my-8 text-4xl font-bold">
					Code: <span class="text-ctp-mauve">{session.code}</span>
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
