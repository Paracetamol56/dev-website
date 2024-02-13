import type { PageLoad } from './$types';
import type { Page } from '$lib/page';

interface File {
	metadata: Omit<Page, 'slug'>;
}

export const load: PageLoad = async ({ url }) => {
  const tag = url.searchParams.get('tag');
	const pages: Page[] = [];

	const paths = import.meta.glob('/src/content/*.md', { eager: true });

	for (const path in paths) {
		const file: File = paths[path] as File;
		if (!file) continue;
		const slug = path.split('/').pop()?.replace(/\.md$/, '');

		if (slug && typeof file === 'object' && 'metadata' in file) {
			const metadata: Omit<Page, 'slug'> = file.metadata;
			const page: Page = { ...metadata, slug } satisfies Page;
			if (page.listed && (!tag || page.tags.includes(tag))) {
        pages.push(page);
      }
		}
	}

  return { tag, pages };
};

