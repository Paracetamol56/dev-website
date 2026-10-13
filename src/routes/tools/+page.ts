import { listTools } from '$lib/content';
import type { PageLoad } from './$types';

export const load: PageLoad = () => ({ tools: listTools().filter((tool) => tool.listed) });
