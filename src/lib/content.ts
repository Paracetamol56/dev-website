import type { Component } from 'svelte';
import type { Page, Tool } from './page';

type Module<T> = { metadata: Omit<T, 'slug'> };

function collect<T>(modules: Record<string, unknown>): T[] {
	return Object.entries(modules).map(([path, module]) => ({
		...(module as Module<T>).metadata,
		slug: path.split('/').pop()!.replace(/\.md$/, '')
	})) as T[];
}

export function listPages(): Page[] {
	return collect<Page>(import.meta.glob('/src/content/pages/*.md', { eager: true }))
		.filter((page) => page.listed)
		.map((page) => ({ ...page, release: new Date(page.release) }))
		.sort((a, b) => b.release.getTime() - a.release.getTime());
}

export function listTools(): Tool[] {
	const modules = import.meta.glob<{ default: Component }>('/src/content/tools/*.md', {
		eager: true
	});
	const sources = import.meta.glob<string>('/src/content/tools/*.md', {
		eager: true,
		query: '?raw',
		import: 'default'
	});

	return collect<Tool>(modules)
		.map((tool) => {
			const path = `/src/content/tools/${tool.slug}.md`;
			const hasBody = sources[path].replace(/^---[\s\S]*?---/, '').trim() !== '';
			return hasBody ? { ...tool, content: modules[path].default } : tool;
		})
		.sort((a, b) => a.order - b.order);
}
