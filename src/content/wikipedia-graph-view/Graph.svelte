<script lang="ts">
	import * as d3 from 'd3';
	import { onMount } from 'svelte';
	import { Search } from 'lucide-svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import axios from 'axios';
	import { type Page, type Link, extractLinks } from './utils';
	import { variants } from '@catppuccin/palette';
	import { user } from '$lib/store';

	let search: string = $page.url.searchParams.get('q') ?? '';
	let searchError: string = '';

	const handleSearch = (event: Event) => {
		event.preventDefault();

		const regex = /\/wiki\/([^#?]+)/;
		const match = search.match(regex);
		if (match) {
			search = match[1];
		}

		if (search !== '' && search !== null) {
			axios
				.get(
					`https://en.wikipedia.org/api/rest_v1/page/html/${search}`
				)
				.then((response) => {
					const newPage: Page = {
						title: search,
					};
					const newPageLinks = extractLinks(response.data);
          console.log(newPageLinks);
          if (!nodes.some((n) => n.title === newPage.title)) {
            nodes.push({
              title: newPage.title
            });
          }
					for (const link of newPageLinks) {
            if (nodes.some((n) => n.title === link.title)) continue;
						nodes.push({
							title: link.title
						});
					}
					for (const link of newPageLinks) {
            if (links.some((l) => l.target === link.title)) continue;
            links.push({
              source: newPage.title,
              target: link.title
            });
          }

					destroySimulation();
					startSimulation();

					$page.url.searchParams.set('q', search);
					goto($page.url.toString());
				})
				.catch((error) => {
					console.error(error);
					searchError = 'No page found with this title';
					$page.url.searchParams.delete('q');
					goto($page.url.toString());
				});
		} else {
			searchError = 'Please enter a valid Wikipedia URL';
			$page.url.searchParams.delete('q');
			goto($page.url.toString());
		}
	};

	let width: number;
	let height: number;
	let svg: SVGSVGElement;

	let hover: { title: string, x: number; y: number } | null = null;
  let summaries = new Map<string, any>();
  const getSummary = (title: string) => {
    if (!summaries.has(title)) {
      axios
        .get(
          `https://en.wikipedia.org/api/rest_v1/page/summary/${title}`
        )
        .then((response) => {
          summaries.set(title, response.data);
          // Recreate summaries map to trigger reactivity
          summaries = new Map<string, any>(summaries);
        })
        .catch((error) => {
          console.error(error);
        });
    }
  }

	const nodes: { title: string }[] = [];
	const links: { source: string; target: string }[] = [];
	let simulation: d3.Simulation<any, any>;

	const startSimulation = () => {
		const dragstart = (event: any) => {
			if (!event.active) simulation.alphaTarget(0.3).restart();
			event.subject.fx = event.subject.x;
			event.subject.fy = event.subject.y;
		};

		const drag = (event: any) => {
			event.subject.fx = event.x;
			event.subject.fy = event.y;
		};

		const dragend = (event: any) => {
			if (!event.active) simulation.alphaTarget(0);
			event.subject.fx = null;
			event.subject.fy = null;
		};

		simulation = d3
			.forceSimulation(nodes)
			.force(
				'link',
				d3.forceLink(links).id((d: any) => d.title)
			)
			.force('charge', d3.forceManyBody().strength(-100))
			.force('X', d3.forceX())
			.force('Y', d3.forceY());

		const link = d3
			.select(svg)
			.append('g')
			.attr('stroke', variants[$user.flavour].surface0.rgb)
			.selectAll('line')
			.data(links)
			.join('line')
			.attr('stroke-opacity', 0.6)
			.attr('stroke-width', 1.5);

		const node = d3
			.select(svg)
			.append('g')
			.attr('stroke', 'none')
			.attr('stroke-width', 1.5)
			.selectAll('circle')
			.data(nodes)
			.join('circle')
			.attr('r', 5)
			.attr('fill', variants[$user.flavour].mauve.rgb)
			.on('mouseenter', (event: any, d: any) => {
				hover = {
					title: d.title,
					x: d.x,
					y: d.y
				};
        getSummary(d.title);
			})
      .on('mouseleave', (event: any, d: any) => {
        if (!event.relatedTarget.closest('#hover-popover')) {
          hover = null;
        }
      })
			

		node.call(
			// @ts-ignore
			d3.drag().on('start', dragstart).on('drag', drag).on('end', dragend)
		);

		simulation.on('tick', () => {
			link
				.attr('x1', (d: any) => d.source.x)
				.attr('y1', (d: any) => d.source.y)
				.attr('x2', (d: any) => d.target.x)
				.attr('y2', (d: any) => d.target.y);
			node.attr('cx', (d: any) => d.x).attr('cy', (d: any) => d.y);
		});
	};

	const destroySimulation = () => {
		simulation.stop();
		d3.select(svg).selectAll('*').remove();
	};

	onMount(() => {
		startSimulation();
	});
</script>

<div class="my-8 mx-auto w-fit">
	<label for="root" class="mb-2 text-sm font-semibold"
		>Paste the Wikipedia URL or name of the root node</label
	>
	<div>
		<div class="flex gap-2">
			<input
				id="root"
				type="text"
				min="0"
				class="flex h-8 items-center justify-between rounded-md bg-ctp-surface0
            px-3 pr-12 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
				bind:value={search}
				on:keydown={(event) => {
					if (event.key === 'Enter') {
						handleSearch(event);
					}
				}}
			/>
			<button
				class="inline-flex square-8 items-center justify-center rounded-md
              bg-ctp-mauve font-medium leading-none text-ctp-mantle
              outline-none hover:opacity-75 active:opacity-50 transition-opacity"
				on:click={handleSearch}
			>
				<Search size="18" />
			</button>
		</div>
		<span class="mb-4 text-left text-sm font-semibold text-ctp-red">{searchError}</span>
	</div>
</div>
<div
	class="relative bg-ctp-mantle rounded-md shadow-md shadow-ctp-crust w-full aspect-video"
	bind:clientWidth={width}
	bind:clientHeight={height}
>
	<svg
		bind:this={svg}
		{width}
		{height}
		viewBox={`${-width / 2} ${-height / 2} ${width} ${height}`}
	/>
	{#if hover}
		<a
      href="https://en.wikipedia.org/wiki/{hover.title}"
      target="_blank"
      id="hover-popover"
			class="content-ignore absolute w-full max-w-72 hover:no-underline z-30"
			style="left: {hover.x + width / 2}px; top: {hover.y + height / 2}px; transform: translate(-50%, 10px);"
      on:mouseleave={() => {
        hover = null;
      }}
		>
      <div class="relative pt-1 px-5 pb-3 bg-ctp-base rounded-md shadow-md shadow-ctp-crust">
        <div
          style="position: absolute; width: 8px; height: 8px; left: calc(50% - 4px); bottom: calc(100% - 4px); transform: rotate(45deg); background-color: inherit; z-index: inherit;"
        />
        <h4 class="content-ignore">{hover.title}</h4>
        {#if summaries.get(hover.title)}
          {#if summaries.get(hover.title).thumbnail}
            <img
              src={summaries.get(hover.title).thumbnail.source}
              alt={summaries.get(hover.title).title}
              class="m-0 w-full object-cover aspect-video rounded-md"
            />
          {/if}
          <p class="line-clamp-5">
            {summaries.get(hover.title).extract ?? ''}
          </p>
        {/if}
      </div>
    </a>
	{/if}
</div>
