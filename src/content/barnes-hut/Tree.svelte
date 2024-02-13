<script lang="ts">
	import * as d3 from 'd3';
	import { type Point, type Node, generatePoints } from './utils';
	import { onMount } from 'svelte';

	let map: SVGSVGElement;
	let mapWidth: number;
	let mapHeight: number;
	let graph: SVGSVGElement;
	let graphWidth: number;
	let graphHeight: number;
	let graphMargin: number = 20;

	let points: Point[] = [];
	let selected: Point[] = [];
	let nodes: Node[] = [];
	let totalPoints: number = 0;

	const quadtree = d3.quadtree().extent([
		[0, 0],
		[1000, 1000]
	]);
	$: mapX = d3.scaleLinear().domain([0, 1000]).range([0, mapWidth]);
	$: mapY = d3.scaleLinear().domain([0, 1000]).range([0, mapHeight]);

	let root: d3.HierarchyNode<unknown>;
	let tree: d3.TreeLayout<unknown>;
	$: {
		tree = d3.tree().size([graphHeight - graphMargin * 2, graphWidth - graphMargin * 2]);
		update();
	}

	function update() {
		nodes = [];
		quadtree.visit((node, x1, y1, x2, y2) => {
			if (node.length) {
				const halfW = (x2 - x1) / 2;
				const halfH = (y2 - y1) / 2;
				nodes.push({
					l1: { x: x1 + halfW, y: y1 },
					l2: { x: x2, y: y1 + halfH },
					l3: { x: x1 + halfW, y: y2 },
					l4: { x: x1, y: y1 + halfH },
					opacity: Math.max(0.25, halfW / 100)
				});
			}
		});
		// Convert the quadtree into a hierarchy
		root = d3.hierarchy(quadtree.root(), (d) => {
			return Array.isArray(d) ? d : undefined;
		});
		tree(root);
	}

	function addPoint(e: MouseEvent) {
		const rect = map.getBoundingClientRect();
		const x = ((e.clientX - rect.left) / mapWidth) * 1000;
		const y = ((e.clientY - rect.top) / mapHeight) * 1000;
		points.push({
			x,
			y
		});
		quadtree.add([x, y]);
		totalPoints = points.length;
		update();
	}

	function nodeEnter(node: d3.HierarchyNode<unknown>) {
		if (node.data === undefined) return;
		// node.data is a tree, loop through it and get all the points as [x, y] arrays
		selected = node
			.descendants()
			.map((d: any) => {
				if (d.data === undefined) return;
				if (d.data.length === 4) return;
				return {
					x: d.data.data[0],
					y: d.data.data[1]
				};
			})
			.filter((d) => d !== undefined) as Point[];
	}

	function nodeLeave(node: d3.HierarchyNode<unknown>) {
		selected = [];
	}

	onMount(() => {
		// Add 3 random points
		const point = generatePoints(3);
		points.push(...point);
		quadtree.addAll(point.map((d) => [d.x, d.y] as [number, number]));
		totalPoints = points.length;
		update();
	});
</script>

<section>
	<div class="flex flex-col lg:flex-row gap-4">
		<div
			class="bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md lg:flex-grow max-w-2xl aspect-square"
			bind:clientWidth={mapWidth}
			bind:clientHeight={mapHeight}
		>
			{#key totalPoints}
				<!-- svelte-ignore a11y-click-events-have-key-events -->
				<!-- svelte-ignore a11y-no-static-element-interactions -->
				<svg bind:this={map} width={mapWidth} height={mapHeight} on:click={addPoint}>
					{#each points as point}
						<circle cx={mapX(point.x)} cy={mapY(point.y)} r="1.5" class="fill-ctp-mauve" />
					{/each}
					{#each selected as point}
						<circle cx={mapX(point.x)} cy={mapY(point.y)} r="2" class="fill-ctp-red" />
					{/each}
					<g fill="none" stroke="currentColor" stroke-width="1.5">
						{#each nodes as node}
							<line
								x1={mapX(node.l1.x)}
								y1={mapY(node.l1.y)}
								x2={mapX(node.l3.x)}
								y2={mapY(node.l3.y)}
								opacity={node.opacity}
							/>
							<line
								x1={mapX(node.l2.x)}
								y1={mapY(node.l2.y)}
								x2={mapX(node.l4.x)}
								y2={mapY(node.l4.y)}
								opacity={node.opacity}
							/>
						{/each}
					</g>
				</svg>
			{/key}
		</div>
		<div
			class="bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md overflow-auto max-w-2xl aspect-square lg:max-w-full min-h-96 lg:flex-1"
			bind:clientWidth={graphWidth}
			bind:clientHeight={graphHeight}
		>
			<svg bind:this={graph} width={graphWidth} height={graphHeight}>
				<g>
					{#if root}
						{#each root.descendants().slice(1) as link}
							<path
								class="stroke-ctp-lavender fill-none"
								d={`
              M ${link.y + graphMargin}, ${link.x + graphMargin}
              C ${(link.y + link.parent.y) / 2 + graphMargin} ${link.x + graphMargin},
                ${(link.y + link.parent.y) / 2 + graphMargin} ${link.parent.x + graphMargin},
                ${link.parent.y + graphMargin} ${link.parent.x + graphMargin}
                `}
							/>
						{/each}
						{#each root.descendants() as node}
							<circle
								cx={node.y + graphMargin}
								cy={node.x + graphMargin}
								r="5"
								class="fill-ctp-mauve hover:fill-ctp-red"
								on:mouseenter={() => nodeEnter(node)}
								on:mouseleave={() => nodeLeave(node)}
								role="tooltip"
							/>
						{/each}
					{/if}
				</g>
			</svg>
		</div>
	</div>
</section>
