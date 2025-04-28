<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import { user } from '$lib/store';
	import { AlertTriangle, Send } from 'lucide-svelte';
	import LoginDialog from '../../../../LoginDialog.svelte';
	import api from '$lib/api';
	import { validateDescription, validateName } from '../../projects';
	import { goto } from '$app/navigation';
	import { addToast } from '../../../../+layout.svelte';

	let name: string = '';
	let nameError: string = '';
	let description: string = '';
	let descriptionError: string = '';

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		api
			.callWithAuth('POST', '/codecarbon/projects', {
				name,
				description
			})
			.then((response) => {
				if (response.status !== 201) throw new Error('Failed to create project');
				// Navigate to the project page
				goto(`/tool/codecarbon-dashboard/project/${response.data.id}`);
			})
			.catch((error) => {
				console.error('Error creating project:', error);
				addToast({
					data: {
						title: 'Error creating project',
						description: 'Please try again later',
						color: 'bg-ctp-red'
					}
				});
			});
	};
</script>

<svelte:head>
	<title>CodeCarbon dashboard - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				CodeCarbon dashboard
			</span>
		</h1>
	</hgroup>
</section>

<section class="container mx-auto">
	<h2 class="mb-4 text-2xl font-semibold text-center">Create a new project</h2>
	<div class="relative">
		<form class="mx-auto max-w-xl {$user.accessToken ? '' : 'blur-sm'}" on:submit={handleSubmit}>
			<div class="flex flex-col gap-y-6">
				<fieldset class="w-full">
					<label for="name" class="mb-2 text-sm font-semibold"> Name </label>
					<input
						id="name"
						type="text"
						name="name"
						class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0
									shadow-md shadow-ctp-crust px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
						bind:value={name}
						on:blur={() => (nameError = validateName(name))}
						disabled={$user.id === null}
					/>
					<p class="text-left text-sm font-semibold text-ctp-red">{nameError}</p>
				</fieldset>
				<fieldset class="sm:col-span-2">
					<label for="description" class="mb-2 text-sm font-semibold">
						Description <small>(optional)</small></label
					>
					<textarea
						id="description"
						name="description"
						class="flex h-32 w-full items-center justify-between rounded-md bg-ctp-surface0
									shadow-md shadow-ctp-crust px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
						bind:value={description}
						on:blur={() => (descriptionError = validateDescription(description))}
						disabled={$user.id === null}
					/>
					<p class="text-left text-sm font-semibold text-ctp-red">{descriptionError}</p>
				</fieldset>
				<div class="sm:col-span-2 flex justify-end">
					<Button type="submit">
						<span>Create</span>
						<Send size="16" />
					</Button>
				</div>
			</div>
		</form>
		{#if $user.accessToken === null}
			<div>
				<div
					class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 flex flex-col items-center"
				>
					<div class="text-ctp-red font-semibold mb-4 flex items-baseline gap-1">
						<AlertTriangle size="16" />
						<p>You must be logged in to create a new session</p>
					</div>
					<LoginDialog />
				</div>
			</div>
		{/if}
	</div>
</section>
