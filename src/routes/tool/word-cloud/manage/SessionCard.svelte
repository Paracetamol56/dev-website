<script lang="ts">
	import type { ManagePageData } from '../utils';

	export let session: ManagePageData;
	export let open: boolean;
</script>

<a
	href="/tool/word-cloud/manage/{session.id}"
	class="group flex flex-col gap-3 rounded-md bg-ctp-mantle p-6 shadow-md shadow-ctp-crust
		transition-colors hover:bg-ctp-surface0 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve"
>
	<div class="flex items-center gap-2 text-sm text-ctp-subtext0">
		<span class="square-2 rounded-full {open ? 'bg-ctp-green' : 'bg-ctp-red'}" />
		<span>{open ? 'Open' : 'Closed'}</span>
		<span class="font-mono font-semibold text-ctp-mauve">{session.code}</span>
		<span class="ml-auto">
			{session.submissions} submission{session.submissions === 1 ? '' : 's'}
		</span>
	</div>
	<h3 class="text-xl font-bold text-ctp-text">{session.name}</h3>
	{#if session.description}
		<p class="line-clamp-3 text-ctp-subtext0">{session.description}</p>
	{/if}
</a>
