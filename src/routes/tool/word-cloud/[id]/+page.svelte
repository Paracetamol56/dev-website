<script lang="ts">
	import type { PageData } from './$types';
	import Button from '$lib/components/Button.svelte';
	import { Send } from 'lucide-svelte';
	import api from '$lib/api';
	import { onDestroy, onMount } from 'svelte';
	import { followSession } from '../socket';

	export let data: PageData;

	let open = data.session.open;
	let text = '';
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
			if (event.type === 'session') open = event.open;
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

	async function submit(retried = false): Promise<void> {
		try {
			const res = await api.call('POST', `/word-cloud/${data.session.id}/word`, {
				text,
				token: await participantToken()
			});
			sent = res.data.text;
			text = '';
		} catch (error: any) {
			const status = error.response?.status;
			const message = error.response?.data?.error;
			if (status === 401 && !retried) {
				token = null;
				storage.set(null);
				return submit(true);
			}
			if (message === 'word cloud closed') open = false;
			else if (status === 429) textError = 'Wait a second before sending another word';
			else textError = ERRORS[message] ?? 'Your word could not be sent, try again';
		}
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!validateWord() || sending) return;
		sending = true;
		sent = null;
		await submit();
		sending = false;
	}
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

	<div class="mx-auto max-w-xl bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust">
		{#if open}
			<form class="flex flex-col items-center gap-6" on:submit={handleSubmit}>
				<div class="w-full">
					<label for="text" class="mb-2 block text-sm font-semibold">Your answer</label>
					<input
						id="text"
						name="text"
						type="text"
						maxlength="100"
						autocomplete="off"
						class="flex h-10 w-full items-center rounded-md bg-ctp-surface0 px-3 shadow-md shadow-ctp-crust
							focus:outline-none transition-colors
							{textError
							? 'ring-2 ring-ctp-red'
							: sent
							? 'ring-2 ring-ctp-green'
							: 'focus:ring-2 focus:ring-ctp-mauve'}"
						bind:value={text}
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
			<p class="mt-6 text-center text-sm text-ctp-subtext0">
				Session code <strong class="font-mono text-ctp-mauve">{data.session.code}</strong>, share it
				with your neighbours.
			</p>
		{:else}
			<p class="py-8 text-center text-lg font-semibold text-ctp-subtext0" role="status">
				This session is closed, thank you for taking part!
			</p>
		{/if}
	</div>
</section>
