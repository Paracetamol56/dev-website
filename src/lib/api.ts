import axios from 'axios';
import { user } from './store';
import { get } from 'svelte/store';
import { addToast } from '../routes/+layout.svelte';
import { goto } from '$app/navigation';
import { browser } from '$app/environment';
import { isExpired } from '$lib/token';

async function login(email: string) {
	await axios
		.post(`/api/auth/login`, { email })
		.then((response) => {
			if (response.status != 200) {
				console.error(response);
				addToast({
					data: {
						title: 'Error',
						description: 'An error occured, please try again later',
						color: 'bg-ctp-red'
					}
				});
				return;
			}
			addToast({
				data: {
					title: 'Magic link sent',
					description: 'Please check your inbox',
					color: 'bg-ctp-green'
				}
			});
		})
		.catch((error) => {
			console.error(error);
			addToast({
				data: {
					title: 'Error',
					description: 'An error occured, please try again later',
					color: 'bg-ctp-red'
				}
			});
		});
}

async function verify(token: string, redirect: string) {
	await axios
		.post(`/api/auth/verify`, { token })
		.then((reponse) => {
			if (reponse.status !== 200) {
				console.error(reponse);
				addToast({
					data: {
						title: 'Invalid token',
						description: 'The token you provided is invalid or has expired, please try again',
						color: 'bg-ctp-red'
					}
				});
				goto(redirect);
				return;
			}
			// Write the user to the store
			user.set({
				accessToken: reponse.data.accessToken,
				refreshToken: reponse.data.refreshToken,
				id: reponse.data.user.id,
				email: reponse.data.user.email,
				flavour: reponse.data.user.flavour
			});

			addToast({
				data: {
					title: 'Successfully logged in',
					description: 'You are now logged in',
					color: 'bg-ctp-green'
				}
			});
			goto(redirect);
		})
		.catch(() => {
			addToast({
				data: {
					title: 'Failed to verify token',
					description: 'An error occured while verifying your token',
					color: 'bg-ctp-red'
				}
			});
			goto(redirect);
		});
}

function logout() {
	// Revoke tokens
	user.set({
		...get(user),
		id: null,
		accessToken: null,
		refreshToken: null
	});
	if (browser) {
		window.location.reload();
	}
}

async function refreshAccessToken() {
	const refreshToken = get(user).refreshToken;
	if (!refreshToken) {
		addToast({
			data: {
				title: 'Outdated token',
				description: 'Your token is outdated, please log in again',
				color: 'bg-ctp-red'
			}
		});
		throw new Error('No refresh token');
	}
	await axios
		.post(`/api/auth/refresh`, { token: refreshToken })
		.then((res) => {
			if (res.status === 200) {
				// Update the user store
				user.set({
					...get(user),
					accessToken: res.data.accessToken,
					refreshToken: res.data.refreshToken
				});
			}
		})
		.catch((err) => {
			console.error(err);
			addToast({
				data: {
					title: 'Failed to refresh token',
					description: 'An error occured while refreshing your access token',
					color: 'bg-ctp-red'
				}
			});
			throw new Error('Failed to refresh token');
		});
}

async function callWithAuth(method: string, path: string, json?: any) {
	let accessToken = get(user).accessToken;
	if (!accessToken) {
		await refreshAccessToken();
		accessToken = get(user).accessToken;
	} else {
		isExpired(accessToken) && (await refreshAccessToken());
		accessToken = get(user).accessToken;
	}
	return axios({
		method,
		url: `/api${path}`,
		headers: {
			Authorization: `Bearer ${accessToken}`
		},
		data: json
	});
}

async function call(method: string, path: string, json?: any) {
	return axios({
		method,
		url: `/api${path}`,
		data: json
	});
}

export default { login, verify, logout, callWithAuth, call };
