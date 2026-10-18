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
	if (editing) {
		deferred = true;
		return;
	}
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
	sendSlide();
	if (selected !== null) mark(selected);
}

const inline = new Set(
	'A ABBR B BDI BDO BR CITE CODE DEL DFN EM FONT I INS KBD MARK Q S SAMP SMALL SPAN STRIKE STRONG SUB SUP TIME U VAR WBR'.split(
		' '
	)
);
const opaque = 'pre, script, style, template, textarea, svg, math, iframe, video, audio, canvas';
let presenting = false;
let readonly = false;
let selected = null;
let editing = null;
let deferred = false;

function sendSlide() {
	const slide = deck.getCurrentSlide();
	const stack = slide?.closest('.slides > section');
	const source = (element) => (element?.hasAttribute('data-cm') ? idOf(element) : null);
	send({ type: 'slide', indices: deck.getIndices(), slide: source(slide), stack: source(stack) });
}

const sourceOf = (node) => node?.closest?.('.slides [data-cm]') ?? null;
const idOf = (element) => Number(element.getAttribute('data-cm'));

function onlyText(element) {
	return [...element.children].every((child) => inline.has(child.tagName) && onlyText(child));
}

function editable(element) {
	return (
		!inline.has(element.tagName) &&
		!element.matches('section') &&
		!element.closest(opaque) &&
		element.textContent.trim() !== '' &&
		onlyText(element)
	);
}

function editableFrom(node) {
	for (let element = sourceOf(node); element; element = sourceOf(element.parentElement)) {
		if (editable(element)) return element;
		if (!inline.has(element.tagName)) return null;
	}
	return null;
}

function mark(id) {
	selected = id;
	slides
		.querySelectorAll('[data-selected]')
		.forEach((element) => element.removeAttribute('data-selected'));
	slides.querySelector(`[data-cm="${id}"]`)?.setAttribute('data-selected', '');
}

function show(id) {
	const element = slides.querySelector(`[data-cm="${id}"]`);
	if (!element) return;
	mark(id);
	const slide = element.closest('.slides > section > section, .slides > section');
	if (slide && !slide.contains(deck.getCurrentSlide())) {
		const { h, v } = deck.getIndices(slide);
		deck.slide(h, v);
	}
}

const clean = (element) => {
	const copy = element.cloneNode(true);
	copy.querySelectorAll('*').forEach((child) => {
		['data-cm', 'data-selected', 'data-hover'].forEach((name) => child.removeAttribute(name));
	});
	copy.querySelectorAll('strike').forEach((strike) => {
		const s = document.createElement('s');
		s.append(...strike.childNodes);
		strike.replaceWith(s);
	});
	return copy.innerHTML;
};

const alignOf = (element) => {
	const value = getComputedStyle(element).textAlign;
	if (value === 'center') return 'center';
	if (value === 'right' || value === 'end') return 'right';
	if (value === 'left' || value === 'start') return 'left';
	return null;
};

const tagOf = (element) => element.tagName.toLowerCase();

function smallChild(element) {
	const nodes = [...element.childNodes].filter(
		(node) => node.nodeType !== Node.TEXT_NODE || node.textContent.trim() !== ''
	);
	return nodes.length === 1 && nodes[0].tagName === 'SMALL' ? nodes[0] : null;
}

function levelOf(element) {
	const tag = tagOf(element);
	if (/^h[1-6]$/.test(tag)) return tag;
	if (tag !== 'p') return null;
	return smallChild(element) ? 'small' : 'normal';
}

// A level is a tag, and "small" is a paragraph wrapped in <small>
function changeLevel(level) {
	const { element } = editing;
	const tag = level === 'small' || level === 'normal' ? 'p' : level;
	let next = element;
	if (tagOf(element) !== tag) {
		next = document.createElement(tag);
		for (const { name, value } of element.attributes) next.setAttribute(name, value);
		next.append(...element.childNodes);
		element.replaceWith(next);
	}
	const small = smallChild(next);
	if (level === 'small' && !small) {
		const wrapper = document.createElement('small');
		wrapper.append(...next.childNodes);
		next.append(wrapper);
	} else if (level !== 'small' && small) {
		small.replaceWith(...small.childNodes);
	}
	editing.element = next;
	next.focus();
	const range = document.createRange();
	range.selectNodeContents(next);
	range.collapse(false);
	const selection = getSelection();
	selection.removeAllRanges();
	selection.addRange(range);
	editing.range = range;
	deck.layout();
}

function report() {
	if (!editing) return;
	const rect = editing.element.getBoundingClientRect();
	send({
		type: 'editing',
		id: editing.id,
		rect: { top: rect.top, left: rect.left, bottom: rect.bottom, right: rect.right },
		format: {
			bold: document.queryCommandState('bold'),
			italic: document.queryCommandState('italic'),
			underline: document.queryCommandState('underline'),
			strikeThrough: document.queryCommandState('strikeThrough'),
			align: alignOf(editing.element),
			level: levelOf(editing.element)
		}
	});
}

