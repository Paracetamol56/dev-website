import type { Page, Tool } from '$lib/page';
import type { PageLoad } from './$types';
import toolsJson from '../content/tools.json';

const tools = toolsJson as Tool[];
interface File {
	metadata: Omit<Page, 'slug'>;
}

export const load: PageLoad = () => {
	const pages: Page[] = [];

	const paths = import.meta.glob('/src/content/*.md', { eager: true });

	for (const path in paths) {
		const file: File = paths[path] as File;
		if (!file) continue;
		const slug = path.split('/').pop()?.replace(/\.md$/, '');

		if (slug && typeof file === 'object' && 'metadata' in file) {
			const metadata: Omit<Page, 'slug'> = file.metadata;
			const page: Page = { ...metadata, slug } satisfies Page;
			if (page.listed) pages.push(page);
		}
	}

	return { pages, tools };
};

export const prerender = true;
