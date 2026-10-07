// Writes build/sitemap.xml listing every prerendered HTML page.
import { readdirSync, writeFileSync } from 'node:fs';

const DOMAIN = 'https://dev.matheo-galuba.com';
const OUT_DIR = 'build';

const urls = readdirSync(OUT_DIR, { recursive: true })
	.filter((file) => file.endsWith('.html'))
	.map((file) => {
		const route = file
			.split(/[\\/]/)
			.map((segment) => encodeURIComponent(segment))
			.join('/')
			.replace(/(^|\/)index\.html$/, '')
			.replace(/\.html$/, '')
			.replace(/\/$/, '');
		return route ? `${DOMAIN}/${route}` : DOMAIN;
	})
	.sort();

const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls.map((url) => `  <url>\n    <loc>${url}</loc>\n  </url>`).join('\n')}
</urlset>
`;

writeFileSync(`${OUT_DIR}/sitemap.xml`, xml);
console.log(`sitemap.xml: ${urls.length} pages`);
