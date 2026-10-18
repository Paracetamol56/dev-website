import type { Tree } from '@lezer/common';

interface OutlineSection {
	element: 'section';
	heading: string | null;
	headingLevel: number | null;
	from: number;
	to: number;
}

// Helper function to extract tag name from element text
function extractTagName(elementText: string): string | null {
	const match = elementText.match(/<(\w+)/);
	return match ? match[1].toLowerCase() : null;
}

// Helper function to extract heading text content
function extractHeadingText(elementText: string): string {
	// Remove HTML tags and get inner text
	const match = elementText.match(/>([^<]*)</);
	return match ? match[1].trim() : '';
}

// Helper function to get heading level (1-6)
function getHeadingLevel(tagName: string): number | null {
	const match = tagName.match(/^h([1-6])$/);
	return match ? parseInt(match[1]) : null;
}

// Main function to get the outline - only direct children of Document that are complete section elements
function getOutline(ast: Tree, sourceText: string): OutlineSection[] {
	const outline: OutlineSection[] = [];

	if (ast.type.name !== 'Document') {
		console.warn('Expected Document node as root');
		return outline;
	}

	// Find all top-level complete section elements
	const topLevelSections: Array<{ from: number; to: number; text: string }> = [];

	// Parse the source text to find complete section elements at the root level
	const sectionRegex = /<section(?:[^>]*)>(.*?)<\/section>/gs;
	let match;

	// Simple approach: find sections that are not nested
	while ((match = sectionRegex.exec(sourceText)) !== null) {
		const fullMatch = match[0];
		const matchStart = match.index!;
		const matchEnd = matchStart + fullMatch.length;

		// Check if this section is at the root level by counting preceding sections
		const textBefore = sourceText.slice(0, matchStart);
		const openSections = (textBefore.match(/<section(?:[^>]*)>/g) || []).length;
		const closeSections = (textBefore.match(/<\/section>/g) || []).length;

		// If we're at root level (equal open/close tags before this section)
		if (openSections === closeSections) {
			topLevelSections.push({
				from: matchStart,
				to: matchEnd,
				text: fullMatch
			});
		}
	}

	// Process each top-level section
	for (const section of topLevelSections) {
		// Find the highest level heading within this section
		let highestHeading: string | null = null;
		let highestLevel: number | null = null;

		const headingRegex = /<(h[1-6])(?:[^>]*)>([^<]*)<\/h[1-6]>/gi;
		let headingMatch;

		while ((headingMatch = headingRegex.exec(section.text)) !== null) {
			const headingTag = headingMatch[1].toLowerCase();
			const headingText = headingMatch[2].trim();
			const level = getHeadingLevel(headingTag);

			if (level && (highestLevel === null || level < highestLevel)) {
				highestLevel = level;
				highestHeading = headingText;
			}
		}

		outline.push({
			element: 'section',
			heading: highestHeading,
			headingLevel: highestLevel,
			from: section.from,
			to: section.to
		});
	}

	return outline;
}

export { getOutline, type OutlineSection };
