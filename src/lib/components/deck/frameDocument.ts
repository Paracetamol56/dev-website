import revealCss from 'reveal.js/dist/reveal.css?url';
import revealJs from 'reveal.js/dist/reveal.js?url';
import highlightJs from 'reveal.js/plugin/highlight/highlight.js?url';
import markdownJs from 'reveal.js/plugin/markdown/markdown.js?url';
import frameJs from './frame.js?raw';

const script = (source: string) => `<script src="${source}"></script>`;

export const frameDocument = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<link rel="stylesheet" href="${revealCss}">
<style id="theme"></style>
<style id="highlight"></style>
</head>
<body>
<div class="reveal"><div class="slides"></div></div>
${script(revealJs)}
${script(markdownJs)}
${script(highlightJs)}
<script>${frameJs}</script>
</body>
</html>`;

export type FrameFont = { family: string; weight: string; style: string; data: ArrayBuffer };
export type FrameTheme = { css: string; fonts: FrameFont[] };

const themeFiles = import.meta.glob<string>('/node_modules/reveal.js/dist/theme/*.css', {
	query: '?inline',
	import: 'default'
});
const highlightFiles = import.meta.glob<string>('/node_modules/reveal.js/plugin/highlight/*.css', {
	query: '?inline',
	import: 'default'
});

const fontFaces = /@font-face\s*{([^}]*)}/g;
const remoteImports = /@import\s*(?:url\(\s*)?['"]?https?:[^;]*;?/g;
const property = (block: string, name: string) =>
	block.match(new RegExp(`${name}\\s*:\\s*['"]?([^;'"]+)`))?.[1].trim();

const fonts = new Map<string, Promise<ArrayBuffer>>();
function fetchFont(url: string) {
	if (!fonts.has(url))
		fonts.set(
			url,
			fetch(url).then((response) => response.arrayBuffer())
		);
	return fonts.get(url) as Promise<ArrayBuffer>;
}

// The frame has an opaque origin: it cannot fetch the theme fonts from the site, and remote
// fonts would tell a third party who is watching. The page fetches the bundled fonts for it.
async function prepare(css: string): Promise<FrameTheme> {
	const faces = [...css.matchAll(fontFaces)].flatMap(([, block]) => {
		const url = block.match(/url\(\s*['"]?([^'")?#]+\.woff)\b/)?.[1];
		const family = property(block, 'font-family');
		if (!url || !family) return [];
		const weight = property(block, 'font-weight') ?? 'normal';
		const style = property(block, 'font-style') ?? 'normal';
		return [fetchFont(url).then((data) => ({ family, weight, style, data }))];
	});
	return {
		css: css.replace(fontFaces, '').replace(remoteImports, ''),
		fonts: await Promise.all(faces)
	};
}

const themes = new Map<string, Promise<FrameTheme>>();
export function loadTheme(name: string): Promise<FrameTheme> {
	if (!themes.has(name)) {
		const file =
			themeFiles[`/node_modules/reveal.js/dist/theme/${name}.css`] ??
			themeFiles['/node_modules/reveal.js/dist/theme/black.css'];
		themes.set(name, file().then(prepare));
	}
	return themes.get(name) as Promise<FrameTheme>;
}

export function loadHighlightTheme(name: string): Promise<string> {
	const file =
		highlightFiles[`/node_modules/reveal.js/plugin/highlight/${name}.css`] ??
		highlightFiles['/node_modules/reveal.js/plugin/highlight/monokai.css'];
	return file();
}
