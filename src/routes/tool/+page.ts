import type { PageLoad } from './$types';
import type { Tool } from '$lib/page';
import toolsJson from '../../content/tools.json';


export const load: PageLoad = async () => {
	const tools: Tool[] = toolsJson.filter((t: Tool) => t.listed);

	return { tools };
};
