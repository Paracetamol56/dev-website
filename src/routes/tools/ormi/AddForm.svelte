<script lang="ts">
	import { Plus } from 'lucide-svelte';
	import api from '$lib/api';
	import { addToast } from '../../+layout.svelte';
	import { invalidateAll } from '$app/navigation';
	import { getLocalTimeZone, today, type DateValue } from '@internationalized/date';
	import { writable, type Writable } from 'svelte/store';
	import MeltDatePicker from '$lib/components/MeltDatePicker.svelte';

	let newTodoName = '';
	const dueDate: Writable<DateValue | undefined> = writable(undefined);

	const showError = (description: string) => {
		addToast({
			data: {
				title: 'Error',
				description,
				color: 'bg-ctp-red'
			}
		});
	};

	const addTodo = () => {
		if (newTodoName === '') return;
		api
			.callWithAuth('POST', '/ormi', {
				title: newTodoName,
				dueDate: $dueDate ? $dueDate.toDate(getLocalTimeZone()) : undefined
			})
			.then(() => invalidateAll())
			.catch((error) => {
				console.error('Failed to add todo:', error);
				showError(error.response?.data?.error ?? 'Failed to add todo.');
			});
		newTodoName = '';
		$dueDate = undefined;
	};
</script>

<form
	class="mb-6 rounded-md bg-ctp-mantle shadow-md shadow-ctp-crust min-h-[40px] px-3 py-2 flex gap-2"
	on:submit|preventDefault={addTodo}
>
	<input
		type="text"
		class="w-full bg-transparent outline-none"
		placeholder="Add a new todo..."
		bind:value={newTodoName}
	/>
	<MeltDatePicker value={dueDate} minValue={today(getLocalTimeZone())} compact />
	<button
		class="transition-opacity hover:opacity-80 active:opacity-60 text-ctp-mauve"
		type="submit"
	>
		<Plus size="18" />
	</button>
</form>
