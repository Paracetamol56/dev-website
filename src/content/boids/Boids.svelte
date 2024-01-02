<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import MeltSlider from '$lib/components/MeltSlider.svelte';
	import { user } from '$lib/store';
	import palette from '@catppuccin/palette';
	import { Plus, RotateCcw } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { writable, type Writable } from 'svelte/store';
	import BoidPreview from './BoidPreview.svelte';
	import MeltPopover from '$lib/components/MeltTooltip.svelte';

	interface Boid {
		x: number;
		y: number;
		vx: number;
		vy: number;
	}

	let canvas: HTMLCanvasElement;
	let width: number;
	let height: number;
	const numBoids: number = 200;

	// Parameters
	let speed: Writable<number[]> = writable([2.5]);
	let cohesion: Writable<number[]> = writable([0.75]);
	let separation: Writable<number[]> = writable([0.75]);
	let alignment: Writable<number[]> = writable([0.5]);
	let vision: Writable<number[]> = writable([100]);
	const margin: number = 50;

	const boids: Boid[] = [];

	const handleReset = () => {
		speed.set([2.5]);
		cohesion.set([0.75]);
		separation.set([0.75]);
		alignment.set([0.5]);
		vision.set([100]);
		boids.splice(0, boids.length);
		initBoids();
	};

	const handleAdd100 = () => {
		for (let i = 0; i < 100; i++) {
			const angle = Math.random() * 2 * Math.PI;
			boids.push({
				x: Math.random() * width,
				y: Math.random() * height,
				vx: Math.cos(angle) * $speed[0],
				vy: Math.sin(angle) * $speed[0]
			});
		}
	};

	const initBoids = () => {
		for (let i = 0; i < numBoids; i++) {
			const angle = Math.random() * 2 * Math.PI;
			boids.push({
				x: Math.random() * width,
				y: Math.random() * height,
				vx: Math.cos(angle) * $speed[0],
				vy: Math.sin(angle) * $speed[0]
			});
		}
	};

	onMount(() => {
		initBoids();
		update();
	});

	function update() {
		const ctx: CanvasRenderingContext2D = canvas.getContext('2d') as CanvasRenderingContext2D;
		ctx.clearRect(0, 0, width, height);
		for (let i = 0; i < boids.length; i++) {
			const v1: number[] = cohesionRule(boids[i], boids);
			const v2: number[] = alignmentRule(boids[i], boids);
			const v3: number[] = separationRule(boids[i], boids);
			boids[i].vx += v1[0] * 0.001 + v2[0] * 0.05 + v3[0] * 0.005;
			boids[i].vy += v1[1] * 0.001 + v2[1] * 0.05 + v3[1] * 0.005;

			// Limit speed
			const invSqrtSpeed: number =
				1 / Math.sqrt(boids[i].vx * boids[i].vx + boids[i].vy * boids[i].vy);
			boids[i].vx = boids[i].vx * invSqrtSpeed * $speed[0];
			boids[i].vy = boids[i].vy * invSqrtSpeed * $speed[0];

			// Update position
			boids[i].x += boids[i].vx;
			boids[i].y += boids[i].vy;
			// Handle edges
			if (boids[i].x >= width - margin) {
				boids[i].vx -= $speed[0] * 0.1;
			} else if (boids[i].x <= margin) {
				boids[i].vx += $speed[0] * 0.1;
			}
			if (boids[i].y >= height - margin) {
				boids[i].vy -= $speed[0] * 0.1;
			} else if (boids[i].y <= margin) {
				boids[i].vy += $speed[0] * 0.1;
			}
			drawBoid(boids[i]);
		}

		requestAnimationFrame(update);
	}

	function rayDistance(boid1: Boid, boid2: Boid): number {
		return Math.sqrt((boid1.x - boid2.x) ** 2 + (boid1.y - boid2.y) ** 2);
	}

	function drawBoid(boid: Boid): void {
		const ctx: CanvasRenderingContext2D = canvas.getContext('2d') as CanvasRenderingContext2D;
		ctx.translate(boid.x, boid.y);
		ctx.rotate(Math.atan2(boid.vy, boid.vx));
		ctx.translate(-boid.x, -boid.y);
		ctx.beginPath();
		ctx.moveTo(boid.x, boid.y);
		ctx.lineTo(boid.x - 10, boid.y + 5);
		ctx.lineTo(boid.x - 10, boid.y - 5);
		ctx.lineTo(boid.x, boid.y);
		ctx.fillStyle = palette.variants[$user.flavour].mauve.hex;
		ctx.fill();
		ctx.setTransform(1, 0, 0, 1, 0, 0);
	}

	function cohesionRule(boid: Boid, otherBoids: Boid[]): number[] {
		const centerOfMass: number[] = [0, 0];
		let numNeighbors: number = 0;
		let distance: number = $vision[0] * $cohesion[0];

		for (let i = 0; i < otherBoids.length; i++) {
			if (rayDistance(boid, otherBoids[i]) < distance) {
				centerOfMass[0] += otherBoids[i].x;
				centerOfMass[1] += otherBoids[i].y;
				numNeighbors++;
			}
		}

		if (numNeighbors === 0) {
			return [0, 0];
		}
		centerOfMass[0] /= numNeighbors;
		centerOfMass[1] /= numNeighbors;

		return [centerOfMass[0] - boid.x, centerOfMass[1] - boid.y];
	}

	function alignmentRule(boid: Boid, otherBoids: Boid[]): number[] {
		const avgV: number[] = [0, 0];
		let numNeighbors: number = 0;
		const distance: number = $vision[0] * $alignment[0] * $cohesion[0];

		for (let i = 0; i < otherBoids.length; i++) {
			if (rayDistance(boid, otherBoids[i]) < distance) {
				avgV[0] += otherBoids[i].vx;
				avgV[1] += otherBoids[i].vy;
				numNeighbors++;
			}
		}

		if (numNeighbors === 0) {
			return [0, 0];
		}
		avgV[0] /= numNeighbors;
		avgV[1] /= numNeighbors;
		return [avgV[0] - boid.vx, avgV[1] - boid.vy];
	}

	function separationRule(boid: Boid, otherBoids: Boid[]): number[] {
		const v: number[] = [0, 0];
		const distance: number = $vision[0] * $separation[0] * $cohesion[0] * $alignment[0];
		for (let i = 0; i < otherBoids.length; i++) {
			const distance: number = rayDistance(boid, otherBoids[i]);
			if (distance === 0) {
				continue;
			}
			if (distance < $vision[0] * $separation[0] * $cohesion[0] * $alignment[0]) {
				v[0] += (boid.x - otherBoids[i].x) / distance;
				v[1] += (boid.y - otherBoids[i].y) / distance;
			}
		}
		return v;
	}
