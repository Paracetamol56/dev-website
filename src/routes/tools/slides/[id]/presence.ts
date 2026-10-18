import { StateEffect, StateField, type ChangeDesc, type Range } from '@codemirror/state';
import { Decoration, EditorView, WidgetType, type DecorationSet } from '@codemirror/view';

export type PeerRange = { anchor: number; head: number };
// movedAt restarts the name label each time the peer moves
export type Peer = { peer: number; name: string; ranges: PeerRange[]; movedAt: number };

const colours = ['peach', 'green', 'blue', 'pink', 'teal', 'yellow', 'sapphire', 'red', 'lavender'];
export const peerColour = (peer: number) => colours[peer % colours.length];

const mapRanges = (ranges: PeerRange[], mapping: ChangeDesc) =>
	ranges.map(({ anchor, head }) => ({
		anchor: mapping.mapPos(anchor, 1),
		head: mapping.mapPos(head, 1)
	}));

// Positions are in the local document; a mapped effect follows the changes made after it
export const setPeer = StateEffect.define<Peer>({
	map: (value, mapping) => ({ ...value, ranges: mapRanges(value.ranges, mapping) })
});
export const removePeer = StateEffect.define<number>();
export const clearPeers = StateEffect.define<null>();

export const peersField = StateField.define<ReadonlyMap<number, Peer>>({
	create: () => new Map(),
	update(peers, transaction) {
		let next = peers;
		if (transaction.docChanged && peers.size > 0) {
			next = new Map(
				[...peers].map(([id, peer]) => [
					id,
					{ ...peer, ranges: mapRanges(peer.ranges, transaction.changes) }
				])
			);
		}
		for (const effect of transaction.effects) {
			if (effect.is(setPeer)) {
				next = new Map(next).set(effect.value.peer, effect.value);
			} else if (effect.is(removePeer)) {
				next = new Map(next);
				(next as Map<number, Peer>).delete(effect.value);
			} else if (effect.is(clearPeers)) {
				next = new Map();
			}
		}
		return next;
	}
});

class CaretWidget extends WidgetType {
	constructor(readonly name: string, readonly colour: string, readonly movedAt: number) {
		super();
	}

	eq(other: CaretWidget) {
		return (
			other.name === this.name && other.colour === this.colour && other.movedAt === this.movedAt
		);
	}

	toDOM() {
		const caret = document.createElement('span');
		caret.className = 'cm-peer-caret';
		caret.style.setProperty('--peer', `var(--ctp-${this.colour})`);
		const label = document.createElement('span');
		label.className = 'cm-peer-name';
		label.textContent = this.name;
		caret.appendChild(label);
		return caret;
	}

	ignoreEvent() {
		return true;
	}
}

const decorations = EditorView.decorations.compute([peersField], (state): DecorationSet => {
	const length = state.doc.length;
	const marks: Range<Decoration>[] = [];
	for (const { peer, name, ranges, movedAt } of state.field(peersField).values()) {
		const colour = peerColour(peer);
		for (const range of ranges) {
			const anchor = Math.min(range.anchor, length);
			const head = Math.min(range.head, length);
			const from = Math.min(anchor, head);
			const to = Math.max(anchor, head);
			if (from < to) {
				marks.push(
					Decoration.mark({
						class: 'cm-peer-selection',
						attributes: { style: `--peer: var(--ctp-${colour})` }
					}).range(from, to)
				);
			}
			marks.push(
				Decoration.widget({ widget: new CaretWidget(name, colour, movedAt), side: 1 }).range(head)
			);
		}
	}
	return Decoration.set(marks, true);
});

const theme = EditorView.baseTheme({
	'.cm-peer-selection': { backgroundColor: 'rgba(var(--peer), 0.3)' },
	'.cm-peer-caret': {
		position: 'relative',
		borderLeft: '2px solid rgb(var(--peer))',
		marginLeft: '-1px',
		marginRight: '-1px',
		pointerEvents: 'none'
	},
	'.cm-peer-name': {
		position: 'absolute',
		bottom: '100%',
		left: '-2px',
		padding: '1px 4px',
		borderRadius: '3px 3px 3px 0',
		backgroundColor: 'rgb(var(--peer))',
		color: 'rgb(var(--ctp-crust))',
		fontFamily: 'ui-sans-serif, system-ui, sans-serif',
		fontSize: '10px',
		fontWeight: '600',
		lineHeight: '1.2',
		whiteSpace: 'nowrap',
		zIndex: '10',
		animation: 'cm-peer-fade 0.3s ease 2s forwards'
	},
	'@keyframes cm-peer-fade': { to: { opacity: '0' } }
});

export const presence = [peersField, decorations, theme];
