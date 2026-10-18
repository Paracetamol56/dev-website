import axios from 'axios';
import { error } from '@sveltejs/kit';
import { get } from 'svelte/store';
import api from '$lib/api';
import { user } from '$lib/store';
import type { Deck } from '../types';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	if (!get(user).accessToken) {
		error(401, 'Log in to edit your slide decks');
	}

	let deck: Deck;
	try {
		deck = (await api.callWithAuth('GET', `/slides/${params.id}`)).data as Deck;
	} catch (err) {
		const status = axios.isAxiosError(err) ? err.response?.status : undefined;
		if (status === 400 || status === 404) error(404, 'This slide deck does not exist');
		error(500, 'The slide deck could not be loaded');
	}
	return { deck };
};
