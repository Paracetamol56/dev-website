import { DebouncedOperation } from '$lib/debouncedOperation';
import { html } from '@codemirror/lang-html';
import { getOutline, type OutlineSection } from './outline';

export class DebouncedOutlineGenerator extends DebouncedOperation<[string], OutlineSection[]> {
	private parser = html().language.parser;

	constructor(delay: number = 5000) {
		super(delay, {
			useStore: true
		});
	}

	protected executeOperation(sourceText: string): OutlineSection[] {
		const ast = this.parser.parse(sourceText);
		return getOutline(ast, sourceText);
	}

	// Custom argument comparison for strings
	protected compareArgs(prev: [string], current: [string]): boolean {
		return prev[0] === current[0];
	}
}
