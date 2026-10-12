import { listPages, listTools } from '$lib/content';
import type { PageLoad } from './$types';

export const load: PageLoad = () => ({ pages: listPages(), tools: listTools() });

export const prerender = true;
