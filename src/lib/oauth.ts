import axios from 'axios';
import { get } from 'svelte/store';
import { goto } from '$app/navigation';
import api from './api';
import { user } from './store';
import { isExpired } from './token';
import { addToast } from '../routes/+layout.svelte';

export type OAuthProvider = {
	name: string;
	clientId: string;
	authorizeUrl: string;
	scope: string;
};

export const providerLabels: Record<string, string> = {
	email: 'Email',
	github: 'GitHub',
	google: 'Google'
};

const STORAGE_KEY = 'oauth';

let providers: Promise<OAuthProvider[]> | null = null;

export const getProviders = (): Promise<OAuthProvider[]> =>
	(providers ??= axios
		.get('/api/auth/providers')
		.then((response) => response.data)
		.catch((error) => {
			console.error(error);
			providers = null;
			return [];
		}));

const redirectUri = (name: string) => `${window.location.origin}/verify/${name}`;

export const startOAuth = (provider: OAuthProvider, returnPath: string) => {
	const state = crypto.randomUUID();
	sessionStorage.setItem(STORAGE_KEY, JSON.stringify({ state, path: returnPath }));

	const params = new URLSearchParams({
		client_id: provider.clientId,
		redirect_uri: redirectUri(provider.name),
		response_type: 'code',
		scope: provider.scope,
		state
	});
	window.location.href = `${provider.authorizeUrl}?${params}`;
};

export const completeOAuth = async (name: string, params: URLSearchParams) => {
	const saved = JSON.parse(sessionStorage.getItem(STORAGE_KEY) ?? 'null');
	sessionStorage.removeItem(STORAGE_KEY);
	const path: string = saved?.path ?? '/';
	const code = params.get('code');

	const fail = (description: string) =>
		addToast({
			data: {
				title: `${providerLabels[name] ?? name} login failed`,
				description,
				color: 'bg-ctp-red'
			}
		});

	if (!code || !saved || saved.state !== params.get('state')) {
		fail('The login request is invalid or has expired, please try again');
		goto(path);
		return;
	}

	// Sent when already logged in, so the API links the identity instead of switching account
	const accessToken = get(user).accessToken;
	const headers =
		accessToken && !isExpired(accessToken) ? { Authorization: `Bearer ${accessToken}` } : {};

	await axios
		.post(`/api/auth/${name}`, { code, redirectUri: redirectUri(name) }, { headers })
		.then((response) => api.persistUser(response))
		.catch((error) => {
			console.error(error);
			fail(error.response?.data?.error ?? 'An error occured, please try again later');
		});
	goto(path);
};
