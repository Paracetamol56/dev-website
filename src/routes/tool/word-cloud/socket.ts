import type { WordCloudWord } from './utils';

export type SessionEvent =
	| { type: 'ready'; owner: boolean }
	| { type: 'word'; word: WordCloudWord }
	| { type: 'session'; open: boolean };

/**
 * Follows a session's live events, reconnecting with backoff. Only the owner's access token
 * unlocks word events. Returns a function that disconnects.
 */
export function followSession(
	id: string,
	onEvent: (event: SessionEvent) => void,
	accessToken: string | null = null
): () => void {
	const url = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${
		location.host
	}/api/word-cloud/${id}/ws`;
	let socket: WebSocket | null = null;
	let retry = 0;
	let timer: ReturnType<typeof setTimeout>;
	let stopped = false;

	const connect = () => {
		socket = new WebSocket(url);
		socket.onopen = () => {
			retry = 0;
			socket?.send(JSON.stringify({ type: 'auth', token: accessToken ?? '' }));
		};
		socket.onmessage = (message) => onEvent(JSON.parse(message.data) as SessionEvent);
		socket.onclose = () => {
			if (stopped) return;
			timer = setTimeout(connect, Math.min(30000, 1000 * 2 ** retry++));
		};
	};
	connect();

	return () => {
		stopped = true;
		clearTimeout(timer);
		socket?.close();
	};
}
