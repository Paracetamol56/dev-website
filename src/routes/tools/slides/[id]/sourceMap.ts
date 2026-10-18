import { htmlLanguage } from '@codemirror/lang-html';
import { StateEffect, StateField, type ChangeDesc, type EditorState } from '@codemirror/state';

export type SourceElement = { from: number; to: number; contentFrom: number; contentTo: number };

export const sourceAttribute = 'data-cm';

const mapElement = (element: SourceElement, changes: ChangeDesc): SourceElement => ({
	from: changes.mapPos(element.from, 1),
	to: changes.mapPos(element.to, -1),
	contentFrom: changes.mapPos(element.contentFrom, -1),
	contentTo: changes.mapPos(element.contentTo, 1)
});

export const setSourceElements = StateEffect.define<SourceElement[]>();

// Ranges of the elements in the last rendered preview, by their id in the frame
export const sourceElements = StateField.define<SourceElement[]>({
	create: () => [],
	update(elements, transaction) {
		for (const effect of transaction.effects) {
			if (effect.is(setSourceElements)) return effect.value;
		}
		if (!transaction.docChanged || !elements.length) return elements;
		return elements.map((element) => mapElement(element, transaction.changes));
	}
});

// Gives every element an id attribute pointing back to its source range
export function annotate(text: string): { html: string; elements: SourceElement[] } {
	const elements: SourceElement[] = [];
	const inserts: { at: number; id: number }[] = [];
	htmlLanguage.parser.parse(text).iterate({
		enter(node) {
			if (node.name !== 'Element') return;
			const open = node.node.getChild('OpenTag') ?? node.node.getChild('SelfClosingTag');
			const name = open?.getChild('TagName');
			if (!open || !name) return;
			const close = node.node.getChild('CloseTag');
			const contentFrom = open.name === 'OpenTag' ? open.to : node.to;
			inserts.push({ at: name.to, id: elements.length });
			elements.push({
				from: node.from,
				to: node.to,
				contentFrom,
				contentTo: close ? close.from : node.to
			});
		}
	});
	let html = '';
	let last = 0;
	for (const { at, id } of inserts) {
		html += `${text.slice(last, at)} ${sourceAttribute}="${id}"`;
		last = at;
	}
	return { html: html + text.slice(last), elements };
}

export function innermostAt(elements: SourceElement[], position: number): number | null {
	let found: number | null = null;
	elements.forEach((element, id) => {
		if (element.from > position || element.to < position) return;
		if (found === null || element.to - element.from < elements[found].to - elements[found].from) {
			found = id;
		}
	});
	return found;
}

const styleAttribute = /\sstyle\s*=\s*(["'])(.*?)\1/is;

// Sets one CSS property in the style attribute of an opening tag
export function withStyle(tag: string, property: string, value: string): string {
	const declaration = `${property}: ${value}`;
	const match = tag.match(styleAttribute);
	if (!match) return tag.replace(/\s*(\/?>)$/, ` style="${declaration}"$1`);
	const declarations = match[2]
		.split(';')
		.map((part) => part.trim())
		.filter((part) => part && part.split(':')[0].trim().toLowerCase() !== property);
	declarations.push(declaration);
	const quote = match[1];
	return tag.replace(match[0], ` style=${quote}${declarations.join('; ')}${quote}`);
}

export const elementById = (state: EditorState, id: number): SourceElement | undefined =>
	state.field(sourceElements, false)?.[id];
