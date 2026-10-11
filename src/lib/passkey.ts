import axios from 'axios';
import { browser } from '$app/environment';
import api from './api';

const decode = (value: string) =>
	Uint8Array.from(atob(value.replace(/-/g, '+').replace(/_/g, '/')), (char) => char.charCodeAt(0));

const encode = (buffer: ArrayBuffer | null) =>
	buffer
		? btoa(String.fromCharCode(...new Uint8Array(buffer)))
				.replace(/\+/g, '-')
				.replace(/\//g, '_')
				.replace(/=+$/, '')
		: undefined;

const decodeIds = (credentials: { id: string }[] | undefined) =>
	(credentials ?? []).map((credential) => ({ ...credential, id: decode(credential.id) }));

export const passkeysSupported = () => browser && !!window.PublicKeyCredential;

export const isCancellation = (error: unknown) =>
	error instanceof DOMException && ['NotAllowedError', 'AbortError'].includes(error.name);

export const loginWithPasskey = async () => {
	const { data } = await axios.post('/api/auth/passkey/begin');
	const options = data.options.publicKey;

	const credential = (await navigator.credentials.get({
		publicKey: {
			...options,
			challenge: decode(options.challenge),
			allowCredentials: decodeIds(options.allowCredentials)
		}
	})) as PublicKeyCredential;
	const assertion = credential.response as AuthenticatorAssertionResponse;

	const response = await axios.post('/api/auth/passkey/finish', {
		sessionId: data.sessionId,
		credential: {
			id: credential.id,
			rawId: encode(credential.rawId),
			type: credential.type,
			response: {
				clientDataJSON: encode(assertion.clientDataJSON),
				authenticatorData: encode(assertion.authenticatorData),
				signature: encode(assertion.signature),
				userHandle: encode(assertion.userHandle)
			},
			clientExtensionResults: credential.getClientExtensionResults()
		}
	});
	api.persistUser(response);
};

export const addPasskey = async (userId: string, name: string) => {
	const { data } = await api.callWithAuth('POST', `/users/${userId}/passkeys/begin`);
	const options = data.options.publicKey;

	const credential = (await navigator.credentials.create({
		publicKey: {
			...options,
			challenge: decode(options.challenge),
			user: { ...options.user, id: decode(options.user.id) },
			excludeCredentials: decodeIds(options.excludeCredentials)
		}
	})) as PublicKeyCredential;
	const attestation = credential.response as AuthenticatorAttestationResponse;

	const response = await api.callWithAuth('POST', `/users/${userId}/passkeys/finish`, {
		sessionId: data.sessionId,
		name,
		credential: {
			id: credential.id,
			rawId: encode(credential.rawId),
			type: credential.type,
			response: {
				clientDataJSON: encode(attestation.clientDataJSON),
				attestationObject: encode(attestation.attestationObject),
				transports: attestation.getTransports?.() ?? []
			},
			clientExtensionResults: credential.getClientExtensionResults()
		}
	});
	return response.data;
};
