import api from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {

	const result = await api.callWithAuth('GET', `/word-cloud/${params.id}`)
		.then((res) => {
			return res.data;
		})
		.catch((err) => {
			console.error(err);
			return [];
		});

	return { sessions: result };
};
