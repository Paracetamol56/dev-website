<script lang="ts">
	import { page } from '$app/stores';
	import { createPinInput, melt } from '@melt-ui/svelte';
	import axios from 'axios';
	import { ArrowRightToLine } from 'lucide-svelte';
	import { addToast } from '../+layout.svelte';
	import type { WordCloudSession } from './utils';
	import { user } from '$lib/store';

	export let joinSession: (session: WordCloudSession) => void;
	let codeError: string = '';

	const {
		elements: { root, input },
		states: { value }
	} = createPinInput({
		placeholder: '•',
		defaultValue: $page.url.searchParams.get('code')?.split('') ?? []
	});

	const validateCode = (code: string[]) => {
		if (code.length !== 5) {
			codeError = 'Code must be 5 characters long';
			return false;
		}
		// Every character must be a number, lowercase letter or uppercase letter
		if (!code.every((char) => /[0-9a-z]/.test(char))) {
			codeError = 'Code must only contain numbers and letters';
			return false;
		}
		codeError = '';
		return true;
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		if (validateCode($value)) {
			axios
				.get('/api/word-cloud', {
					params: {
						code: $value.join('')
					}
				})
				.then((res) => {
					if (res.data) {
						joinSession(res.data);
					} else {
						codeError = 'An error occured';
					}
				})
				.catch((error) => {
					console.error(error);
					if (error.response.status === 404) {
						codeError = 'Session not found';
						addToast({
							data: {
								title: 'Error',
								description: 'The code you provided does not match any open session',
								color: 'bg-ctp-red'
							}
						});
					} else {
						codeError = 'An error occured';
						addToast({
							data: {
								title: 'Error',
								description: 'An error occured while joining the session',
								color: 'bg-ctp-red'
							}
						});
					}
				});
		}
	};
</script>

<form on:submit={handleSubmit}>
	<div class="flex flex-col gap-8 items-center">
		<div class="w-fit">
			<label for="code" class="mb-2 text-sm font-semibold"> Session code </label>
			<div use:melt={$root} class="flex items-center gap-2">
				{#each Array.from({ length: 5 }) as _}
					<input
						id="code"
						class="rounded-md bg-ctp-surface0 text-center text-lg text-ctp-text shadow-sm square-12
                  focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
						use:melt={$input()}
					/>
				{/each}
			</div>
			<p class="text-left text-sm font-semibold text-ctp-red">{codeError}</p>
		</div>
		<button
			class="flex justify-center items-center rounded-md bg-ctp-mauve px-3 py-1 font-medium
            text-ctp-surface0 hover:opacity-75 active:opacity-50 transition-opacity"
			type="submit"
		>
			Join&nbsp;<ArrowRightToLine size="18" />
		</button>
	</div>
</form>

<div class="mt-8 flex justify-center gap-4">
	<a class="font-semibold hover:text-ctp-blue transition-colors" href="/word-cloud/new"
		>Create a new session</a
	>
	{#if $user.id !== null}
		<a class="font-semibold hover:text-ctp-blue transition-colors" href="/word-cloud/all"
			>View your sessions</a
		>
	{/if}
</div>
