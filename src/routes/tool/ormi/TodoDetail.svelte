<script lang="ts">
	import api from '$lib/api';
	import { addToast } from '../../+layout.svelte';
	import type { MeltElement } from '@melt-ui/svelte/internal/helpers';
	import type { Todo } from './+page';
	import Button from '$lib/components/Button.svelte';
	import MeltDatePicker from '$lib/components/MeltDatePicker.svelte';
	import MeltTags from '$lib/components/MeltTags.svelte';
	import { melt, type Tag } from '@melt-ui/svelte';
	import { Save } from 'lucide-svelte';
	import {
		fromDate,
		getLocalTimeZone,
		toCalendarDate,
		today,
		type DateValue
	} from '@internationalized/date';
	import { onMount } from 'svelte';
	import { writable, type Writable } from 'svelte/store';
	import { invalidateAll } from '$app/navigation';
	import { todoStates, type TodoState } from './states';
	import { suggestLabels } from './labels';

	export let todoId: string;
	export let title: MeltElement<any, any, any, any>;
	export let description: MeltElement<any, any, any, any>;

	let todo: Todo | null = null;
	let newTitle: string = '';
	let titleError: string = '';
	let newDescription: string = '';
	let descriptionError: string = '';
	let newState: TodoState = 'TODO';
	const states = Object.keys(todoStates) as TodoState[];
	const dueDate: Writable<DateValue | undefined> = writable(undefined);
	let savedDueDate: string | undefined;
	const labels: Writable<Tag[]> = writable([]);

	const showError = (description: string) => {
		addToast({
			data: {
				title: 'Error',
				description,
				color: 'bg-ctp-red'
			}
		});
	};

	onMount(() => {
		api
			.callWithAuth('GET', `/ormi/${todoId}`)
			.then((response) => {
				todo = response.data;
				newTitle = todo?.title ?? '';
				newDescription = todo?.description ?? '';
				newState = todo?.state ?? 'TODO';
				$labels = (todo?.labels ?? []).map((label) => ({ id: label, value: label }));
				$dueDate = todo?.dueDate
					? toCalendarDate(fromDate(new Date(todo.dueDate), getLocalTimeZone()))
					: undefined;
				savedDueDate = $dueDate?.toString();
			})
			.catch((error) => {
				console.error('Failed to fetch todo:', error);
				showError('Failed to fetch todo.');
			});
	});

	function validateTitle(title: string) {
		if (title.length === 0) {
			titleError = 'Title is required';
			return false;
		}
		if (title.length < 2) {
			titleError = 'Title must be at least 2 characters long';
			return false;
		}
		if (title.length > 100) {
			titleError = 'Title must be less than 100 characters long';
			return false;
		}
		titleError = '';
		return true;
	}

	function validateDescription(desc: string) {
		if (desc.length > 1000) {
			descriptionError = 'Description must be less than 1000 characters long';
			return false;
		}
		descriptionError = '';
		return true;
	}

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		if (!validateTitle(newTitle) || !validateDescription(newDescription)) {
			addToast({
				data: {
					title: 'Invalid form',
					description: 'Please check your inputs',
					color: 'bg-ctp-red'
				}
			});
			return;
		}

		api
			.callWithAuth('PATCH', `/ormi/${todoId}`, {
				title: newTitle,
				description: newDescription,
				state: newState,
				labels: $labels.map((label) => label.value),
				...($dueDate?.toString() !== savedDueDate && {
					dueDate: $dueDate ? $dueDate.toDate(getLocalTimeZone()) : null
				})
			})
			.then((response) => {
				todo = response.data;
				savedDueDate = $dueDate?.toString();
				addToast({
					data: {
						title: 'Success',
						description: 'Todo updated successfully.',
						color: 'bg-ctp-green'
					}
				});
				invalidateAll();
			})
			.catch((error) => {
				console.error('Failed to update todo:', error);
				showError(error.response?.data?.error ?? 'Failed to update todo.');
			});
	};
</script>

{#if todo !== null}
	<form use:melt={$description} class="mb-5 mt-2 flex flex-col gap-6" on:submit={handleSubmit}>
		<fieldset>
			<input
				use:melt={$title}
				id="title"
				name="title"
				class="bg-transparent outline-none text-lg font-semibold text-ctp-text"
				bind:value={newTitle}
				on:blur={() => validateTitle(newTitle)}
			/>
			<p class="text-left text-sm font-semibold text-ctp-red">{titleError}</p>
		</fieldset>
		<fieldset>
			<span class="mb-2 text-sm font-semibold">State</span>
			<div class="flex flex-wrap gap-2" role="radiogroup" aria-label="State">
				{#each states as state}
					<button
						type="button"
						role="radio"
						aria-checked={newState === state}
						class="flex items-center gap-1 rounded-md bg-ctp-surface0 px-2 py-1 text-sm font-semibold
							aria-checked:ring-2 aria-checked:ring-ctp-mauve hover:bg-ctp-surface1"
						on:click={() => (newState = state)}
					>
						<svelte:component
							this={todoStates[state].icon}
							size="14"
							class={todoStates[state].color}
						/>
						{todoStates[state].label}
					</button>
				{/each}
			</div>
		</fieldset>
		<fieldset>
			<label for="description" class="mb-2 text-sm font-semibold">Description</label>
			<textarea
				id="description"
				name="description"
				class="flex h-32 w-full items-center justify-between rounded-md bg-ctp-surface0
				shadow-md shadow-ctp-crust px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
				bind:value={newDescription}
				on:blur={() => validateDescription(newDescription)}
			/>
			<p class="text-left text-sm font-semibold text-ctp-red">{descriptionError}</p>
		</fieldset>
		<fieldset>
			<MeltDatePicker value={dueDate} minValue={today(getLocalTimeZone())} label="Due Date" />
		</fieldset>
		<fieldset>
			<span class="mb-2 text-sm font-semibold">Labels</span>
			<MeltTags tags={labels} placeholder="Enter labels..." suggest={suggestLabels} />
		</fieldset>

		<div class="flex flex-row-reverse gap-2">
			<Button type="submit">
				<Save size={16} />
				<span>Save</span>
			</Button>
		</div>
	</form>

	<h3 class="mb-2 text-sm font-semibold">History</h3>
	<ol class="flex flex-col gap-3 border-l-2 border-ctp-surface0 pl-4">
		{#each [...todo.history].reverse() as event}
			<li class="text-sm">
				<p class="flex items-center gap-1 font-semibold">
					<svelte:component
						this={todoStates[event.newState].icon}
						size="14"
						class={todoStates[event.newState].color}
					/>
					{#if event.previousState === ''}
						Created
					{:else}
						{todoStates[event.previousState].label} → {todoStates[event.newState].label}
					{/if}
				</p>
				<time class="text-ctp-subtext0" datetime={String(event.updatedAt)}>
					{new Date(event.updatedAt).toLocaleString(undefined, {
						dateStyle: 'medium',
						timeStyle: 'short'
					})}
				</time>
			</li>
		{/each}
	</ol>
{:else}
	<h2 use:melt={$title} class="mb-0 text-lg font-semibold text-ctp-text">Loading...</h2>
{/if}
