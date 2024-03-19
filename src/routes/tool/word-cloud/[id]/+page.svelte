<script lang="ts">
	import axios from "axios";
	import { onMount } from "svelte";
	import type { PageData } from "./$types";
	import Button from "$lib/components/Button.svelte";
	import { Send } from "lucide-svelte";

  export let data: PageData;
  let ip: string | null = null;
	let text: string = '';
	let textError: string = '';
	let textSuccess: boolean = false;

	/*onMount(() => {
		axios
			.get('https://api.ipify.org?format=json')
			.then((res) => {
				ip = res.data.ip;
			})
			.catch((error) => {
				console.error(error);
			});
		
		// Connect to the websocket
		const ws = new WebSocket(`/api/word-cloud/${data.session!.id}/ws`);
		ws.onopen = () => {
			ws.send(JSON.stringify({ type: 'join', code: data.session!.id }));
		};
	});*/

	const validateWord = () => {
		if (text.length === 0) {
			textError = 'A word is required';
			return false;
		}
		if (text.length > 100) {
			textError = 'Your word must be less than 100 characters long';
			return false;
		}
		if (JSON.parse(sessionStorage.getItem(data.session!.id)!)?.includes(text.toLowerCase())) {
			textError = 'This word has already been sent';
			return false;
		}
		textError = '';
		return true;
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		if (!validateWord()) return;
		// TODO: Send the word to the server
	};
</script>

<div class="mx-auto max-w-xl">
	<h2 class="mb-4 text-4xl font-bold text-center">
		{data.session?.name}
	</h2>
	<p class="text-center">
		{data.session?.description}
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
			<Button type="submit">
				<span>Send</span>
				<Send size="18" />
			</Button>
		</div>
	</form>

	<p class="mt-8 text-center">
		The session code is <strong>{data.session?.code}</strong>,<br />share it with your neighbors.
	</p>
</div>