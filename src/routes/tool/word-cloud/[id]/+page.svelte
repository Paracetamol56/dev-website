<script lang="ts">
	import type { PageData } from './$types';
	import Button from '$lib/components/Button.svelte';
	import { Send, ThumbsUp } from 'lucide-svelte';
	import api from '$lib/api';
	import { onDestroy, onMount } from 'svelte';
	import { followSession } from '../socket';
	import { clientInfo } from '../client';
	import WordCloud from '../WordCloud.svelte';
	import Card from '../Card.svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import { writable } from 'svelte/store';
	import { countWords } from '../utils';

	export let data: PageData;

	let open = data.session.open;
	let deleted = false;
	let audienceWords: string[] = data.session.words ?? [];
	$: if (data.session.words) audienceWords = data.session.words;
	$: showCloud = open && data.session.settings?.showCloudToAudience;
	const answer = writable('');
	$: text = $answer;
	let textError = '';
	let sent: string | null = null;
	let sending = false;

	const tokenKey = `word-cloud:participant:${data.session.id}`;
	const storage = {
		get: () => {
			try {
				return localStorage.getItem(tokenKey);
			} catch {
				return null;
			}
		},
		set: (token: string | null) => {
			try {
				if (token) localStorage.setItem(tokenKey, token);
				else localStorage.removeItem(tokenKey);
			} catch {
				// Without storage the token only lasts for this visit
			}
		}
	};
	let token = storage.get();

	async function participantToken(): Promise<string> {
		if (!token) {
			const res = await api.call('POST', `/word-cloud/${data.session.id}/participant`);
			token = res.data.token as string;
			storage.set(token);
		}
		return token!;
	}

	let stopFollowing = () => {};
	onMount(() => {
		stopFollowing = followSession(data.session.id, (event) => {
			if (event.type === 'word') {
				audienceWords = [...audienceWords, event.word.text];
				recent = [event.word.text, ...recent.filter((word) => word !== event.word.text)];
				return;
			}
			if (event.type !== 'session') return;
			if (event.session) {
				data.session = event.session;
				open = event.session.open;
			} else {
				deleted = true;
				open = false;
			}
		});
	});
	onDestroy(() => stopFollowing());

	const ERRORS: Record<string, string> = {
		'word already submitted': 'You already sent this word',
		'word limit reached': 'You reached the maximum number of words for this session',
		'word is empty': 'Use letters or numbers'
	};

	const validateWord = () => {
		if (text.trim().length === 0) {
			textError = 'A word is required';
			return false;
		}
		if (text.length > 100) {
			textError = 'Your word must be less than 100 characters long';
			return false;
		}
		textError = '';
		return true;
	};

	async function submit(word: string, retried = false): Promise<void> {
		try {
			const res = await api.call('POST', `/word-cloud/${data.session.id}/word`, {
				text: word,
				token: await participantToken(),
				client: clientInfo()
			});
			sent = res.data.text;
			mine = [...mine, res.data.text];
			saveMine();
		} catch (error: any) {
			const status = error.response?.status;
			const message = error.response?.data?.error;
			if (status === 401 && !retried) {
				token = null;
				storage.set(null);
				return submit(word, true);
			}
			if (message === 'word cloud closed') open = false;
			else if (status === 429) textError = 'Wait a second before sending another word';
			else textError = ERRORS[message] ?? 'Your word could not be sent, try again';
		}
	}

	async function send(word: string) {
		if (sending) return;
		sending = true;
		sent = null;
		textError = '';
		await submit(word);
		sending = false;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!validateWord()) return;
		await send(text);
		if (sent) answer.set('');
	}

	const mineKey = `word-cloud:sent:${data.session.id}`;
	let mine: string[] = (() => {
		try {
			return JSON.parse(localStorage.getItem(mineKey) ?? '[]');
		} catch {
			return [];
		}
	})();
	function saveMine() {
		try {
			localStorage.setItem(mineKey, JSON.stringify(mine));
		} catch {
			// Own words then only stay hidden from the suggestions during this visit
		}
	}

	const UPVOTE_CHOICES = 3;
	let recent: string[] = data.session.recent ?? [];
	$: if (data.session.recent) recent = data.session.recent;
	$: suggestions = recent.filter((word) => !mine.includes(word)).slice(0, UPVOTE_CHOICES);
	$: limit = data.session.settings.maxWords ?? 0;
	$: wordsLeft = limit > 0 ? Math.max(0, limit - mine.length) : Infinity;
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

	<Card class="mx-auto max-w-xl">
		{#if open}
			<form class="flex flex-col items-center gap-6" on:submit={handleSubmit}>
				<div class="w-full">
					<TextInput
						label="Your answer"
						value={answer}
						maxlength={100}
						class={textError ? 'ring-2 ring-ctp-red' : sent ? 'ring-2 ring-ctp-green' : ''}
						on:input={() => {
							sent = null;
							textError = '';
						}}
					/>
					<p class="mt-1 text-left text-sm font-semibold text-ctp-red" aria-live="polite">
						{textError}
					</p>
					{#if sent}
						<p class="mt-1 text-left text-sm font-semibold text-ctp-green" aria-live="polite">
							“{sent}” was sent
						</p>
					{/if}
				</div>
				<Button type="submit" disabled={sending}>
					<span>Send</span>
					<Send size="18" />
				</Button>
			</form>
			{#if data.session.settings.allowUpvotes && suggestions.length}
				<div class="mt-6 flex flex-col gap-2">
					<p class="text-sm font-semibold">Or agree with someone else</p>
					<ul class="flex flex-wrap gap-2">
						{#each suggestions as word (word)}
							<li>
								<button
									type="button"
									class="flex items-center gap-1 rounded-md bg-ctp-surface0 px-3 py-1 font-semibold shadow-md shadow-ctp-crust hover:bg-ctp-mauve hover:text-ctp-base disabled:opacity-50"
									disabled={sending || wordsLeft === 0}
									aria-label="Upvote “{word}”"
									on:click={() => send(word)}
								>
									<ThumbsUp size="14" />
									{word}
								</button>
							</li>
						{/each}
					</ul>
				</div>
			{/if}
			{#if limit > 0}
				<p class="mt-4 text-center text-xs text-ctp-subtext0">
					{wordsLeft} word{wordsLeft === 1 ? '' : 's'} left
				</p>
			{/if}
			<p class="mt-6 text-center text-sm text-ctp-subtext0">
				Session code <strong class="font-mono text-ctp-mauve">{data.session.code}</strong>, share it
				with your neighbours.
			</p>
		{:else}
			<p class="py-8 text-center text-lg font-semibold text-ctp-subtext0" role="status">
				{deleted
					? 'This session no longer exists.'
					: 'This session is closed, thank you for taking part!'}
			</p>
		{/if}
	</Card>

	{#if showCloud}
		<Card title="Live answers" class="mx-auto mt-8 max-w-3xl">
			<WordCloud data={countWords(audienceWords)} exportable={false} />
		</Card>
	{/if}
</section>
