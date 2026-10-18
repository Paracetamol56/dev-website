<script lang="ts">
	import { createToolbar, melt } from '@melt-ui/svelte';
	import {
		AlertCircle,
		ArrowLeft,
		Check,
		Code,
		Heading,
		Image,
		List,
		Loader2,
		Pencil,
		Presentation,
		Redo2,
		SquarePlus,
		Type,
		Undo2,
		WifiOff
	} from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import type { Readable } from 'svelte/store';
	import ThemePopover from '$lib/components/ThemePopover.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { Tone } from '$lib/components/ui/styles';
	import type { BlockKind } from './blocks';
	import { peerColour, type Peer } from './presence';
	import type { Connection } from './socket';

	export let title: string;
	export let connection: Readable<Connection>;
	export let unsaved: boolean;
	export let peers: Peer[];

	const dispatch = createEventDispatcher<{
		rename: void;
		present: void;
		insert: BlockKind | 'image';
		undo: void;
		redo: void;
	}>();

	const history = [
		{ action: 'undo', label: 'Undo', shortcut: 'Ctrl+Z', icon: Undo2 },
		{ action: 'redo', label: 'Redo', shortcut: 'Ctrl+Shift+Z', icon: Redo2 }
	] as const;
	const inserts = [
		{ kind: 'slide', label: 'Slide', icon: SquarePlus },
		{ kind: 'text', label: 'Text', icon: Type },
		{ kind: 'heading', label: 'Heading', icon: Heading },
		{ kind: 'list', label: 'List', icon: List },
		{ kind: 'image', label: 'Image', icon: Image },
		{ kind: 'code', label: 'Code', icon: Code }
	] as const;

	const {
		elements: { root, button, separator }
	} = createToolbar();
</script>

<header class="flex flex-col gap-3">
	<div class="flex flex-wrap items-center gap-x-3 gap-y-2">
		<a
			href="/tools/slides"
			class="grid place-items-center rounded-md p-1.5 text-ctp-subtext0 hover:bg-ctp-surface0 hover:text-ctp-text"
			aria-label="Back to your decks"
		>
			<ArrowLeft size="18" />
		</a>
		<button
			type="button"
			class="group flex min-w-0 max-w-xs items-center gap-2 rounded-md px-2 py-1 hover:bg-ctp-surface0"
			on:click={() => dispatch('rename')}
		>
			<h1 class="truncate text-lg font-bold text-ctp-text">{title}</h1>
			<Pencil size="14" class="shrink-0 text-ctp-subtext0 opacity-0 group-hover:opacity-100" />
			<span class="sr-only">Rename</span>
		</button>

		<p class="flex items-center gap-1.5 text-sm text-ctp-subtext0" role="status">
			{#if $connection === 'refused'}
				<AlertCircle size="14" class="text-ctp-red" /> Disconnected
			{:else if $connection === 'closed'}
				<WifiOff size="14" class="text-ctp-yellow" /> Offline, reconnecting…
			{:else if $connection === 'connecting'}
				<Loader2 size="14" class="animate-spin" /> Connecting…
			{:else if unsaved}
				<Loader2 size="14" class="animate-spin" /> <span class="sr-only">Saving…</span>
			{:else}
				<Check size="14" class="text-ctp-green" /> <span class="sr-only">Saved</span>
			{/if}
		</p>

		<div class="ml-auto flex items-center gap-3">
			{#if peers.length}
				<ul class="flex -space-x-1.5" aria-label="Also editing">
					{#each peers as peer (peer.peer)}
						<li
							class="grid h-7 w-7 place-items-center rounded-full border-2 border-ctp-mantle text-xs font-bold text-ctp-crust"
							style="background-color: rgb(var(--ctp-{peerColour(peer.peer)}))"
							title="{peer.name} is editing"
						>
							{peer.name.slice(0, 1).toUpperCase()}
						</li>
					{/each}
				</ul>
			{/if}
			<ThemePopover />
			<Button tone={Tone.Accent} on:click={() => dispatch('present')}>
				<Presentation size="18" />
				Present
			</Button>
		</div>
	</div>

	<div
		use:melt={$root}
		aria-label="Edit the deck"
		class="flex flex-wrap items-center gap-1 rounded-md bg-ctp-base p-1 shadow-md shadow-ctp-crust"
	>
		{#each history as item}
			<button
				type="button"
				class="tool"
				aria-label={item.label}
				title="{item.label} ({item.shortcut})"
				use:melt={$button}
				on:click={() => dispatch(item.action)}
			>
				<svelte:component this={item.icon} size="16" />
			</button>
		{/each}
		<div class="mx-0.5 w-px self-stretch bg-ctp-surface1" use:melt={$separator} />
		{#each inserts as item}
			<button
				type="button"
				class="tool"
				aria-label="Add {item.label.toLowerCase()}"
				title="Add {item.label.toLowerCase()}"
				use:melt={$button}
				on:click={() => dispatch('insert', item.kind)}
			>
				<svelte:component this={item.icon} size="16" />
				<span class="hidden md:inline">{item.label}</span>
			</button>
		{/each}
	</div>
</header>

<style lang="postcss">
	.tool {
		@apply flex items-center gap-1.5 rounded-md px-2 py-1.5 text-sm font-semibold text-ctp-subtext1 transition-colors
			hover:bg-ctp-surface0 hover:text-ctp-text focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve;
	}
</style>
