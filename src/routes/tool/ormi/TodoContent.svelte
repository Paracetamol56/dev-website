<script lang="ts">
	import { CalendarClock, CalendarX2 } from 'lucide-svelte';
	import type { Todo } from './+page';
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';

	export let todo: Todo;

	const startOfDay = (date: Date) =>
		new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();

	function daysUntil(date: Date) {
		return Math.round(
			(startOfDay(new Date(date)) - startOfDay(new Date())) / (1000 * 60 * 60 * 24)
		);
	}

	function toRelative(days: number) {
		if (days === 0) {
			return 'today';
		} else if (days === 1) {
			return 'tomorrow';
		} else if (days === -1) {
			return 'yesterday';
		} else if (days > 0) {
			return `in ${days} days`;
		} else {
			return `${-days} days ago`;
		}
	}

	$: days = todo.dueDate ? daysUntil(todo.dueDate) : null;
</script>

<span class="font-semibold mr-auto text-left">{todo.title}</span>
{#each todo.labels ?? [] as label}
	<span class="rounded-md bg-ctp-mauve/20 px-1.5 text-sm font-semibold text-ctp-mauve">{label}</span
	>
{/each}
{#if days !== null}
	<MeltTooltip text={toRelative(days)}>
		{#if days < 0}
			<CalendarX2 size="16" />
		{:else}
			<CalendarClock size="16" />
		{/if}
	</MeltTooltip>
{/if}
