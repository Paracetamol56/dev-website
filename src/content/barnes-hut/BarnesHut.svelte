<script lang="ts">
	import * as d3 from 'd3';
	import { onMount } from 'svelte';
	import { type Point, type Node, generatePoints } from './utils';

	let svg: SVGSVGElement;
	let width: number;
	let height: number;

	let points: Point[] = [];
	let nodes: Node[] = [];
	let totalPoints: number = 0;

	const tree = d3
		.quadtree()
		//.cover(d3.min(points, (d) => d.x)!, d3.min(points, (d) => d.y)!)
		//.cover(d3.max(points, (d) => d.x)!, d3.max(points, (d) => d.y)!)
		.extent([
			[0, 0],
			[1000, 1000]
		]);
	$: x = d3.scaleLinear().domain([0, 1000]).range([0, width]);
	$: y = d3.scaleLinear().domain([0, 1000]).range([0, height]);

	function update() {
		nodes = [];

		tree.visit((node, x1, y1, x2, y2) => {
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
	}

	onMount(() => {
		const intervalId = setInterval(() => {
			if (totalPoints < 500) {
				points.push(...generatePoints(5));
				tree.addAll(points.map((d) => [d.x, d.y] as [number, number]));
				update();
				totalPoints = points.length;
			} else {
				clearInterval(intervalId);
			}
		}, 50);

		update();
	});
</script>

<section>
	<div class="flex justify-center items-center">
		<div
			class="bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md w-full max-w-2xl aspect-square mb-8"
			bind:clientWidth={width}
			bind:clientHeight={height}
		>
			{#key totalPoints}
				<svg bind:this={svg} {width} {height}>
					{#each points as point}
						<circle cx={x(point.x)} cy={y(point.y)} r="1.5" class="fill-ctp-mauve" />
					{/each}
					<g fill="none" stroke="currentColor" stroke-width="1.5">
						{#each nodes as node}
							<line
								x1={x(node.l1.x)}
								y1={y(node.l1.y)}
								x2={x(node.l3.x)}
								y2={y(node.l3.y)}
								opacity={node.opacity}
							/>
							<line
								x1={x(node.l2.x)}
								y1={y(node.l2.y)}
								x2={x(node.l4.x)}
								y2={y(node.l4.y)}
								opacity={node.opacity}
							/>
						{/each}
					</g>
				</svg>
			{/key}
		</div>
	</div>
</section>
