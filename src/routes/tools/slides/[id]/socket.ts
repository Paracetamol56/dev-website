import { writable } from 'svelte/store';
import api from '$lib/api';
import type { DeckSettings } from '$lib/components/deck/settings';

export type WireUpdate = { clientID: string; changes: unknown };
export type WireRange = { anchor: number; head: number };

export type ServerMessage =
	| { type: 'state'; title: string; settings: DeckSettings; peer: number }
	| { type: 'updates'; version: number; updates: WireUpdate[] }
	| { type: 'reset'; version: number; doc: string }
	| { type: 'pushed'; ok: boolean }
	| { type: 'title'; title: string }
	| { type: 'settings'; settings: DeckSettings }
	| { type: 'presence'; peer: number; name: string; version: number; ranges: WireRange[] }
	| { type: 'joined'; peer: number; name: string }
	| { type: 'left'; peer: number }
	| { type: 'deleted' }
	| { type: 'error'; error: string };

export type ClientMessage =
	| { type: 'push'; version: number; updates: WireUpdate[] }
	| { type: 'pull'; version: number }
	| { type: 'title'; title: string }
	| { type: 'settings'; settings: DeckSettings }
	| { type: 'presence'; version: number; ranges: WireRange[] };

type Handlers = {
	[K in ServerMessage['type']]?: (message: Extract<ServerMessage, { type: K }>) => void;
};
type Listener = (message: ServerMessage) => void;

export type Connection = 'connecting' | 'open' | 'closed' | 'refused';

// The server closes with these codes when retrying cannot help
const final = new Set([4401, 4404, 4410]);

export class DeckSocket {
	readonly connection = writable<Connection>('connecting');
	private socket: WebSocket | null = null;
	private listeners = new Set<Handlers>();
	private retry = 0;
	private timer: ReturnType<typeof setTimeout> | undefined;
	private stopped = false;
	private ready = false;

	constructor(private id: string, private version: () => number) {}

	on(handlers: Handlers): () => void {
		this.listeners.add(handlers);
		return () => this.listeners.delete(handlers);
	}

	connect() {
		this.connection.set('connecting');
		this.ready = false;
		const url = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/api/slides/${
			this.id
		}/ws`;
		const socket = new WebSocket(url);
		this.socket = socket;
		socket.onopen = async () => {
			const token = await api.accessToken().catch(() => null);
			if (socket.readyState !== WebSocket.OPEN) return;
			socket.send(JSON.stringify({ type: 'auth', token, version: this.version() }));
		};
		socket.onmessage = (event) => {
			const message = JSON.parse(event.data) as ServerMessage;
			if (message.type === 'state') {
				this.retry = 0;
				this.ready = true;
				this.connection.set('open');
			}
			for (const handlers of this.listeners) {
				(handlers[message.type] as Listener | undefined)?.(message);
			}
		};
		socket.onclose = (event) => {
			if (this.socket !== socket) return;
			this.ready = false;
			if (this.stopped || final.has(event.code)) {
				this.connection.set(final.has(event.code) ? 'refused' : 'closed');
				return;
			}
			this.connection.set('closed');
			this.timer = setTimeout(() => this.connect(), Math.min(15000, 500 * 2 ** this.retry++));
		};
	}

	send(message: ClientMessage): boolean {
		if (!this.ready || this.socket?.readyState !== WebSocket.OPEN) return false;
		this.socket.send(JSON.stringify(message));
		return true;
	}

	close() {
		this.stopped = true;
		clearTimeout(this.timer);
		this.socket?.close();
	}
}
