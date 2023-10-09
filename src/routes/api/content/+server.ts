import { json } from '@sveltejs/kit'
import type { Page } from '$lib/page'

async function getContent() {
	const pages: Page[] = []

	const paths = import.meta.glob('/src/content/*.md', { eager: true })

	for (const path in paths) {
		const file = paths[path]
		const slug = path.split('/').at(-1)?.replace('.md', '')

		if (file && typeof file === 'object' && 'metadata' in file && slug) {
			const metadata = file.metadata as Omit<Page, 'slug'>
			const page = { ...metadata, slug } satisfies Page
			page.published && pages.push(page)
		}
	}

	// pages = pages.sort((first, second) =>
  //   new Date(second.date).getTime() - new Date(first.date).getTime()
	// )

	return pages
}

export async function GET() {
	const pages = await getContent()
	return json(pages)
}
