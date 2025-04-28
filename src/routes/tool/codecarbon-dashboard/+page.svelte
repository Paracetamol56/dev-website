<script lang="ts">
	import { type Project, validateDescription, validateName } from './projects';
	import api from '$lib/api';
	import { Pencil, Plus, Save, Trash2, X } from 'lucide-svelte';
	import Button from '$lib/components/Button.svelte';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { fade, fly } from 'svelte/transition';
	import { addToast } from '../../+layout.svelte';
	import type { PageData } from './$types';
	import { invalidateAll } from '$app/navigation';

	export let data: PageData;

	let editedProjectDraft: Project | null = null;
	let nameError: string = '';
	let descriptionError: string = '';

	const {
		elements: { trigger, overlay, content, title, close, portalled },
		states: { open }
	} = createDialog({
		forceVisible: true
	});

	const {
		elements: {
			trigger: triggerNested,
			overlay: overlayNested,
			content: contentNested,
			title: titleNested,
			description: descriptionNested,
			close: closeNested,
			portalled: portalledNested
		},
		states: { open: openNested }
	} = createDialog({ forceVisible: true });

	function editProject(e: Event) {
		e.preventDefault();

		if (editedProjectDraft === null) return;

		nameError = validateName(editedProjectDraft.name);
		descriptionError = validateDescription(editedProjectDraft.description);

		if (nameError || descriptionError) return;

		api
			.callWithAuth('PATCH', `/codecarbon/projects/${editedProjectDraft.id}`, {
				name: editedProjectDraft.name,
				description: editedProjectDraft.description
			})
			.then((response) => {
				if (response.status !== 200) {
					throw new Error('Failed to update project');
				}

				invalidateAll();

				$open = false;
			})
			.catch((error) => {
				console.error('Error updating project:', error);
				addToast({
					data: {
						title: 'Error',
						description: 'Failed to update project',
						color: 'bg-ctp-red'
					}
				});
			});
	}

	function deleteProject(e: Event) {
		e.preventDefault();

		if (editedProjectDraft === null) return;

		api
			.callWithAuth('DELETE', `/codecarbon/projects/${editedProjectDraft.id}`)
			.then((response) => {
				if (response.status !== 204) {
					throw new Error('Failed to delete project');
				}

				invalidateAll();

				$openNested = false;
				$open = false;
			})
			.catch((error) => {
				console.error('Error deleting project:', error);
				addToast({
					data: {
						title: 'Error',
						description: 'Failed to delete project',
						color: 'bg-ctp-red'
					}
				});
			});
	}
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
	<div class="mb-8 flex justify-left gap-4">
		<h2 class="text-2xl font-bold text-ctp-lavender">Statistics</h2>
	</div>
</section>

