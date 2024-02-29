<script lang="ts">
	import { Send } from 'lucide-svelte';
	import type { WordCloudSession } from './utils';
	import axios from 'axios';
	import { addToast } from '../../+layout.svelte';
	import { onMount } from 'svelte';

	export let session: WordCloudSession;

	let ip: string | null = null;
	let text: string = '';
	let textError: string = '';
	let textSuccess: boolean = false;

	onMount(() => {
		axios
			.get('https://api.ipify.org?format=json')
			.then((res) => {
				ip = res.data.ip;
			})
			.catch((error) => {
				console.error(error);
			});
		
		// Connect to the websocket
		const ws = new WebSocket('/api/word-cloud/ws');
		ws.onopen = () => {
			ws.send(JSON.stringify({ type: 'join', code: session.id }));
		};
	});

	const validateWord = () => {
		if (text.length === 0) {
			textError = 'A word is required';
			return false;
		}
		if (text.length > 100) {
			textError = 'Your word must be less than 100 characters long';
			return false;
		}
		if (JSON.parse(sessionStorage.getItem(session.id)!)?.includes(text.toLowerCase())) {
			textError = 'This word has already been sent';
			return false;
		}
		textError = '';
		return true;
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		if (!validateWord()) return;
		axios
			.put(`/api/word-cloud/${session.id}`, { text, ip })
			.then(() => {
				const words = sessionStorage.getItem(session.id)
					? JSON.parse(sessionStorage.getItem(session.id)!)
					: [];
				words.push(text.toLowerCase());
				sessionStorage.setItem(session.id, JSON.stringify(words));
				text = '';
				textError = '';
				textSuccess = true;
			})
			.catch((error) => {
				console.error(error);
				textError = 'An error occured';
				addToast({
					data: {
						title: 'Error',
						description: 'An error occured while sending your word',
						color: 'bg-ctp-red'
					}
				});
			});
	};
</script>

<div class="mx-auto max-w-xl">
	<h2 class="mb-4 text-4xl font-bold text-center">
		{session.name}
	</h2>
	<p class="text-center">
		{session.description}
	</p>

	<form class="my-16" on:submit={handleSubmit}>
		<div class="flex flex-col gap-8 items-center">
			<div class="w-full">
				<label for="text" class="mb-2 text-sm font-semibold"> Write something </label>
				<input
					id="text"
					name="text"
					type="text"
					class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0
                px-3 pr-12 focus:outline-none
                {textError ? 'ring-2 ring-ctp-red' : ''}
                {textSuccess
						? 'ring-2 ring-ctp-green'
						: 'focus:ring-2 focus:ring-ctp-mauve'} transition-colors"
					bind:value={text}
					on:blur={() => validateWord()}
					on:input={() => (textSuccess = false)}
				/>
				<p class="text-left text-sm font-semibold text-ctp-red">{textError}</p>
			</div>
			<button
				class="flex justify-center items-center rounded-md bg-ctp-mauve px-3 py-1 font-medium
              text-ctp-surface0 hover:opacity-75 active:opacity-50 transition-opacity"
				type="submit"
			>
				Send&nbsp;<Send size="18" />
			</button>
		</div>
	</form>

	<p class="mt-8 text-center">
		The session code is <strong>{session.code}</strong>,<br />share it with your neighbors.
	</p>
</div>
