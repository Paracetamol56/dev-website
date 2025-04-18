<script lang="ts">
	import { projectsStore } from './projects';
</script>

<svelte:head>
	<title>CodeCarbon dashboard - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				CodeCarbon dashboard
			</span>
		</h1>
	</hgroup>
</section>

<section class="container mx-auto">
	{#if $projectsStore.length === 0}
		<p>You don't have any open project.</p>
	{:else}
		<div class="mb-8 grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
			{#each $projectsStore as project (project.id)}
				<div class="p-8 bg-ctp-mantle rounded-md shadow-md shadow-ctp-crust">
					<div class="flex justify-start items-baseline mb-4">
						{#if project.closedAt !== null}
							<div class="relative square-2 mr-2 bg-ctp-green rounded-full">
								<span
									class="animate-ping absolute top-0.5 right-0.5 block square-1 rounded-full ring-2 ring-ctp-green bg-ctp-green"
								/>
							</div>
							<p class="text-ctp-subtext0 text-sm">Open</p>
						{:else}
							<div class="relative square-2 mr-2 bg-ctp-red rounded-full" />
							<p class="text-ctp-subtext0 text-sm">Archived</p>
						{/if}
						<p class="ml-auto text-ctp-subtext0 text-sm">
							{project.runsCount} run{project.runsCount !== 1 ? 's' : ''}
						</p>
					</div>
					<a href="/tool/codecarbon-dashboard/{project.id}">
						<h4 class="mb-4 text-2xl font-bold hover:opacity-75 transition-opacity">
							{project.name}
						</h4>
					</a>
					<p class="text-ctp-subtext0">{project.description}</p>
				</div>
			{/each}
		</div>
	{/if}
</section>
