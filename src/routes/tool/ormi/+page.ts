import api from '$lib/api';
import { user } from '$lib/store';
import { error } from '@sveltejs/kit';
import { get } from 'svelte/store';
import type { PageLoad } from './$types';

export type Todo = {
	id: string;
	user: string;
	title: string;
	description: string | null;
	dueDate: Date | null;
	labels: string[];
	history: {
		previousState: string;
		newState: string;
		updatedAt: Date;
	}[];
	state: string;
	position: number;
	gitURL: string | null;
	gitIssue: number | null;
	createdAt: Date;
};

export type DayCount = {
	date: string;
	count: number;
};

export const load: PageLoad = async () => {
	if (!get(user).id) {
		return { loggedIn: false, todos: [] as Todo[], stats: [] as DayCount[] };
	}

	try {
		const timezone = encodeURIComponent(Intl.DateTimeFormat().resolvedOptions().timeZone);
		const todos = await api.callWithAuth('GET', '/ormi?state=TODO');
		const stats = await api.callWithAuth('GET', `/ormi/stats?tz=${timezone}`);
		return { loggedIn: true, todos: todos.data as Todo[], stats: stats.data as DayCount[] };
	} catch (err: any) {
		console.error(err);
		error(500, 'Failed to fetch items');
	}
};
