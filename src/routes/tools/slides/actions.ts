import axios from 'axios';
import api from '$lib/api';
import { addToast } from '$lib/toast';
import type { Deck } from './types';

export function errorMessage(error: unknown, fallback: string): string {
	if (axios.isAxiosError(error) && typeof error.response?.data?.error === 'string') {
		return error.response.data.error;
	}
	return fallback;
}

export function notifyError(description: string) {
	addToast({ data: { title: 'Error', description, color: 'bg-ctp-red' } });
}

async function run<T>(action: () => Promise<T>, failure: string): Promise<T | null> {
	try {
		return await action();
	} catch (error) {
		notifyError(errorMessage(error, failure));
		return null;
	}
}

export const createDeck = (title: string) =>
	run(
		async () => (await api.callWithAuth('POST', '/slides', { title })).data as Deck,
		'The deck could not be created'
	);

export const duplicateDeck = (id: string) =>
	run(
		async () => (await api.callWithAuth('POST', `/slides/${id}/duplicate`)).data as Deck,
		'The deck could not be duplicated'
	);

export const deleteDeck = (id: string) =>
	run(async () => {
		await api.callWithAuth('DELETE', `/slides/${id}`);
		return true;
	}, 'The deck could not be deleted');
