<script lang="ts">
	import { user } from '$lib/store';
	import { AlertTriangle, Send } from 'lucide-svelte';
	import LoginDialog from '../../../LoginDialog.svelte';
	import { addToast } from '../../../+layout.svelte';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import { writable } from 'svelte/store';
	import api from '$lib/api';
	import SettingsFields from '../SettingsFields.svelte';
	import { DEFAULT_SETTINGS, type Settings } from '../utils';

	const nameStore = writable('');
	$: name = $nameStore;
	let nameError: string = '';
	let description: string = '';
	let descriptionError: string = '';
	let settings: Settings;

	const validateName = (name: string) => {
		name = name.trim();
		if (name.length === 0) {
			nameError = 'Name is required';
			return false;
		}
		if (name.length < 3) {
			nameError = 'Name must be at least 3 characters long';
			return false;
		}
		if (name.length > 100) {
			nameError = 'Name must be less than 100 characters long';
			return false;
		}
		nameError = '';
		return true;
	};

	const validateDescription = (description: string) => {
		description = description.trim();
		if (description.length === 0) {
			descriptionError = '';
			return true;
		}
		if (description.length < 10) {
			descriptionError = 'Description must be at least 10 characters long';
			return false;
		}
		if (description.length > 1000) {
			descriptionError = 'Description must be less than 1000 characters long';
			return false;
		}
		descriptionError = '';
		return true;
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		if (!validateName(name) || !validateDescription(description)) {
			addToast({
				data: {
					title: 'Error',
					description: 'Please check your inputs',
					color: 'bg-ctp-red'
				}
			});
			return;
		}
		api
			.callWithAuth('POST', '/word-cloud', {
				name,
				description,
				settings
			})
			.then((res) => {
				addToast({
					data: {
						title: 'Success',
						description: 'Word cloud created',
						color: 'bg-ctp-green'
					}
				});
				goto(`/tool/word-cloud/manage/${res.data.id}`);
			})
			.catch((err) => {
				addToast({
					data: {
						title: 'Error',
						description: 'An error occured while creating the session',
						color: 'bg-ctp-red'
					}
				});
			});
	};
</script>

<svelte:head>
	<title>New word cloud - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-4xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				New word cloud
			</span>
		</h1>
		<p class="text-center text-ctp-subtext0 mb-8">
			Name your question; your audience joins with a code or a QR code.
		</p>
	</hgroup>

	<div class="relative mx-auto max-w-xl">
		<form
			class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col gap-6
				{$user.accessToken ? '' : 'blur-sm'}"
			on:submit={handleSubmit}
			inert={$user.accessToken === null}
		>
			<div>
				<TextInput
					label="Question or title"
					value={nameStore}
					maxlength={100}
					on:blur={() => validateName(name)}
				/>
				<p class="text-left text-sm font-semibold text-ctp-red" aria-live="polite">{nameError}</p>
			</div>
			<div>
				<label for="description" class="mb-1 block text-sm font-semibold text-ctp-text">
					Description <small class="text-ctp-subtext0">(optional)</small>
				</label>
				<textarea
					id="description"
					name="description"
					maxlength="1000"
					class="flex h-32 w-full rounded-md bg-ctp-surface0 shadow-md shadow-ctp-crust px-3 py-2
						focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
					bind:value={description}
					on:blur={() => validateDescription(description)}
				/>
				<p class="text-left text-sm font-semibold text-ctp-red" aria-live="polite">
					{descriptionError}
				</p>
			</div>
			<SettingsFields initial={DEFAULT_SETTINGS} bind:settings />
			<div class="flex justify-end">
				<Button type="submit">
					<span>Create</span>
					<Send size="16" />
				</Button>
			</div>
		</form>
		{#if $user.accessToken === null}
			<div class="absolute inset-0 flex flex-col items-center justify-center gap-4 text-center">
				<p class="flex items-baseline gap-1 font-semibold text-ctp-red">
					<AlertTriangle size="16" />
					You must be logged in to create a session
				</p>
				<LoginDialog />
			</div>
		{/if}
	</div>
</section>
