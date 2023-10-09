import adapter from '@sveltejs/adapter-node';
import sequence from 'svelte-sequential-preprocessor'
import { vitePreprocess } from '@sveltejs/kit/vite';
import { escapeSvelte, mdsvex } from 'mdsvex'
import { preprocessMeltUI } from '@melt-ui/pp'
import shiki from 'shiki'
import fs from 'fs'

/** @type {import('mdsvex').MdsvexOptions} */
const mdsvexOptions = {
	extensions: ['.md'],
	highlight: {
		highlighter: async (code, lang = 'text') => {
			const highlighter = await shiki.getHighlighter({ theme: JSON.parse(fs.readFileSync('./src/themes/mocha.json', 'utf-8')) })
			const html = escapeSvelte(highlighter.codeToHtml(code, { lang }))
			return `{@html \`${html}\` }`
		}
	},
}


/** @type {import('@sveltejs/kit').Config} */
const config = {
	extensions: ['.svelte', '.md'],
	preprocess: sequence([
		vitePreprocess({}),
		mdsvex(mdsvexOptions),
		preprocessMeltUI()
	]),
	kit: {
		adapter: adapter()
	},
};

export default config;
