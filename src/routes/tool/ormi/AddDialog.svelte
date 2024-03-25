<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import MeltTags from '$lib/components/MeltTags.svelte';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { Plus, X } from 'lucide-svelte';
	import { writable, type Writable } from 'svelte/store';
	import { fade, fly } from 'svelte/transition';

	const dialogOpen: Writable<boolean> = writable(false);
	const {
		elements: { trigger, overlay, content, title, close, portalled },
		states: { open }
	} = createDialog({
		forceVisible: true,
		open: dialogOpen
	});

	let titleValue = '';
	let titleError = '';
	let descriptionValue = '';
	let descriptionError = '';

	function validateTitle(): boolean {
		if (titleValue.length == 0) {
			titleError = 'Title is required';
			return false;
		}
		if (titleValue.length < 2) {
			titleError = 'Title must be at least 2 characters long';
			return false;
		}
		if (titleValue.length > 100) {
			titleError = 'Title must be less than 100 characters long';
			return false;
		}
		titleError = '';
		return true;
	}

	function validateDescription(): boolean {
		if (descriptionValue.length > 1000) {
			descriptionError = 'Description must be less than 1000 characters long';
			return false;
		}
		descriptionError = '';
		return true;
	}

	const handleSubmit = (event: Event) => {
		event.preventDefault();
		console.log('submit');
		dialogOpen.set(false);
	};
</script>

<!-- Button -->
<button
	class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
  font-semibold text-ctp-mantle
  shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60"
	use:melt={$trigger}
>
	<span>Add</span>
	<Plus size="16" stroke-width="3" />
</button>

<!-- Dialog -->
<div use:melt={$portalled}>
	{#if $open}
		<div
			use:melt={$overlay}
			class="fixed inset-0 z-30 bg-black/50"
			transition:fade={{ duration: 200 }}
		/>
		<div
			class="fixed left-[50%] top-[50%] z-50 max-h-[85vh] w-[90vw]
            max-w-[450px] translate-x-[-50%] translate-y-[-50%] rounded-lg bg-ctp-base
            p-6 shadow-md"
			transition:fly={{ duration: 200, y: 10 }}
			use:melt={$content}
		>
			<h2 use:melt={$title} class="m-0 text-lg font-medium text-ctp-text">Add a new task</h2>

			<form on:submit={handleSubmit}>
				<fieldset class="w-full">
					<label class="mb-2 text-sm font-semibold" for="title"> Title* </label>
					<input
						class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0
                shadow-md shadow-ctp-crust px-3 pr-12 focus:outline-none focus:ring-2 focus:ring-ctp-mauve {titleError !=
						''
							? 'ring-2 ring-ctp-red'
							: ''}"
						id="title"
						type="text"
						name="title"
						bind:value={titleValue}
						on:blur={validateTitle}
					/>
					<p class="text-left text-sm font-semibold text-ctp-red">{titleError}</p>
				</fieldset>

				<fieldset class="w-full">
					<label class="mb-2 text-sm font-semibold" for="description"> Description </label>
					<textarea
						class="flex h-32 w-full items-center justify-between rounded-md bg-ctp-surface0
								shadow-md shadow-ctp-crust px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve {descriptionError !=
						''
							? 'ring-2 ring-ctp-red'
							: ''}"
						id="description"
						name="description"
						bind:value={descriptionValue}
						on:blur={validateDescription}
					/>
					<p class="text-left text-sm font-semibold text-ctp-red">{descriptionError}</p>
				</fieldset>
				<fieldset class="w-full">
					<label class="mb-2 text-sm font-semibold" for="description"> Labels </label>
					<MeltTags />
				</fieldset>

				<div class="mt-6 flex justify-end gap-4">
					<Button type="submit">
						<span>Add</span>
						<Plus size="16" />
					</Button>
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
	{/if}
</div>
