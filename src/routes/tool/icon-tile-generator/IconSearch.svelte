<script lang="ts">
	import { Search } from 'lucide-svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import MeltCheckbox from '$lib/components/MeltCheckbox.svelte';
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import api from '$lib/api';
	import { onDestroy } from 'svelte';
	import { derived, writable, type Writable } from 'svelte/store';
	import { iconUrl } from './icon-fetcher';
	import type { Icon, IconSource } from './types';

	export let iconName: Writable<string>;
	export let iconSource: Writable<IconSource | ''>;

	const LIMIT = 48;

	const sources: { value: IconSource; label: string; checked: Writable<boolean> }[] = [
		{ value: 'lucide', label: 'Lucide', checked: writable(true) },
		{ value: 'simpleicons', label: 'Simple Icons', checked: writable(true) }
	];
	const sourceLabels = Object.fromEntries(sources.map((s) => [s.value, s.label]));
	const selectedSources = derived(
		sources.map((s) => s.checked),
		($checked) => sources.filter((_, i) => $checked[i]).map((s) => s.value)
	);

	const query = writable($iconName);

	let results: Icon[] = [];
	let total = 0;
	let loading = false;
	let error: string | null = null;
	let timer: ReturnType<typeof setTimeout>;
	let requestId = 0;

	async function search(q: string, selected: IconSource[], offset: number) {
		const id = ++requestId;
		if (selected.length === 0) {
			results = [];
			total = 0;
			loading = false;
			return;
		}
		loading = true;
		error = null;
		const params = new URLSearchParams({ q, limit: String(LIMIT), offset: String(offset) });
		selected.forEach((source) => params.append('source', source));
		try {
			const { data } = await api.call('GET', `/icons/search?${params}`);
			if (id !== requestId) return;
			results = offset === 0 ? data.items : [...results, ...data.items];
			total = data.total;
		} catch {
			if (id !== requestId) return;
			error = 'Could not search icons.';
			if (offset === 0) results = [];
		} finally {
			if (id === requestId) loading = false;
		}
	}

	// Debounce typing; changing the selected sources searches immediately
	let lastSources = $selectedSources;
	$: {
		const q = $query.trim();
		const selected = $selectedSources;
		clearTimeout(timer);
		if (typeof window !== 'undefined') {
			timer = setTimeout(() => search(q, selected, 0), selected === lastSources ? 250 : 0);
		}
		lastSources = selected;
	}

	onDestroy(() => clearTimeout(timer));

	let selectedTitle: string | null = null;

	function select(icon: Icon) {
		selectedTitle = icon.title;
		iconName.set(icon.name);
		iconSource.set(icon.source);
	}

	// Icons loaded from a link only know their name until picked from the results
	$: selectedLabel = $iconSource
		? `${selectedTitle ?? $iconName} (${sourceLabels[$iconSource]})`
		: $iconName;
</script>

<div class="flex flex-col gap-3">
	<div class="relative">
		<Search class="absolute left-3 bottom-2 text-ctp-subtext0" size="16" />
		<TextInput
			label="Search icons"
			value={query}
			placeholder="Search by name, title or tag..."
			class="pl-9"
		/>
	</div>

	<div class="flex flex-wrap items-center gap-x-4 gap-y-2">
		{#each sources as source}
			<MeltCheckbox
				name="icon-source-{source.value}"
				label={source.label}
				checked={source.checked}
			/>
		{/each}
		<p class="ml-auto text-sm text-ctp-subtext0" aria-live="polite">
			{#if $selectedSources.length === 0}
				Select at least one source
			{:else if error}
				<span class="text-ctp-red">{error}</span>
			{:else if loading && results.length === 0}
				Searching...
			{:else}
				{total} icon{total === 1 ? '' : 's'} found
			{/if}
		</p>
	</div>

	{#if results.length > 0}
		<ul class="grid grid-cols-[repeat(auto-fill,minmax(2.75rem,1fr))] gap-1.5">
			{#each results as icon (icon.id)}
				{@const selected = icon.name === $iconName && icon.source === $iconSource}
				{@const label = `${icon.title} (${sourceLabels[icon.source]})`}
				<li>
					<MeltTooltip text={label}>
						<button
							type="button"
							on:click={() => select(icon)}
							aria-pressed={selected}
							aria-label={label}
							class="w-full aspect-square grid place-items-center rounded-md bg-ctp-surface0 hocus:bg-ctp-surface1
								focus:outline-none focus:ring-2 focus:ring-ctp-mauve
								{selected ? 'ring-2 ring-ctp-mauve' : ''}"
						>
							<!-- Masked so every icon takes the theme's text colour, whatever its source -->
							<span
								class="square-5 bg-ctp-text"
								style="mask: url({iconUrl(
									icon
								)}) center / contain no-repeat; -webkit-mask: url({iconUrl(
									icon
								)}) center / contain no-repeat;"
								aria-hidden="true"
							/>
						</button>
					</MeltTooltip>
				</li>
			{/each}
		</ul>
	{/if}

	<div class="flex items-center justify-between gap-4 text-sm">
		<p class="text-ctp-subtext0 truncate">
			Selected: <span class="font-semibold text-ctp-text">{selectedLabel}</span>
		</p>
		{#if results.length < total}
			<button
				type="button"
				class="shrink-0 text-ctp-mauve hover:underline disabled:opacity-50"
				disabled={loading}
				on:click={() => search($query.trim(), $selectedSources, results.length)}
			>
				{loading ? 'Loading...' : 'Show more'}
			</button>
		{/if}
	</div>
</div>
