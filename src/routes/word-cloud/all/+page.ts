import axios from 'axios';
import type { PageData } from '../$types';

export const load: PageData = async () => {
	return axios
		.get('/api/word-cloud')
		.then((res) => {
			return { sessions: res.data };
		})
		.catch((err) => {
			console.error(err);
			return { sessions: [] };
	});
};
