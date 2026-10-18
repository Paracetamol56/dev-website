import type { EditorState } from '@codemirror/state';
import type { CurrentSlide } from '$lib/components/deck/Slides.svelte';
import { sourceElements } from './sourceMap';

export type BlockKind = 'slide' | 'text' | 'heading' | 'list' | 'code';

// focus starts the element to edit once it is on screen
const blocks: Record<BlockKind, { lines: string[]; focus?: string }> = {
	slide: { lines: ['<section>', '\t<h2>New slide</h2>', '</section>'], focus: '<h2>' },
	text: { lines: ['<p>New text</p>'], focus: '<p>' },
	heading: { lines: ['<h2>New heading</h2>'], focus: '<h2>' },
	list: {
		lines: ['<ul>', '\t<li>First item</li>', '\t<li>Second item</li>', '</ul>'],
		focus: '<li>'
	},
	code: {
		lines: ['<pre><code data-trim class="language-js">', "console.log('Hello');", '</code></pre>']
	}
};

export type Insertion = { from: number; insert: string; focus: number | null };

const indentAt = (state: EditorState, position: number) =>
	state.doc.lineAt(position).text.match(/^\s*/)?.[0] ?? '';

function place(at: number, indent: string, lines: string[], focus?: string): Insertion {
	const insert = lines.map((line) => `\n${indent}${line}`).join('');
	const offset = focus ? insert.indexOf(focus) : -1;
	return { from: at, insert, focus: offset < 0 ? null : at + offset };
}

// A slide goes after the current top-level slide, anything else at the end of the slide on screen
export function insertion(state: EditorState, kind: BlockKind, current: CurrentSlide): Insertion {
	const elements = state.field(sourceElements);
	const stack = current.stack === null ? undefined : elements[current.stack];
	const slide = current.slide === null ? undefined : elements[current.slide];
	const block = blocks[kind];

	if (kind !== 'slide' && slide && slide.contentTo > slide.contentFrom) {
		let at = slide.contentTo;
		while (at > slide.contentFrom && /\s/.test(state.sliceDoc(at - 1, at))) at--;
		return place(at, indentAt(state, slide.from) + '\t', block.lines, block.focus);
	}

	const lines =
		kind === 'slide'
			? block.lines
			: ['<section>', ...block.lines.map((line) => `\t${line}`), '</section>'];
	const at = stack ? stack.to : state.doc.length;
	const indent = stack ? indentAt(state, stack.from) : '';
	const placed = place(at, indent, lines, block.focus);
	return at === 0
		? { ...placed, insert: placed.insert.slice(1), focus: placed.focus && placed.focus - 1 }
		: placed;
}
