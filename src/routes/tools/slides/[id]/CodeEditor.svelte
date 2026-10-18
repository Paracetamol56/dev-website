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
	import type { DeckSocket } from './socket';

	export let value: string;
	export let version: number;
	export let socket: DeckSocket;

	const dispatch = createEventDispatcher<{ change: string; pending: boolean; peers: Peer[] }>();
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
				collaboration(
					socket,
					startVersion,
					(pending) => dispatch('pending', pending),
					(peers) => dispatch('peers', peers)
				),
				EditorView.updateListener.of((update) => {
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

	onMount(() => {
		view = new EditorView({ parent, state: createState(value, version) });
		return () => view?.destroy();
	});

	$: view?.dispatch({ effects: theme.reconfigure(themes[$user.flavour] ?? catppuccinMocha) });
</script>

<div bind:this={parent} class="h-full" />
