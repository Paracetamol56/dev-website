import { user } from '$lib/store';
import { get } from 'svelte/store';
import api from '$lib/api';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import type { ManagePageData } from '../utils';

export const load: PageLoad = async () => {
	// If the user is not logged in, throw a 401 error
	if (!get(user).accessToken) {
		error(401, 'Unauthorized');
	}
	const result: ManagePageData[] = await api
		.callWithAuth('GET', `/word-cloud?user=${get(user).id}`)
		.then((res) => {
			if (res.data === null) {
				return [];
			}

			return res.data as ManagePageData[];
		})
		.catch((err) => {
			console.error(err);
			return [];
		});

	return { sessions: result };
};
