import api from '$lib/api';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import type { WordCloudSessionAdmin } from '../../utils';

export const load: PageLoad = async ({ params }) => {
	const session = await api
		.callWithAuth('GET', `/word-cloud/${params.id}`)
		.then((res) => res.data as WordCloudSessionAdmin)
		.catch((err) => {
			if (err.response?.status === 404 || err.response?.status === 400) {
				error(404, 'This session does not exist');
			}
			error(500, 'An error occured while fetching the session');
		});
	if (!Array.isArray(session.words)) error(403, 'You are not allowed to access this session');

	return { session };
};
