import adapter from '@sveltejs/adapter-static';
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
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html',
			strict: false,
			precompress: true,
			enableSourcemap: true
		}),
		prerender: {
			crawl: true,
			concurrency: 4,
			entries: [
				'/',
				'/contact',
				'/embed/mars-missions',
				'/embed/moore-s-law',
				'/embed/sync-vs-async',
				'/page',
				'/page/ai-activation-functions',
				'/page/ai-activation-functions/og.png',
				'/page/barnes-hut',
				'/page/barnes-hut/og.png',
				'/page/boids',
				'/page/boids/og.png',
				'/page/convolution',
				'/page/convolution/og.png',
				'/page/hertzsprung-russell-diagram',
				'/page/hertzsprung-russell-diagram/og.png',
				'/page/l-informatique-dans-la-conquete-spatiale-et-l-astronomie',
				'/page/l-informatique-dans-la-conquete-spatiale-et-l-astronomie/og.png',
				'/page/maze-solving',
				'/page/maze-solving/og.png',
				'/page/mnist-classification',
				'/page/mnist-classification/og.png',
				'/page/moore-s-law',
				'/page/moore-s-law/og.png',
				'/page/privacy-policy',
				'/page/privacy-policy/og.png',
				'/page/raspberry-pi-fan-controller',
				'/page/raspberry-pi-fan-controller/og.png',
				'/page/what-s-on-my-homelab-part-1',
				'/page/what-s-on-my-homelab-part-1/og.png',
				'/page/what-s-on-my-homelab-part-2',
				'/page/what-s-on-my-homelab-part-2/og.png',
				'/page/what-s-on-my-homelab-part-3',
				'/page/what-s-on-my-homelab-part-3/og.png',
				'/page/what-s-on-my-homelab-part-4',
				'/page/what-s-on-my-homelab-part-4/og.png',
				'/page/wikipedia-graph-view',
				'/page/wikipedia-graph-view/og.png',
				'/settings',
				'/tool',
				'/verify/email',
				'/verify/github',
			],
			handleHttpError: 'warn',
			handleMissingId: 'warn'
		}
	}
};

export default config;
