import api from '$lib/api';
import { user } from '$lib/store';
import { error } from '@sveltejs/kit';
import { get } from 'svelte/store';
import type { PageLoad } from './$types';
import type { TodoState } from './states';

export type Todo = {
	id: string;
	user: string;
	title: string;
	description: string | null;
	dueDate: Date | null;
	labels: string[];
	history: {
		previousState: TodoState | '';
		newState: TodoState;
		updatedAt: Date;
	}[];
	state: TodoState;
	position: number;
	gitURL: string | null;
	gitIssue: number | null;
	createdAt: Date;
};

export type DayCount = {
	date: string;
	count: number;
};

export type TodoStats = {
	days: DayCount[];
	firstYear: number;
	currentStreak: number;
	longestStreak: number;
	todo: number;
	inProgress: number;
	standby: number;
	overdue: number;
};

export const load: PageLoad = async () => {
	if (!get(user).id) {
		return { loggedIn: false, stats: null };
	}

	try {
		const timezone = encodeURIComponent(Intl.DateTimeFormat().resolvedOptions().timeZone);
		const stats = await api.callWithAuth('GET', `/ormi/stats?tz=${timezone}`);
		return { loggedIn: true, stats: stats.data as TodoStats };
	} catch (err: any) {
		console.error(err);
		error(500, 'Failed to fetch items');
	}
};
