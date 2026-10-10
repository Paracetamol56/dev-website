<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import api from '$lib/api';
	import Button from '$lib/components/Button.svelte';
	import { ArrowRightToLine, List, Plus } from 'lucide-svelte';
	import { addToast } from '../../+layout.svelte';
	import { createPinInput, melt } from '@melt-ui/svelte';
	import { user } from '$lib/store';

	let codeError: string = '';

	const {
		elements: { root, input },
		states: { value }
	} = createPinInput({
		placeholder: '•',
		defaultValue: $page.url.searchParams.get('code')?.split('') ?? []
	});

	const validateCode = (code: string[]) => {
		if (code.length !== 5 || code.some((char) => char === '')) {
			codeError = 'Code must be 5 characters long';
			return false;
		}
		// Every character must be a number, lowercase letter or uppercase letter
		if (!code.every((char) => /[a-zA-Z0-9]/.test(char))) {
			codeError = 'Code must only contain numbers and letters';
			return false;
		}
		codeError = '';
		return true;
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		if (validateCode($value)) {
			api
				.call('GET', `/word-cloud?code=${$value.join('')}`)
				.then((res) => {
					goto(`/tool/word-cloud/${res.data.id}`);
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

<svelte:head>
	<title>Word cloud - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-4xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				Word cloud
			</span>
		</h1>
		<p class="text-center text-ctp-subtext0 mb-8">
			Ask your audience a question and watch their answers grow into a live word cloud.
		</p>
	</hgroup>

	<div class="mx-auto flex max-w-xl flex-col gap-8">
		<form
			class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col items-center gap-6"
			on:submit={handleSubmit}
		>
			<h2 class="text-2xl font-bold text-ctp-text">Join a session</h2>
			<div class="w-fit">
				<label for="code" class="mb-2 block text-sm font-semibold">Session code</label>
				<div use:melt={$root} class="flex items-center gap-2">
					{#each Array.from({ length: 5 }) as _}
						<input
							id="code"
							autocomplete="off"
							type="text"
							maxlength="1"
							class="rounded-md bg-ctp-surface0 text-center text-lg uppercase text-ctp-text square-12
								shadow-md shadow-ctp-crust focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
							on:keydown={(e) => {
								if (e.key === 'Enter') handleSubmit(e);
							}}
							use:melt={$input()}
						/>
					{/each}
				</div>
				<p class="mt-1 text-left text-sm font-semibold text-ctp-red" aria-live="polite">
					{codeError}
				</p>
			</div>
			<Button type="submit">
				<span>Join</span>
				<ArrowRightToLine size="18" />
			</Button>
		</form>

		<div
			class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-wrap items-center justify-between gap-4"
		>
			<p class="text-ctp-subtext0">Running a talk or a workshop?</p>
			<div class="flex flex-wrap gap-2">
				<Button link="/tool/word-cloud/new">
					<Plus size="18" />
					<span>Create a session</span>
				</Button>
				{#if $user.id !== null}
					<Button link="/tool/word-cloud/manage">
						<List size="18" />
						<span>Your sessions</span>
					</Button>
				{/if}
			</div>
		</div>
	</div>
</section>
