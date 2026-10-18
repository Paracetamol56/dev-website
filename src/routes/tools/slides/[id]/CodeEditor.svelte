<script lang="ts">
	import { basicSetup } from 'codemirror';
	import { getSyncedVersion, sendableUpdates } from '@codemirror/collab';
	import { indentWithTab } from '@codemirror/commands';
	import { html } from '@codemirror/lang-html';
	import { Compartment, EditorState, Prec } from '@codemirror/state';
	import { EditorView, keymap } from '@codemirror/view';
	import {
		catppuccinFrappe,
		catppuccinLatte,
		catppuccinMacchiato,
		catppuccinMocha
	} from '@catppuccin/codemirror';
	import { createEventDispatcher, onMount } from 'svelte';
	import { user } from '$lib/store';
	import { collaboration } from './collab';
	import { presence, type Peer } from './presence';
	import {
		annotate as annotateText,
		elementById,
		innermostAt,
		setSourceElements,
		sourceElements,
		withStyle
	} from './sourceMap';
	import type { DeckSocket } from './socket';
	import type { SlideEdit } from '$lib/components/deck/Slides.svelte';

	export let value: string;
	export let version: number;
	export let socket: DeckSocket;

	const dispatch = createEventDispatcher<{
		change: string;
		pending: boolean;
		peers: Peer[];
		cursor: number | null;
	}>();
	const themes = {
		latte: catppuccinLatte,
		frappe: catppuccinFrappe,
		macchiato: catppuccinMacchiato,
		mocha: catppuccinMocha
	};
	const theme = new Compartment();

	let parent: HTMLDivElement;
	let view: EditorView | null = null;

	function createState(doc: string, startVersion: number) {
		return EditorState.create({
			doc,
			extensions: [
				Prec.high(keymap.of([{ key: 'Mod-s', preventDefault: true, run: () => true }])),
				basicSetup,
				keymap.of([indentWithTab]),
				html(),
				EditorView.lineWrapping,
				EditorView.theme({ '&': { height: '100%' }, '.cm-scroller': { overflow: 'auto' } }),
				theme.of(themes[$user.flavour] ?? catppuccinMocha),
				presence,
				sourceElements,
				collaboration(
					socket,
					startVersion,
					(pending) => dispatch('pending', pending),
					(peers) => dispatch('peers', peers)
				),
				EditorView.updateListener.of((update) => {
					if (update.selectionSet && update.view.hasFocus) {
						const elements = update.state.field(sourceElements);
						dispatch('cursor', innermostAt(elements, update.state.selection.main.head));
					}
					if (!update.docChanged) return;
					value = update.state.doc.toString();
					dispatch('change', value);
				})
			]
		});
	}

	export function syncedVersion(): number {
		return view ? getSyncedVersion(view.state) : version;
	}

	export function hasPending(): boolean {
		return view ? sendableUpdates(view.state).length > 0 : false;
	}

	// Starts again from the server's document, dropping changes it never confirmed
	export function reset(doc: string, startVersion: number) {
		view?.setState(createState(doc, startVersion));
		dispatch('peers', []);
		value = doc;
		dispatch('change', doc);
	}

	// Gives the preview's html, with each element pointing back to its place in the code
	export function annotate(): string {
		if (!view) return value;
		const { html, elements } = annotateText(view.state.doc.toString());
		view.dispatch({ effects: setSourceElements.of(elements) });
		return html;
	}

	export function reveal(id: number) {
		const element = view && elementById(view.state, id);
		if (!view || !element) return;
		view.dispatch({
			selection: { anchor: element.from, head: element.to },
			effects: EditorView.scrollIntoView(element.from, { y: 'center' })
		});
	}

	export function replace({ id, html, align, tag }: SlideEdit) {
		const element = view && elementById(view.state, id);
		if (!view || !element || element.contentFrom === element.from) return;
		const changes = [];
		if (align || tag) {
			let open = view.state.sliceDoc(element.from, element.contentFrom);
			if (tag) open = open.replace(/^<[^\s/>]+/, `<${tag}`);
			if (align) open = withStyle(open, 'text-align', align);
			changes.push({ from: element.from, to: element.contentFrom, insert: open });
		}
		if (html !== null)
			changes.push({ from: element.contentFrom, to: element.contentTo, insert: html });
		if (tag) changes.push({ from: element.contentTo, to: element.to, insert: `</${tag}>` });
		if (changes.length) view.dispatch({ changes, userEvent: 'input.preview' });
	}

	onMount(() => {
		view = new EditorView({ parent, state: createState(value, version) });
		return () => view?.destroy();
	});

	$: view?.dispatch({ effects: theme.reconfigure(themes[$user.flavour] ?? catppuccinMocha) });
</script>

<div bind:this={parent} class="h-full" />
