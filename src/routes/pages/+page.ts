import { listPages } from '$lib/content';
import type { PageLoad } from './$types';

export const load: PageLoad = () => ({ pages: listPages() });

export const prerender = true;
export const ssr = true;
