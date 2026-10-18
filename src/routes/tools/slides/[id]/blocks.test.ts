import { EditorState } from '@codemirror/state';
import { describe, expect, it } from 'vitest';
import { insertion } from './blocks';
import { annotate, setSourceElements, sourceElements } from './sourceMap';

function stateOf(text: string) {
	const state = EditorState.create({ doc: text, extensions: sourceElements });
	return state.update({ effects: setSourceElements.of(annotate(text).elements) }).state;
}

function apply(
	text: string,
	kind: Parameters<typeof insertion>[1],
	slide: number | null,
	stack = slide
) {
	const state = stateOf(text);
	const { from, insert, focus } = insertion(state, kind, { slide, stack });
	const result = state.update({ changes: { from, insert } }).state.doc.toString();
	return { result, focused: focus === null ? null : result.slice(focus, focus + 4) };
}

const deck = '<section>\n\t<h2>One</h2>\n</section>\n<section>\n\t<h2>Two</h2>\n</section>';

describe('insertion', () => {
	it('adds a text at the end of the current slide, indented like its content', () => {
		const { result, focused } = apply(deck, 'text', 0);
		expect(result).toBe(
			'<section>\n\t<h2>One</h2>\n\t<p>New text</p>\n</section>\n<section>\n\t<h2>Two</h2>\n</section>'
		);
		expect(focused).toBe('<p>N');
	});

	it('adds a slide after the current top-level slide', () => {
		const { result, focused } = apply(deck, 'slide', 0);
		expect(result).toBe(
			'<section>\n\t<h2>One</h2>\n</section>\n<section>\n\t<h2>New slide</h2>\n</section>\n<section>\n\t<h2>Two</h2>\n</section>'
		);
		expect(focused).toBe('<h2>');
	});

	it('adds a slide after the whole stack when a vertical slide is on screen', () => {
		const text = '<section>\n\t<section>A</section>\n\t<section>B</section>\n</section>';
		const { result } = apply(text, 'slide', 2, 0);
		expect(result.endsWith('</section>\n<section>\n\t<h2>New slide</h2>\n</section>')).toBe(true);
	});

	it('focuses the first item of a new list and nothing in a code block', () => {
		expect(apply(deck, 'list', 3).focused).toBe('<li>');
		expect(apply(deck, 'code', 3).focused).toBeNull();
	});

	it('wraps a block in a new slide when no slide is on screen', () => {
		expect(apply('', 'heading', null).result).toBe('<section>\n\t<h2>New heading</h2>\n</section>');
		expect(apply('', 'slide', null).focused).toBe('<h2>');
	});
});
