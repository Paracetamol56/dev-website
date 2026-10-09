<script lang="ts">
	import { CalendarClock, CalendarX2 } from 'lucide-svelte';
	import type { Todo } from './+page';
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import { closedStates, todoStates } from './states';

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

	$: closed = closedStates.includes(todo.state);
	$: closedAt = todo.history?.at(-1)?.updatedAt;
	$: days = todo.dueDate && !closed ? daysUntil(todo.dueDate) : null;
</script>

<span class="font-semibold mr-auto text-left {closed ? 'text-ctp-subtext0' : ''}">{todo.title}</span
>
{#if todo.state !== 'TODO'}
	<span class="flex items-center gap-1 text-sm font-semibold text-ctp-subtext0">
		<svelte:component
			this={todoStates[todo.state].icon}
			size="14"
			class={todoStates[todo.state].color}
		/>
		{todoStates[todo.state].label}
		{#if closed && closedAt}
			· {new Date(closedAt).toLocaleDateString(undefined, { dateStyle: 'medium' })}
		{/if}
	</span>
{/if}
{#each todo.labels ?? [] as label}
	<span class="rounded-md bg-ctp-mauve/20 px-1.5 text-sm font-semibold text-ctp-mauve">{label}</span
	>
{/each}
{#if days !== null}
	<MeltTooltip text={toRelative(days)}>
		{#if days < 0}
			<CalendarX2 size="16" class="text-ctp-red" />
		{:else}
			<CalendarClock size="16" />
		{/if}
	</MeltTooltip>
{/if}
