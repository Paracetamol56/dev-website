import api from '$lib/api';
import { error } from '@sveltejs/kit';
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
	state: string | null;
	gitURL: string | null;
	gitIssue: string | null;
	createdAt: Date;
};

export const load: PageLoad = async () => {
	let todos: Todo[];

	try {
		todos = await api
			.callWithAuth('GET', '/ormi?state=TODO')
			.then((res) => {
				return res.data as Todo[];
			});
	} catch (err: any) {
		console.error(err);
		error(500, 'Failed to fetch items');
		return;
	}

	return { todos };
};
