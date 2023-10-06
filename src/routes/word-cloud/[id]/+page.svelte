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

  export let data: PageLoad;
  const distribution: {text: string, occurence: number}[] = [];
  let session: WordCloudSession;

  if (browser) {
    axios
      .get(`/api/word-cloud/${data.id}`)
      .then((res) => {
        session = res.data;
        if (session.words === undefined) {
          addToast({data: {
            title: "Unauthorized",
            description: "You are not the owner of this session",
            color: "bg-ctp-red",
          }})
          goto("/word-cloud");
        }
        if (session.words.length === 0) {
          distribution.push({ text: "No words yet", occurence: 1 });
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
      console.log(qrWidthAvailable);
      QRCode.toCanvas(
        canvas,
        url,
        {
          errorCorrectionLevel: 'H',
          scale: 3 + qrWidthAvailable / 175,
        },
        (error) => {
          if (error) console.error(error)
        }
      );
    }
  }

  const humanReadableDate = (date: string) => {
		return new Date(date).toLocaleString();
	};

  const {
    elements: {
      trigger,
      overlay,
      content,
      title,
      close,
      portalled,
    },
    states: { open },
  } = createDialog({
    forceVisible: true,
  });
</script>

<section class="container mx-auto my-16">
  <h1 class="mb-16 text-6xl font-bold text-center">
		<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">Word cloud</span>&nbsp;{session?.name || ""}
	</h1>

  {#if session}
    <WordCloud data={distribution}/>
    <div class="mb-4 p-4 w-full bg-ctp-mantle rounded-md">
      <p><strong>Participants:</strong> {session.words.length}</p>
      <p><strong>Unique words:</strong> {distribution.length}</p>
      <p><strong>Created at:</strong> {humanReadableDate(session.createdAt.toString())}</p>

      <div class="mt-2 flex justify-start gap-2">
        <button
          class="rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-mantle flex items-center gap-1 hover:opacity-75 active:opacity-50 transition-opacity"
          use:melt={$trigger}
        >
          Session info <QrCode size="18" />
        </button>
        <button
          class="rounded-md bg-ctp-red px-3 py-1 font-medium text-ctp-mantle flex items-center gap-1 hover:opacity-75 active:opacity-50 transition-opacity"
          use:melt={$trigger}
        >
          Close session <X size="18" />
        </button>
      </div>
    </div>
    <BarChart data={distribution}/>
    <Table data={session.words} id={session.id} />
  {/if}
</section>

<div use:melt={$portalled}>
  {#if $open}
    <div use:melt={$overlay} class="fixed inset-0 z-10 bg-black/50" transition:fade={{duration: 200}} />
    <div
      class="fixed left-[50%] top-[50%] z-50 h-[90vh] w-[90vw]
            translate-x-[-50%] translate-y-[-50%] rounded-md bg-ctp-base
            p-6 shadow-md overflow-auto"
      transition:fly={{ duration: 200, y: 10 }}
      use:melt={$content}
    >
      <h2 use:melt={$title} class="mt-4 mb-8 text-4xl text-center font-bold">
        Join the session !
      </h2>
      <div class="w-full flex justify-center" bind:clientWidth={qrWidthAvailable}>
        <canvas bind:this={canvas} />
      </div>
      <p class="my-8 text-3xl text-center font-medium">
        {window.location.origin}/word-cloud
      </p>
      <p class="my-8 text-4xl text-center font-bold">
        Code: <span class="text-ctp-mauve">{session.code}</span>
      </p>

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