<section class="container mx-auto">
	<div class="mb-8 flex justify-left gap-4">
		<h2 class="text-2xl font-bold text-ctp-lavender">Recent projects</h2>
		<Button link="/tool/codecarbon-dashboard/project/new">
			<span>New project</span>
			<Plus size="18" />
		</Button>
	</div>
	{#if data.projects.length === 0}
		<p>You don't have any open project.</p>
	{:else}
		<div class="mb-8 grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
			{#each data.projects as project (project.id)}
				<div class="p-8 bg-ctp-mantle rounded-md shadow-md shadow-ctp-crust">
					<div class="flex justify-start items-baseline mb-4">
						{#if project.closed_at !== null}
							<div class="relative square-2 mr-2 bg-ctp-green rounded-full">
								<span
									class="animate-ping absolute top-0.5 right-0.5 block square-1 rounded-full ring-2 ring-ctp-green bg-ctp-green"
								/>
							</div>
							<p class="text-ctp-subtext0 text-sm">Open</p>
						{:else}
							<div class="relative square-2 mr-2 bg-ctp-red rounded-full" />
							<p class="text-ctp-subtext0 text-sm">Archived</p>
						{/if}
						<p class="ml-auto text-ctp-subtext0 text-sm">
							{project.experiments.length} experiment{project.experiments.length > 1 ? 's' : ''}
						</p>
						<button
							type="button"
							class="ml-2 text-ctp-subtext0"
							use:melt={$trigger}
							on:m-click={(e) => {
								editedProjectDraft = Object.assign({}, project);
								console.log('Edit project', editedProjectDraft);
							}}
						>
							<Pencil size="16" />
						</button>
					</div>
					<a href="/tool/codecarbon-dashboard/{project.id}">
						<h4 class="mb-4 text-2xl font-bold hover:opacity-75 transition-opacity">
							{project.name}
						</h4>
					</a>
					<p class="text-ctp-subtext0">{project.description}</p>
				</div>
			{/each}
		</div>
	{/if}
</section>

{#if $open && editedProjectDraft !== null}
	<div use:melt={$portalled}>
		<div
			use:melt={$overlay}
			class="fixed inset-0 z-30 bg-black/50"
			transition:fade={{ duration: 200 }}
		/>
		<div
			class="fixed left-[50%] top-[50%] z-40 max-h-[85vh] w-[90vw]
            max-w-[450px] translate-x-[-50%] translate-y-[-50%] rounded-lg bg-ctp-base
            p-6 shadow-md"
			transition:fly={{ duration: 200, y: 10 }}
			use:melt={$content}
		>
			<h2 use:melt={$title} class="m-0 text-lg font-medium text-ctp-text">Login</h2>

			<form on:submit={editProject}>
				<div class="flex flex-col gap-y-6">
					<fieldset>
						<label for="name" class="mb-2 text-sm font-semibold"> Name </label>
						<input
							id="name"
							type="text"
							name="name"
							class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0
									shadow-md shadow-ctp-crust px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
							bind:value={editedProjectDraft.name}
							on:blur={() => (nameError = validateName(editedProjectDraft.name))}
						/>
						<p class="text-left text-sm font-semibold text-ctp-red">{nameError}</p>
					</fieldset>
					<fieldset>
						<label for="description" class="mb-2 text-sm font-semibold">
							Description <small>(optional)</small></label
						>
						<textarea
							id="description"
							name="description"
							class="flex h-32 w-full items-center justify-between rounded-md bg-ctp-surface0
									shadow-md shadow-ctp-crust px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
							bind:value={editedProjectDraft.description}
							on:blur={() =>
								(descriptionError = validateDescription(editedProjectDraft.description))}
						/>
						<p class="text-left text-sm font-semibold text-ctp-red">{descriptionError}</p>
					</fieldset>

					<div class="flex justify-end gap-4">
						<button
							use:melt={$triggerNested}
							class="flex items-center gap-1 rounded-md bg-ctp-red px-3 py-1
					font-semibold text-ctp-mantle
					shadow-md shadow-ctp-crust transition-opacity
					hover:opacity-80 active:opacity-60"
						>
							<span>Delete</span>
							<Trash2 size="16" />
						</button>
						<Button type="submit">
							<span>Save</span>
							<Save size="16" />
						</Button>
					</div>
				</div>
			</form>

			<button
				use:melt={$close}
				aria-label="close"
				class="absolute right-4 top-4 inline-flex h-6 w-6 appearance-none
                items-center justify-center rounded-full p-1 text-base
                hover:bg-ctp-mauve hover:text-ctp-base transition-colors"
			>
				<X class="square-4" />
			</button>
		</div>
	</div>
{/if}

{#if $openNested}
	<div class="" use:melt={$portalledNested}>
		<div
			use:melt={$overlayNested}
			class="fixed inset-0 z-40 bg-black/50"
			transition:fade={{ duration: 200 }}
		/>
		<div
			class="fixed left-[50%] top-[50%] z-50 max-h-[85vh] w-[90vw]
            max-w-[450px] translate-x-[-50%] translate-y-[-50%] rounded-lg bg-ctp-base
            p-6 shadow-md"
			transition:fly={{ duration: 200, y: 10 }}
			use:melt={$contentNested}
		>
			<h2 use:melt={$titleNested} class="m-0 text-lg font-medium text-ctp-text">Delete project</h2>
			<p use:melt={$descriptionNested} class="mb-5 mt-2 leading-normal text-ctp-text">
				You are about to delete this project. This action is irreversible.<br />
				<span class="font-semibold">Are you sure you want to continue?</span>
			</p>

			<form on:submit={deleteProject}>
				<div class="mt-6 flex justify-end gap-4">
					<button
						use:melt={$closeNested}
						aria-label="close"
						class="flex items-center gap-1 rounded-md border-2 border-ctp-mauve px-3 py-1
							font-semibold text-ctp-mauve
							shadow-md shadow-ctp-crust transition-opacity
							hover:opacity-80 active:opacity-60"
					>
						<span>Cancel</span>
						<X size="16" />
					</button>
					<Button type="submit">
						<span>Confirm</span>
						<Trash2 size="16" />
					</Button>
				</div>
			</form>
			<button
				use:melt={$closeNested}
				aria-label="close"
				class="absolute right-4 top-4 inline-flex h-6 w-6 appearance-none
                items-center justify-center rounded-full p-1 text-base
                hover:bg-ctp-mauve hover:text-ctp-base transition-colors"
			>
				<X class="square-4" />
			</button>
		</div>
	</div>
{/if}
