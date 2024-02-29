import { user } from '$lib/store';
import { get } from 'svelte/store';
import api from '$lib/api';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	// If the user is not logged in, throw a 401 error
	if (!get(user).accessToken) {
		error(401, 'Unauthorized');
	}
	const result = await api.callWithAuth('GET', `/word-cloud?user=${get(user).id}`)
		.then((res) => {
			return res.data;
		})
		.catch((err) => {
			console.error(err);
			return [];
		});
	
	return { sessions: result};
};
