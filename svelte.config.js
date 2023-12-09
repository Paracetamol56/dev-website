import adapter from '@sveltejs/adapter-node';
import sequence from 'svelte-sequential-preprocessor';
import { vitePreprocess } from '@sveltejs/kit/vite';
import { escapeSvelte, mdsvex } from 'mdsvex';
import { preprocessMeltUI } from '@melt-ui/pp';
import shiki from 'shiki';
import fs from 'fs';
import remarkUnwrapImages from 'remark-unwrap-images';
import remarkToc from 'remark-toc';
import remarkMermaid from 'remark-mermaidjs';
import remarkMath from 'remark-math';
import rehypeSlug from 'rehype-slug';
import rehypeKatex from 'rehype-katex-svelte';

/** @type {import('mdsvex').MdsvexOptions} */
const mdsvexOptions = {
	extensions: ['.md'],
	highlight: {
		highlighter: async (code, lang = 'text') => {
			const highlighter = await shiki.getHighlighter({
				theme: JSON.parse(fs.readFileSync('./src/themes/mocha.json', 'utf-8'))
			});
			let html = escapeSvelte(highlighter.codeToHtml(code, { lang }));
			html = html.replace('<pre ', '<pre data-lang="' + lang + '" ');
			return `{@html \`${html}\` }`;
		}
	},
	remarkPlugins: [remarkUnwrapImages, remarkToc, remarkMermaid, remarkMath],
	rehypePlugins: [rehypeSlug, rehypeKatex]
};

/** @type {import('@sveltejs/kit').Config} */
const config = {
	extensions: ['.svelte', '.md'],
	preprocess: sequence([vitePreprocess({}), mdsvex(mdsvexOptions), preprocessMeltUI()]),
	kit: {
		adapter: adapter()
	}
};

export default config;