function startEditing(element, selectAll = false) {
	mark(idOf(element));
	const selection = getSelection();
	let range = selection.rangeCount ? selection.getRangeAt(0) : null;
	if (selectAll) {
		range = document.createRange();
		range.selectNodeContents(element);
	}
	editing = {
		element,
		source: element,
		tag: tagOf(element),
		id: idOf(element),
		html: clean(element),
		original: element.innerHTML,
		align: alignOf(element),
		style: element.getAttribute('style'),
		range
	};
	element.contentEditable = 'true';
	element.focus();
	if (range && element.contains(range.commonAncestorContainer)) {
		selection.removeAllRanges();
		selection.addRange(range);
	}
	report();
}

function stopEditing(keep) {
	if (!editing) return;
	const { element, source, tag, id, html, original, align, style } = editing;
	editing = null;
	element.removeAttribute('contenteditable');
	getSelection().removeAllRanges();
	const next = clean(element);
	const nextAlign = alignOf(element);
	const nextTag = tagOf(element);
	const changed = keep && (next !== html || nextAlign !== align || nextTag !== tag);
	if (changed) {
		send({
			type: 'edit',
			id,
			html: next !== html ? next : null,
			align: nextAlign !== align ? nextAlign : null,
			tag: nextTag !== tag ? nextTag : null
		});
	} else {
		if (element !== source) element.replaceWith(source);
		source.removeAttribute('contenteditable');
		source.innerHTML = original;
		if (style === null) source.removeAttribute('style');
		else source.setAttribute('style', style);
	}
	send({ type: 'editing', id: null });
	if (deferred && !changed) {
		deferred = false;
		enqueue(() => render(deck.getIndices()));
	}
	deferred = false;
}

function format(command, value) {
	if (!editing) return;
	if (command === 'level') {
		changeLevel(value);
		report();
		return;
	}
	const { element, range } = editing;
	element.focus();
	if (range) {
		const selection = getSelection();
		selection.removeAllRanges();
		selection.addRange(range);
	}
	if (command === 'align') element.style.textAlign = value;
	else document.execCommand(command, false);
	report();
}

slides.addEventListener('click', (event) => {
	if (event.target.closest('a[href]:not([href^="#"])')) event.preventDefault();
	if (presenting) return;
	const element = sourceOf(event.target);
	if (!element || editing?.element.contains(element)) return;
	mark(idOf(element));
	send({ type: 'select', id: idOf(element) });
});

slides.addEventListener('dblclick', (event) => {
	if (presenting || readonly || editing) return;
	const element = editableFrom(event.target);
	if (element) startEditing(element);
});

slides.addEventListener('pointerover', (event) => {
	slides
		.querySelectorAll('[data-hover]')
		.forEach((element) => element.removeAttribute('data-hover'));
	if (!presenting) sourceOf(event.target)?.setAttribute('data-hover', '');
});

slides.addEventListener('pointerleave', () => {
	slides
		.querySelectorAll('[data-hover]')
		.forEach((element) => element.removeAttribute('data-hover'));
});

document.addEventListener(
	'keydown',
	(event) => {
		if (!editing || (event.key !== 'Escape' && (event.key !== 'Enter' || event.shiftKey))) return;
		event.preventDefault();
		event.stopPropagation();
		stopEditing(event.key === 'Enter');
	},
	true
);

document.addEventListener('focusout', () => {
	setTimeout(() => {
		if (editing && document.hasFocus() && !editing.element.contains(document.activeElement)) {
			stopEditing(true);
		}
	});
});

document.addEventListener('selectionchange', () => {
	if (!editing) return;
	const selection = getSelection();
	if (
		selection.rangeCount &&
		editing.element.contains(selection.getRangeAt(0).commonAncestorContainer)
	) {
		editing.range = selection.getRangeAt(0);
	}
	report();
});

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
	} else if (message.type === 'show') {
		if (!editing && message.id !== null) show(message.id);
	} else if (message.type === 'format') {
		format(message.command, message.value);
	} else if (message.type === 'start') {
		enqueue(() => {
			const element = slides.querySelector(`[data-cm="${message.id}"]`);
			if (presenting || readonly || !element) return;
			stopEditing(true);
			show(message.id);
			if (editable(element)) startEditing(element, true);
		});
	} else if (message.type === 'commit') {
		stopEditing(true);
	} else if (message.type === 'present') {
		presenting = message.on;
		document.documentElement.classList.toggle('presenting', presenting);
		if (presenting) stopEditing(true);
	} else if (message.type === 'readonly') {
		readonly = message.on;
		if (readonly) stopEditing(false);
	} else if (message.type === 'slide') {
		const current = deck.getIndices();
		const { h, v, f } = message.indices;
		if (current.h !== h || current.v !== v) deck.slide(h, v, f);
	}
});

deck.initialize().then(() => {
	deck.on('slidechanged', sendSlide);
	deck.on('slidetransitionend', report);
	deck.on('resize', report);
	send({ type: 'ready' });
});
