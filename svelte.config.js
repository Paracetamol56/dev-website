import adapter from '@sveltejs/adapter-static';
import sequence from 'svelte-sequential-preprocessor';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
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

const highlighterPromise = shiki.getHighlighter({
	theme: JSON.parse(fs.readFileSync('./src/themes/mocha.json', 'utf-8'))
});

/** @type {import('mdsvex').MdsvexOptions} */
const mdsvexOptions = {
	extensions: ['.md'],
	highlight: {
		highlighter: async (code, lang = 'text') => {
			const highlighter = await highlighterPromise;
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
				'/embed/moore-s-law',
				'/pages',
				'/pages/ai-activation-functions',
				'/pages/ai-activation-functions/og.png',
				'/pages/barnes-hut',
				'/pages/barnes-hut/og.png',
				'/pages/boids',
				'/pages/boids/og.png',
				'/pages/convolution',
				'/pages/convolution/og.png',
				'/pages/hertzsprung-russell-diagram',
				'/pages/hertzsprung-russell-diagram/og.png',
				'/pages/maze-solving',
				'/pages/maze-solving/og.png',
				'/pages/mnist-classification',
				'/pages/mnist-classification/og.png',
				'/pages/moore-s-law',
				'/pages/moore-s-law/og.png',
				'/pages/privacy-policy',
				'/pages/privacy-policy/og.png',
				'/pages/raspberry-pi-fan-controller',
				'/pages/raspberry-pi-fan-controller/og.png',
				'/pages/what-s-on-my-homelab-part-1',
				'/pages/what-s-on-my-homelab-part-1/og.png',
				'/pages/what-s-on-my-homelab-part-2',
				'/pages/what-s-on-my-homelab-part-2/og.png',
				'/pages/what-s-on-my-homelab-part-3',
				'/pages/what-s-on-my-homelab-part-3/og.png',
				'/pages/what-s-on-my-homelab-part-4',
				'/pages/what-s-on-my-homelab-part-4/og.png',
				'/pages/wikipedia-graph-view',
				'/pages/wikipedia-graph-view/og.png',
				'/settings',
				'/tools',
				'/verify/email',
				'/verify/github',
				'/verify/google'
			],
			handleHttpError: 'warn',
			handleMissingId: 'warn'
		}
	}
};

export default config;
