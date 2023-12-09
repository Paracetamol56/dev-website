<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import { Cog } from 'lucide-svelte';
	import { addToast } from '../../routes/+layout.svelte';
	import axios from 'axios';

	let prompt = '';
	let promptError = '';
	let loading = false;

	let generated: any | null = null;

	const validatePrompt = () => {
		if (prompt.length > 0 && prompt.length > 50) {
			promptError = 'Prompt cannot be longer than 50 characters';
			return false;
		}
		promptError = '';
		return true;
	};

	const handleSubmit = (event: Event) => {
		event.preventDefault();

		if (!validatePrompt()) {
			addToast({
				data: {
					title: 'Invalid form',
					description: 'Please check your inputs',
					color: 'bg-ctp-red'
				}
			});
			return;
		}

		loading = true;
		axios
			.post('/api/topic-generator', {
				prompt: prompt
			})
			.then((response) => {
				generated = response.data;
				promptError = '';
				loading = false;
			})
			.catch((error) => {
				addToast({
					data: {
						title: 'An error occured while generating your topic',
						description: error.response.data,
						color: 'bg-ctp-red'
					}
				});
				loading = false;
			});
	};
</script>

<section class="my-16">
	<form class="mx-auto mb-8 max-w-xl" on:submit={handleSubmit}>
		<fieldset class="mb-4 w-full">
			<label for="prompt" class="mb-2 text-sm font-semibold" data-svelte-h="svelte-9baxrn"
				>Domain (optional)</label
			>
			<input
				id="prompt"
				type="text"
				name="prompt"
				bind:value={prompt}
				on:blur={validatePrompt}
				class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
			<p class="text-left text-sm font-semibold text-ctp-red">
				{promptError}
			</p>
		</fieldset>
		<div class="flex justify-center">
			<Button type="submit">
				<span>Generate</span>
				<Cog stroke-width="3" />
			</Button>
		</div>
	</form>
	<div class="mx-auto max-w-xl">
		{#if generated}
			<p class="text-center text-2xl font-semibold">
				{generated.choices[0].message.content}
			</p>
		{:else if loading}
			<p class="text-center text-3xl font-semibold text-ctp-peach">Thinking...</p>
		{/if}
	</div>
</section>
