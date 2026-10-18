// @ts-nocheck
const slides = document.querySelector('.slides');
const deck = new Reveal({
	hash: false,
	history: false,
	respondToHashChanges: false,
	scrollActivationWidth: 0,
	plugins: [RevealMarkdown, RevealHighlight]
});

const send = (message) => parent.postMessage(message, '*');
const fonts = new Set();
let queue = Promise.resolve();
let html = '';
let plugins = { markdown: true, highlight: true };

async function render(indices) {
	slides.innerHTML = html;
	if (plugins.markdown) {
		const markdown = deck.getPlugin('markdown');
		await markdown.processSlides(slides).then(markdown.convertSlides);
	}
	if (plugins.highlight) {
		const highlight = deck.getPlugin('highlight');
		slides
			.querySelectorAll('pre code:not([data-highlighted])')
			.forEach((block) => highlight.highlightBlock(block));
	}
	deck.sync();
	const last = Math.max(deck.getHorizontalSlides().length - 1, 0);
	deck.slide(Math.min(indices.h, last), indices.v);
	send({ type: 'slide', indices: deck.getIndices() });
}

function addFonts(list) {
	for (const { family, data, weight, style } of list) {
		const key = `${family}/${weight}/${style}`;
		if (fonts.has(key)) continue;
		fonts.add(key);
		const face = new FontFace(family, data, { weight, style });
		document.fonts.add(face);
		face.load().then(() => deck.layout());
	}
}

function configure({ settings, theme, highlightCss }) {
	document.getElementById('theme').textContent = theme.css;
	document.getElementById('highlight').textContent = highlightCss;
	addFonts(theme.fonts);
	plugins = { markdown: settings.markdown, highlight: settings.highlight };
	deck.configure({
		transition: settings.transition,
		transitionSpeed: settings.transitionSpeed,
		backgroundTransition: settings.backgroundTransition,
		controls: settings.controls,
		progress: settings.progress,
		slideNumber: settings.slideNumber,
		center: settings.center,
		loop: settings.loop,
		autoAnimate: settings.autoAnimate,
		autoAnimateEasing: settings.autoAnimateEasing,
		autoAnimateDuration: settings.autoAnimateDuration,
		width: settings.width,
		height: settings.height
	});
}

function enqueue(task) {
	queue = queue.then(task).catch(console.error);
}

addEventListener('message', (event) => {
	if (event.source !== parent) return;
	const message = event.data;
	if (message.type === 'render') {
		html = message.html;
		enqueue(() => render(message.indices));
	} else if (message.type === 'settings') {
		enqueue(() => {
			configure(message);
			return render(deck.getIndices());
		});
	} else if (message.type === 'slide') {
		const current = deck.getIndices();
		const { h, v, f } = message.indices;
		if (current.h !== h || current.v !== v) deck.slide(h, v, f);
	}
});

deck.initialize().then(() => {
	deck.on('slidechanged', () => send({ type: 'slide', indices: deck.getIndices() }));
	send({ type: 'ready' });
});
