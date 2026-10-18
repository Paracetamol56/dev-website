import { EditorState } from '@codemirror/state';
import { describe, expect, it } from 'vitest';
import {
	annotate,
	elementById,
	innermostAt,
	setSourceElements,
	sourceElements,
	withStyle
} from './sourceMap';

describe('annotate', () => {
	it('gives every element an id pointing to its source range', () => {
		const text = '<section class="a"><h2>Hi <b>x</b></h2><img src="a.png"></section>';
		const { html, elements } = annotate(text);
		expect(html).toBe(
			'<section data-cm="0" class="a"><h2 data-cm="1">Hi <b data-cm="2">x</b></h2><img data-cm="3" src="a.png"></section>'
		);
		const heading = elements[1];
		expect(text.slice(heading.from, heading.to)).toBe('<h2>Hi <b>x</b></h2>');
		expect(text.slice(heading.contentFrom, heading.contentTo)).toBe('Hi <b>x</b>');
		const image = elements[3];
		expect(image.contentFrom).toBe(image.to);
	});

	it('ends the content of an element without closing tag at its end', () => {
		const text = '<ul><li>a<li>b</ul>';
		const { elements } = annotate(text);
		expect(text.slice(elements[1].contentFrom, elements[1].contentTo)).toBe('a');
	});

	it('leaves the html of templates alone', () => {
		const text = '<section data-markdown><script type="text/template"><b>md</b></script></section>';
		expect(annotate(text).html).toContain(
			'<script data-cm="1" type="text/template"><b>md</b></script>'
		);
	});
});

describe('sourceElements', () => {
	it('follows the changes made after the preview was rendered', () => {
		const text = '<p>one</p><p>two</p>';
		let state = EditorState.create({ doc: text, extensions: sourceElements });
		state = state.update({ effects: setSourceElements.of(annotate(text).elements) }).state;
		state = state.update({ changes: { from: 0, insert: '<h1>new</h1>' } }).state;
		const second = elementById(state, 1)!;
		expect(state.sliceDoc(second.from, second.to)).toBe('<p>two</p>');
		state = state.update({ changes: { from: second.contentFrom, insert: '>' } }).state;
		const edited = elementById(state, 1)!;
		expect(state.sliceDoc(edited.contentFrom, edited.contentTo)).toBe('>two');
	});
});

describe('innermostAt', () => {
	it('picks the smallest element around a position', () => {
		const text = '<section><p>a <b>b</b></p></section>';
		const { elements } = annotate(text);
		expect(innermostAt(elements, text.indexOf('b</b>'))).toBe(2);
		expect(innermostAt(elements, text.indexOf('a '))).toBe(1);
		expect(innermostAt(elements, text.length + 1)).toBeNull();
	});
});

describe('withStyle', () => {
	it('adds a style attribute', () => {
		expect(withStyle('<p class="a">', 'text-align', 'left')).toBe(
			'<p class="a" style="text-align: left">'
		);
		expect(withStyle('<img src="a"/>', 'width', '50%')).toBe('<img src="a" style="width: 50%"/>');
	});

	it('replaces the property and keeps the other ones', () => {
		expect(withStyle(`<p style='color: red; Text-Align : center'>`, 'text-align', 'right')).toBe(
			`<p style='color: red; text-align: right'>`
		);
	});
});
