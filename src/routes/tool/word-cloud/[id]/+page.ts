import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import api from '$lib/api';
import { error } from '@sveltejs/kit';
import { addToast } from '../../../+layout.svelte';
import type { WordCloudSessionUser } from '../utils';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	let session: WordCloudSessionUser | null;
	session = await api
		.call('GET', `/word-cloud/${params.id}`)
		.then((res) => {
			if (res.data.open === false) {
				error(400, 'This session is closed');
			}
			return res.data as WordCloudSessionUser;
		})
		.catch((err) => {
			if (err.response.status === 404) {
				error(404, 'This session does not exist');
			} else {
				error(500, 'An error occured while fetching the session');
			}
			return null;
		});

	console.log(session);

	return {
		session
	};
};