</script>

<section>
	<div
		class="bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md p-4 flex flex-col lg:flex-row gap-2 md:gap-4 justify-center items-center my-8"
	>
		<div class="flex flex-col sm:flex-row justify-center items-start gap-2 lg:4">
			<div class="flex flex-col justify-center items-start gap-2">
				<MeltSlider name="Cohesion" min={0} max={1} value={cohesion} step={0.01} />
				<MeltSlider name="Alignment" min={0} max={1} value={alignment} step={0.01} />
				<MeltSlider name="Separation" min={0} max={1} value={separation} step={0.01} />
			</div>
			<div class="flex flex-col justify-center items-start gap-2">
				<MeltSlider name="Speed" min={0} max={10} value={speed} step={0.1} />
				<MeltSlider name="Vision" min={0} max={100} value={vision} step={1} />
			</div>
		</div>
		<div>
			<BoidPreview {cohesion} {separation} {alignment} {vision} />
		</div>
	</div>
	<div
		class="bg-ctp-mantle shadow-md shadow-ctp-crust rounded-md aspect-video mb-8"
		bind:clientWidth={width}
		bind:clientHeight={height}
	>
		<div class="absolute top-2 right-2 flex gap-2">
			<MeltPopover text="Reset">
				<button
					class="flex items-center gap-1 rounded-md bg-ctp-mauve p-1
								font-semibold text-ctp-mantle
								shadow-md shadow-ctp-crust transition-opacity
								hover:opacity-80 active:opacity-60"
					on:click={handleReset}
					type="button"
				>
					<RotateCcw size="18" />
				</button>
			</MeltPopover>
			<MeltPopover text="Add 100 boids">
				<button
					class="flex items-center gap-1 rounded-md bg-ctp-mauve p-1
								font-semibold text-ctp-mantle
								shadow-md shadow-ctp-crust transition-opacity
								hover:opacity-80 active:opacity-60"
					on:click={handleAdd100}
					type="button"
				>
					<Plus size="18" />
				</button>
			</MeltPopover>
		</div>
		<canvas class="rounded-md" bind:this={canvas} {width} {height} />
	</div>
</section>
