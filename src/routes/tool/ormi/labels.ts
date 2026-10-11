import api from '$lib/api';

export const suggestLabels = (query: string): Promise<string[]> =>
	api
		.callWithAuth('GET', `/ormi/labels?q=${encodeURIComponent(query)}`)
		.then((response) => response.data);
