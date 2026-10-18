import axios from 'axios';
import api from '$lib/api';
import { addToast } from '$lib/toast';
import type { Deck, LinkRole, Links } from './types';

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

export const joinUrl = (token: string) => `${location.origin}/tools/slides/join/${token}`;

export const createLink = (id: string, role: LinkRole) =>
	run(
		async () => (await api.callWithAuth('PUT', `/slides/${id}/links/${role}`)).data as Links,
		'The link could not be created'
	);

export const revokeLink = (id: string, role: LinkRole) =>
	run(async () => {
		await api.callWithAuth('DELETE', `/slides/${id}/links/${role}`);
		return true;
	}, 'The link could not be turned off');

export const emailLink = (id: string, role: LinkRole, email: string) =>
	run(async () => {
		await api.callWithAuth('POST', `/slides/${id}/links/${role}/email`, { email });
		return true;
	}, 'The email could not be sent');

export const setMemberRole = (id: string, userId: string, role: LinkRole) =>
	run(async () => {
		await api.callWithAuth('PATCH', `/slides/${id}/members/${userId}`, { role });
		return true;
	}, 'The role could not be changed');

export const removeMember = (id: string, userId: string) =>
	run(async () => {
		await api.callWithAuth('DELETE', `/slides/${id}/members/${userId}`);
		return true;
	}, 'The member could not be removed');
