import { describe, expect, it } from 'vitest';
import { buildOutline } from './outline';

describe('buildOutline', () => {
	it('lists top-level slides with the text of their first heading', () => {
		const text =
			'<section><h2>Intro <b>bold</b> &amp; more</h2><p>Body</p></section>\n<section><p>No heading</p></section>';
		expect(buildOutline(text)).toEqual([
			{ title: 'Intro bold & more', from: 0, children: [] },
			{ title: 'No heading', from: text.indexOf('<section><p>'), children: [] }
		]);
	});

	it('nests the vertical slides of a stack, which takes the title of its first one', () => {
		const text =
			'<section>\n\t<section><h3>A</h3></section>\n\t<section><h3>B</h3></section>\n</section>';
		const [stack] = buildOutline(text);
		expect(stack.title).toBe('A');
		expect(stack.children.map((slide) => slide.title)).toEqual(['A', 'B']);
		expect(stack.children[1].from).toBe(text.indexOf('<section><h3>B'));
	});

	it('reads the first markdown heading of a template', () => {
		const script =
			'<section data-markdown><script type="text/template">\n# Script title\n</script></section>';
		const textarea =
			'<section data-markdown><textarea data-template>\nText\n## Area title\n</textarea></section>';
		expect(buildOutline(script)[0].title).toBe('Script title');
		expect(buildOutline(textarea)[0].title).toBe('Area title');
	});

	it('names an empty slide and ignores anything outside sections', () => {
		expect(buildOutline('<p>loose</p><section></section>')).toEqual([
			{ title: 'Untitled', from: 12, children: [] }
		]);
		expect(buildOutline('')).toEqual([]);
	});
});
