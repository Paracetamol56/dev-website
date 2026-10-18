import { collab, getSyncedVersion, receiveUpdates, sendableUpdates } from '@codemirror/collab';
import { ChangeSet, type Extension } from '@codemirror/state';
import { ViewPlugin, type EditorView, type ViewUpdate } from '@codemirror/view';
import { clearPeers, peersField, removePeer, setPeer, type Peer } from './presence';
import type { DeckSocket, WireRange, WireUpdate } from './socket';

const keptChanges = 500;
const presenceDelay = 50;

// Keeps the editor in step with the deck's hub: local changes are pushed from the last
// version the server confirmed, and updates from others are rebased under unconfirmed ones.
// Selections travel in positions of the last confirmed version, and are mapped on each side.
export function collaboration(
	socket: DeckSocket,
	startVersion: number,
	onPending: (pending: boolean) => void,
	onPeers: (peers: Peer[]) => void
): Extension {
	const plugin = ViewPlugin.fromClass(
		class {
			pushing = false;
			confirmed: ChangeSet[] = [];
			confirmedFrom = startVersion;
			presenceTimer: ReturnType<typeof setTimeout> | undefined;
			stop: () => void;

			constructor(readonly view: EditorView) {
				this.stop = socket.on({
					state: () => {
						this.pushing = false;
						this.view.dispatch({ effects: clearPeers.of(null) });
						this.sendPresence();
					},
					updates: ({ version, updates }) => this.receive(version, updates),
					pushed: ({ ok }) => {
						this.pushing = false;
						if (ok) this.push();
						else socket.send({ type: 'pull', version: getSyncedVersion(this.view.state) });
					},
					presence: ({ peer, name, version, ranges }) => {
						const local = this.toLocal(version, ranges);
						if (local)
							this.view.dispatch({
								effects: setPeer.of({ peer, name, ranges: local, movedAt: Date.now() })
							});
					},
					joined: () => this.sendPresence(),
					left: ({ peer }) => this.view.dispatch({ effects: removePeer.of(peer) })
				});
				onPending(false);
			}

			update(update: ViewUpdate) {
				if (update.docChanged) this.push();
				if (update.selectionSet || update.docChanged) this.schedulePresence();
				onPending(sendableUpdates(update.state).length > 0);
				if (update.startState.field(peersField) !== update.state.field(peersField)) {
					onPeers([...update.state.field(peersField).values()]);
				}
			}

			receive(version: number, updates: WireUpdate[]) {
				const synced = getSyncedVersion(this.view.state);
				if (version > synced) {
					socket.send({ type: 'pull', version: synced });
					return;
				}
				const missing = updates.slice(synced - version).map((update) => ({
					clientID: update.clientID,
					changes: ChangeSet.fromJSON(update.changes)
				}));
				if (missing.length) {
					this.remember(
						synced,
						missing.map((update) => update.changes)
					);
					this.view.dispatch(receiveUpdates(this.view.state, missing));
				}
				this.push();
			}

			remember(version: number, changes: ChangeSet[]) {
				if (version !== this.confirmedFrom + this.confirmed.length) {
					this.confirmed = [];
					this.confirmedFrom = version;
				}
				this.confirmed.push(...changes);
				const excess = this.confirmed.length - keptChanges;
				if (excess > 0) {
					this.confirmed.splice(0, excess);
					this.confirmedFrom += excess;
				}
			}

			// Maps positions of a confirmed version to this editor's document, or gives up
			toLocal(version: number, ranges: WireRange[]) {
				const state = this.view.state;
				const synced = getSyncedVersion(state);
				if (version > synced || version < this.confirmedFrom) return null;
				const later = [
					...this.confirmed.slice(version - this.confirmedFrom, synced - this.confirmedFrom),
					...sendableUpdates(state).map((update) => update.changes)
				];
				const start = later.length ? later[0].length : state.doc.length;
				const map = (position: number) =>
					later.reduce((mapped, changes) => changes.mapPos(mapped, 1), Math.min(position, start));
				return ranges.map(({ anchor, head }) => ({ anchor: map(anchor), head: map(head) }));
			}

			schedulePresence() {
				clearTimeout(this.presenceTimer);
				this.presenceTimer = setTimeout(() => this.sendPresence(), presenceDelay);
			}

			// Sends the selection in positions of the last confirmed version
			sendPresence() {
				const state = this.view.state;
				const unconfirmed = sendableUpdates(state)
					.map((update) => update.changes.invertedDesc)
					.reverse();
				const back = (position: number) =>
					unconfirmed.reduce((mapped, changes) => changes.mapPos(mapped, 1), position);
				socket.send({
					type: 'presence',
					version: getSyncedVersion(state),
					ranges: state.selection.ranges.map((range) => ({
						anchor: back(range.anchor),
						head: back(range.head)
					}))
				});
			}

			push() {
				const updates = sendableUpdates(this.view.state);
				if (this.pushing || !updates.length) return;
				this.pushing = socket.send({
					type: 'push',
					version: getSyncedVersion(this.view.state),
					updates: updates.map((update) => ({
						clientID: update.clientID,
						changes: update.changes.toJSON()
					}))
				});
			}

			destroy() {
				clearTimeout(this.presenceTimer);
				this.stop();
			}
		}
	);
	return [collab({ startVersion }), plugin];
}
