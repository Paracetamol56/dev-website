<script lang="ts">
	import api from '$lib/api';
	import Button from '$lib/components/Button.svelte';
	import MeltCheckbox from '$lib/components/MeltCheckbox.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import { downloadBlob } from '$lib/download';
	import { ArrowDown, ArrowUp, Eye, EyeOff, FileSpreadsheet, Merge, Search } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { derived, get, writable } from 'svelte/store';
	import { addToast } from '../../../../+layout.svelte';
	import Dialog from '../../Dialog.svelte';
	import { countWords, type WordCloudWord } from '../../utils';

	export let id: string;
	export let words: WordCloudWord[];
	/** Participant numbers, in order of first submission */
	export let participants: Map<string, number>;

	const dispatch = createEventDispatcher<{ words: WordCloudWord[]; participant: string }>();
	const PER_PAGE = 25;

	type SortKey = 'createdAt' | 'text' | 'participant';
	const mergeInto = writable('');
	const COLUMNS: [SortKey, string][] = [
		['text', 'Word'],
		['participant', 'Participant'],
		['createdAt', 'Time']
	];
	const query = writable('');
	const showVisible = writable(true);
	const showHidden = writable(true);
	const page = writable(1);
	let sortKey: SortKey = 'createdAt';
	let descending = true;
	const selected = writable(new Set<string>());

	$: needle = $query.trim().toLowerCase();
	$: filtered = words.filter(
		(word) =>
			(word.hidden ? $showHidden : $showVisible) &&
			(!needle ||
				word.text.includes(needle) ||
				word.originalText?.toLowerCase().includes(needle) ||
				`#${participants.get(word.uuid)}` === needle)
	);
	$: sorted = [...filtered].sort((a, b) => {
		const order =
			sortKey === 'text'
				? a.text.localeCompare(b.text)
				: sortKey === 'participant'
				? (participants.get(a.uuid) ?? 0) - (participants.get(b.uuid) ?? 0)
				: a.createdAt.localeCompare(b.createdAt);
		return descending ? -order : order;
	});
	$: needle, $showVisible, $showHidden, page.set(1);
	$: shown = sorted.slice(($page - 1) * PER_PAGE, $page * PER_PAGE);
	$: selected.update(
		(ids) => new Set([...ids].filter((wordId) => words.some((word) => word.id === wordId)))
	);

	const shownIds = writable<string[]>([]);
	$: shownIds.set(shown.map((word) => word.id));
	const allShownSelected = {
		subscribe: derived(
			[selected, shownIds],
			([$selected, $ids]) => $ids.length > 0 && $ids.every((wordId) => $selected.has(wordId))
		).subscribe,
		set(value: boolean) {
			selected.update((ids) => {
				for (const wordId of get(shownIds)) value ? ids.add(wordId) : ids.delete(wordId);
				return new Set(ids);
			});
		},
		update(updater: (value: boolean) => boolean) {
			this.set(updater(get(this)));
		}
	};
	const rowSelection = (wordId: string) => ({
		subscribe: derived(selected, ($selected) => $selected.has(wordId)).subscribe,
		set(value: boolean) {
			selected.update((ids) => {
				value ? ids.add(wordId) : ids.delete(wordId);
				return new Set(ids);
			});
		},
		update(updater: (value: boolean) => boolean) {
			this.set(updater(get(this)));
		}
	});

	function sortBy(key: SortKey) {
		if (sortKey === key) descending = !descending;
		else [sortKey, descending] = [key, key === 'createdAt'];
	}

	function fail(description: string) {
		addToast({ data: { title: 'Error', description, color: 'bg-ctp-red' } });
	}

	async function setHidden(word: WordCloudWord, hidden: boolean) {
		try {
			const res = await api.callWithAuth('PATCH', `/word-cloud/${id}/words/${word.id}`, { hidden });
			dispatch('words', res.data.words);
		} catch {
			fail(`“${word.text}” could not be ${hidden ? 'hidden' : 'shown'}`);
		}
	}

	const merging = writable(false);
	$: selectedTexts = countWords(
		words.filter((word) => $selected.has(word.id)).map((word) => word.text)
	);
	function startMerge() {
		mergeInto.set(selectedTexts[0]?.text ?? '');
		merging.set(true);
	}
	async function merge(e: Event) {
		e.preventDefault();
		if (!$mergeInto.trim()) return;
		try {
			const res = await api.callWithAuth('POST', `/word-cloud/${id}/words/merge`, {
				from: selectedTexts.map((entry) => entry.text),
				into: $mergeInto
			});
			dispatch('words', res.data.words);
			selected.set(new Set());
			merging.set(false);
		} catch {
			fail('The words could not be merged');
		}
	}

	// Quoted for commas and newlines; a leading = + - @ would run as a spreadsheet formula
	const csvCell = (value: unknown) => {
		const text = String(value ?? '');
		return `"${(/^[=+\-@]/.test(text) ? `'${text}` : text).replaceAll('"', '""')}"`;
	};

	function exportCSV() {
		const rows = [
			['Word', 'Submitted as', 'Hidden', 'Participant', 'Date'],
			...words.map((word) => [
				word.text,
				word.originalText ?? word.text,
				word.hidden,
				participants.get(word.uuid),
				new Date(word.createdAt).toISOString()
			])
		];
		const csv = rows.map((row) => row.map(csvCell).join(',')).join('\n');
		downloadBlob(new Blob([csv], { type: 'text/csv' }), `word-cloud-${id}.csv`);
	}

	const time = (date: string) =>
		new Date(date).toLocaleTimeString([], {
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});
	const header = 'px-3 py-2 text-start font-semibold';
</script>

<div class="flex flex-col gap-4">
	<div class="grid items-end gap-3 md:grid-cols-[1fr_auto_auto]">
		<div class="relative">
			<Search class="absolute left-3 bottom-2 text-ctp-subtext0" size="16" />
			<TextInput
				label="Search"
				value={query}
				placeholder="Words, or #3 for a participant"
				class="pl-9"
			/>
		</div>
		<div class="flex h-8 items-center gap-4">
			<MeltCheckbox name="show-visible" label="Visible" checked={showVisible} />
			<MeltCheckbox name="show-hidden" label="Hidden" checked={showHidden} />
		</div>
		<div class="flex gap-2">
			<Button on:click={startMerge} disabled={selectedTexts.length < 1}>
				<Merge size="16" />
				<span>Merge{$selected.size ? ` ${$selected.size}` : ''}</span>
			</Button>
			<Button on:click={exportCSV} disabled={!words.length}>
				<FileSpreadsheet size="16" />
				<span>CSV</span>
			</Button>
		</div>
	</div>

	<div class="overflow-x-auto">
		<table class="w-full table-auto text-sm">
			<thead class="border-b border-ctp-surface1 text-ctp-subtext0">
				<tr>
					<th class="w-8 px-3 py-2">
						<MeltCheckbox
							name="select-page"
							label="Select the words on this page"
							hideLabel
							checked={allShownSelected}
						/>
					</th>
					{#each COLUMNS as [key, label]}
						<th
							class={header}
							aria-sort={sortKey === key ? (descending ? 'descending' : 'ascending') : 'none'}
						>
							<button
								type="button"
								class="flex items-center gap-1 hover:text-ctp-text"
								on:click={() => sortBy(key)}
							>
								{label}
								{#if sortKey === key}
									<svelte:component this={descending ? ArrowDown : ArrowUp} size="14" />
								{/if}
							</button>
						</th>
					{/each}
					<th class="{header} text-end">Visibility</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-ctp-surface0">
				{#each shown as word (word.id)}
					<tr class="hover:bg-ctp-crust {word.hidden ? 'text-ctp-overlay1' : ''}">
						<td class="px-3 py-1.5">
							<MeltCheckbox
								name="select-{word.id}"
								label="Select “{word.text}”"
								hideLabel
								checked={rowSelection(word.id)}
							/>
						</td>
						<td class="px-3 py-1.5">
							<span class={word.hidden ? 'line-through' : 'font-semibold'}>{word.text}</span>
							{#if word.originalText && word.originalText !== word.text}
								<span class="ml-1 text-xs text-ctp-subtext0">from “{word.originalText}”</span>
							{/if}
						</td>
						<td class="px-3 py-1.5">
							<button
								type="button"
								class="rounded-md bg-ctp-surface0 px-2 py-0.5 font-mono text-xs hover:bg-ctp-mauve hover:text-ctp-base"
								on:click={() => dispatch('participant', word.uuid)}
								aria-label="All words of participant {participants.get(word.uuid)}"
							>
								#{participants.get(word.uuid)}
							</button>
						</td>
						<td class="px-3 py-1.5 font-mono text-xs text-ctp-subtext0">{time(word.createdAt)}</td>
						<td class="px-3 py-1.5 text-end">
							<MeltTooltip text={word.hidden ? 'Show this word' : 'Hide this word'}>
								<button
									type="button"
									class="rounded-md p-1 hover:bg-ctp-surface0 {word.hidden
										? 'text-ctp-overlay1'
										: 'text-ctp-text'}"
									aria-label="{word.hidden ? 'Show' : 'Hide'} “{word.text}”"
									on:click={() => setHidden(word, !word.hidden)}
								>
									<svelte:component this={word.hidden ? EyeOff : Eye} size="16" />
								</button>
							</MeltTooltip>
						</td>
					</tr>
				{:else}
					<tr>
						<td colspan="5" class="py-6 text-center text-ctp-subtext0">
							{words.length ? 'No submission matches' : 'No submissions yet'}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	{#if sorted.length > PER_PAGE}
		<Pagination {page} count={sorted.length} perPage={PER_PAGE} />
	{/if}
</div>

<Dialog
	open={merging}
	title="Merge {selectedTexts.length} word{selectedTexts.length === 1 ? '' : 's'}"
>
	<form class="flex flex-col gap-4" on:submit={merge}>
		<p class="text-sm text-ctp-subtext0">
			Every submission of
			{#each selectedTexts as entry, i}
				<strong class="text-ctp-text">{entry.text}</strong>{i < selectedTexts.length - 1
					? ', '
					: ''}
			{/each}
			becomes the word below. The original text stays in the table and the CSV.
		</p>
		<TextInput label="Merge into" value={mergeInto} />
		<div class="flex justify-end">
			<Button type="submit">
				<Merge size="16" />
				<span>Merge</span>
			</Button>
		</div>
	</form>
</Dialog>
