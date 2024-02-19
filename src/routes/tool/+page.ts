import type { PageLoad } from './$types';
import type { Tool } from '$lib/page';
import toolsJson from '../../content/tools.json';


export const load: PageLoad = async ({ url }) => {
	const tag = url.searchParams.get('tag');
	const tools: Tool[] = [];
	
	for (const tool of toolsJson as Tool[]) {
		if (tool.listed && (!tag || tool.tags.includes(tag))) {
			tools.push(tool);
		}
	}

	return { tag, tools };
};
