import { htmlLanguage } from '@codemirror/lang-html';
import type { SyntaxNode } from '@lezer/common';

export type OutlineSlide = { title: string; from: number; children: OutlineSlide[] };

const tagName = (element: SyntaxNode, text: string) => {
	const name = (element.getChild('OpenTag') ?? element.getChild('SelfClosingTag'))?.getChild(
		'TagName'
	);
	return name ? text.slice(name.from, name.to).toLowerCase() : null;
};

const sections = (parent: SyntaxNode, text: string) =>
	parent.getChildren('Element').filter((element) => tagName(element, text) === 'section');

const plain = (html: string) =>
	html
		.replace(/<[^>]*>/g, ' ')
		.replace(/&nbsp;/g, ' ')
		.replace(/&amp;/g, '&')
		.replace(/&lt;/g, '<')
		.replace(/&gt;/g, '>')
		.replace(/\s+/g, ' ')
		.trim();

function find(node: SyntaxNode, match: (node: SyntaxNode) => boolean): SyntaxNode | null {
	for (let child = node.firstChild; child; child = child.nextSibling) {
		if (match(child)) return child;
		const found = find(child, match);
		if (found) return found;
	}
	return null;
}

function titleOf(section: SyntaxNode, text: string): string {
	const heading = find(
		section,
		(node) => node.name === 'Element' && /^h[1-6]$/.test(tagName(node, text) ?? '')
	);
	const title = heading && plain(text.slice(heading.from, heading.to));
	if (title) return title;
	const template = find(
		section,
		(node) => node.name === 'ScriptText' || node.name === 'TextareaText'
	);
	const markdown =
		template && text.slice(template.from, template.to).match(/^\s*#{1,6}\s+(.+)$/m)?.[1];
	if (markdown) return plain(markdown);
	return plain(text.slice(section.from, section.to)).slice(0, 60) || 'Untitled';
}

// Slides in the order reveal.js shows them: top-level sections, and the sections nested in a stack
export function buildOutline(text: string): OutlineSlide[] {
	const document = htmlLanguage.parser.parse(text).topNode;
	return sections(document, text).map((stack) => {
		const children = sections(stack, text).map((slide) => ({
			title: titleOf(slide, text),
			from: slide.from,
			children: []
		}));
		return {
			title: children.length ? children[0].title : titleOf(stack, text),
			from: stack.from,
			children
		};
	});
}
