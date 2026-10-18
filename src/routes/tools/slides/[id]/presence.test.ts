import { EditorState } from '@codemirror/state';
import { describe, expect, it } from 'vitest';
import { clearPeers, peersField, removePeer, setPeer } from './presence';

const peer = { peer: 1, name: 'Ada', ranges: [{ anchor: 2, head: 5 }], movedAt: 0 };

describe('peersField', () => {
	it('keeps the selections of others in place when the document changes', () => {
		let state = EditorState.create({ doc: 'abcdef', extensions: peersField });
		state = state.update({ effects: setPeer.of(peer) }).state;
		state = state.update({ changes: { from: 0, insert: 'xx' } }).state;
		expect(state.field(peersField).get(1)?.ranges).toEqual([{ anchor: 4, head: 7 }]);
	});

	it('maps a selection received alongside a change', () => {
		const state = EditorState.create({ doc: 'abcdef', extensions: peersField });
		const changes = state.changes({ from: 0, insert: 'xx' });
		const next = state.update({ changes, effects: setPeer.of(peer).map(changes) }).state;
		expect(next.field(peersField).get(1)?.ranges).toEqual([{ anchor: 4, head: 7 }]);
	});

	it('removes one peer or all of them', () => {
		let state = EditorState.create({ doc: 'abcdef', extensions: peersField });
		state = state.update({ effects: [setPeer.of(peer), setPeer.of({ ...peer, peer: 2 })] }).state;
		state = state.update({ effects: removePeer.of(1) }).state;
		expect([...state.field(peersField).keys()]).toEqual([2]);
		state = state.update({ effects: clearPeers.of(null) }).state;
		expect(state.field(peersField).size).toBe(0);
	});
});
