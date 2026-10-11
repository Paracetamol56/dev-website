<script lang="ts">
	import { user } from '$lib/store';
	import { Maximize, MessageSquareText, Minimize, Moon, Sun, Users, X } from 'lucide-svelte';
	import Kpi from '../../../Kpi.svelte';
	import QRCode from 'qrcode';
	import { onDestroy, onMount } from 'svelte';
	import WordCloud from '../../../WordCloud.svelte';
	import Sentiment from '../Sentiment.svelte';
	import { followSession } from '../../../socket';
	import { countWords, type WordCloudSessionAdmin, type WordCloudWord } from '../../../utils';
	import type { PageData } from './$types';

	export let data: PageData & { session: WordCloudSessionAdmin };

	let session: WordCloudSessionAdmin = data.session;
	let deleted = false;
	$: visible = session.words.filter((word) => !word.hidden);
	$: distribution = countWords(visible.map((word) => word.text));
	$: participantCount = new Set(session.words.map((word) => word.uuid)).size;

	let joinUrl = '';
	let qr = '';
	$: if (joinUrl)
		QRCode.toDataURL(`${joinUrl}?code=${session.code}`, { margin: 1, width: 480 }).then(
			(url: string) => (qr = url)
		);

	let stage: HTMLElement;
	let fullscreen = false;
	const toggleFullscreen = () =>
		document.fullscreenElement ? document.exitFullscreen() : stage.requestFullscreen();
	const onFullscreen = () => (fullscreen = document.fullscreenElement !== null);

	const toggleTheme = () => ($user.flavour = $user.flavour === 'latte' ? 'mocha' : 'latte');

	let stopFollowing = () => {};
	onMount(() => {
		joinUrl = `${location.origin}/tool/word-cloud`;
		document.addEventListener('fullscreenchange', onFullscreen);
		stopFollowing = followSession(
			session.id,
			(event) => {
				if (event.type === 'word') session.words = [...session.words, event.word as WordCloudWord];
				else if (event.type === 'session') {
					if (!event.session) deleted = true;
					else
						session = {
							...session,
							...event.session,
							settings: { ...session.settings, ...event.session.settings, ...event.settings },
							words: event.words ?? session.words
						};
				}
			},
			$user.accessToken
		);
	});
	onDestroy(() => {
		stopFollowing();
		if (typeof document !== 'undefined')
			document.removeEventListener('fullscreenchange', onFullscreen);
	});

	const control =
		'grid square-9 place-items-center rounded-md bg-ctp-surface0 text-ctp-text hover:bg-ctp-surface1 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve';
</script>

<svelte:head>
	<title>{session.name} - Presenter - Word cloud</title>
</svelte:head>

<section
	bind:this={stage}
	class="relative flex min-h-[calc(100vh-6rem)] flex-col gap-6 bg-ctp-base p-6 lg:flex-row {fullscreen
		? 'h-screen'
		: 'container mx-auto rounded-md'}"
>
	<div class="flex min-w-0 flex-1 flex-col gap-4">
		<h1 class="text-center text-3xl font-bold text-ctp-text lg:text-5xl">{session.name}</h1>
		<div class="flex flex-1 items-center">
			{#if deleted}
				<p class="w-full text-center text-2xl text-ctp-subtext0">This session was deleted.</p>
			{:else}
				<div class="w-full"><WordCloud data={distribution} exportable={false} /></div>
			{/if}
		</div>
	</div>

	<aside class="flex shrink-0 flex-col items-center justify-center gap-6 pt-12 lg:w-80">
		{#if session.open}
			<p class="text-center text-lg text-ctp-subtext0">
				Join at <strong class="text-ctp-text">{joinUrl.replace(/^https?:\/\//, '')}</strong>
			</p>
			<p
				class="font-mono text-6xl font-bold tracking-widest text-ctp-mauve"
				aria-label="Code {session.code}"
			>
				{session.code}
			</p>
			{#if qr}
				<img
					src={qr}
					alt="QR code to join the session"
					class="w-56 rounded-md bg-white p-2 lg:w-full"
				/>
			{/if}
		{:else}
			<p class="text-center text-2xl font-semibold text-ctp-subtext0">Answers are closed</p>
		{/if}
		{#if session.settings.sentiment && session.settings.sentimentInPresenter}
			<div class="w-full rounded-md bg-ctp-mantle p-4">
				<Sentiment words={visible} compact />
			</div>
		{/if}
		<div class="grid w-full grid-cols-2 gap-4">
			<Kpi icon={Users} iconClass="text-ctp-peach" label="Participants" value={participantCount} />
			<Kpi
				icon={MessageSquareText}
				iconClass="text-ctp-blue"
				label="Answers"
				value={visible.length}
			/>
		</div>
	</aside>

	<div class="absolute right-4 top-4 flex gap-2">
		<button
			type="button"
			class={control}
			on:click={toggleTheme}
			aria-label="Switch light or dark theme"
		>
			<svelte:component this={$user.flavour === 'latte' ? Moon : Sun} size="18" />
		</button>
		<button
			type="button"
			class={control}
			on:click={toggleFullscreen}
			aria-label={fullscreen ? 'Exit full screen' : 'Full screen'}
		>
			<svelte:component this={fullscreen ? Minimize : Maximize} size="18" />
		</button>
		{#if !fullscreen}
			<a
				href="/tool/word-cloud/manage/{session.id}"
				class={control}
				aria-label="Back to the dashboard"
			>
				<X size="18" />
			</a>
		{/if}
	</div>
</section>
