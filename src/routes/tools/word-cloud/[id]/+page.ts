import api from '$lib/api';
import { error } from '@sveltejs/kit';
import type { WordCloudSessionUser } from '../utils';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const session = await api
		.call('GET', `/word-cloud/${params.id}`)
		.then((res) => res.data as WordCloudSessionUser)
		.catch((err) => {
			if (err.response?.status === 404 || err.response?.status === 400) {
				error(404, 'This session does not exist');
			}
			error(500, 'An error occured while fetching the session');
		});

	return { session };
};
