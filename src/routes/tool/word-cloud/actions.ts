import api from '$lib/api';
import { addToast } from '../../+layout.svelte';
import type { WordCloudSessionUser } from './utils';

type SessionPatch = { name?: string; description?: string; open?: boolean };

function notify(ok: boolean, description: string) {
	addToast({
		data: {
			title: ok ? 'Success' : 'Error',
			description,
			color: ok ? 'bg-ctp-green' : 'bg-ctp-red'
		}
	});
}

async function run<T>(
	action: () => Promise<T>,
	success: string,
	failure: string
): Promise<T | null> {
	try {
		const result = await action();
		notify(true, success);
		return result;
	} catch {
		notify(false, failure);
		return null;
	}
}

export const updateSession = (id: string, patch: SessionPatch) =>
	run(
		async () =>
			(await api.callWithAuth('PATCH', `/word-cloud/${id}`, patch)).data as WordCloudSessionUser,
		patch.open === false
			? 'The session is closed'
			: patch.open
			? 'The session is open again'
			: 'The session was updated',
		'The session could not be updated'
	);

export const duplicateSession = (id: string) =>
	run(
		async () => (await api.callWithAuth('POST', `/word-cloud/${id}/duplicate`)).data.id as string,
		'The session was duplicated',
		'The session could not be duplicated'
	);

export const deleteSession = (id: string) =>
	run(
		async () => {
			await api.callWithAuth('DELETE', `/word-cloud/${id}`);
			return true;
		},
		'The session was deleted',
		'The session could not be deleted'
	);
