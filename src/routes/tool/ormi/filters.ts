import { browser } from '$app/environment';
import type { DateValue } from '@internationalized/date';
import type { SelectOption, Tag } from '@melt-ui/svelte';
import { writable, type Writable } from 'svelte/store';
import { openStates, type TodoState } from './states';

export const sortOptions: SelectOption<string>[] = [
	{ value: 'position', label: 'My order' },
	{ value: 'due', label: 'Due date' },
	{ value: 'created', label: 'Creation date' },
	{ value: 'updated', label: 'Last change' }
];

export const states: Writable<TodoState[]> = writable([...openStates]);
export const labels: Writable<Tag[]> = writable([]);
export const from: Writable<DateValue | undefined> = writable(undefined);
export const to: Writable<DateValue | undefined> = writable(undefined);
export const sort = writable(
	sortOptions.find((option) => browser && option.value === localStorage.getItem('ormi-sort')) ??
		sortOptions[0]
);

type Filters = {
	states?: TodoState[];
	from?: DateValue;
	to?: DateValue;
	sort?: string;
};

export const setFilters = (filters: Filters = {}) => {
	states.set(filters.states ?? [...openStates]);
	labels.set([]);
	from.set(filters.from);
	to.set(filters.to);
	const option = sortOptions.find((option) => option.value === filters.sort);
	if (option) sort.set(option);
};

export const showTodos = (filters: Filters) => {
	setFilters(filters);
	document.getElementById('todos')?.scrollIntoView({ behavior: 'smooth' });
};
