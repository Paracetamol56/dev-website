<script lang="ts">
	import { Flame, CircleCheck, Play, ListTodo, Trophy } from 'lucide-svelte';
	import type { TodoStats } from './+page';
	import { addDays, toKey } from './dates';
	import Calendar from './Calendar.svelte';
	import { showTodos } from './filters';
	import { getLocalTimeZone, today as todayDate } from '@internationalized/date';

	export let stats: TodoStats;

	const today = new Date();

	$: counts = new Map(stats.days.map((day) => [day.date, day.count]));
	const sumDays = (counts: Map<string, number>, from: Date, length: number) =>
		Array.from({ length }, (_, i) => counts.get(toKey(addDays(from, i))) ?? 0).reduce(
			(sum, count) => sum + count,
			0
		);

	$: lastWeek = sumDays(counts, addDays(today, -6), 7);
	$: previousWeek = sumDays(counts, addDays(today, -13), 7);
	$: record = stats.currentStreak > 0 && stats.currentStreak === stats.longestStreak;
	const day = (offset: number) => todayDate(getLocalTimeZone()).add({ days: offset });
	const showStreak = () =>
		showTodos({
			states: ['DONE'],
			from: day(-Math.max(stats.currentStreak, 1)),
			to: day(0),
			sort: 'updated'
		});

	const plural = (count: number, word: string) => `${count} ${word}${count === 1 ? '' : 's'}`;
</script>

<div class="mb-16 grid grid-cols-2 gap-4 lg:grid-cols-4">
	{#if record}
		<button
			type="button"
			class="rounded-md bg-gradient-to-r from-ctp-peach via-ctp-red to-ctp-mauve p-0.5 text-left shadow-lg shadow-ctp-peach/30"
			on:click={showStreak}
		>
			<span
				class="block h-full rounded bg-ctp-mantle p-3.5 transition-colors hover:bg-ctp-surface0"
			>
				<span class="flex items-center gap-2 text-sm font-semibold text-ctp-subtext0">
					<Flame size="16" class="text-ctp-peach" />
					Current streak
				</span>
				<span
					class="block w-fit bg-gradient-to-r from-ctp-peach to-ctp-red bg-clip-text text-3xl font-bold text-transparent"
				>
					{plural(stats.currentStreak, 'day')}
				</span>
				<span class="flex items-center gap-1 text-sm font-semibold text-ctp-subtext0">
					<Trophy size="14" class="text-ctp-yellow" />
					Personal best, keep going!
				</span>
			</span>
		</button>
	{:else}
		<button
			type="button"
			class="rounded-md bg-ctp-mantle p-4 text-left transition-colors hover:bg-ctp-surface0"
			on:click={showStreak}
		>
			<span class="flex items-center gap-2 text-sm font-semibold text-ctp-subtext0">
				<Flame size="16" class="text-ctp-peach" />
				Current streak
			</span>
			<span class="block text-3xl font-bold">{plural(stats.currentStreak, 'day')}</span>
			<span class="block text-sm text-ctp-subtext0"
				>Longest: {plural(stats.longestStreak, 'day')}</span
			>
		</button>
	{/if}
	<button
		type="button"
		class="rounded-md bg-ctp-mantle p-4 text-left transition-colors hover:bg-ctp-surface0"
		on:click={() => showTodos({ states: ['DONE'], from: day(-6), to: day(0), sort: 'updated' })}
	>
		<span class="flex items-center gap-2 text-sm font-semibold text-ctp-subtext0">
			<CircleCheck size="16" class="text-ctp-green" />
			Completed, last 7 days
		</span>
		<span class="block text-3xl font-bold">{lastWeek}</span>
		<span class="block text-sm text-ctp-subtext0">
			{lastWeek >= previousWeek ? '+' : ''}{lastWeek - previousWeek} vs the 7 days before
		</span>
	</button>
	<button
		type="button"
		class="rounded-md bg-ctp-mantle p-4 text-left transition-colors hover:bg-ctp-surface0"
		on:click={() => showTodos({ states: ['IN_PROGRESS'] })}
	>
		<span class="flex items-center gap-2 text-sm font-semibold text-ctp-subtext0">
			<Play size="16" class="text-ctp-blue" />
			In progress
		</span>
		<span class="block text-3xl font-bold">{stats.inProgress}</span>
		<span class="block text-sm text-ctp-subtext0">{stats.standby} on standby</span>
	</button>
	<button
		type="button"
		class="rounded-md bg-ctp-mantle p-4 text-left transition-colors hover:bg-ctp-surface0"
		on:click={() => showTodos({ states: ['TODO'] })}
	>
		<span class="flex items-center gap-2 text-sm font-semibold text-ctp-subtext0">
			<ListTodo size="16" class="text-ctp-mauve" />
			To do
		</span>
		<span class="block text-3xl font-bold">{stats.todo}</span>
		<span class="block text-sm text-ctp-subtext0">{stats.overdue} overdue</span>
	</button>

	<div class="col-span-2 lg:col-span-4">
		<Calendar stats={stats.days} firstYear={stats.firstYear} />
	</div>
</div>
