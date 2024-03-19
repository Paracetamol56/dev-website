import api from '$lib/api';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import type { WordCloudSession } from '../../utils';

export const load: PageLoad = async ({ params }) => {
	const result = await api.callWithAuth('GET', `/word-cloud/${params.id}`)
		.then((res) => {
			const session: WordCloudSession = res.data as WordCloudSession;
			const distribution: { text: string; occurence: number }[] = [];
			for (const word of res.data.words) {
				const wDistribition = distribution.find((w) => w.text.toLowerCase() === word.text);
				if (wDistribition) {
					wDistribition.occurence++;
				} else {
					distribution.push({ text: word.text.toLowerCase(), occurence: 1 });
				}
				distribution.sort((a, b) => b.occurence - a.occurence);
			}
			return { session, distribution };
		})
		.catch((err) => {
			console.error(err);
			if (err.response.status === 404)
				throw error(404, 'This session does not exist');
			else if (err.response.status === 403)
				throw error(403, 'You are not allowed to access this session');
			else
				throw error(500, 'An error occured while fetching the session');
		});

	return {
		session: result.session,
		distribution: result.distribution
	};
};
