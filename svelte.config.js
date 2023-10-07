import adapter from '@sveltejs/adapter-node';
import sequence from 'svelte-sequential-preprocessor'
import { preprocessMeltUI } from '@melt-ui/pp'
import { vitePreprocess } from '@sveltejs/kit/vite';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter()
	},
	preprocess: sequence([
		vitePreprocess({}),
		preprocessMeltUI()
	]),
};

export default config;
