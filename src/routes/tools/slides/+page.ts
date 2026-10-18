import { error, redirect } from '@sveltejs/kit';
import { get } from 'svelte/store';
import api from '$lib/api';
import { user } from '$lib/store';
import { perPage, type DeckPage } from './types';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	if (!get(user).accessToken) {
		error(401, 'Log in to manage your slide decks');
	}

	const page = Math.max(1, Math.floor(Number(url.searchParams.get('page'))) || 1);
	const search = url.searchParams.get('search')?.slice(0, 100) ?? '';
	const query = new URLSearchParams({ page: String(page), limit: String(perPage) });
	if (search) query.set('search', search);

	let decks: DeckPage;
	try {
		decks = (await api.callWithAuth('GET', `/slides?${query}`)).data as DeckPage;
	} catch {
		error(500, 'Your slide decks could not be loaded');
	}
	if (decks.items.length === 0 && decks.total > 0) {
		query.set('page', String(Math.ceil(decks.total / perPage)));
		query.delete('limit');
		redirect(307, `/tools/slides?${query}`);
	}
	return { decks, search };
};
