import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const prerender = false;
export const ssr = false;

const LOCAL_HOSTS = ['localhost', '127.0.0.1', '[::1]'];

export const load: PageLoad = ({ url }) => {
	if (!LOCAL_HOSTS.includes(url.hostname)) error(404, 'Not Found');
};
